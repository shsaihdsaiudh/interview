package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"interview-server/internal/model"
	"interview-server/internal/repository"
)

var (
	ErrInvalidEmail    = errors.New("仅支持 .edu 邮箱注册")
	ErrWrongPassword   = errors.New("密码错误")
	ErrEmailNotVerified = errors.New("邮箱未验证，请先验证邮箱")
	ErrInvalidToken    = errors.New("无效的验证 token")

	jwtSecret = []byte("interview-platform-secret-key") // TODO: 生产环境用环境变量
)

// AuthService 认证业务逻辑
type AuthService struct {
	repo *repository.UserRepo
}

// NewAuthService 创建认证服务
func NewAuthService(repo *repository.UserRepo) *AuthService {
	return &AuthService{repo: repo}
}

// Register 注册新用户
func (s *AuthService) Register(req model.RegisterRequest) (*model.UserResponse, error) {
	// 校验 .edu 邮箱
	if !strings.HasSuffix(req.Email, ".edu") {
		return nil, ErrInvalidEmail
	}

	// 检查邮箱是否已注册
	if _, err := s.repo.FindByEmail(req.Email); err == nil {
		return nil, repository.ErrUserAlreadyExists
	}

	// bcrypt 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 生成邮箱验证 token
	verifyToken, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("生成验证 token 失败: %w", err)
	}

	user := &model.User{
		Email:         req.Email,
		PasswordHash:  string(hash),
		Nickname:      req.Nickname,
		StudentID:     req.StudentID,
		EmailVerified: false,
		VerifyToken:   verifyToken,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	// 开发阶段：打印验证链接到日志
	verifyURL := fmt.Sprintf("http://localhost:8080/api/v1/auth/verify-email?token=%s", verifyToken)
	log.Printf("📧 [DEV] 邮箱验证链接: %s", verifyURL)

	resp := user.ToResponse()
	return &resp, nil
}

// VerifyEmail 验证邮箱
func (s *AuthService) VerifyEmail(token string) error {
	user, err := s.repo.FindByVerifyToken(token)
	if err != nil {
		return ErrInvalidToken
	}

	user.EmailVerified = true
	user.VerifyToken = "" // 一次性使用后清除

	return s.repo.Update(user)
}

// Login 登录
func (s *AuthService) Login(req model.LoginRequest) (*model.AuthResponse, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, repository.ErrUserNotFound
	}

	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrWrongPassword
	}

	// 检查邮箱是否已验证
	if !user.EmailVerified {
		return nil, ErrEmailNotVerified
	}

	// 签发 JWT
	jwtToken, err := s.generateJWT(user.Email)
	if err != nil {
		return nil, fmt.Errorf("JWT 生成失败: %w", err)
	}

	return &model.AuthResponse{
		Token: jwtToken,
		User:  user.ToResponse(),
	}, nil
}

// GetMe 获取当前用户信息
func (s *AuthService) GetMe(email string) (*model.UserResponse, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	resp := user.ToResponse()
	return &resp, nil
}

// generateJWT 生成 JWT token（24 小时有效）
func (s *AuthService) generateJWT(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseJWT 解析 JWT token，返回用户邮箱
func ParseJWT(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名方法: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("无效的 token")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", errors.New("token 中缺少 email 字段")
	}

	return email, nil
}

// generateToken 生成随机 hex token（用于邮箱验证）
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
