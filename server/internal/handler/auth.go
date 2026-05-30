package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"interview-server/internal/model"
	"interview-server/internal/repository"
	"interview-server/internal/service"
)

// AuthHandler 认证相关 HTTP 处理器
type AuthHandler struct {
	svc *service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	user, err := h.svc.Register(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, repository.ErrUserAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败，请稍后重试"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "注册成功，请查看邮箱验证链接（开发阶段验证链接已打印到服务端日志）",
		"user":    user,
	})
}

// VerifyEmail 邮箱验证
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少验证 token"})
		return
	}

	if err := h.svc.VerifyEmail(token); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidToken):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "验证失败，请稍后重试"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "邮箱验证成功，现在可以登录了"})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	resp, err := h.svc.Login(req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码错误"})
		case errors.Is(err, service.ErrWrongPassword):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码错误"})
		case errors.Is(err, service.ErrEmailNotVerified):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败，请稍后重试"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Me 获取当前登录用户信息
func (h *AuthHandler) Me(c *gin.Context) {
	// email 由 JWT 中间件注入
	email, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	user, err := h.svc.GetMe(email.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, user)
}
