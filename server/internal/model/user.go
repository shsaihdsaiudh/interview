package model

import "time"

// User 用户数据模型
type User struct {
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"` // 不序列化到 JSON
	Nickname       string    `json:"nickname"`
	StudentID      string    `json:"student_id"`
	Department     string    `json:"department"`    // 院系
	Tags           []string  `json:"tags"`          // 面试方向标签
	Avatar         string    `json:"avatar"`        // 头像 URL
	ContactInfo    string    `json:"contact_info"`  // 联系方式
	EmailVerified  bool      `json:"email_verified"`
	VerifyToken    string    `json:"-"` // 邮箱验证 token，不暴露
	CreatedAt      time.Time `json:"created_at"`
}

// ── 请求结构体 ──

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	Nickname  string `json:"nickname" binding:"required"`
	StudentID string `json:"student_id" binding:"required"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ── 响应结构体 ──

// UserResponse 对外展示的用户信息（不含密码等敏感字段）
type UserResponse struct {
	Email         string    `json:"email"`
	Nickname      string    `json:"nickname"`
	StudentID     string    `json:"student_id"`
	Department    string    `json:"department"`
	Tags          []string  `json:"tags"`
	Avatar        string    `json:"avatar"`
	ContactInfo   string    `json:"contact_info,omitempty"` // 仅双方确认预约后可见
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

// UpdateProfileRequest 更新个人资料请求
type UpdateProfileRequest struct {
	Nickname    string   `json:"nickname"`
	StudentID   string   `json:"student_id"`
	Department  string   `json:"department"`
	Tags        []string `json:"tags"`
	Avatar      string   `json:"avatar"`
	ContactInfo string   `json:"contact_info"`
}

// AuthResponse 登录成功返回的 JWT 响应
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// ToResponse 将内部 User 转为对外 UserResponse（不含敏感信息）
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		Email:         u.Email,
		Nickname:      u.Nickname,
		StudentID:     u.StudentID,
		Department:    u.Department,
		Tags:          u.Tags,
		Avatar:        u.Avatar,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
	}
}

// ToResponseWithContact 包含联系方式（仅双方确认预约后使用）
func (u *User) ToResponseWithContact() UserResponse {
	r := u.ToResponse()
	r.ContactInfo = u.ContactInfo
	return r
}
