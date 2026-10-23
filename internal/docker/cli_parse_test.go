package docker_test

import (
	"testing"

	"AHaSSHTools/internal/docker"
)

func TestParseDockerPSJSONLines(t *testing.T) {
	in := `{"ID":"abc123def456","Names":"web","Image":"nginx","Status":"Up 1 minute","Ports":"80/tcp","CreatedAt":"2024-01-02 03:04:05 +0000 UTC"}
{"ID":"fff111aaa222","Names":"db","Image":"postgres","Status":"Exited (0)","Ports":"","CreatedAt":"2024-01-01 00:00:00 +0000 UTC"}`
	list, err := docker.ParseDockerPSJSONLines(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].Name != "web" || list[1].Status == "" {
		t.Fatalf("%+v", list)
	}
}

func TestParseDockerPSJSONLinesEmpty(t *testing.T) {
	list, err := docker.ParseDockerPSJSONLines("\n")
	if err != nil || len(list) != 0 {
		t.Fatalf("list=%v err=%v", list, err)
	}
}

func TestParseDockerPSJSONLinesNormalizesIDAndName(t *testing.T) {
	in := `{"ID":"abc123def456789012345","Names":"/web","Image":"nginx","Status":"Up","Ports":"","CreatedAt":"2024-01-02 03:04:05 +0000 UTC"}`
	list, err := docker.ParseDockerPSJSONLines(in)
	if err != nil {
		t.Fatal(err)
	}
	if list[0].ID != "abc123def456" || list[0].Name != "web" {
		t.Fatalf("%+v", list[0])
	}
}

func TestParseDockerPSJSONLinesNamesArrayAndCommaList(t *testing.T) {
	in := `{"ID":"aaaaaaaaaaaa","Names":["/api","/alias"],"Image":"x","Status":"Up","Ports":["80/tcp"]}
{"ID":"bbbbbbbbbbbb","Names":"db,db-alias","Image":"y","Status":"Up","Ports":"5432/tcp"}`
	list, err := docker.ParseDockerPSJSONLines(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].Name != "api" || list[1].Name != "db" {
		t.Fatalf("%+v", list)
	}
	if list[0].Ports != "80/tcp" {
		t.Fatalf("ports=%q", list[0].Ports)
	}
}

func TestParseDockerPSJSONLinesRealDockerOutputWithExtraFields(t *testing.T) {
	in := `{"Command":"\"/docker-entrypoint.…\"","CreatedAt":"2024-05-06 07:08:09 +0800 CST","ID":"0123456789ab","Image":"nginx:latest","Labels":"a=b","LocalVolumes":"0","Mounts":"","Names":"web","Networks":"bridge","Ports":"0.0.0.0:80->80/tcp","RunningFor":"2 hours ago","Size":"0B","State":"running","Status":"Up 2 hours"}`
	list, err := docker.ParseDockerPSJSONLines(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Created == 0 || list[0].Ports != "0.0.0.0:80->80/tcp" {
		t.Fatalf("%+v", list)
	}
}
