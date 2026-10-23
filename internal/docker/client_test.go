package docker_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"AHaSSHTools/internal/docker"
)

func TestPingAndListContainersOverHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_ping":
			w.Header().Set("API-Version", "1.44")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		case "/containers/json":
			if r.URL.Query().Get("all") != "1" {
				t.Fatalf("expected all=1, got %q", r.URL.Query().Get("all"))
			}
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{
					"Id":      "abcdef0123456789",
					"Names":   []string{"/web"},
					"Image":   "nginx:latest",
					"Status":  "Up 2 hours",
					"Created": int64(1700000000),
					"Ports":   []map[string]any{{"PublicPort": 8080, "PrivatePort": 80, "Type": "tcp"}},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	host, port := mustHostPort(t, srv.URL)
	c, err := docker.NewClient(docker.Config{Transport: docker.TransportTCP, Host: host, Port: port})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := c.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	list, err := c.ListContainers(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Name != "web" || list[0].Image != "nginx:latest" {
		t.Fatalf("unexpected list: %+v", list)
	}
}

func mustHostPort(t *testing.T, raw string) (string, int) {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}
	return u.Hostname(), port
}
