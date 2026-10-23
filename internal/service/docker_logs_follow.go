package service

import (
	"context"
	"fmt"
	"io"
	"sync"
	"unicode/utf8"

	"AHaSSHTools/internal/docker"
)

const dockerLogReadBufSize = 4096

// dockerFollow 是一个正在运行的日志跟流。stop 之后保证不会再有 emit 开始。
type dockerFollow struct {
	mu      sync.Mutex
	stopped bool
	cancel  context.CancelFunc
	rc      io.ReadCloser
}

func (f *dockerFollow) stop() {
	f.mu.Lock()
	already := f.stopped
	f.stopped = true
	f.mu.Unlock()
	if already {
		return
	}
	f.cancel()
	_ = f.rc.Close()
}

// emit 在未停止时调用回调；持锁保证 stop 返回后不再触发回调。
func (f *dockerFollow) emit(fn func(string), chunk string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.stopped {
		return
	}
	fn(chunk)
}

// StartDockerContainerLogs 开始跟流容器日志（tail<=0 表示从头）。数据通过 emit 以
// UTF-8 完整的文本块回调。同一会话+容器重复调用会先停止旧的跟流。
func (s *DockerService) StartDockerContainerLogs(sessionID, id string, tail int, emit func(chunk string)) error {
	sess, err := s.lookupSession(sessionID)
	if err != nil {
		return err
	}
	if id, err = validateContainerID(id); err != nil {
		return err
	}
	if emit == nil {
		return fmt.Errorf("docker: 日志回调不能为空")
	}

	ctx, cancel := context.WithCancel(context.Background())
	var rc io.ReadCloser
	if sess.mode == dockerModeCLI {
		if sess.ssh == nil {
			cancel()
			return fmt.Errorf("docker: CLI 模式缺少 SSH 会话")
		}
		rc, err = sess.ssh.StreamOutput(docker.BuildDockerLogsFollowCommand(id, tail))
	} else {
		rc, err = sess.apiClient.FollowContainerLogs(ctx, id, tail)
	}
	if err != nil {
		cancel()
		return err
	}
	f := &dockerFollow{cancel: cancel, rc: rc}

	sess.followMu.Lock()
	if sess.followsClosed {
		sess.followMu.Unlock()
		f.stop()
		return fmt.Errorf("docker: 会话 %q 已关闭", sessionID)
	}
	if sess.follows == nil {
		sess.follows = make(map[string]*dockerFollow)
	}
	old := sess.follows[id]
	sess.follows[id] = f
	sess.followMu.Unlock()
	if old != nil {
		old.stop()
	}

	go sess.runFollow(id, f, emit)
	return nil
}

// StopDockerContainerLogs 停止日志跟流。未在跟流或会话不存在均视为成功。
func (s *DockerService) StopDockerContainerLogs(sessionID, id string) error {
	s.mu.Lock()
	sess := s.sessions[sessionID]
	s.mu.Unlock()
	if sess == nil {
		return nil
	}
	id, err := validateContainerID(id)
	if err != nil {
		return err
	}
	sess.stopFollow(id)
	return nil
}

func (sess *dockerSession) stopFollow(id string) {
	sess.followMu.Lock()
	f := sess.follows[id]
	delete(sess.follows, id)
	sess.followMu.Unlock()
	if f != nil {
		f.stop()
	}
}

// stopAllFollows 停止全部跟流并拒绝之后的新跟流（会话关闭时调用）。
func (sess *dockerSession) stopAllFollows() {
	sess.followMu.Lock()
	sess.followsClosed = true
	follows := sess.follows
	sess.follows = nil
	sess.followMu.Unlock()
	for _, f := range follows {
		f.stop()
	}
}

func (sess *dockerSession) runFollow(id string, f *dockerFollow, emit func(string)) {
	defer func() {
		sess.followMu.Lock()
		if sess.follows[id] == f {
			delete(sess.follows, id)
		}
		sess.followMu.Unlock()
		f.stop()
	}()
	buf := make([]byte, dockerLogReadBufSize)
	var carry []byte
	for {
		n, err := f.rc.Read(buf)
		if n > 0 {
			data := append(carry, buf[:n]...)
			complete, rest := splitIncompleteUTF8(data)
			carry = append([]byte(nil), rest...)
			if len(complete) > 0 {
				f.emit(emit, string(complete))
			}
		}
		if err != nil {
			if len(carry) > 0 {
				f.emit(emit, string(carry))
			}
			return
		}
	}
}

// splitIncompleteUTF8 把末尾被截断的多字节序列切到 rest，complete 可安全转成字符串。
func splitIncompleteUTF8(b []byte) (complete, rest []byte) {
	for i := 1; i < utf8.UTFMax && i <= len(b); i++ {
		if utf8.RuneStart(b[len(b)-i]) {
			if !utf8.FullRune(b[len(b)-i:]) {
				return b[:len(b)-i], b[len(b)-i:]
			}
			break
		}
	}
	return b, nil
}
