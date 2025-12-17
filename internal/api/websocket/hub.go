package websocket

import (
	"sync"
)

// Hub manages WebSocket client connections and message broadcasting
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Broadcast messages to clients
	broadcast chan *Message

	// Register client connections
	register chan *Client

	// Unregister client connections
	unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				// Filter messages based on client subscriptions
				if client.shouldReceive(message) {
					select {
					case client.send <- message:
					default:
						// Client is slow or disconnected, remove it
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToSession broadcasts a message to all clients subscribed to a specific session
func (h *Hub) BroadcastToSession(sessionID string, eventType string, data interface{}) {
	message := &Message{
		Type:      eventType,
		SessionID: sessionID,
		Data:      data,
		Timestamp: getCurrentTimestamp(),
	}
	h.broadcast <- message
}

// BroadcastToTransfer broadcasts a message to all clients subscribed to a specific transfer
func (h *Hub) BroadcastToTransfer(transferID string, data interface{}) {
	message := &Message{
		Type:       "transfer:progress",
		TransferID: transferID,
		Data:       data,
		Timestamp:  getCurrentTimestamp(),
	}
	h.broadcast <- message
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(message *Message) {
	h.broadcast <- message
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// getCurrentTimestamp returns current Unix timestamp
func getCurrentTimestamp() int64 {
	return int64(0) // Will be set in Message constructor
}
