package docker_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"AHaSSHTools/internal/docker"
)

func newTestDockerClient(t *testing.T, handler http.HandlerFunc) (*docker.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	host, port := mustHostPort(t, srv.URL)
	c, err := docker.NewClient(docker.Config{Transport: docker.TransportTCP, Host: host, Port: port})
	if err != nil {
		t.Fatal(err)
	}
	return c, srv
}

func TestStartStopRestartContainerHTTP(t *testing.T) {
	const id = "abc123def456"
	var got []string
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.Method+" "+r.URL.Path)
		switch r.URL.Path {
		case "/containers/" + id + "/start":
			if r.Method != http.MethodPost {
				t.Fatalf("start: want POST, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		case "/containers/" + id + "/stop":
			if r.Method != http.MethodPost {
				t.Fatalf("stop: want POST, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		case "/containers/" + id + "/restart":
			if r.Method != http.MethodPost {
				t.Fatalf("restart: want POST, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNotModified)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	defer c.Close()

	ctx := context.Background()
	if err := c.StartContainer(ctx, id); err != nil {
		t.Fatal(err)
	}
	if err := c.StopContainer(ctx, id); err != nil {
		t.Fatal(err)
	}
	if err := c.RestartContainer(ctx, id); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"POST /containers/" + id + "/start",
		"POST /containers/" + id + "/stop",
		"POST /containers/" + id + "/restart",
	}
	if len(got) != len(want) {
		t.Fatalf("requests: got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("request[%d]: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRemoveContainerForceQuery(t *testing.T) {
	const id = "deadbeef"
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/containers/"+id {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodDelete {
			t.Fatalf("want DELETE, got %s", r.Method)
		}
		if r.URL.Query().Get("force") != "1" {
			t.Fatalf("expected force=1, got %q", r.URL.Query().Get("force"))
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	defer c.Close()

	if err := c.RemoveContainer(context.Background(), id, true); err != nil {
		t.Fatal(err)
	}
}

func TestInspectContainerMapsDetail(t *testing.T) {
	const id = "abcdef0123456789"
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/containers/"+id+"/json" || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"Id":      id,
			"Name":    "/my-app",
			"Image":   "sha256:abc",
			"Created": "2023-11-14T22:13:30.123456789Z",
			"State": map[string]any{
				"Status": "running",
			},
			"Config": map[string]any{
				"Image":      "nginx:latest",
				"Env":        []string{"FOO=bar"},
				"Cmd":        []string{"nginx", "-g", "daemon off;"},
				"Entrypoint": []string{"/docker-entrypoint.sh"},
			},
			"HostConfig": map[string]any{
				"NetworkMode": "bridge",
			},
			"Mounts": []map[string]any{
				{"Source": "/host/data", "Destination": "/data"},
			},
			"NetworkSettings": map[string]any{
				"Ports": map[string]any{
					"80/tcp": []map[string]any{{"HostIp": "0.0.0.0", "HostPort": "8080"}},
				},
			},
		})
	}))
	defer srv.Close()
	defer c.Close()

	detail, err := c.InspectContainer(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Name != "my-app" {
		t.Fatalf("Name: got %q, want my-app", detail.Name)
	}
	if detail.ID != "abcdef012345" {
		t.Fatalf("ID: got %q", detail.ID)
	}
	if detail.Image != "nginx:latest" {
		t.Fatalf("Image: got %q", detail.Image)
	}
	if detail.State != "running" || detail.Status != "running" {
		t.Fatalf("state/status: %+v", detail)
	}
	if detail.NetworkMode != "bridge" {
		t.Fatalf("NetworkMode: got %q", detail.NetworkMode)
	}
	if detail.Created != 1700000010 {
		t.Fatalf("Created: got %d, want 1700000010", detail.Created)
	}
	if len(detail.Mounts) != 1 || detail.Mounts[0] != "/host/data -> /data" {
		t.Fatalf("Mounts: %+v", detail.Mounts)
	}
	if len(detail.Env) != 1 || detail.Env[0] != "FOO=bar" {
		t.Fatalf("Env: %+v", detail.Env)
	}
	if !strings.Contains(detail.Ports, "8080") || !strings.Contains(detail.Ports, "80") {
		t.Fatalf("Ports: got %q", detail.Ports)
	}
}
