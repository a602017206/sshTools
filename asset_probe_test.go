package main

import (
	"net"
	"testing"
	"time"

	"AHaSSHTools/internal/config"
)

func TestAssetProbeTargetsFromConnectionsSkipsExcluded(t *testing.T) {
	conns := []config.ConnectionConfig{
		{ID: "keep", Host: "10.0.0.1", Port: 22},
		{ID: "skip", Host: "10.0.0.2", Port: 22, Metadata: map[string]string{AssetWallExcludedMetaKey: "true"}},
		{ID: "keep2", Host: "10.0.0.3", Port: 3306, Metadata: map[string]string{AssetWallExcludedMetaKey: "false"}},
	}
	targets := assetProbeTargetsFromConnections(conns)
	if len(targets) != 2 || targets[0].id != "keep" || targets[1].id != "keep2" {
		t.Fatalf("excluded asset not skipped: %+v", targets)
	}
}

func TestProbeAssetTargetReachable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port

	result := probeAssetTarget(assetProbeTarget{id: "a1", host: "127.0.0.1", port: port}, time.Second)
	if !result.Reachable {
		t.Fatalf("expected reachable, got error %q", result.Error)
	}
	if result.ID != "a1" || result.CheckedAt <= 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestProbeAssetTargetUnreachable(t *testing.T) {
	// 先占一个端口再释放，得到几乎必然关闭的端口
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	result := probeAssetTarget(assetProbeTarget{id: "a2", host: "127.0.0.1", port: port}, time.Second)
	if result.Reachable {
		t.Fatal("expected unreachable for closed port")
	}
	if result.Error == "" {
		t.Fatal("expected error detail for unreachable target")
	}
}

func TestProbeAssetTargetMissingHost(t *testing.T) {
	result := probeAssetTarget(assetProbeTarget{id: "a3", host: "", port: 22}, time.Second)
	if result.Reachable || result.Error == "" {
		t.Fatalf("expected config error, got %+v", result)
	}
}

func TestProbeAssetTargetsKeepsOrderAndCount(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	openPort := listener.Addr().(*net.TCPAddr).Port

	targets := []assetProbeTarget{
		{id: "open", host: "127.0.0.1", port: openPort},
		{id: "bad", host: "", port: 0},
	}
	results := probeAssetTargets(targets, time.Second)
	if len(results) != 2 || results[0].ID != "open" || results[1].ID != "bad" {
		t.Fatalf("order/count broken: %+v", results)
	}
	if !results[0].Reachable || results[1].Reachable {
		t.Fatalf("reachability broken: %+v", results)
	}
}
