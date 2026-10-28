package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"AHaSSHTools/internal/docker"
)

// validateDockerImageRef 校验镜像引用：允许 registry/路径、":" tag、"@" digest。
func validateDockerImageRef(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("docker: 镜像引用不能为空")
	}
	if ref[0] == '-' {
		return "", fmt.Errorf("docker: 镜像引用 %q 非法", ref)
	}
	for _, r := range ref {
		if !isDockerNameRune(r) && r != ':' && r != '/' && r != '@' {
			return "", fmt.Errorf("docker: 镜像引用 %q 含非法字符", ref)
		}
	}
	return ref, nil
}

// validateDockerResourceName 校验网络/卷的名称或 ID：非空、不以 "-" 开头、
// 仅限 [A-Za-z0-9_.-]（不含空白、引号、斜杠，避免注入 CLI 参数或 API 路径）。
func validateDockerResourceName(kind, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("docker: %s名称不能为空", kind)
	}
	if name[0] == '-' {
		return "", fmt.Errorf("docker: %s名称 %q 非法", kind, name)
	}
	for _, r := range name {
		if !isDockerNameRune(r) {
			return "", fmt.Errorf("docker: %s名称 %q 含非法字符", kind, name)
		}
	}
	return name, nil
}

func isDockerNameRune(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return true
	}
	return r == '_' || r == '.' || r == '-'
}

// dockerResourceCall 解析会话并分发到 API / CLI 分支；list/inspect 与 remove 超时不同。
func dockerResourceCall[T any](
	s *DockerService, sessionID string, timeout time.Duration,
	api func(ctx context.Context, c *docker.Client) (T, error),
	cli func(c docker.CLISession) (T, error),
) (T, error) {
	var zero T
	sess, err := s.lookupSession(sessionID)
	if err != nil {
		return zero, err
	}
	if sess.mode == dockerModeCLI {
		return cli(sess.cliSession(timeout))
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return api(ctx, sess.apiClient)
}

// ListDockerImages 列出镜像。
func (s *DockerService) ListDockerImages(sessionID string) ([]docker.ImageSummary, error) {
	return dockerResourceCall(s, sessionID, dockerOpTimeout,
		func(ctx context.Context, c *docker.Client) ([]docker.ImageSummary, error) { return c.ListImages(ctx) },
		func(c docker.CLISession) ([]docker.ImageSummary, error) {
			return docker.CLIListImages(c)
		})
}

// ListDockerNetworks 列出网络。
func (s *DockerService) ListDockerNetworks(sessionID string) ([]docker.NetworkSummary, error) {
	return dockerResourceCall(s, sessionID, dockerOpTimeout,
		func(ctx context.Context, c *docker.Client) ([]docker.NetworkSummary, error) {
			return c.ListNetworks(ctx)
		},
		func(c docker.CLISession) ([]docker.NetworkSummary, error) {
			return docker.CLIListNetworks(c)
		})
}

// ListDockerVolumes 列出卷。
func (s *DockerService) ListDockerVolumes(sessionID string) ([]docker.VolumeSummary, error) {
	return dockerResourceCall(s, sessionID, dockerOpTimeout,
		func(ctx context.Context, c *docker.Client) ([]docker.VolumeSummary, error) { return c.ListVolumes(ctx) },
		func(c docker.CLISession) ([]docker.VolumeSummary, error) {
			return docker.CLIListVolumes(c)
		})
}

// InspectDockerImage 返回镜像详情。
func (s *DockerService) InspectDockerImage(sessionID, ref string) (docker.ImageDetail, error) {
	if _, err := s.lookupSession(sessionID); err != nil {
		return docker.ImageDetail{}, err
	}
	ref, err := validateDockerImageRef(ref)
	if err != nil {
		return docker.ImageDetail{}, err
	}
	return dockerResourceCall(s, sessionID, dockerOpTimeout,
		func(ctx context.Context, c *docker.Client) (docker.ImageDetail, error) {
			return c.InspectImage(ctx, ref)
		},
		func(c docker.CLISession) (docker.ImageDetail, error) {
			d, err := docker.CLIInspectImage(c, ref)
			d.Via = dockerViaCLI
			return d, err
		})
}

// InspectDockerNetwork 返回网络详情。
func (s *DockerService) InspectDockerNetwork(sessionID, idOrName string) (docker.NetworkDetail, error) {
	if _, err := s.lookupSession(sessionID); err != nil {
		return docker.NetworkDetail{}, err
	}
	id, err := validateDockerResourceName("网络", idOrName)
	if err != nil {
		return docker.NetworkDetail{}, err
	}
	return dockerResourceCall(s, sessionID, dockerOpTimeout,
		func(ctx context.Context, c *docker.Client) (docker.NetworkDetail, error) {
			return c.InspectNetwork(ctx, id)
		},
		func(c docker.CLISession) (docker.NetworkDetail, error) {
			d, err := docker.CLIInspectNetwork(c, id)
			d.Via = dockerViaCLI
			return d, err
		})
}

// InspectDockerVolume 返回卷详情。
func (s *DockerService) InspectDockerVolume(sessionID, name string) (docker.VolumeDetail, error) {
	if _, err := s.lookupSession(sessionID); err != nil {
		return docker.VolumeDetail{}, err
	}
	name, err := validateDockerResourceName("卷", name)
	if err != nil {
		return docker.VolumeDetail{}, err
	}
	return dockerResourceCall(s, sessionID, dockerOpTimeout,
		func(ctx context.Context, c *docker.Client) (docker.VolumeDetail, error) {
			return c.InspectVolume(ctx, name)
		},
		func(c docker.CLISession) (docker.VolumeDetail, error) {
			d, err := docker.CLIInspectVolume(c, name)
			d.Via = dockerViaCLI
			return d, err
		})
}

// RemoveDockerImage 删除镜像，force 对应 docker rmi -f。
func (s *DockerService) RemoveDockerImage(sessionID, ref string, force bool) error {
	if _, err := s.lookupSession(sessionID); err != nil {
		return err
	}
	ref, err := validateDockerImageRef(ref)
	if err != nil {
		return err
	}
	_, err = dockerResourceCall(s, sessionID, dockerLifecycleTimeout,
		func(ctx context.Context, c *docker.Client) (struct{}, error) {
			return struct{}{}, c.RemoveImage(ctx, ref, force)
		},
		func(c docker.CLISession) (struct{}, error) {
			return struct{}{}, docker.CLIRemoveImage(c, ref, force)
		})
	return err
}

// RemoveDockerNetwork 删除网络。
func (s *DockerService) RemoveDockerNetwork(sessionID, idOrName string) error {
	if _, err := s.lookupSession(sessionID); err != nil {
		return err
	}
	id, err := validateDockerResourceName("网络", idOrName)
	if err != nil {
		return err
	}
	_, err = dockerResourceCall(s, sessionID, dockerLifecycleTimeout,
		func(ctx context.Context, c *docker.Client) (struct{}, error) {
			return struct{}{}, c.RemoveNetwork(ctx, id)
		},
		func(c docker.CLISession) (struct{}, error) {
			return struct{}{}, docker.CLIRemoveNetwork(c, id)
		})
	return err
}

// RemoveDockerVolume 删除卷；始终 force=false，避免误删使用中的数据卷。
func (s *DockerService) RemoveDockerVolume(sessionID, name string) error {
	if _, err := s.lookupSession(sessionID); err != nil {
		return err
	}
	name, err := validateDockerResourceName("卷", name)
	if err != nil {
		return err
	}
	_, err = dockerResourceCall(s, sessionID, dockerLifecycleTimeout,
		func(ctx context.Context, c *docker.Client) (struct{}, error) {
			return struct{}{}, c.RemoveVolume(ctx, name, false)
		},
		func(c docker.CLISession) (struct{}, error) {
			return struct{}{}, docker.CLIRemoveVolume(c, name, false)
		})
	return err
}
