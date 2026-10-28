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

func TestListNetworksHTTP(t *testing.T) {
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/networks" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"Id":"abc123","Name":"bridge","Driver":"bridge","Scope":"local"},{"Id":"def","Name":"host","Driver":"host","Scope":"local"}]`))
	}))
	defer srv.Close()
	defer c.Close()

	nets, err := c.ListNetworks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(nets) != 2 || nets[0].ID != "abc123" || nets[0].Name != "bridge" || nets[0].Driver != "bridge" || nets[0].Scope != "local" || nets[0].Via != "" {
		t.Fatalf("unexpected networks: %+v", nets)
	}
}

func TestInspectNetworkHTTP(t *testing.T) {
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/networks/abc123" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"Id":"abc123","Name":"bridge","Driver":"bridge","Scope":"local","Internal":true,"Attachable":true,"Ingress":false,"IPAM":{"Config":[{"Subnet":"172.17.0.0/16","Gateway":"172.17.0.1"},{"Subnet":"10.0.0.0/24"}]}}`))
	}))
	defer srv.Close()
	defer c.Close()

	d, err := c.InspectNetwork(context.Background(), "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if d.ID != "abc123" || d.Name != "bridge" || !d.Internal || !d.Attachable || d.Ingress || d.IPAM != "172.17.0.0/16 gw 172.17.0.1, 10.0.0.0/24" {
		t.Fatalf("unexpected detail: %+v", d)
	}
}

func TestInspectNetworkNotFound(t *testing.T) {
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"network not found"}`, http.StatusNotFound)
	}))
	defer srv.Close()
	defer c.Close()
	if _, err := c.InspectNetwork(context.Background(), "nope"); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("want 404 error, got %v", err)
	}
}

func TestRemoveNetworkHTTP(t *testing.T) {
	var got string
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Method + " " + r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	defer c.Close()
	if err := c.RemoveNetwork(context.Background(), "abc123"); err != nil {
		t.Fatal(err)
	}
	if got != "DELETE /networks/abc123" {
		t.Fatalf("request: %s", got)
	}
}

func TestRemoveNetworkConflictError(t *testing.T) {
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"network has active endpoints"}`, http.StatusForbidden)
	}))
	defer srv.Close()
	defer c.Close()
	if err := c.RemoveNetwork(context.Background(), "x"); err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("want 403 error, got %v", err)
	}
}

func TestListVolumesHTTP(t *testing.T) {
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/volumes" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"Volumes":[{"Name":"data","Driver":"local","Mountpoint":"/var/lib/docker/volumes/data/_data","CreatedAt":"2023-11-14T22:13:20Z"}],"Warnings":null}`))
	}))
	defer srv.Close()
	defer c.Close()

	vols, err := c.ListVolumes(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(vols) != 1 || vols[0].Name != "data" || vols[0].Driver != "local" || vols[0].Mountpoint != "/var/lib/docker/volumes/data/_data" || vols[0].Created != 1700000000 {
		t.Fatalf("unexpected volumes: %+v", vols)
	}
}

func TestListVolumesNullVolumes(t *testing.T) {
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Volumes":null,"Warnings":null}`))
	}))
	defer srv.Close()
	defer c.Close()
	vols, err := c.ListVolumes(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if vols == nil || len(vols) != 0 {
		t.Fatalf("want empty non-nil slice, got %#v", vols)
	}
}

func TestInspectVolumeHTTP(t *testing.T) {
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/volumes/data" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"Name":"data","Driver":"local","Mountpoint":"/mnt/data","CreatedAt":"2023-11-14T22:13:20Z","Labels":{"app":"web"},"Scope":"local"}`))
	}))
	defer srv.Close()
	defer c.Close()

	d, err := c.InspectVolume(context.Background(), "data")
	if err != nil {
		t.Fatal(err)
	}
	if d.Name != "data" || d.Scope != "local" || d.Labels["app"] != "web" || d.Created != 1700000000 {
		t.Fatalf("unexpected detail: %+v", d)
	}
}

func TestInspectVolumeNilLabels(t *testing.T) {
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Name":"v","Driver":"local","Labels":null}`))
	}))
	defer srv.Close()
	defer c.Close()
	d, err := c.InspectVolume(context.Background(), "v")
	if err != nil {
		t.Fatal(err)
	}
	if d.Labels == nil {
		t.Fatal("nil labels should map to empty map")
	}
}

func TestRemoveVolumeForceQuery(t *testing.T) {
	var queries []string
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/volumes/data" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		queries = append(queries, r.URL.RawQuery)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	defer c.Close()
	ctx := context.Background()
	if err := c.RemoveVolume(ctx, "data", true); err != nil {
		t.Fatal(err)
	}
	if err := c.RemoveVolume(ctx, "data", false); err != nil {
		t.Fatal(err)
	}
	if len(queries) != 2 || queries[0] != "force=1" || queries[1] != "" {
		t.Fatalf("queries: %v", queries)
	}
}

func TestRemoveVolumeConflictError(t *testing.T) {
	c, srv := newTestDockerClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"volume in use"}`, http.StatusConflict)
	}))
	defer srv.Close()
	defer c.Close()
	if err := c.RemoveVolume(context.Background(), "x", false); err == nil || !strings.Contains(err.Error(), "409") {
		t.Fatalf("want 409 error, got %v", err)
	}
}
