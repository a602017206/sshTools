package service

import "AHaSSHTools/internal/config"

type DriverView struct {
	ID                 string                     `json:"id"`
	Name               string                     `json:"name"`
	RecommendedVersion string                     `json:"recommendedVersion"`
	Installed          bool                       `json:"installed"`
	Profiles           []config.JDBCDriverProfile `json:"profiles"`
}

type RuntimeStatus struct {
	Kind     RuntimeKind `json:"kind"`
	JavaPath string      `json:"javaPath"`
	Version  string      `json:"version"`
}

type JDBCAgentState string

const (
	JDBCAgentStateStopped  JDBCAgentState = "stopped"
	JDBCAgentStateStarting JDBCAgentState = "starting"
	JDBCAgentStateRunning  JDBCAgentState = "running"
	JDBCAgentStateFailed   JDBCAgentState = "failed"
)

type JDBCAgentStatus struct {
	State       JDBCAgentState `json:"state"`
	RuntimeKind RuntimeKind    `json:"runtimeKind"`
	LastError   string         `json:"lastError"`
}
