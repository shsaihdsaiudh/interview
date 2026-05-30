package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"interview-server/internal/model"
	"interview-server/internal/service"
)

// AppointmentHandler 预约 HTTP 处理器
type AppointmentHandler struct {
	apptSvc *service.AppointmentService
}

// NewAppointmentHandler 创建预约处理器
func NewAppointmentHandler(apptSvc *service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{apptSvc: apptSvc}
}

// CreateAppointment 发起预约
func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	email := c.GetString("user_email")

	var req model.CreateAppointmentRequest
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

// GetMyAppointments 获取我的预约列表
func (h *AppointmentHandler) GetMyAppointments(c *gin.Context) {
	email := c.GetString("user_email")
	role := c.DefaultQuery("role", "student") // 默认查发出的

	apps, err := h.apptSvc.GetMyAppointments(email, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 填充用户信息
	resp := make([]*model.AppointmentResponse, 0, len(apps))
	for _, a := range apps {
		r, err := h.apptSvc.ResolveUsers(a)
		if err != nil {
			continue // 跳过数据不完整的预约
		}
		resp = append(resp, r)
	}

	c.JSON(http.StatusOK, gin.H{"appointments": resp})
}

// AcceptAppointment 接受预约
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

// RejectAppointment 拒绝预约
func (h *AppointmentHandler) RejectAppointment(c *gin.Context) {
	email := c.GetString("user_email")
	apptID := c.Param("id")

	var req model.RejectAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许不传 reason
		req.Reason = ""
	}

	appt, err := h.apptSvc.RejectAppointment(email, apptID, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, _ := h.apptSvc.ResolveUsers(appt)
	c.JSON(http.StatusOK, resp)
}
