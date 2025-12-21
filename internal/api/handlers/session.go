package handlers

import (
	"fmt"
	"net/http"

	"AHaSSHTools/internal/api/dto"
	"AHaSSHTools/internal/api/websocket"
	"AHaSSHTools/internal/service"
	"github.com/gin-gonic/gin"
)

// SessionHandler handles SSH session-related HTTP requests
type SessionHandler struct {
	service *service.SessionService
	wsHub   *websocket.Hub
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(s *service.SessionService, hub *websocket.Hub) *SessionHandler {
	return &SessionHandler{
		service: s,
		wsHub:   hub,
	}
}

// ConnectRequest represents the request body for connecting to SSH
type ConnectRequest struct {
	SessionID  string `json:"session_id" binding:"required"`
	Host       string `json:"host" binding:"required"`
	Port       int    `json:"port" binding:"required"`
	User       string `json:"user" binding:"required"`
	AuthType   string `json:"auth_type" binding:"required"`
	AuthValue  string `json:"auth_value" binding:"required"`
	Passphrase string `json:"passphrase"`
	Cols       int    `json:"cols" binding:"required"`
	Rows       int    `json:"rows" binding:"required"`
}

// Connect handles POST /api/v1/sessions/connect
func (h *SessionHandler) Connect(c *gin.Context) {
	var req ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse("Invalid request body"))
		return
	}

	// Connect SSH with WebSocket output callback
	err := h.service.ConnectSSH(
		req.SessionID,
		req.Host,
		req.Port,
		req.User,
		req.AuthType,
		req.AuthValue,
		req.Passphrase,
		req.Cols,
		req.Rows,
		func(data []byte) {
			// Broadcast SSH output via WebSocket
			h.wsHub.BroadcastToSession(req.SessionID, "ssh:output", string(data))
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(map[string]string{
		"session_id": req.SessionID,
		"message":    fmt.Sprintf("SSH session started: %s@%s:%d", req.User, req.Host, req.Port),
	}))
}

// SendDataRequest represents the request body for sending data to SSH session
type SendDataRequest struct {
	Data string `json:"data" binding:"required"`
}

// SendData handles POST /api/v1/sessions/:id/send
func (h *SessionHandler) SendData(c *gin.Context) {
	sessionID := c.Param("id")

	var req SendDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse("Invalid request body"))
		return
	}

	if err := h.service.SendData(sessionID, req.Data); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessMessageResponse("Data sent successfully"))
}

// ResizeRequest represents the request body for resizing terminal
type ResizeRequest struct {
	Cols int `json:"cols" binding:"required"`
	Rows int `json:"rows" binding:"required"`
}

// Resize handles POST /api/v1/sessions/:id/resize
func (h *SessionHandler) Resize(c *gin.Context) {
	sessionID := c.Param("id")

	var req ResizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse("Invalid request body"))
		return
	}

	if err := h.service.ResizeTerminal(sessionID, req.Cols, req.Rows); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessMessageResponse("Terminal resized successfully"))
}

// Disconnect handles DELETE /api/v1/sessions/:id
func (h *SessionHandler) Disconnect(c *gin.Context) {
	sessionID := c.Param("id")

	if err := h.service.CloseSession(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessMessageResponse("Session closed successfully"))
}

// ListSessions handles GET /api/v1/sessions
func (h *SessionHandler) ListSessions(c *gin.Context) {
	sessions := h.service.ListSessions()
	c.JSON(http.StatusOK, dto.NewSuccessResponse(sessions))
}
