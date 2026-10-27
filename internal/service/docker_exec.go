package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/google/uuid"

	"AHaSSHTools/internal/docker"
)

const (
	dockerExecReadBufSize = 4096
	dockerExecDefaultCols = 80
	dockerExecDefaultRows = 24
)

// dockerExec 是一个正在运行的交互 exec（API hijack 流或 CLI PTY 流）。
// stop 之后保证不会再有 emit 开始，Write 也会失败。
type dockerExec struct {
	id        string
	sessionID string // 所属 DockerService 会话 ID
	sess      *dockerSession

	stream io.ReadWriteCloser
	resize func(cols, rows int) error

	writeMu sync.Mutex // 串行化 Write

	mu       sync.Mutex // 保护 stopped；emit 也持锁
	stopped  bool
	onClosed func(execSessionID string)
}

func (e *dockerExec) stop() {
	e.mu.Lock()
	already := e.stopped
	e.stopped = true
	e.mu.Unlock()
	if already {
		return
	}
	_ = e.stream.Close()
	if e.onClosed != nil {
		e.onClosed(e.id)
	}
}

func (e *dockerExec) emit(fn func([]byte), data []byte) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.stopped {
		return
	}
	fn(data)
}

func (e *dockerExec) isStopped() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.stopped
}

func (e *dockerExec) write(data []byte) error {
	if e.isStopped() {
		return fmt.Errorf("docker: exec %q 已关闭", e.id)
	}
	e.writeMu.Lock()
	defer e.writeMu.Unlock()
	_, err := e.stream.Write(data)
	return err
}

func newDockerExecSessionID() string {
	return "dexec-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12]
}

// SetDockerExecClosedHandler 设置 exec 结束（EOF/错误/主动关闭/会话关闭）时的回调，
// 可用于向前端发送 docker:exec-closed:{execSessionID}。仅影响之后新建的 exec。
func (s *DockerService) SetDockerExecClosedHandler(fn func(execSessionID string)) {
	s.mu.Lock()
	s.execClosed = fn
	s.mu.Unlock()
}

// StartDockerExec 在容器内启动交互 shell（/bin/sh）。输出以原始字节通过 emit 回调。
func (s *DockerService) StartDockerExec(dockerSessionID, containerID string, cols, rows int, emit func([]byte)) (string, error) {
	sess, err := s.lookupSession(dockerSessionID)
	if err != nil {
		return "", err
	}
	if containerID, err = validateContainerID(containerID); err != nil {
		return "", err
	}
	if emit == nil {
		return "", fmt.Errorf("docker: exec 输出回调不能为空")
	}
	if cols <= 0 {
		cols = dockerExecDefaultCols
	}
	if rows <= 0 {
		rows = dockerExecDefaultRows
	}

	s.mu.Lock()
	onClosed := s.execClosed
	s.mu.Unlock()

	e := &dockerExec{
		id:        newDockerExecSessionID(),
		sessionID: dockerSessionID,
		sess:      sess,
		onClosed:  onClosed,
	}
	if sess.mode == dockerModeCLI {
		if sess.ssh == nil {
			return "", fmt.Errorf("docker: CLI 模式缺少 SSH 会话")
		}
		pty, err := sess.ssh.StartPTYCommand(docker.BuildDockerExecCommand(containerID), cols, rows)
		if err != nil {
			return "", err
		}
		e.stream = pty
		e.resize = pty.Resize
	} else {
		if sess.apiClient == nil {
			return "", fmt.Errorf("docker: 会话 %q 缺少 API 客户端", dockerSessionID)
		}
		client := sess.apiClient
		ctx, cancel := context.WithTimeout(context.Background(), dockerOpTimeout)
		defer cancel()
		execID, err := client.CreateExec(ctx, containerID, docker.ExecCreateOptions{
			Cmd:          []string{"/bin/sh"},
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
		})
		if err != nil {
			return "", err
		}
		stream, err := client.StartExec(ctx, execID, cols, rows)
		if err != nil {
			return "", err
		}
		e.stream = stream
		e.resize = func(c, r int) error {
			rctx, rcancel := context.WithTimeout(context.Background(), dockerOpTimeout)
			defer rcancel()
			return client.ResizeExec(rctx, execID, c, r)
		}
	}

	if err := sess.registerExec(e); err != nil {
		e.stop()
		return "", err
	}
	go sess.runExec(e, emit)
	return e.id, nil
}

// SendDockerExecData 向 exec 写入 stdin 数据。
func (s *DockerService) SendDockerExecData(execSessionID string, data []byte) error {
	e, err := s.lookupExec(execSessionID)
	if err != nil {
		return err
	}
	return e.write(data)
}

// ResizeDockerExec 调整 exec 终端尺寸。
func (s *DockerService) ResizeDockerExec(execSessionID string, cols, rows int) error {
	e, err := s.lookupExec(execSessionID)
	if err != nil {
		return err
	}
	if cols <= 0 || rows <= 0 {
		return fmt.Errorf("docker: 非法终端尺寸 %dx%d", cols, rows)
	}
	return e.resize(cols, rows)
}

// CloseDockerExec 关闭 exec。未知 exec 视为成功（可能已自行结束）。
func (s *DockerService) CloseDockerExec(execSessionID string) error {
	e, err := s.lookupExec(execSessionID)
	if err != nil {
		return nil
	}
	e.sess.removeExec(e)
	e.stop()
	return nil
}

// lookupExec 在已登记的会话中查找 exec。会话关闭后其 exec 已被清理，因此返回错误。
func (s *DockerService) lookupExec(execSessionID string) (*dockerExec, error) {
	s.mu.Lock()
	sessions := make([]*dockerSession, 0, len(s.sessions))
	for _, sess := range s.sessions {
		sessions = append(sessions, sess)
	}
	s.mu.Unlock()
	for _, sess := range sessions {
		sess.execMu.Lock()
		e := sess.execs[execSessionID]
		sess.execMu.Unlock()
		if e != nil {
			return e, nil
		}
	}
	return nil, fmt.Errorf("docker: exec 会话 %q 不存在或已关闭", execSessionID)
}

func (sess *dockerSession) registerExec(e *dockerExec) error {
	sess.execMu.Lock()
	defer sess.execMu.Unlock()
	if sess.execsClosed {
		return fmt.Errorf("docker: 会话 %q 已关闭", e.sessionID)
	}
	if sess.execs == nil {
		sess.execs = make(map[string]*dockerExec)
	}
	sess.execs[e.id] = e
	return nil
}

func (sess *dockerSession) removeExec(e *dockerExec) {
	sess.execMu.Lock()
	if sess.execs[e.id] == e {
		delete(sess.execs, e.id)
	}
	sess.execMu.Unlock()
}

// stopAllExecs 关闭该会话下全部 exec 并拒绝之后新建的 exec（会话关闭时调用）。
func (sess *dockerSession) stopAllExecs() {
	sess.execMu.Lock()
	sess.execsClosed = true
	execs := sess.execs
	sess.execs = nil
	sess.execMu.Unlock()
	for _, e := range execs {
		e.stop()
	}
}

func (sess *dockerSession) runExec(e *dockerExec, emit func([]byte)) {
	defer func() {
		sess.removeExec(e)
		e.stop()
	}()
	buf := make([]byte, dockerExecReadBufSize)
	for {
		n, err := e.stream.Read(buf)
		if n > 0 {
			e.emit(emit, append([]byte(nil), buf[:n]...))
		}
		if err != nil {
			return
		}
	}
}
