package service

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"AHaSSHTools/internal/config"
)

type recordedReq struct{ method, uri string }

// newResourceAPIService 启动 mock Docker API（TCP），并返回已连接的 DockerService。
func newResourceAPIService(t *testing.T) (*DockerService, *[]recordedReq) {
	t.Helper()
	var mu sync.Mutex
	reqs := &[]recordedReq{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		*reqs = append(*reqs, recordedReq{r.Method, r.URL.RequestURI()})
		mu.Unlock()
		switch {
		case r.URL.Path == "/_ping":
			_, _ = w.Write([]byte("OK"))
		case r.Method == http.MethodGet && r.URL.Path == "/images/json":
			_, _ = w.Write([]byte(`[{"Id":"sha256:aa","RepoTags":["nginx:latest"],"Size":10,"Created":1}]`))
		case r.Method == http.MethodGet && r.URL.Path == "/images/registry.io/team/app:1.0/json":
			_, _ = w.Write([]byte(`{"Id":"sha256:bb","RepoTags":["registry.io/team/app:1.0"],"Architecture":"amd64","Os":"linux"}`))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/images/"):
			_, _ = w.Write([]byte(`[]`))
		case r.Method == http.MethodGet && r.URL.Path == "/networks":
			_, _ = w.Write([]byte(`[{"Id":"n1","Name":"bridge","Driver":"bridge","Scope":"local"}]`))
		case r.Method == http.MethodGet && r.URL.Path == "/networks/bridge":
			_, _ = w.Write([]byte(`{"Id":"n1","Name":"bridge","Driver":"bridge","Scope":"local","Internal":true}`))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/networks/"):
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/volumes":
			_, _ = w.Write([]byte(`{"Volumes":[{"Name":"data","Driver":"local","Mountpoint":"/v/data"}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/volumes/data":
			_, _ = w.Write([]byte(`{"Name":"data","Driver":"local","Scope":"local","Labels":{"a":"b"}}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/volumes/busy":
			http.Error(w, "volume is in use", http.StatusConflict)
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/volumes/"):
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	host, portStr, _ := net.SplitHostPort(srv.Listener.Addr().String())
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	svc := NewDockerService(nil)
	if err := svc.ConnectDocker("api", config.ConnectionConfig{Host: host, Port: port, Type: "docker"}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = svc.CloseAll() })
	return svc, reqs
}

func TestDockerResourcesAPI(t *testing.T) {
	svc, reqs := newResourceAPIService(t)

	imgs, err := svc.ListDockerImages("api")
	if err != nil || len(imgs) != 1 || imgs[0].Tags[0] != "nginx:latest" || imgs[0].Via != "" {
		t.Fatalf("images=%+v err=%v", imgs, err)
	}
	nets, err := svc.ListDockerNetworks("api")
	if err != nil || len(nets) != 1 || nets[0].Name != "bridge" {
		t.Fatalf("nets=%+v err=%v", nets, err)
	}
	vols, err := svc.ListDockerVolumes("api")
	if err != nil || len(vols) != 1 || vols[0].Name != "data" {
		t.Fatalf("vols=%+v err=%v", vols, err)
	}

	img, err := svc.InspectDockerImage("api", "registry.io/team/app:1.0")
	if err != nil || img.Architecture != "amd64" {
		t.Fatalf("img=%+v err=%v", img, err)
	}
	nd, err := svc.InspectDockerNetwork("api", "bridge")
	if err != nil || !nd.Internal {
		t.Fatalf("net=%+v err=%v", nd, err)
	}
	vd, err := svc.InspectDockerVolume("api", "data")
	if err != nil || vd.Labels["a"] != "b" {
		t.Fatalf("vol=%+v err=%v", vd, err)
	}

	*reqs = nil
	if err := svc.RemoveDockerImage("api", "nginx:latest", false); err != nil {
		t.Fatal(err)
	}
	if err := svc.RemoveDockerImage("api", "nginx:latest", true); err != nil {
		t.Fatal(err)
	}
	if err := svc.RemoveDockerNetwork("api", "bridge"); err != nil {
		t.Fatal(err)
	}
	if err := svc.RemoveDockerVolume("api", "data"); err != nil {
		t.Fatal(err)
	}
	got := make([]string, len(*reqs))
	for i, r := range *reqs {
		got[i] = r.method + " " + r.uri
	}
	want := []string{
		"DELETE /images/nginx:latest",
		"DELETE /images/nginx:latest?force=1",
		"DELETE /networks/bridge",
		"DELETE /volumes/data", // force=false：不带 force 参数
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("requests=%v", got)
	}

	err = svc.RemoveDockerVolume("api", "busy")
	if err == nil || !strings.Contains(err.Error(), "in use") {
		t.Fatalf("err=%v", err)
	}
}

func TestDockerResourcesCLI(t *testing.T) {
	sess := &fakeDockerSSH{dialErr: errors.New("x"), output: cliJSONLine}
	svc := newCLIDockerService(t, sess)

	sess.output = `{"ID":"abc","Repository":"nginx","Tag":"latest","Size":"10MB"}` + "\n"
	imgs, err := svc.ListDockerImages("s1")
	if err != nil || len(imgs) != 1 || imgs[0].Via != "cli" || imgs[0].Tags[0] != "nginx:latest" {
		t.Fatalf("imgs=%+v err=%v", imgs, err)
	}
	sess.output = `{"ID":"n1","Name":"bridge","Driver":"bridge","Scope":"local"}` + "\n"
	nets, err := svc.ListDockerNetworks("s1")
	if err != nil || len(nets) != 1 || nets[0].Via != "cli" {
		t.Fatalf("nets=%+v err=%v", nets, err)
	}
	sess.output = `{"Name":"data","Driver":"local"}` + "\n"
	vols, err := svc.ListDockerVolumes("s1")
	if err != nil || len(vols) != 1 || vols[0].Via != "cli" {
		t.Fatalf("vols=%+v err=%v", vols, err)
	}

	sess.output = `[{"Id":"sha256:bb","RepoTags":["a:1"],"Architecture":"arm64"}]`
	img, err := svc.InspectDockerImage("s1", "a:1")
	if err != nil || img.Architecture != "arm64" || img.Via != "cli" {
		t.Fatalf("img=%+v err=%v", img, err)
	}
	sess.output = `[{"Id":"n1","Name":"bridge"}]`
	if nd, err := svc.InspectDockerNetwork("s1", "bridge"); err != nil || nd.Via != "cli" {
		t.Fatalf("nd=%+v err=%v", nd, err)
	}
	sess.output = `[{"Name":"data"}]`
	if vd, err := svc.InspectDockerVolume("s1", "data"); err != nil || vd.Via != "cli" {
		t.Fatalf("vd=%+v err=%v", vd, err)
	}

	sess.output = ""
	sess.commands = nil
	for _, err := range []error{
		svc.RemoveDockerImage("s1", "nginx:latest", false),
		svc.RemoveDockerImage("s1", "nginx:latest", true),
		svc.RemoveDockerNetwork("s1", "bridge"),
		svc.RemoveDockerVolume("s1", "data"),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	want := []string{
		"docker rmi 'nginx:latest'",
		"docker rmi -f 'nginx:latest'",
		"docker network rm 'bridge'",
		"docker volume rm 'data'", // 不带 -f
	}
	if strings.Join(sess.commands, "|") != strings.Join(want, "|") {
		t.Fatalf("commands=%v", sess.commands)
	}

	sess.output = "Error response from daemon: network has active endpoints"
	sess.outputErr = errors.New("exit status 1")
	if err := svc.RemoveDockerNetwork("s1", "bridge"); err == nil || !strings.Contains(err.Error(), "active endpoints") {
		t.Fatalf("err=%v", err)
	}
}

func TestDockerResourcesUnknownSession(t *testing.T) {
	svc := NewDockerService(nil)
	if _, err := svc.ListDockerImages("nope"); err == nil {
		t.Fatal("list images")
	}
	if _, err := svc.ListDockerNetworks("nope"); err == nil {
		t.Fatal("list networks")
	}
	if _, err := svc.ListDockerVolumes("nope"); err == nil {
		t.Fatal("list volumes")
	}
	if _, err := svc.InspectDockerImage("nope", "a"); err == nil {
		t.Fatal("inspect image")
	}
	if err := svc.RemoveDockerVolume("nope", "a"); err == nil {
		t.Fatal("remove volume")
	}
}

func TestDockerResourcesValidation(t *testing.T) {
	sess := &fakeDockerSSH{dialErr: errors.New("x"), output: cliJSONLine}
	svc := newCLIDockerService(t, sess)

	badImages := []string{"", "  ", "-f", "a b", "a;rm -rf /", "a'b", "a$(x)"}
	for _, ref := range badImages {
		if err := svc.RemoveDockerImage("s1", ref, false); err == nil {
			t.Errorf("image %q should be rejected", ref)
		}
		if _, err := svc.InspectDockerImage("s1", ref); err == nil {
			t.Errorf("inspect image %q should be rejected", ref)
		}
	}
	badNames := []string{"", " ", "-x", "a b", "a/b", "a:b", "a;b", "a'b", "a\nb"}
	for _, n := range badNames {
		if err := svc.RemoveDockerNetwork("s1", n); err == nil {
			t.Errorf("network %q should be rejected", n)
		}
		if _, err := svc.InspectDockerNetwork("s1", n); err == nil {
			t.Errorf("inspect network %q should be rejected", n)
		}
		if err := svc.RemoveDockerVolume("s1", n); err == nil {
			t.Errorf("volume %q should be rejected", n)
		}
		if _, err := svc.InspectDockerVolume("s1", n); err == nil {
			t.Errorf("inspect volume %q should be rejected", n)
		}
	}
	if len(sess.commands) != 0 {
		t.Fatalf("invalid input must not reach CLI: %v", sess.commands)
	}

	for _, ref := range []string{"nginx", "nginx:1.25", "registry.io:5000/team/app:1.0", "nginx@sha256:abcdef0123", "sha256:abc"} {
		if _, err := validateDockerImageRef(ref); err != nil {
			t.Errorf("image %q should be valid: %v", ref, err)
		}
	}
}
