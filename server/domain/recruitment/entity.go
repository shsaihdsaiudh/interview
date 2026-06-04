// Package recruitment 定义招募卡片领域的聚合根、值对象和 DTO。
// 本包零外部依赖，只使用 Go 标准库。
package recruitment

import "time"

// ── 角色常量 ──

const (
	RoleInterviewee = "interviewee"
	RoleInterviewer = "interviewer"
	RoleBoth        = "both"
)

// ── RecruitmentCard 聚合根 ──

// RecruitmentCard 是用户的招募名片/卡片，一名用户只能拥有一张卡片。
type RecruitmentCard struct {
	ID              string
	UserID          string
	Nickname        string   // 从 users 表 JOIN 获取，非 persistence 字段
	Avatar          string   // 从 users 表 JOIN 获取，非 persistence 字段
	Skills          []string
	TargetCompanies []string
	Role            string // interviewee / interviewer / both
	ExperienceYears int
	Bio             string
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ── 聚合根行为方法 ──

// Activate 开放被预约。
func (c *RecruitmentCard) Activate() {
	c.IsActive = true
}

// Deactivate 关闭被预约。
func (c *RecruitmentCard) Deactivate() {
	c.IsActive = false
}

// CanBeManagedBy 检查用户是否有权管理此卡片。
func (c *RecruitmentCard) CanBeManagedBy(userID string) bool {
	return c.UserID == userID
}

// ── DTO ──

// CardResponse 对外展示的卡片信息。
type CardResponse struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Nickname        string    `json:"nickname"`
	Avatar          string    `json:"avatar"`
	Skills          []string  `json:"skills"`
	TargetCompanies []string  `json:"target_companies"`
	Role            string    `json:"role"`
	ExperienceYears int       `json:"experience_years"`
	Bio             string    `json:"bio"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ToResponse 将聚合根转为对外展示对象。
func (c *RecruitmentCard) ToResponse() CardResponse {
	return CardResponse{
		ID:              c.ID,
		UserID:          c.UserID,
		Nickname:        c.Nickname,
		Avatar:          c.Avatar,
		Skills:          safeSlice(c.Skills),
		TargetCompanies: safeSlice(c.TargetCompanies),
		Role:            c.Role,
		ExperienceYears: c.ExperienceYears,
		Bio:             c.Bio,
		IsActive:        c.IsActive,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}

// safeSlice 确保 nil slice 返回空数组而不是 null。
func safeSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// ── 请求/响应 DTO ──

// CreateOrUpdateCardRequest 创建/更新邀请卡片的请求体。
type CreateOrUpdateCardRequest struct {
	Skills          []string `json:"skills"`
	TargetCompanies []string `json:"target_companies"`
	Role            string   `json:"role" binding:"required"`
	ExperienceYears int      `json:"experience_years"`
	Bio             string   `json:"bio"`
	IsActive        *bool    `json:"is_active"` // nil 表示使用默认值 true
}

// ListCardsFilter 列表查询筛选条件。
type ListCardsFilter struct {
	Skill   string // 按技能筛选（单个）
	Company string // 按目标公司筛选（单个）
	Role    string // 按角色筛选
	Page    int    // 分页页码，从 1 开始
	Size    int    // 每页条数
}

// CardListResponse 卡片列表分页响应。
type CardListResponse struct {
	Cards    []CardResponse `json:"cards"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}
