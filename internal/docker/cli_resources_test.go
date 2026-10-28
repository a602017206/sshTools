package docker_test

import (
	"strings"
	"testing"

	"AHaSSHTools/internal/docker"
)

func TestBuildDockerResourceCommands(t *testing.T) {
	tests := []struct{ name, got, want string }{
		{"images", docker.BuildDockerImagesCommand(), "docker images --format '{{json .}}'"},
		{"networks", docker.BuildDockerNetworksCommand(), "docker network ls --format '{{json .}}'"},
		{"volumes", docker.BuildDockerVolumesCommand(), "docker volume ls --format '{{json .}}'"},
		{"rmi", docker.BuildDockerRmiCommand("nginx:1", false), "docker rmi 'nginx:1'"},
		{"rmi force", docker.BuildDockerRmiCommand("nginx:1", true), "docker rmi -f 'nginx:1'"},
		{"network rm", docker.BuildDockerNetworkRmCommand("n1"), "docker network rm 'n1'"},
		{"volume rm", docker.BuildDockerVolumeRmCommand("v1", false), "docker volume rm 'v1'"},
		{"volume rm force", docker.BuildDockerVolumeRmCommand("v1", true), "docker volume rm -f 'v1'"},
		{"image inspect", docker.BuildDockerImageInspectCommand("i1"), "docker image inspect 'i1'"},
		{"network inspect", docker.BuildDockerNetworkInspectCommand("n1"), "docker network inspect 'n1'"},
		{"volume inspect", docker.BuildDockerVolumeInspectCommand("v1"), "docker volume inspect 'v1'"},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

func TestBuildDockerResourceCommandsShellEscape(t *testing.T) {
	evil := "x'; echo pwned; '"
	quoted := `'x'\''; echo pwned; '\'''`
	for name, got := range map[string]string{
		"rmi":     docker.BuildDockerRmiCommand(evil, false),
		"network": docker.BuildDockerNetworkRmCommand(evil),
		"volume":  docker.BuildDockerVolumeRmCommand(evil, false),
		"iinsp":   docker.BuildDockerImageInspectCommand(evil),
		"ninsp":   docker.BuildDockerNetworkInspectCommand(evil),
		"vinsp":   docker.BuildDockerVolumeInspectCommand(evil),
	} {
		if !strings.HasSuffix(got, " "+quoted) {
			t.Errorf("%s not quoted: %q", name, got)
		}
	}
}

const sampleImagesLines = `WARNING: something
{"Containers":"N/A","CreatedAt":"2023-11-14 22:13:30 +0000 UTC","CreatedSince":"3 weeks ago","Digest":"<none>","ID":"abc123def456","Repository":"nginx","SharedSize":"N/A","Size":"187MB","Tag":"latest","UniqueSize":"N/A","VirtualSize":"187.1MB"}

{"ID":"fff000fff000","Repository":"<none>","Tag":"<none>","Size":"1.5GB"}
`

func TestParseDockerImagesJSONLines(t *testing.T) {
	imgs, err := docker.ParseDockerImagesJSONLines(sampleImagesLines)
	if err != nil {
		t.Fatal(err)
	}
	if len(imgs) != 2 {
		t.Fatalf("len=%d: %+v", len(imgs), imgs)
	}
	if imgs[0].ID != "abc123def456" || len(imgs[0].Tags) != 1 || imgs[0].Tags[0] != "nginx:latest" {
		t.Fatalf("img0: %+v", imgs[0])
	}
	if imgs[0].Size != 187_000_000 || imgs[0].Created != 1700000010 {
		t.Fatalf("img0 size/created: %+v", imgs[0])
	}
	if imgs[0].Via != "cli" {
		t.Fatalf("via: %q", imgs[0].Via)
	}
	if len(imgs[1].Tags) != 0 || imgs[1].Size != 1_500_000_000 {
		t.Fatalf("img1: %+v", imgs[1])
	}
}

func TestParseDockerImagesMergesTagsByID(t *testing.T) {
	out := `{"ID":"abc123","Repository":"nginx","Tag":"latest","Size":"187MB"}` + "\n" +
		`{"ID":"abc123","Repository":"nginx","Tag":"1.25","Size":"187MB"}` + "\n" +
		`{"ID":"abc123","Repository":"nginx","Tag":"latest","Size":"187MB"}` + "\n" +
		`{"ID":"def456","Repository":"redis","Tag":"7","Size":"100MB"}`
	imgs, err := docker.ParseDockerImagesJSONLines(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(imgs) != 2 {
		t.Fatalf("len=%d: %+v", len(imgs), imgs)
	}
	if imgs[0].ID != "abc123" || len(imgs[0].Tags) != 2 || imgs[0].Tags[0] != "nginx:latest" || imgs[0].Tags[1] != "nginx:1.25" {
		t.Fatalf("merged tags: %+v", imgs[0])
	}
	if imgs[1].ID != "def456" || len(imgs[1].Tags) != 1 {
		t.Fatalf("img1: %+v", imgs[1])
	}
}

func TestParseDockerImagesSizeFallback(t *testing.T) {
	imgs, err := docker.ParseDockerImagesJSONLines(`{"ID":"a","Repository":"r","Tag":"t","Size":"weird"}` + "\n" + `{"ID":"b","Repository":"r","Tag":"t","Size":"512kB"}`)
	if err != nil {
		t.Fatal(err)
	}
	if imgs[0].Size != 0 || imgs[1].Size != 512_000 {
		t.Fatalf("sizes: %d %d", imgs[0].Size, imgs[1].Size)
	}
}

func TestParseDockerNetworksJSONLines(t *testing.T) {
	out := "{\"ID\":\"n1\",\"Name\":\"bridge\",\"Driver\":\"bridge\",\"Scope\":\"local\",\"IPv6\":\"false\"}\nnot json\n{\"ID\":\"n2\",\"Name\":\"host\",\"Driver\":\"host\",\"Scope\":\"local\"}\n"
	nets, err := docker.ParseDockerNetworksJSONLines(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(nets) != 2 || nets[0].Name != "bridge" || nets[1].Driver != "host" || nets[0].Via != "cli" {
		t.Fatalf("nets: %+v", nets)
	}
}

func TestParseDockerVolumesJSONLines(t *testing.T) {
	out := `{"Driver":"local","Name":"data","Mountpoint":"/var/lib/docker/volumes/data/_data","Scope":"local"}` + "\n"
	vols, err := docker.ParseDockerVolumesJSONLines(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(vols) != 1 || vols[0].Name != "data" || vols[0].Driver != "local" || vols[0].Via != "cli" {
		t.Fatalf("vols: %+v", vols)
	}
}

func TestParseDockerResourcesBadJSON(t *testing.T) {
	if _, err := docker.ParseDockerImagesJSONLines("{bad"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCLIListResources(t *testing.T) {
	sess := &fakeCLISession{output: `{"ID":"n1","Name":"bridge","Driver":"bridge","Scope":"local"}`}
	nets, err := docker.CLIListNetworks(sess)
	if err != nil || len(nets) != 1 {
		t.Fatalf("nets=%v err=%v", nets, err)
	}
	if sess.commands[0] != "docker network ls --format '{{json .}}'" {
		t.Fatalf("cmd: %q", sess.commands[0])
	}
	sess = &fakeCLISession{output: `{"ID":"i","Repository":"r","Tag":"t","Size":"1MB"}`}
	imgs, err := docker.CLIListImages(sess)
	if err != nil || len(imgs) != 1 || imgs[0].Via != "cli" {
		t.Fatalf("imgs=%v err=%v", imgs, err)
	}
	sess = &fakeCLISession{output: `{"Name":"v","Driver":"local"}`}
	vols, err := docker.CLIListVolumes(sess)
	if err != nil || len(vols) != 1 {
		t.Fatalf("vols=%v err=%v", vols, err)
	}
	if _, err := docker.CLIListImages(nil); err == nil {
		t.Fatal("expected nil session error")
	}
}

func TestCLIInspectResources(t *testing.T) {
	sess := &fakeCLISession{output: `[{"Id":"sha256:abc","RepoTags":["nginx:latest"],"Size":100,"Created":"2023-11-14T22:13:30Z","Architecture":"amd64","Os":"linux"}]`}
	img, err := docker.CLIInspectImage(sess, "nginx")
	if err != nil || img.Architecture != "amd64" || img.Tags[0] != "nginx:latest" || img.Created == 0 {
		t.Fatalf("img=%+v err=%v", img, err)
	}
	sess = &fakeCLISession{output: `[{"Id":"n1","Name":"b","Driver":"bridge","Scope":"local","Internal":true,"IPAM":{"Config":[{"Subnet":"172.17.0.0/16","Gateway":"172.17.0.1"}]}}]`}
	net, err := docker.CLIInspectNetwork(sess, "n1")
	if err != nil || !net.Internal || net.IPAM != "172.17.0.0/16 gw 172.17.0.1" {
		t.Fatalf("net=%+v err=%v", net, err)
	}
	sess = &fakeCLISession{output: `[{"Name":"v","Driver":"local","Labels":null,"Scope":"local","CreatedAt":"2023-11-14T22:13:30Z"}]`}
	vol, err := docker.CLIInspectVolume(sess, "v")
	if err != nil || vol.Scope != "local" || vol.Labels == nil || vol.Created == 0 {
		t.Fatalf("vol=%+v err=%v", vol, err)
	}
	sess = &fakeCLISession{output: "[]"}
	if _, err := docker.CLIInspectImage(sess, "x"); err == nil {
		t.Fatal("expected empty error")
	}
}

func TestCLIRemoveResources(t *testing.T) {
	sess := &fakeCLISession{}
	if err := docker.CLIRemoveImage(sess, "i", true); err != nil {
		t.Fatal(err)
	}
	if err := docker.CLIRemoveNetwork(sess, "n"); err != nil {
		t.Fatal(err)
	}
	if err := docker.CLIRemoveVolume(sess, "v", false); err != nil {
		t.Fatal(err)
	}
	want := []string{"docker rmi -f 'i'", "docker network rm 'n'", "docker volume rm 'v'"}
	for i, w := range want {
		if sess.commands[i] != w {
			t.Fatalf("cmd %d: %q want %q", i, sess.commands[i], w)
		}
	}
}
