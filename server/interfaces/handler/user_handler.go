package handler

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

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

// UploadAvatar 上传头像。
// 接收 multipart/form-data，限制 JPEG/PNG/WebP ≤2MB。
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	email := c.GetString("user_email")

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择头像文件"})
		return
	}
	defer file.Close()

	// ── 校验文件类型 ──
	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取文件"})
		return
	}
	contentType := http.DetectContentType(buf)
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	if !allowed[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 JPEG、PNG、WebP 格式的头像"})
		return
	}

	// ── 校验文件大小 ──
	const maxSize = 2 << 20 // 2MB
	if header.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "头像文件大小不能超过 2MB"})
		return
	}

	// ── 读取完整文件内容 ──
	if _, err := file.Seek(0, 0); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}
	fileData := make([]byte, header.Size)
	if _, err := file.Read(fileData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}

	// ── 重新验证实际内容（防止伪造 Content-Type）──
	actualType := http.DetectContentType(fileData)
	if !allowed[actualType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 JPEG、PNG、WebP 格式的头像"})
		return
	}

	// ── 生成文件名并保存 ──
	emailHash := sha256Hex(email)
	ext := extensionFromContentType(actualType)
	filename := fmt.Sprintf("%s_%d.%s", emailHash, time.Now().UnixMilli(), ext)

	uploadDir := "server/uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}

	destPath := filepath.Join(uploadDir, filename)
	if err := os.WriteFile(destPath, fileData, 0o644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存头像文件失败"})
		return
	}

	// ── 更新用户 avatar 字段 ──
	avatarURL := "/uploads/avatars/" + filename
	_, err = h.userSvc.UpdateProfile(email, user.UpdateProfileRequest{Avatar: avatarURL})
	if err != nil {
		// 文件已保存，但数据库更新失败 — 记录错误但仍返回 URL
		c.JSON(http.StatusOK, gin.H{
			"avatar_url": avatarURL,
			"warning":    "头像文件已上传，但更新资料失败，请重试保存",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"avatar_url": avatarURL})
}

// ListUsers 用户列表（公开，支持分页）。
// Query 参数: page（默认 1），page_size（默认 20，最大 100）
func (h *UserHandler) ListUsers(c *gin.Context) {
	page := 1
	pageSize := 20

	if v := c.Query("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}
	if v := c.Query("page_size"); v != "" {
		if ps, err := strconv.Atoi(v); err == nil && ps > 0 {
			if ps > 100 {
				ps = 100
			}
			pageSize = ps
		}
	}

	resp, err := h.userSvc.GetAllUsers(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}
	c.JSON(http.StatusOK, resp)
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

// =============================================================================
// 辅助函数
// =============================================================================

// sha256Hex 返回字符串的 SHA-256 哈希前 16 个十六进制字符。
func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)[:16]
}

// extensionFromContentType 根据 MIME 类型返回文件扩展名。
func extensionFromContentType(ct string) string {
	switch ct {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	default:
		return "bin"
	}
}
