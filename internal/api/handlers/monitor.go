package handlers

import (
	"net/http"

	"AHaSSHTools/internal/api/dto"
	"AHaSSHTools/internal/service"
	"github.com/gin-gonic/gin"
)

// MonitorHandler exposes per-session system metrics over HTTP，
// 供移动端监控看板轮询（桌面端走 Wails 绑定）。
type MonitorHandler struct {
	service *service.MonitorService
}

func NewMonitorHandler(s *service.MonitorService) *MonitorHandler {
	return &MonitorHandler{service: s}
}

// GetMetrics handles GET /api/v1/sessions/:id/monitor
func (h *MonitorHandler) GetMetrics(c *gin.Context) {
	if h.service == nil {
		c.JSON(http.StatusNotImplemented, dto.NewErrorMessageResponse("monitor service not available"))
		return
	}
	data, err := h.service.GetMonitoringData(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResponse(data))
}
