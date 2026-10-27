package docker

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// CLISession runs a shell command on the remote host (e.g. dedicated Docker SSH client).
type CLISession interface {
	CombinedOutput(cmd string) ([]byte, error)
}

func BuildDockerStartCommand(containerID string) string {
	return "docker start " + shellQuote(containerID)
}

func BuildDockerStopCommand(containerID string) string {
	return "docker stop " + shellQuote(containerID)
}

func BuildDockerRestartCommand(containerID string) string {
	return "docker restart " + shellQuote(containerID)
}

func BuildDockerRemoveCommand(containerID string, force bool) string {
	if force {
		return "docker rm -f " + shellQuote(containerID)
	}
	return "docker rm " + shellQuote(containerID)
}

func BuildDockerInspectCommand(containerID string) string {
	return "docker inspect " + shellQuote(containerID)
}

func BuildDockerLogsTailCommand(containerID string, tail int) string {
	if tail > 0 {
		return "docker logs --tail " + strconv.Itoa(tail) + " " + shellQuote(containerID)
	}
	return "docker logs " + shellQuote(containerID)
}

// BuildDockerLogsFollowCommand builds docker logs -f for service-layer streaming (Task 4/5).
func BuildDockerLogsFollowCommand(containerID string, tail int) string {
	base := "docker logs -f"
	if tail > 0 {
		base += " --tail " + strconv.Itoa(tail)
	}
	return base + " " + shellQuote(containerID)
}

// BuildDockerExecCommand builds an interactive shell command for a container
// (docker exec -i -t <id> /bin/sh). The caller must run it under a PTY.
func BuildDockerExecCommand(containerID string) string {
	return "docker exec -i -t " + shellQuote(containerID) + " /bin/sh"
}

func CLIStart(sess CLISession, id string) error {
	return cliRun(sess, BuildDockerStartCommand(id))
}

func CLIStop(sess CLISession, id string) error {
	return cliRun(sess, BuildDockerStopCommand(id))
}

func CLIRestart(sess CLISession, id string) error {
	return cliRun(sess, BuildDockerRestartCommand(id))
}

func CLIRemove(sess CLISession, id string, force bool) error {
	return cliRun(sess, BuildDockerRemoveCommand(id, force))
}

func CLIInspect(sess CLISession, id string) (ContainerDetail, error) {
	if sess == nil {
		return ContainerDetail{}, fmt.Errorf("docker: CLI 模式缺少 SSH 会话")
	}
	out, err := sess.CombinedOutput(BuildDockerInspectCommand(id))
	if err != nil {
		return ContainerDetail{}, cliCommandError(out, err)
	}
	return ParseDockerInspectJSON(out)
}

func CLILogsTail(sess CLISession, id string, tail int) (string, error) {
	if sess == nil {
		return "", fmt.Errorf("docker: CLI 模式缺少 SSH 会话")
	}
	out, err := sess.CombinedOutput(BuildDockerLogsTailCommand(id, tail))
	if err != nil {
		return "", cliCommandError(out, err)
	}
	return string(out), nil
}

// ParseDockerInspectJSON parses stdout from docker inspect (JSON array, first element).
func ParseDockerInspectJSON(data []byte) (ContainerDetail, error) {
	data = bytesTrimSpace(data)
	var arr []containerInspectJSON
	if err := json.Unmarshal(data, &arr); err != nil {
		return ContainerDetail{}, fmt.Errorf("docker: parse inspect JSON: %w", err)
	}
	if len(arr) == 0 {
		return ContainerDetail{}, fmt.Errorf("docker: inspect returned no containers")
	}
	return mapContainerDetail(arr[0]), nil
}

func cliRun(sess CLISession, cmd string) error {
	if sess == nil {
		return fmt.Errorf("docker: CLI 模式缺少 SSH 会话")
	}
	out, err := sess.CombinedOutput(cmd)
	if err != nil {
		return cliCommandError(out, err)
	}
	return nil
}

func cliCommandError(out []byte, err error) error {
	msg := strings.TrimSpace(string(out))
	if msg == "" {
		msg = err.Error()
	}
	return fmt.Errorf("docker: %s", msg)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func bytesTrimSpace(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}
