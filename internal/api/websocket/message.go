package websocket

import "time"

// Message represents a server-to-client WebSocket message
type Message struct {
	Type       string      `json:"type"`                  // "ssh:output", "transfer:progress", etc.
	SessionID  string      `json:"session_id,omitempty"`  // For SSH output messages
	TransferID string      `json:"transfer_id,omitempty"` // For transfer progress messages
	Data       interface{} `json:"data"`                  // Message payload
	Timestamp  int64       `json:"timestamp"`             // Unix timestamp
}

// ClientMessage represents a client-to-server WebSocket message
type ClientMessage struct {
	Action string `json:"action"` // "subscribe" or "unsubscribe"
	Target string `json:"target"` // sessionID or transferID
}

// NewSSHOutputMessage creates a new SSH output message
func NewSSHOutputMessage(sessionID string, data string) *Message {
	return &Message{
		Type:      "ssh:output",
		SessionID: sessionID,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// NewTransferProgressMessage creates a new transfer progress message
func NewTransferProgressMessage(transferID string, data interface{}) *Message {
	return &Message{
		Type:       "transfer:progress",
		TransferID: transferID,
		Data:       data,
		Timestamp:  time.Now().Unix(),
	}
}
