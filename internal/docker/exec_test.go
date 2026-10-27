package docker_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"AHaSSHTools/internal/docker"
)

func TestCreateExecHTTP(t *testing.T) {
	var gotBody map[string]any
	c, srv := newTestDockerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/containers/abc/exec" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"Id":"exec-1"}`))
	})
	defer srv.Close()
	defer c.Close()

	id, err := c.CreateExec(context.Background(), "abc", docker.ExecCreateOptions{
		AttachStdin: true, AttachStdout: true, AttachStderr: true, Tty: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if id != "exec-1" {
		t.Fatalf("want exec-1, got %q", id)
	}
	if gotBody["Tty"] != true || gotBody["AttachStdin"] != true {
		t.Fatalf("body missing flags: %v", gotBody)
	}
	cmd, _ := gotBody["Cmd"].([]any)
	if len(cmd) != 1 || cmd[0] != "/bin/sh" {
		t.Fatalf("want default Cmd [/bin/sh], got %v", gotBody["Cmd"])
	}
}

func TestCreateExecCustomCmdAndError(t *testing.T) {
	var gotBody map[string]any
	status := http.StatusCreated
	c, srv := newTestDockerClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(status)
		if status == http.StatusCreated {
			_, _ = w.Write([]byte(`{"Id":"x"}`))
		} else {
			_, _ = w.Write([]byte(`{"message":"no such container"}`))
		}
	})
	defer srv.Close()
	defer c.Close()

	if _, err := c.CreateExec(context.Background(), "abc", docker.ExecCreateOptions{Cmd: []string{"bash", "-l"}}); err != nil {
		t.Fatal(err)
	}
	cmd, _ := gotBody["Cmd"].([]any)
	if len(cmd) != 2 || cmd[0] != "bash" {
		t.Fatalf("custom Cmd not sent: %v", gotBody["Cmd"])
	}

	status = http.StatusNotFound
	_, err := c.CreateExec(context.Background(), "abc", docker.ExecCreateOptions{})
	if err == nil || !strings.Contains(err.Error(), "no such container") {
		t.Fatalf("want error with body, got %v", err)
	}
}

func TestResizeExecHTTP(t *testing.T) {
	c, srv := newTestDockerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/exec/exec-1/resize" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("h") != "24" || r.URL.Query().Get("w") != "80" {
			t.Errorf("bad query: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()
	defer c.Close()

	if err := c.ResizeExec(context.Background(), "exec-1", 80, 24); err != nil {
		t.Fatal(err)
	}
}

func TestResizeExecError(t *testing.T) {
	c, srv := newTestDockerClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	defer srv.Close()
	defer c.Close()
	if err := c.ResizeExec(context.Background(), "exec-1", 80, 24); err == nil {
		t.Fatal("want error")
	}
}

func TestStartExecHijackRawStream(t *testing.T) {
	var resized int32
	c, srv := newTestDockerClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/exec/exec-1/start":
			if r.Method != http.MethodPost {
				t.Errorf("start: want POST, got %s", r.Method)
			}
			if !strings.EqualFold(r.Header.Get("Connection"), "Upgrade") || r.Header.Get("Upgrade") != "tcp" {
				t.Errorf("missing upgrade headers: %v", r.Header)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode: %v", err)
			}
			if body["Detach"] != false || body["Tty"] != true {
				t.Errorf("bad start body: %v", body)
			}
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Error("no hijacker")
				return
			}
			conn, buf, err := hj.Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			defer conn.Close()
			_, _ = buf.WriteString("HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.raw-stream\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\nwelcome$ ")
			_ = buf.Flush()
			// echo until closed (raw, no multiplex)
			tmp := make([]byte, 64)
			for {
				n, err := buf.Read(tmp)
				if n > 0 {
					_, _ = conn.Write(tmp[:n])
				}
				if err != nil {
					return
				}
			}
		case "/exec/exec-1/resize":
			atomic.AddInt32(&resized, 1)
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	})
	defer srv.Close()
	defer c.Close()

	rwc, err := c.StartExec(context.Background(), "exec-1", 100, 30)
	if err != nil {
		t.Fatal(err)
	}
	defer rwc.Close()

	readN := func(n int) string {
		buf := make([]byte, n)
		_ = setDeadline(rwc, time.Second*3)
		if _, err := io.ReadFull(rwc, buf); err != nil {
			t.Fatalf("read: %v", err)
		}
		return string(buf)
	}
	if got := readN(len("welcome$ ")); got != "welcome$ " {
		t.Fatalf("greeting = %q", got)
	}
	if _, err := rwc.Write([]byte("ls\n")); err != nil {
		t.Fatal(err)
	}
	if got := readN(3); got != "ls\n" {
		t.Fatalf("echo = %q", got)
	}
	deadline := time.Now().Add(2 * time.Second)
	for atomic.LoadInt32(&resized) == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if atomic.LoadInt32(&resized) == 0 {
		t.Fatal("initial resize not sent")
	}
}

func setDeadline(rwc io.ReadWriteCloser, d time.Duration) error {
	if dl, ok := rwc.(interface{ SetReadDeadline(time.Time) error }); ok {
		return dl.SetReadDeadline(time.Now().Add(d))
	}
	return nil
}

func TestStartExecNon2xxReturnsError(t *testing.T) {
	c, srv := newTestDockerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"No such exec instance"}`))
	})
	defer srv.Close()
	defer c.Close()

	rwc, err := c.StartExec(context.Background(), "exec-1", 80, 24)
	if err == nil {
		_ = rwc.Close()
		t.Fatal("want error")
	}
	if !strings.Contains(err.Error(), "No such exec instance") {
		t.Fatalf("error should include body, got %v", err)
	}
}
