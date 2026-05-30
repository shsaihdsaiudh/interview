package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"interview-server/application"
	"interview-server/domain/appointment"
)

// AppointmentHandler 预约和空闲时间相关的 HTTP 处理器。
type AppointmentHandler struct {
	apptSvc *application.AppointmentService
}

// NewAppointmentHandler 创建预约处理器。
func NewAppointmentHandler(apptSvc *application.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{apptSvc: apptSvc}
}

// =============================================================================
// 空闲时间管理
// =============================================================================

// GetMyAvailability 获取自己的空闲时间列表。
func (h *AppointmentHandler) GetMyAvailability(c *gin.Context) {
	email := c.GetString("user_email")
	slots := h.apptSvc.GetMyAvailability(email)
	c.JSON(http.StatusOK, gin.H{"availabilities": slots})
}

// AddAvailability 添加空闲时间。
func (h *AppointmentHandler) AddAvailability(c *gin.Context) {
	email := c.GetString("user_email")

	var req appointment.AddAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	slot, err := h.apptSvc.AddAvailability(email, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, slot)
}

// DeleteAvailability 删除空闲时间。
func (h *AppointmentHandler) DeleteAvailability(c *gin.Context) {
	email := c.GetString("user_email")
	slotID := c.Param("id")

	if err := h.apptSvc.DeleteAvailability(email, slotID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// =============================================================================
// 预约管理
// =============================================================================

// CreateAppointment 发起预约。
func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	email := c.GetString("user_email")

	var req appointment.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	appt, err := h.apptSvc.CreateAppointment(email, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, _ := h.apptSvc.ResolveUsers(appt)
	c.JSON(http.StatusCreated, resp)
}

// GetMyAppointments 获取我的预约列表。
func (h *AppointmentHandler) GetMyAppointments(c *gin.Context) {
	email := c.GetString("user_email")
	role := c.DefaultQuery("role", "student")

	apps, err := h.apptSvc.GetMyAppointments(email, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 填充用户信息
	resp := make([]*appointment.AppointmentResponse, 0, len(apps))
	for _, a := range apps {
		r, err := h.apptSvc.ResolveUsers(a)
		if err != nil {
			continue // 跳过数据不完整的预约
		}
		resp = append(resp, r)
	}

	c.JSON(http.StatusOK, gin.H{"appointments": resp})
}

// AcceptAppointment 接受预约。
func (h *AppointmentHandler) AcceptAppointment(c *gin.Context) {
	email := c.GetString("user_email")
	apptID := c.Param("id")

	appt, err := h.apptSvc.AcceptAppointment(email, apptID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, _ := h.apptSvc.ResolveUsers(appt)
	c.JSON(http.StatusOK, resp)
}

// RejectAppointment 拒绝预约。
func (h *AppointmentHandler) RejectAppointment(c *gin.Context) {
	email := c.GetString("user_email")
	apptID := c.Param("id")

	var req appointment.RejectAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Reason = "" // 允许不传 reason
	}

	appt, err := h.apptSvc.RejectAppointment(email, apptID, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, _ := h.apptSvc.ResolveUsers(appt)
	c.JSON(http.StatusOK, resp)
}
