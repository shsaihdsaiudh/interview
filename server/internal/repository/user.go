package repository

import (
	"errors"
	"sync"

	"interview-server/internal/model"
)

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserAlreadyExists = errors.New("用户已存在")
)

// UserRepo 用户数据仓库（内存存储）
type UserRepo struct {
	mu    sync.RWMutex
	users map[string]*model.User // key = email
}

// NewUserRepo 创建新的用户仓库
func NewUserRepo() *UserRepo {
	return &UserRepo{
		users: make(map[string]*model.User),
	}
}

// Create 创建用户，邮箱已存在则返回错误
func (r *UserRepo) Create(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Email]; exists {
		return ErrUserAlreadyExists
	}

	r.users[user.Email] = user
	return nil
}

// FindByEmail 按邮箱查找用户
func (r *UserRepo) FindByEmail(email string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// FindByVerifyToken 按邮箱验证 token 查找用户
func (r *UserRepo) FindByVerifyToken(token string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.VerifyToken == token {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

// Update 更新用户信息
func (r *UserRepo) Update(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Email]; !exists {
		return ErrUserNotFound
	}

	r.users[user.Email] = user
	return nil
}
