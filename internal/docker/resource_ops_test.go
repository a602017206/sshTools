package docker_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestListImagesHTTP(t *testing.T) {
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/images/json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"Id":"sha256:abc","RepoTags":["nginx:latest"],"Size":1234,"Created":1700000000},{"Id":"sha256:def","RepoTags":null,"Size":5,"Created":1}]`))
	}))
	defer srv.Close()
	defer c.Close()

	imgs, err := c.ListImages(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(imgs) != 2 {
		t.Fatalf("len: got %d", len(imgs))
	}
	if imgs[0].ID != "sha256:abc" || len(imgs[0].Tags) != 1 || imgs[0].Tags[0] != "nginx:latest" || imgs[0].Size != 1234 || imgs[0].Created != 1700000000 {
		t.Fatalf("unexpected image: %+v", imgs[0])
	}
	if imgs[1].Tags == nil || len(imgs[1].Tags) != 0 {
		t.Fatalf("nil tags should map to empty slice: %+v", imgs[1])
	}
}

func TestInspectImageHTTP(t *testing.T) {
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/images/registry.local/app/web:1.0/json" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"Id":"sha256:abc","RepoTags":["registry.local/app/web:1.0"],"Size":99,"Created":"2023-11-14T22:13:20Z","Architecture":"arm64","Os":"linux","Author":"me","Comment":"hi"}`))
	}))
	defer srv.Close()
	defer c.Close()

	d, err := c.InspectImage(context.Background(), "registry.local/app/web:1.0")
	if err != nil {
		t.Fatal(err)
	}
	if d.ID != "sha256:abc" || d.Architecture != "arm64" || d.Os != "linux" || d.Author != "me" || d.Comment != "hi" || d.Size != 99 || d.Created != 1700000000 {
		t.Fatalf("unexpected detail: %+v", d)
	}
}

func TestInspectImageNotFound(t *testing.T) {
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"No such image"}`, http.StatusNotFound)
	}))
	defer srv.Close()
	defer c.Close()
	if _, err := c.InspectImage(context.Background(), "nope"); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("want 404 error, got %v", err)
	}
}

func TestRemoveImageForceQuery(t *testing.T) {
	var queries []string
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/images/nginx:latest" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		queries = append(queries, r.URL.RawQuery)
		_, _ = w.Write([]byte(`[{"Untagged":"nginx:latest"}]`))
	}))
	defer srv.Close()
	defer c.Close()

	ctx := context.Background()
	if err := c.RemoveImage(ctx, "nginx:latest", true); err != nil {
		t.Fatal(err)
	}
	if err := c.RemoveImage(ctx, "nginx:latest", false); err != nil {
		t.Fatal(err)
	}
	if len(queries) != 2 || queries[0] != "force=1" || queries[1] != "" {
		t.Fatalf("queries: %v", queries)
	}
}

func TestRemoveImageConflictError(t *testing.T) {
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"image is being used"}`, http.StatusConflict)
	}))
	defer srv.Close()
	defer c.Close()
	if err := c.RemoveImage(context.Background(), "x", false); err == nil || !strings.Contains(err.Error(), "409") {
		t.Fatalf("want 409 error, got %v", err)
	}
}
