// Package application 编排用户相关用例。
// 依赖领域层接口，不依赖具体技术实现。
package application

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"interview-server/domain/appointment"
	"interview-server/domain/user"
	"interview-server/infrastructure/auth"
)

// UserService 用户用例编排。
// 依赖 UserRepository 接口而非具体实现 — 依赖反转。
type UserService struct {
	userRepo user.UserRepository
	apptRepo appointment.AppointmentRepository
}

// NewUserService 创建用户服务。
func NewUserService(userRepo user.UserRepository, apptRepo appointment.AppointmentRepository) *UserService {
	return &UserService{userRepo: userRepo, apptRepo: apptRepo}
}

// Register 注册新用户。
func (s *UserService) Register(req user.RegisterRequest) (*user.UserResponse, error) {
	// 校验 .edu 邮箱
	if !strings.HasSuffix(req.Email, ".edu") {
		return nil, user.ErrInvalidEmail
	}

	// 检查邮箱是否已注册
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, user.ErrEmailAlreadyExists
	}

	// bcrypt 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 生成邮箱验证 token
	verifyToken, err := generateRandomHex(32)
	if err != nil {
		return nil, fmt.Errorf("生成验证 token 失败: %w", err)
	}

	u := &user.User{
		Email:         req.Email,
		PasswordHash:  string(hash),
		Nickname:      req.Nickname,
		StudentID:     req.StudentID,
		EmailVerified: false,
		VerifyToken:   verifyToken,
		CreatedAt:     time.Now(),
	}

	if err := s.userRepo.Create(u); err != nil {
		return nil, err
	}

	// 开发阶段：打印验证链接到日志
	verifyURL := fmt.Sprintf("http://localhost:8080/api/v1/auth/verify-email?token=%s", verifyToken)
	log.Printf("📧 [DEV] 邮箱验证链接: %s", verifyURL)

	resp := u.ToResponse()
	return &resp, nil
}

// VerifyEmail 验证邮箱。
func (s *UserService) VerifyEmail(token string) error {
	u, err := s.userRepo.FindByVerifyToken(token)
	if err != nil {
		return user.ErrInvalidToken
	}

	u.MarkVerified() // 使用聚合根方法
	return s.userRepo.Update(u)
}

// Login 登录。
func (s *UserService) Login(req user.LoginRequest) (*user.AuthResponse, error) {
	u, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, user.ErrUserNotFound
	}

	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, user.ErrWrongPassword
	}

	// 检查邮箱是否已验证
	if !u.IsVerified() {
		return nil, user.ErrEmailNotVerified
	}

	// 签发 JWT
	jwtToken, err := auth.GenerateJWT(u.Email)
	if err != nil {
		return nil, fmt.Errorf("JWT 生成失败: %w", err)
	}

	return &user.AuthResponse{
		Token: jwtToken,
		User:  u.ToResponse(),
	}, nil
}

// GetMe 获取当前用户信息。
func (s *UserService) GetMe(email string) (*user.UserResponse, error) {
	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	resp := u.ToResponse()
	return &resp, nil
}

// UpdateProfile 更新个人资料。
func (s *UserService) UpdateProfile(email string, req user.UpdateProfileRequest) (*user.UserResponse, error) {
	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, user.ErrUserNotFound
	}

	// 使用聚合根方法更新
	u.UpdateProfile(req.Nickname, req.StudentID, req.Department, req.Tags, req.Avatar, req.ContactInfo)

	if err := s.userRepo.Update(u); err != nil {
		return nil, err
	}

	resp := u.ToResponse()
	return &resp, nil
}

// GetAllUsers 获取所有已验证用户列表。
func (s *UserService) GetAllUsers() []user.UserResponse {
	users := s.userRepo.FindAll()
	result := make([]user.UserResponse, 0, len(users))
	for _, u := range users {
		result = append(result, u.ToResponse())
	}
	return result
}

// UserDetailResponse 用户详情响应（含空闲时间）。
type UserDetailResponse struct {
	User           user.UserResponse            `json:"user"`
	Availabilities []*appointment.Availability   `json:"availabilities"`
}

// GetUserDetail 获取用户详情（含空闲时间，可能含联系方式）。
func (s *UserService) GetUserDetail(email string, requesterEmail string) (*UserDetailResponse, error) {
	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, user.ErrUserNotFound
	}

	resp := u.ToResponse()
	availabilities := s.apptRepo.FindAvailabilitiesByUserID(email)
	if availabilities == nil {
		availabilities = []*appointment.Availability{}
	}

	// 检查 requester 是否与该用户有 accepted 的预约 — 有则展示联系方式
	if requesterEmail != "" && s.hasAcceptedAppointment(requesterEmail, email) {
		resp.ContactInfo = u.ContactInfo
	}

	return &UserDetailResponse{
		User:           resp,
		Availabilities: availabilities,
	}, nil
}

// hasAcceptedAppointment 检查两个用户之间是否有已接受的预约。
func (s *UserService) hasAcceptedAppointment(user1, user2 string) bool {
	apps := s.apptRepo.FindAppointmentsByStudentID(user1)
	for _, a := range apps {
		if a.MentorID == user2 && a.IsAccepted() {
			return true
		}
	}
	apps = s.apptRepo.FindAppointmentsByMentorID(user1)
	for _, a := range apps {
		if a.StudentID == user2 && a.IsAccepted() {
			return true
		}
	}
	return false
}

// =============================================================================
// 辅助函数
// =============================================================================

// generateRandomHex 生成指定字节数的随机 hex 字符串。
func generateRandomHex(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
