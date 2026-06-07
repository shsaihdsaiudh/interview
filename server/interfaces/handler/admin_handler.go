package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"interview-server/application"
	"interview-server/domain/user"
)

// AdminHandler 管理员后台 HTTP 处理器。
type AdminHandler struct {
	adminSvc *application.AdminService
}

// NewAdminHandler 创建管理员处理器。
func NewAdminHandler(adminSvc *application.AdminService) *AdminHandler {
	return &AdminHandler{adminSvc: adminSvc}
}

// ── 仪表盘 ──

// GetStats 获取仪表盘统计。
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.adminSvc.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计数据失败"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// ── 用户管理 ──

// ListUsers 管理员查询用户列表。
func (h *AdminHandler) ListUsers(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	page, pageSize := parsePagination(c)

	resp, err := h.adminSvc.ListUsers(keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateUserRole 修改用户角色。
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少邮箱参数"})
		return
	}

	var req application.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	u, err := h.adminSvc.UpdateUserRole(email, req.Role)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		if errors.Is(err, user.ErrInvalidRole) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "修改角色失败"})
		return
	}
	c.JSON(http.StatusOK, u)
}

// BanUser 封禁用户。
func (h *AdminHandler) BanUser(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少邮箱参数"})
		return
	}

	u, err := h.adminSvc.BanUser(email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "封禁失败"})
		return
	}
	c.JSON(http.StatusOK, u)
}

// UnbanUser 解封用户。
func (h *AdminHandler) UnbanUser(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少邮箱参数"})
		return
	}

	u, err := h.adminSvc.UnbanUser(email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解封失败"})
		return
	}
	c.JSON(http.StatusOK, u)
}

// ── 名片管理 ──

// ListCards 管理员查询名片列表。
func (h *AdminHandler) ListCards(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	page, pageSize := parsePagination(c)

	resp, err := h.adminSvc.ListCards(keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取名片列表失败"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteCard 删除名片。
func (h *AdminHandler) DeleteCard(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少名片 ID"})
		return
	}

	if err := h.adminSvc.DeleteCard(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "名片不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "名片已删除"})
}

// ── 预约管理 ──

// ListAppointments 管理员查询预约列表。
func (h *AdminHandler) ListAppointments(c *gin.Context) {
	page, pageSize := parsePagination(c)

	resp, err := h.adminSvc.ListAppointments(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取预约列表失败"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteAppointment 删除预约。
func (h *AdminHandler) DeleteAppointment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少预约 ID"})
		return
	}

	if err := h.adminSvc.DeleteAppointment(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "预约不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "预约已取消"})
}

// ── 辅助 ──

// parsePagination 解析分页参数，默认 page=1, pageSize=20。
func parsePagination(c *gin.Context) (int, int) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}
