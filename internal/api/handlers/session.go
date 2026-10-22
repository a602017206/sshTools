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
	service     *service.SessionService
	wsHub       *websocket.Hub
	connections *service.ConnectionService
}

// NewSessionHandler creates a new session handler.
// connections 可为 nil；提供时用于按 connection_id 回退已保存的凭据，
// 让移动端这类不持有明文密码的客户端也能直接开会话。
func NewSessionHandler(s *service.SessionService, hub *websocket.Hub, connections *service.ConnectionService) *SessionHandler {
	return &SessionHandler{
		service:     s,
		wsHub:       hub,
		connections: connections,
	}
}

// ConnectRequest represents the request body for connecting to SSH
type ConnectRequest struct {
	SessionID  string `json:"session_id" binding:"required"`
	Host       string `json:"host" binding:"required"`
	Port       int    `json:"port" binding:"required"`
	User       string `json:"user" binding:"required"`
	AuthType   string `json:"auth_type" binding:"required"`
	AuthValue  string `json:"auth_value"`
	Passphrase string `json:"passphrase"`
	// ConnectionID 可选：auth_value 为空时按它查已保存的密码/密钥路径，
	// 口径与连接测试接口（handlers/connection.go）一致。
	ConnectionID string `json:"connection_id"`
	Cols         int    `json:"cols" binding:"required"`
	Rows         int    `json:"rows" binding:"required"`
}

// Connect handles POST /api/v1/sessions/connect
func (h *SessionHandler) Connect(c *gin.Context) {
	var req ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse(invalidRequestBodyMessage))
		return
	}

	authValue := req.AuthValue
	if authValue == "" && req.ConnectionID != "" && h.connections != nil {
		if req.AuthType == "key" {
			if conn, err := h.connections.GetConnection(req.ConnectionID); err == nil {
				authValue = conn.KeyPath
			}
		} else if password, err := h.connections.GetPassword(req.ConnectionID); err == nil {
			authValue = password
		}
	}
	if authValue == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse("missing credential: provide auth_value or save a password first"))
		return
	}

	// Connect SSH with WebSocket output callback
	err := h.service.ConnectSSH(
		req.SessionID,
		req.Host,
		req.Port,
		req.User,
		req.AuthType,
		authValue,
		req.Passphrase,
		req.Cols,
		req.Rows,
		func(data []byte) {
			// Broadcast SSH output via WebSocket
			h.wsHub.BroadcastToSession(req.SessionID, "ssh:output", string(data))
		},
		func(err error) {
			payload := map[string]string{"reason": "closed"}
			if err != nil {
				payload["error"] = err.Error()
			}
			h.wsHub.BroadcastToSession(req.SessionID, "ssh:closed", payload)
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}

	// 临时输入的新密码开会话成功后顺手保存，口径与测试连接接口一致
	if req.AuthValue != "" && req.ConnectionID != "" && req.AuthType != "key" && h.connections != nil {
		_ = h.connections.SavePassword(req.ConnectionID, req.AuthValue)
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
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse(invalidRequestBodyMessage))
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
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse(invalidRequestBodyMessage))
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
