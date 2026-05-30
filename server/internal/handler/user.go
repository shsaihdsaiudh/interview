package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"interview-server/internal/model"
	"interview-server/internal/service"
)

// UserHandler 用户相关 HTTP 处理器
type UserHandler struct {
	userSvc *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// GetProfile 获取自己的完整资料（含联系方式）
func (h *UserHandler) GetProfile(c *gin.Context) {
	email := c.GetString("user_email")

	detail, err := h.userSvc.GetUserDetail(email, email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// UpdateProfile 更新个人资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	email := c.GetString("user_email")

	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	user, err := h.userSvc.UpdateProfile(email, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListUsers 用户列表（公开）
func (h *UserHandler) ListUsers(c *gin.Context) {
	users := h.userSvc.GetAllUsers()
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// GetUser 用户详情（公开）
func (h *UserHandler) GetUser(c *gin.Context) {
	email := c.Param("id")

	// 获取当前登录用户（可能为空）
	requesterEmail, _ := c.Get("user_email")

	var requester string
	if v, ok := requesterEmail.(string); ok {
		requester = v
	}

	detail, err := h.userSvc.GetUserDetail(email, requester)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, detail)
}
