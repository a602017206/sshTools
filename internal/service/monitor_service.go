package service

import (
	"sshTools/internal/ssh"
)

// MonitorService handles system monitoring operations
type MonitorService struct {
	sessionManager   *ssh.SessionManager
	monitorCollector *ssh.MonitorCollector
}

// NewMonitorService creates a new monitor service
func NewMonitorService(sm *ssh.SessionManager) *MonitorService {
	return &MonitorService{
		sessionManager:   sm,
		monitorCollector: ssh.NewMonitorCollector(sm),
	}
}

// GetMonitoringData retrieves monitoring data for a session
func (s *MonitorService) GetMonitoringData(sessionID string) (*ssh.MonitoringData, error) {
	return s.monitorCollector.CollectMetrics(sessionID)
}
