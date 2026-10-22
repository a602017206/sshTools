package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"AHaSSHTools/internal/config"
)

// AssetProbeResult 是单个资产的一次 TCP 探活结果（状态墙用）。
// 只做网络层可达性探测，不登录、不需要凭据。
type AssetProbeResult struct {
	ID        string `json:"id"`
	Reachable bool   `json:"reachable"`
	LatencyMs int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
	CheckedAt int64  `json:"checked_at"`
}

type assetProbeTarget struct {
	id   string
	host string
	port int
}

const (
	assetProbeTimeout     = 2 * time.Second
	assetProbeConcurrency = 16
)

// AssetWallExcludedMetaKey 是连接 metadata 中「不参与状态墙探活」的标记键。
const AssetWallExcludedMetaKey = "status_wall_excluded"

// ProbeAllAssets 对全部已保存连接做一轮 TCP 探活，供 PC 状态墙展示在线状态与延迟。
// 在连接上标记了 status_wall_excluded=true 的资产会被跳过。
func (a *App) ProbeAllAssets() []AssetProbeResult {
	conns, _ := a.connectionService.GetConnections()
	return probeAssetTargets(assetProbeTargetsFromConnections(conns), assetProbeTimeout)
}

// SetAssetProbeExcluded 设置单个资产是否参与状态墙探活，持久化在连接 metadata 中。
func (a *App) SetAssetProbeExcluded(id string, excluded bool) error {
	conn, err := a.connectionService.GetConnection(id)
	if err != nil {
		return err
	}
	if conn.Metadata == nil {
		conn.Metadata = map[string]string{}
	}
	if excluded {
		conn.Metadata[AssetWallExcludedMetaKey] = "true"
	} else {
		delete(conn.Metadata, AssetWallExcludedMetaKey)
	}
	return a.connectionService.UpdateConnection(conn)
}

func assetProbeTargetsFromConnections(conns []config.ConnectionConfig) []assetProbeTarget {
	targets := make([]assetProbeTarget, 0, len(conns))
	for _, conn := range conns {
		if conn.Metadata[AssetWallExcludedMetaKey] == "true" {
			continue
		}
		targets = append(targets, assetProbeTarget{id: conn.ID, host: conn.Host, port: conn.Port})
	}
	return targets
}

func probeAssetTargets(targets []assetProbeTarget, timeout time.Duration) []AssetProbeResult {
	results := make([]AssetProbeResult, len(targets))
	sem := make(chan struct{}, assetProbeConcurrency)
	var wg sync.WaitGroup
	for i, target := range targets {
		wg.Add(1)
		go func(i int, target assetProbeTarget) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i] = probeAssetTarget(target, timeout)
		}(i, target)
	}
	wg.Wait()
	return results
}

func probeAssetTarget(target assetProbeTarget, timeout time.Duration) AssetProbeResult {
	result := AssetProbeResult{ID: target.id, CheckedAt: time.Now().UnixMilli()}
	if target.host == "" || target.port <= 0 {
		result.Error = "未配置主机或端口"
		return result
	}
	address := net.JoinHostPort(target.host, strconv.Itoa(target.port))
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		result.Error = fmt.Sprintf("连接失败: %v", err)
		return result
	}
	_ = conn.Close()
	result.Reachable = true
	result.LatencyMs = time.Since(start).Milliseconds()
	return result
}
