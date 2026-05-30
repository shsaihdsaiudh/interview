package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"interview-server/internal/model"
	"interview-server/internal/service"
)

// AvailabilityHandler 空闲时间 HTTP 处理器
type AvailabilityHandler struct {
	availSvc *service.AvailabilityService
}

// NewAvailabilityHandler 创建空闲时间处理器
func NewAvailabilityHandler(availSvc *service.AvailabilityService) *AvailabilityHandler {
	return &AvailabilityHandler{availSvc: availSvc}
}

// GetMyAvailability 获取自己的空闲时间列表
func (h *AvailabilityHandler) GetMyAvailability(c *gin.Context) {
	email := c.GetString("user_email")
	slots := h.availSvc.GetMyAvailability(email)
	c.JSON(http.StatusOK, gin.H{"availabilities": slots})
}

// AddAvailability 添加空闲时间
func (h *AvailabilityHandler) AddAvailability(c *gin.Context) {
	email := c.GetString("user_email")

	var req model.AddAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	slot, err := h.availSvc.AddAvailability(email, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, slot)
}

// DeleteAvailability 删除空闲时间
func (h *AvailabilityHandler) DeleteAvailability(c *gin.Context) {
	email := c.GetString("user_email")
	slotID := c.Param("id")

	if err := h.availSvc.DeleteAvailability(email, slotID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
