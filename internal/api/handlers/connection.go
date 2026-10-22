package handlers

import (
	"net/http"

	"AHaSSHTools/internal/api/dto"
	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const invalidRequestBodyMessage = "Invalid request body"

// ConnectionHandler handles connection-related HTTP requests
type ConnectionHandler struct {
	service *service.ConnectionService
}

// NewConnectionHandler creates a new connection handler
func NewConnectionHandler(s *service.ConnectionService) *ConnectionHandler {
	return &ConnectionHandler{service: s}
}

// connectionRequest 是新增/更新连接的请求体：在连接配置之外可选带密码。
// 密码不写进 config.json，而是经 CredentialStore 加密单独存储，
// 与桌面端 SavePassword 的行为保持一致。
type connectionRequest struct {
	config.ConnectionConfig
	Password string `json:"password,omitempty"`
}

// connectionResponse 是连接的响应体，额外暴露是否已保存密码，
// 供客户端决定要不要再向用户索要密码（密码本身永不回传）。
type connectionResponse struct {
	config.ConnectionConfig
	HasPassword bool `json:"has_password"`
}

func (h *ConnectionHandler) toResponse(conn config.ConnectionConfig) connectionResponse {
	return connectionResponse{
		ConnectionConfig: conn,
		HasPassword:      h.service.HasPassword(conn.ID),
	}
}

// GetConnections handles GET /api/v1/connections
func (h *ConnectionHandler) GetConnections(c *gin.Context) {
	connections, err := h.service.GetConnections()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}

	responses := make([]connectionResponse, 0, len(connections))
	for _, conn := range connections {
		responses = append(responses, h.toResponse(conn))
	}
	c.JSON(http.StatusOK, dto.NewSuccessResponse(responses))
}

// AddConnection handles POST /api/v1/connections
func (h *ConnectionHandler) AddConnection(c *gin.Context) {
	var req connectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse(invalidRequestBodyMessage))
		return
	}

	conn := req.ConnectionConfig
	// 客户端未提供 ID 时由服务端生成；早期移动端传空 ID 导致
	// 后续按 ID 删除/更新全部 404，这里从源头堵住。
	if conn.ID == "" {
		conn.ID = uuid.NewString()
	}

	if err := h.service.AddConnection(conn); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}

	if req.Password != "" {
		if err := h.service.SavePassword(conn.ID, req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
			return
		}
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse(h.toResponse(conn)))
}

// UpdateConnection handles PUT /api/v1/connections/:id
func (h *ConnectionHandler) UpdateConnection(c *gin.Context) {
	id := c.Param("id")

	var req connectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse(invalidRequestBodyMessage))
		return
	}

	// Ensure the ID matches
	conn := req.ConnectionConfig
	conn.ID = id

	if err := h.service.UpdateConnection(conn); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}

	// 密码留空表示不修改已保存的密码（编辑表单的约定）
	if req.Password != "" {
		if err := h.service.SavePassword(conn.ID, req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
			return
		}
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(h.toResponse(conn)))
}

// DeleteConnection handles DELETE /api/v1/connections/:id
func (h *ConnectionHandler) DeleteConnection(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.RemoveConnection(id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessMessageResponse("Connection deleted successfully"))
}

// TestConnectionRequest represents the request body for testing a connection
type TestConnectionRequest struct {
	Host       string `json:"host" binding:"required"`
	Port       int    `json:"port" binding:"required"`
	User       string `json:"user" binding:"required"`
	AuthType   string `json:"auth_type" binding:"required"`
	AuthValue  string `json:"auth_value"`
	Passphrase string `json:"passphrase"`
	// ConnectionID 可选：auth_value 为空时按它查已保存的密码/密钥路径，
	// 这样客户端保存过密码后测试连接不必让用户重复输入。
	ConnectionID string `json:"connection_id"`
}

// TestConnection handles POST /api/v1/connections/test
func (h *ConnectionHandler) TestConnection(c *gin.Context) {
	var req TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse(invalidRequestBodyMessage))
		return
	}

	authValue := req.AuthValue
	if authValue == "" && req.ConnectionID != "" {
		if req.AuthType == "key" {
			if conn, err := h.service.GetConnection(req.ConnectionID); err == nil {
				authValue = conn.KeyPath
			}
		} else if password, err := h.service.GetPassword(req.ConnectionID); err == nil {
			authValue = password
		}
	}
	if authValue == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse("missing credential: provide auth_value or save a password first"))
		return
	}

	err := h.service.TestConnection(req.Host, req.Port, req.User, req.AuthType, authValue, req.Passphrase)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err))
		return
	}

	// 用户为测试临时输入了新密码且测试成功时顺手保存，
	// 与桌面端「输入密码测试成功后回存」的行为一致。
	if req.AuthValue != "" && req.ConnectionID != "" && req.AuthType != "key" {
		_ = h.service.SavePassword(req.ConnectionID, req.AuthValue)
	}

	c.JSON(http.StatusOK, dto.NewSuccessMessageResponse("Connection test successful"))
}
