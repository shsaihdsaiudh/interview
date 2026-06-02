package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"interview-server/application"
	"interview-server/domain/user"
)

// UserHandler 用户和认证相关的 HTTP 处理器。
type UserHandler struct {
	userSvc *application.UserService
}

// NewUserHandler 创建用户处理器。
func NewUserHandler(userSvc *application.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// =============================================================================
// 认证相关
// =============================================================================

// SendCode 发送邮箱验证码。
func (h *UserHandler) SendCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	if err := h.userSvc.SendCode(req.Email); err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, user.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发送验证码失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送（开发阶段请看服务端日志）"})
}

// Register 用户注册。验证码通过后直接返回 JWT。
func (h *UserHandler) Register(c *gin.Context) {
	var req user.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	resp, err := h.userSvc.Register(req)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, user.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, user.ErrInvalidCode):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败，请稍后重试"})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// VerifyEmail 邮箱验证。
func (h *UserHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少验证 token"})
		return
	}

	if err := h.userSvc.VerifyEmail(token); err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidToken):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "验证失败，请稍后重试"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "邮箱验证成功，现在可以登录了"})
}

// Login 用户登录。
func (h *UserHandler) Login(c *gin.Context) {
	var req user.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	resp, err := h.userSvc.Login(req)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserNotFound):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码错误"})
		case errors.Is(err, user.ErrWrongPassword):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码错误"})
		case errors.Is(err, user.ErrEmailNotVerified):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败，请稍后重试"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ForgotPassword 忘记密码 — 发送重置验证码。
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req user.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	if err := h.userSvc.ForgotPassword(req.Email); err != nil {
		switch {
		case errors.Is(err, user.ErrUserNotFound):
			// 为安全起见不暴露用户是否存在，统一返回成功
			c.JSON(http.StatusOK, gin.H{"message": "如果该邮箱已注册，验证码已发送"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发送验证码失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "重置验证码已发送（开发阶段请看服务端日志）"})
}

// ResetPassword 重置密码。
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req user.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	if err := h.userSvc.ResetPassword(req.Email, req.Code, req.Password); err != nil {
		switch {
		case errors.Is(err, user.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		case errors.Is(err, user.ErrInvalidCode):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "重置密码失败，请稍后重试"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码重置成功，请登录"})
}

// Me 获取当前登录用户信息。
func (h *UserHandler) Me(c *gin.Context) {
	email, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	resp, err := h.userSvc.GetMe(email.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ChangePassword 修改密码（已登录）。
func (h *UserHandler) ChangePassword(c *gin.Context) {
	email := c.GetString("user_email")

	var req user.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	if err := h.userSvc.ChangePassword(email, req); err != nil {
		switch {
		case errors.Is(err, user.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		case errors.Is(err, user.ErrWrongOldPassword):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "修改密码失败，请稍后重试"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// =============================================================================
// 账号管理
// =============================================================================

// DeleteAccount 注销账号。需提供密码确认身份。
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	email := c.GetString("user_email")

	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	if err := h.userSvc.DeleteAccount(email, req.Password); err != nil {
		switch {
		case errors.Is(err, user.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		case errors.Is(err, user.ErrWrongPassword):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		case errors.Is(err, user.ErrCannotDeleteWithActiveAppointments):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "注销失败，请稍后重试"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "账号已注销"})
}

// =============================================================================
// 用户资料
// =============================================================================

// GetProfile 获取自己的完整资料（含联系方式）。
func (h *UserHandler) GetProfile(c *gin.Context) {
	email := c.GetString("user_email")

	detail, err := h.userSvc.GetProfile(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// UpdateProfile 更新个人资料。
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	email := c.GetString("user_email")

	var req user.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	resp, err := h.userSvc.UpdateProfile(email, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListUsers 用户列表（公开）。
func (h *UserHandler) ListUsers(c *gin.Context) {
	users := h.userSvc.GetAllUsers()
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// GetUser 用户详情（公开，可能含联系方式）。
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
