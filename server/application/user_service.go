// Package application 编排用户相关用例。
// 依赖领域层接口，不依赖具体技术实现。
package application

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"interview-server/domain/appointment"
	"interview-server/domain/user"
	"interview-server/infrastructure/auth"
)

// UserService 用户用例编排。
// 依赖 UserRepository 接口而非具体实现 — 依赖反转。
type UserService struct {
	userRepo          user.UserRepository
	apptRepo          appointment.AppointmentRepository
	verificationCodes map[string]verificationCode // email → code（开发阶段内存存储）
	mu                sync.Mutex
}

type verificationCode struct {
	Code      string
	ExpiresAt time.Time
}

const verificationCodeTTL = 10 * time.Minute

// NewUserService 创建用户服务。
func NewUserService(userRepo user.UserRepository, apptRepo appointment.AppointmentRepository) *UserService {
	return &UserService{
		userRepo:          userRepo,
		apptRepo:          apptRepo,
		verificationCodes: make(map[string]verificationCode),
	}
}

// SendCode 发送邮箱验证码（开发阶段打印到日志）。
func (s *UserService) SendCode(email string) error {
	email = normalizeEmail(email)
	if !strings.HasSuffix(email, "@std.uestc.edu.cn") {
		return user.ErrInvalidEmail
	}

	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return user.ErrEmailAlreadyExists
	} else if !errors.Is(err, user.ErrUserNotFound) {
		return err
	}

	code, err := generateVerificationCode()
	if err != nil {
		return fmt.Errorf("生成验证码失败: %w", err)
	}

	s.mu.Lock()
	s.verificationCodes[email] = verificationCode{
		Code:      code,
		ExpiresAt: time.Now().Add(verificationCodeTTL),
	}
	s.mu.Unlock()

	log.Printf("📧 [DEV] 验证码已发送到 %s: %s", email, code)
	return nil
}

// Register 注册新用户。校验验证码，通过后直接创建已验证用户并返回 JWT。
func (s *UserService) Register(req user.RegisterRequest) (*user.AuthResponse, error) {
	req.Email = normalizeEmail(req.Email)

	// 校验 @std.uestc.edu.cn 邮箱（仅限本校学生）
	if !strings.HasSuffix(req.Email, "@std.uestc.edu.cn") {
		return nil, user.ErrInvalidEmail
	}

	// 检查邮箱是否已注册
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, user.ErrEmailAlreadyExists
	} else if !errors.Is(err, user.ErrUserNotFound) {
		return nil, err
	}

	// 校验验证码
	s.mu.Lock()
	storedCode, ok := s.verificationCodes[req.Email]
	if !ok || storedCode.Code != req.Code || time.Now().After(storedCode.ExpiresAt) {
		if ok && time.Now().After(storedCode.ExpiresAt) {
			delete(s.verificationCodes, req.Email)
		}
		s.mu.Unlock()
		return nil, user.ErrInvalidCode
	}
	delete(s.verificationCodes, req.Email) // 一次性使用
	s.mu.Unlock()

	// bcrypt 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Email // 默认昵称为邮箱前缀
	}

	u := &user.User{
		Email:         req.Email,
		PasswordHash:  string(hash),
		Nickname:      nickname,
		StudentID:     req.StudentID,
		Tags:          []string{},
		EmailVerified: true, // 已验证
		CreatedAt:     time.Now(),
	}

	if err := s.userRepo.Create(u); err != nil {
		return nil, err
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
	req.Email = normalizeEmail(req.Email)

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
	email = normalizeEmail(email)

	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	resp := u.ToResponse()
	return &resp, nil
}

// GetProfile 获取自己的完整资料。
func (s *UserService) GetProfile(email string) (*UserDetailResponse, error) {
	email = normalizeEmail(email)

	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, user.ErrUserNotFound
	}

	availabilities := s.apptRepo.FindAvailabilitiesByUserID(email)
	if availabilities == nil {
		availabilities = []*appointment.Availability{}
	}

	return &UserDetailResponse{
		User:           u.ToResponseWithContact(),
		Availabilities: availabilities,
	}, nil
}

// UpdateProfile 更新个人资料。
func (s *UserService) UpdateProfile(email string, req user.UpdateProfileRequest) (*user.UserResponse, error) {
	email = normalizeEmail(email)

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
	User           user.UserResponse           `json:"user"`
	Availabilities []*appointment.Availability `json:"availabilities"`
}

// GetUserDetail 获取用户详情（含空闲时间，可能含联系方式）。
func (s *UserService) GetUserDetail(email string, requesterEmail string) (*UserDetailResponse, error) {
	email = normalizeEmail(email)
	requesterEmail = normalizeEmail(requesterEmail)

	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, user.ErrUserNotFound
	}

	resp := u.ToResponse()
	availabilities := s.apptRepo.FindAvailabilitiesByUserID(email)
	if availabilities == nil {
		availabilities = []*appointment.Availability{}
	}

	// requester 是本人，或双方有 accepted 预约时展示联系方式。
	if requesterEmail != "" && (requesterEmail == email || s.hasAcceptedAppointment(requesterEmail, email)) {
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

// VerificationCodeForTest 仅供测试使用：获取指定邮箱的验证码。
func (s *UserService) VerificationCodeForTest(email string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.verificationCodes[normalizeEmail(email)].Code
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// generateRandomHex 生成指定字节数的随机 hex 字符串。
func generateRandomHex(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// generateVerificationCode 生成 6 位数字验证码。
func generateVerificationCode() (string, error) {
	code := ""
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code, nil
}
