package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"interview-server/application"
	"interview-server/domain/recruitment"
)

// RecruitmentHandler 招募卡片相关的 HTTP 处理器。
type RecruitmentHandler struct {
	recruitSvc *application.RecruitmentService
}

// NewRecruitmentHandler 创建招募卡片处理器。
func NewRecruitmentHandler(recruitSvc *application.RecruitmentService) *RecruitmentHandler {
	return &RecruitmentHandler{recruitSvc: recruitSvc}
}

// =============================================================================
// 卡片管理
// =============================================================================

// CreateOrUpdateCard 创建或更新自己的招募卡片（需登录）。
func (h *RecruitmentHandler) CreateOrUpdateCard(c *gin.Context) {
	email := c.GetString("user_email")

	var req recruitment.CreateOrUpdateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不合法: " + err.Error()})
		return
	}

	resp, err := h.recruitSvc.CreateOrUpdateCard(email, req)
	if err != nil {
		switch {
		case errors.Is(err, recruitment.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, recruitment.ErrInvalidRole):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetCardByUserID 获取指定用户的招募卡片。
func (h *RecruitmentHandler) GetCardByUserID(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 user_id 参数"})
		return
	}

	resp, err := h.recruitSvc.GetCardByUserID(userID)
	if err != nil {
		if errors.Is(err, recruitment.ErrCardNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListCards 列表搜索招募卡片。
// Query 参数: skill, company, role, page, size
func (h *RecruitmentHandler) ListCards(c *gin.Context) {
	page := 1
	size := 20

	if v := c.Query("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}
	if v := c.Query("size"); v != "" {
		if s, err := strconv.Atoi(v); err == nil && s > 0 {
			if s > 100 {
				s = 100
			}
			size = s
		}
	}

	filter := recruitment.ListCardsFilter{
		Skill:   c.Query("skill"),
		Company: c.Query("company"),
		Role:    c.Query("role"),
		Page:    page,
		Size:    size,
	}

	resp, err := h.recruitSvc.ListCards(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
