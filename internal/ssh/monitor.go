package ssh

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func roundTo2Decimals(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

// SystemInfo contains basic system information
type SystemInfo struct {
	Hostname  string `json:"hostname"`
	Uptime    string `json:"uptime"`
	OS        string `json:"os"`
	Kernel    string `json:"kernel"`
	Username  string `json:"username"`
	Processes int    `json:"processes"`
}

// CPUMetrics contains CPU usage information
type CPUMetrics struct {
	Overall     float64   `json:"overall"`      // Overall CPU %
	User        float64   `json:"user"`         // User mode %
	System      float64   `json:"system"`       // System mode %
	IOWait      float64   `json:"iowait"`       // IO wait %
	Idle        float64   `json:"idle"`         // Idle %
	PerCore     []float64 `json:"per_core"`     // Per-core usage
	LoadAverage []float64 `json:"load_average"` // 1, 5, 15 min
}

// MemoryMetrics contains memory usage information
type MemoryMetrics struct {
	Total       uint64  `json:"total"`        // Total memory (bytes)
	Used        uint64  `json:"used"`         // Used memory (bytes)
	Free        uint64  `json:"free"`         // Free memory (bytes)
	Available   uint64  `json:"available"`    // Available memory (bytes)
	UsedPercent float64 `json:"used_percent"` // Used %
	SwapTotal   uint64  `json:"swap_total"`
	SwapUsed    uint64  `json:"swap_used"`
	SwapFree    uint64  `json:"swap_free"`
}

// NetworkMetrics contains network statistics
type NetworkMetrics struct {
	TotalRxBytes uint64  `json:"total_rx_bytes"` // Total received
	TotalTxBytes uint64  `json:"total_tx_bytes"` // Total sent
	RxRate       float64 `json:"rx_rate"`        // RX rate (bytes/sec)
	TxRate       float64 `json:"tx_rate"`        // TX rate (bytes/sec)
}

// PartitionInfo contains disk partition information
type PartitionInfo struct {
	MountPoint  string  `json:"mount_point"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

// DiskMetrics contains disk usage information
type DiskMetrics struct {
	Partitions []PartitionInfo `json:"partitions"`
}

// MonitoringData aggregates all monitoring metrics
type MonitoringData struct {
	Timestamp int64          `json:"timestamp"`
	System    SystemInfo     `json:"system"`
	CPU       CPUMetrics     `json:"cpu"`
	Memory    MemoryMetrics  `json:"memory"`
	Network   NetworkMetrics `json:"network"`
	Disk      DiskMetrics    `json:"disk"`
}

// MonitorCollector collects system metrics via SSH
type MonitorCollector struct {
	sessionManager *SessionManager
	prevNetStats   map[string]*NetworkMetrics // sessionID -> previous stats
	prevTimestamp  map[string]int64           // sessionID -> timestamp
}

// NewMonitorCollector creates a new monitor collector
func NewMonitorCollector(sm *SessionManager) *MonitorCollector {
	return &MonitorCollector{
		sessionManager: sm,
		prevNetStats:   make(map[string]*NetworkMetrics),
		prevTimestamp:  make(map[string]int64),
	}
}

// CollectMetrics collects all metrics for a session
func (mc *MonitorCollector) CollectMetrics(sessionID string) (*MonitoringData, error) {
	data := &MonitoringData{
		Timestamp: time.Now().Unix(),
	}

	timeout := 5 * time.Second

	// Collect system info
	sysInfo, err := mc.collectSystemInfo(sessionID, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to collect system info: %w", err)
	}
	data.System = *sysInfo

	// Collect CPU metrics
	cpuMetrics, err := mc.collectCPUMetrics(sessionID, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to collect CPU metrics: %w", err)
	}
	data.CPU = *cpuMetrics

	// Collect memory metrics
	memMetrics, err := mc.collectMemoryMetrics(sessionID, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to collect memory metrics: %w", err)
	}
	data.Memory = *memMetrics

	// Collect network metrics
	netMetrics, err := mc.collectNetworkMetrics(sessionID, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to collect network metrics: %w", err)
	}
	data.Network = *netMetrics

	// Collect disk metrics
	diskMetrics, err := mc.collectDiskMetrics(sessionID, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to collect disk metrics: %w", err)
	}
	data.Disk = *diskMetrics

	return data, nil
}

// collectSystemInfo gathers basic system information
func (mc *MonitorCollector) collectSystemInfo(sessionID string, timeout time.Duration) (*SystemInfo, error) {
	info := &SystemInfo{}

	// Hostname
	stdout, _, err := mc.sessionManager.ExecuteCommand(sessionID, "hostname", timeout)
	if err == nil {
		info.Hostname = strings.TrimSpace(stdout)
	}

	// Uptime
	stdout, _, err = mc.sessionManager.ExecuteCommand(sessionID, "uptime -p", timeout)
	if err == nil {
		info.Uptime = strings.TrimSpace(stdout)
	}

	// OS
	stdout, _, err = mc.sessionManager.ExecuteCommand(sessionID, `cat /etc/os-release | grep PRETTY_NAME | cut -d '"' -f2`, timeout)
	if err == nil {
		info.OS = strings.TrimSpace(stdout)
	}

	// Kernel
	stdout, _, err = mc.sessionManager.ExecuteCommand(sessionID, "uname -r", timeout)
	if err == nil {
		info.Kernel = strings.TrimSpace(stdout)
	}

	// Username
	stdout, _, err = mc.sessionManager.ExecuteCommand(sessionID, "whoami", timeout)
	if err == nil {
		info.Username = strings.TrimSpace(stdout)
	}

	stdout, _, err = mc.sessionManager.ExecuteCommand(sessionID, `ps aux | wc -l`, timeout)
	if err == nil {
		count, _ := strconv.Atoi(strings.TrimSpace(stdout))
		info.Processes = count
	}

	return info, nil
}

// collectCPUMetrics gathers CPU usage statistics
func (mc *MonitorCollector) collectCPUMetrics(sessionID string, timeout time.Duration) (*CPUMetrics, error) {
	metrics := &CPUMetrics{}

	// Get overall CPU usage from top
	stdout, _, err := mc.sessionManager.ExecuteCommand(sessionID, `top -bn1 | grep "Cpu(s)"`, timeout)
	if err == nil {
		// Parse: "%Cpu(s):  0.8 us,  0.8 sy,  0.0 ni, 98.5 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st"
		// Use regex to extract each metric value reliably
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s+(\w+),?`)
		matches := re.FindAllStringSubmatch(stdout, -1)

		for _, match := range matches {
			if len(match) >= 3 {
				value, _ := strconv.ParseFloat(match[1], 64)
				key := match[2]

				switch key {
				case "us":
					metrics.User = value
				case "sy":
					metrics.System = value
				case "id":
					metrics.Idle = value
				case "wa":
					metrics.IOWait = value
				}
			}
		}

		metrics.Overall = 100.0 - metrics.Idle

		// Round all CPU metrics to 2 decimal places
		metrics.User = roundTo2Decimals(metrics.User)
		metrics.System = roundTo2Decimals(metrics.System)
		metrics.IOWait = roundTo2Decimals(metrics.IOWait)
		metrics.Idle = roundTo2Decimals(metrics.Idle)
		metrics.Overall = roundTo2Decimals(metrics.Overall)
	}

	// Load average
	stdout, _, err = mc.sessionManager.ExecuteCommand(sessionID, `uptime | awk -F'load average:' '{print $2}'`, timeout)
	if err == nil {
		parts := strings.Split(strings.TrimSpace(stdout), ",")
		for _, part := range parts {
			val, _ := strconv.ParseFloat(strings.TrimSpace(part), 64)
			metrics.LoadAverage = append(metrics.LoadAverage, val)
		}
	}

	return metrics, nil
}

// collectMemoryMetrics gathers memory usage statistics
func (mc *MonitorCollector) collectMemoryMetrics(sessionID string, timeout time.Duration) (*MemoryMetrics, error) {
	metrics := &MemoryMetrics{}

	// Physical memory
	stdout, _, err := mc.sessionManager.ExecuteCommand(sessionID, `free -b | grep Mem`, timeout)
	if err == nil {
		fields := strings.Fields(stdout)
		if len(fields) >= 7 {
			metrics.Total, _ = strconv.ParseUint(fields[1], 10, 64)
			metrics.Used, _ = strconv.ParseUint(fields[2], 10, 64)
			metrics.Free, _ = strconv.ParseUint(fields[3], 10, 64)
			metrics.Available, _ = strconv.ParseUint(fields[6], 10, 64)
			if metrics.Total > 0 {
				metrics.UsedPercent = float64(metrics.Used) / float64(metrics.Total) * 100.0
			}
		}
	}

	// Swap
	stdout, _, err = mc.sessionManager.ExecuteCommand(sessionID, `free -b | grep Swap`, timeout)
	if err == nil {
		fields := strings.Fields(stdout)
		if len(fields) >= 4 {
			metrics.SwapTotal, _ = strconv.ParseUint(fields[1], 10, 64)
			metrics.SwapUsed, _ = strconv.ParseUint(fields[2], 10, 64)
			metrics.SwapFree, _ = strconv.ParseUint(fields[3], 10, 64)
		}
	}

	return metrics, nil
}

// collectNetworkMetrics gathers network statistics
func (mc *MonitorCollector) collectNetworkMetrics(sessionID string, timeout time.Duration) (*NetworkMetrics, error) {
	metrics := &NetworkMetrics{}

	// Get current network stats
	stdout, _, err := mc.sessionManager.ExecuteCommand(sessionID, `cat /proc/net/dev | grep -v "lo:" | awk 'NR>2 {rx+=$2; tx+=$10} END {print rx, tx}'`, timeout)
	if err == nil {
		fields := strings.Fields(stdout)
		if len(fields) >= 2 {
			metrics.TotalRxBytes, _ = strconv.ParseUint(fields[0], 10, 64)
			metrics.TotalTxBytes, _ = strconv.ParseUint(fields[1], 10, 64)
		}
	}

	// Calculate rate if we have previous data
	now := time.Now().Unix()
	if prev, exists := mc.prevNetStats[sessionID]; exists {
		prevTime := mc.prevTimestamp[sessionID]
		timeDelta := float64(now - prevTime)
		if timeDelta > 0 {
			metrics.RxRate = float64(metrics.TotalRxBytes-prev.TotalRxBytes) / timeDelta
			metrics.TxRate = float64(metrics.TotalTxBytes-prev.TotalTxBytes) / timeDelta
		}
	}

	// Save current stats for next calculation
	mc.prevNetStats[sessionID] = &NetworkMetrics{
		TotalRxBytes: metrics.TotalRxBytes,
		TotalTxBytes: metrics.TotalTxBytes,
	}
	mc.prevTimestamp[sessionID] = now

	return metrics, nil
}

// collectDiskMetrics gathers disk usage statistics
func (mc *MonitorCollector) collectDiskMetrics(sessionID string, timeout time.Duration) (*DiskMetrics, error) {
	metrics := &DiskMetrics{
		Partitions: []PartitionInfo{},
	}

	stdout, _, err := mc.sessionManager.ExecuteCommand(sessionID, `df -B1 | grep "^/dev" | awk '{print $6, $2, $3, $4, $5}'`, timeout)
	if err == nil {
		lines := strings.Split(strings.TrimSpace(stdout), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				partition := PartitionInfo{
					MountPoint: fields[0],
				}
				partition.Total, _ = strconv.ParseUint(fields[1], 10, 64)
				partition.Used, _ = strconv.ParseUint(fields[2], 10, 64)
				partition.Free, _ = strconv.ParseUint(fields[3], 10, 64)
				percentStr := strings.TrimSuffix(fields[4], "%")
				partition.UsedPercent, _ = strconv.ParseFloat(percentStr, 64)

				metrics.Partitions = append(metrics.Partitions, partition)
			}
		}
	}

	return metrics, nil
}
