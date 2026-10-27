package docker_test

import (
	"errors"
	"strings"
	"testing"

	"AHaSSHTools/internal/docker"
)

type fakeCLISession struct {
	commands []string
	output   string
	err      error
}

func (f *fakeCLISession) CombinedOutput(cmd string) ([]byte, error) {
	f.commands = append(f.commands, cmd)
	return []byte(f.output), f.err
}

func TestBuildDockerLifecycleCommands(t *testing.T) {
	const id = "abc123"
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"start", docker.BuildDockerStartCommand(id), "docker start 'abc123'"},
		{"stop", docker.BuildDockerStopCommand(id), "docker stop 'abc123'"},
		{"restart", docker.BuildDockerRestartCommand(id), "docker restart 'abc123'"},
		{"rm", docker.BuildDockerRemoveCommand(id, false), "docker rm 'abc123'"},
		{"rm force", docker.BuildDockerRemoveCommand(id, true), "docker rm -f 'abc123'"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q, want %q", tc.got, tc.want)
			}
		})
	}
}

func TestBuildDockerLifecycleCommandsShellEscape(t *testing.T) {
	id := "x'; echo pwned; '"
	cmd := docker.BuildDockerStartCommand(id)
	want := `docker start 'x'\''; echo pwned; '\'''`
	if cmd != want {
		t.Fatalf("got %q, want %q", cmd, want)
	}
}

func TestBuildDockerInspectAndLogsCommands(t *testing.T) {
	if got := docker.BuildDockerInspectCommand("c1"); got != "docker inspect 'c1'" {
		t.Fatalf("inspect: %q", got)
	}
	if got := docker.BuildDockerLogsTailCommand("c1", 50); got != "docker logs --tail 50 'c1'" {
		t.Fatalf("logs tail: %q", got)
	}
	if got := docker.BuildDockerLogsTailCommand("c1", 0); got != "docker logs 'c1'" {
		t.Fatalf("logs no tail: %q", got)
	}
}

const sampleInspectArray = `[{
  "Id": "abcdef0123456789",
  "Created": "2023-11-14T22:13:30.123456789Z",
  "Name": "/my-app",
  "Image": "sha256:abc",
  "State": {"Status": "running"},
  "Config": {
    "Image": "nginx:latest",
    "Env": ["FOO=bar"],
    "Cmd": ["nginx", "-g", "daemon off;"],
    "Entrypoint": ["/docker-entrypoint.sh"]
  },
  "HostConfig": {"NetworkMode": "bridge"},
  "Mounts": [{"Source": "/host/data", "Destination": "/data"}],
  "NetworkSettings": {
    "Ports": {"80/tcp": [{"HostIp": "0.0.0.0", "HostPort": "8080"}]}
  }
}]`

func TestParseDockerInspectJSON(t *testing.T) {
	detail, err := docker.ParseDockerInspectJSON([]byte(sampleInspectArray))
	if err != nil {
		t.Fatal(err)
	}
	if detail.Name != "my-app" || detail.ID != "abcdef012345" {
		t.Fatalf("detail: %+v", detail)
	}
	if detail.Image != "nginx:latest" || detail.State != "running" {
		t.Fatalf("image/state: %+v", detail)
	}
	if len(detail.Mounts) != 1 || detail.Mounts[0] != "/host/data -> /data" {
		t.Fatalf("mounts: %+v", detail.Mounts)
	}
}

func TestParseDockerInspectJSONEmpty(t *testing.T) {
	_, err := docker.ParseDockerInspectJSON([]byte("[]"))
	if err == nil {
		t.Fatal("expected error for empty array")
	}
}

func TestCLIStartStopRestartRemove(t *testing.T) {
	sess := &fakeCLISession{}
	for _, fn := range []struct {
		name string
		run  func() error
		want string
	}{
		{"start", func() error { return docker.CLIStart(sess, "id1") }, "docker start 'id1'"},
		{"stop", func() error { return docker.CLIStop(sess, "id2") }, "docker stop 'id2'"},
		{"restart", func() error { return docker.CLIRestart(sess, "id3") }, "docker restart 'id3'"},
		{"remove", func() error { return docker.CLIRemove(sess, "id4", true) }, "docker rm -f 'id4'"},
	} {
		t.Run(fn.name, func(t *testing.T) {
			sess.commands = nil
			if err := fn.run(); err != nil {
				t.Fatal(err)
			}
			if len(sess.commands) != 1 || sess.commands[0] != fn.want {
				t.Fatalf("commands: %v", sess.commands)
			}
		})
	}
}

func TestCLILifecyclePropagatesRemoteError(t *testing.T) {
	sess := &fakeCLISession{output: "Error: No such container: x\n", err: errors.New("exit 1")}
	err := docker.CLIStart(sess, "x")
	if err == nil || !strings.Contains(err.Error(), "No such container") {
		t.Fatalf("err: %v", err)
	}
}

func TestCLIInspect(t *testing.T) {
	sess := &fakeCLISession{output: sampleInspectArray}
	detail, err := docker.CLIInspect(sess, "abcdef0123456789")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Name != "my-app" {
		t.Fatalf("detail: %+v", detail)
	}
	if len(sess.commands) != 1 || !strings.Contains(sess.commands[0], "docker inspect") {
		t.Fatalf("cmd: %v", sess.commands)
	}
}

func TestCLILogsTail(t *testing.T) {
	sess := &fakeCLISession{output: "line1\nline2\n"}
	got, err := docker.CLILogsTail(sess, "logbox", 10)
	if err != nil {
		t.Fatal(err)
	}
	if got != "line1\nline2\n" {
		t.Fatalf("logs: %q", got)
	}
	wantCmd := "docker logs --tail 10 'logbox'"
	if sess.commands[0] != wantCmd {
		t.Fatalf("cmd: %q", sess.commands[0])
	}
}

func TestBuildDockerExecCommand(t *testing.T) {
	got := docker.BuildDockerExecCommand("abc123")
	if got != "docker exec -i -t 'abc123' /bin/sh" {
		t.Fatalf("cmd: %q", got)
	}
	evil := docker.BuildDockerExecCommand("x'; rm -rf / #")
	if evil != "docker exec -i -t 'x'\\''; rm -rf / #' /bin/sh" {
		t.Fatalf("not quoted: %q", evil)
	}
}
