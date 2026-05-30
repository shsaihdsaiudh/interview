package model

import "time"

// User 用户数据模型
type User struct {
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"` // 不序列化到 JSON
	Nickname       string    `json:"nickname"`
	StudentID      string    `json:"student_id"`
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
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

// AuthResponse 登录成功返回的 JWT 响应
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// ToResponse 将内部 User 转为对外 UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		Email:         u.Email,
		Nickname:      u.Nickname,
		StudentID:     u.StudentID,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
	}
}
