// Package user 定义用户领域的聚合根、值对象和领域错误。
// 本包零外部依赖，只使用 Go 标准库。
package user

import "time"

// User 是用户聚合根。
// 以 Email 作为唯一标识符（同时也是 PostgreSQL 主键）。
type User struct {
	Email         string
	PasswordHash  string
	Nickname      string
	StudentID     string
	Department    string
	Tags          []string
	Avatar        string
	ContactInfo   string
	EmailVerified bool
	VerifyToken   string
	CreatedAt     time.Time
}

// ── 聚合根行为方法 ──

// IsVerified 检查邮箱是否已验证。
func (u *User) IsVerified() bool {
	return u.EmailVerified
}

// ClearVerifyToken 清除验证 token（一次性使用）。
func (u *User) ClearVerifyToken() {
	u.VerifyToken = ""
}

// MarkVerified 标记邮箱为已验证并清除 token。
func (u *User) MarkVerified() {
	u.EmailVerified = true
	u.VerifyToken = ""
}

// UpdateProfile 更新个人资料。
// 只更新非空字段，避免意外覆盖。
func (u *User) UpdateProfile(nickname, studentID, department string, tags []string, avatar, contactInfo string) {
	if nickname != "" {
		u.Nickname = nickname
	}
	if studentID != "" {
		u.StudentID = studentID
	}
	u.Department = department
	if tags != nil {
		u.Tags = tags
	}
	u.Avatar = avatar
	u.ContactInfo = contactInfo
}

// ── DTO 转换方法 ──

// UserResponse 是对外展示的用户信息（不含密码等敏感字段）。
type UserResponse struct {
	Email         string    `json:"email"`
	Nickname      string    `json:"nickname"`
	StudentID     string    `json:"student_id"`
	Department    string    `json:"department"`
	Tags          []string  `json:"tags"`
	Avatar        string    `json:"avatar"`
	ContactInfo   string    `json:"contact_info,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

// ToResponse 将聚合根转为对外展示对象（不含联系方式）。
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

// ToResponseWithContact 包含联系方式（仅双方确认预约后可见）。
func (u *User) ToResponseWithContact() UserResponse {
	r := u.ToResponse()
	r.ContactInfo = u.ContactInfo
	return r
}

// ── 请求/响应 DTO ──

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

// AuthResponse 登录成功返回
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Nickname    string   `json:"nickname"`
	StudentID   string   `json:"student_id"`
	Department  string   `json:"department"`
	Tags        []string `json:"tags"`
	Avatar      string   `json:"avatar"`
	ContactInfo string   `json:"contact_info"`
}
