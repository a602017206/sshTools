package handlers

import (
	"net/http"

	"AHaSSHTools/internal/api/dto"
	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/service"
	"github.com/gin-gonic/gin"
)

// ConnectionHandler handles connection-related HTTP requests
type ConnectionHandler struct {
	service *service.ConnectionService
}

// NewConnectionHandler creates a new connection handler
func NewConnectionHandler(s *service.ConnectionService) *ConnectionHandler {
	return &ConnectionHandler{service: s}
}

// GetConnections handles GET /api/v1/connections
func (h *ConnectionHandler) GetConnections(c *gin.Context) {
	connections, err := h.service.GetConnections()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(connections))
}

// AddConnection handles POST /api/v1/connections
func (h *ConnectionHandler) AddConnection(c *gin.Context) {
	var conn config.ConnectionConfig
	if err := c.ShouldBindJSON(&conn); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse("Invalid request body"))
		return
	}

	if err := h.service.AddConnection(conn); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse(conn))
}

// UpdateConnection handles PUT /api/v1/connections/:id
func (h *ConnectionHandler) UpdateConnection(c *gin.Context) {
	id := c.Param("id")

	var conn config.ConnectionConfig
	if err := c.ShouldBindJSON(&conn); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse("Invalid request body"))
		return
	}

	// Ensure the ID matches
	conn.ID = id

	if err := h.service.UpdateConnection(conn); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(conn))
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
	AuthValue  string `json:"auth_value" binding:"required"`
	Passphrase string `json:"passphrase"`
}

// TestConnection handles POST /api/v1/connections/test
func (h *ConnectionHandler) TestConnection(c *gin.Context) {
	var req TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse("Invalid request body"))
		return
	}

	err := h.service.TestConnection(req.Host, req.Port, req.User, req.AuthType, req.AuthValue, req.Passphrase)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessMessageResponse("Connection test successful"))
}
