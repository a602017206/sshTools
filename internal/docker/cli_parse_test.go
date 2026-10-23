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
