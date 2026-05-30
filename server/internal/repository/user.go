package repository

import (
	"errors"

	"interview-server/internal/model"
)

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserAlreadyExists = errors.New("用户已存在")
)

// UserRepo 用户数据仓库
type UserRepo struct {
	store *Store
}

// NewUserRepo 创建新的用户仓库
func NewUserRepo(store *Store) *UserRepo {
	return &UserRepo{store: store}
}

// Create 创建用户，邮箱已存在则返回错误
func (r *UserRepo) Create(user *model.User) error {
	r.store.Lock()
	defer r.store.Unlock()

	if _, exists := r.store.Users[user.Email]; exists {
		return ErrUserAlreadyExists
	}

	r.store.Users[user.Email] = user
	if user.VerifyToken != "" {
		r.store.VerifyTokens[user.VerifyToken] = user.Email
	}
	return nil
}

// FindByEmail 按邮箱查找用户
func (r *UserRepo) FindByEmail(email string) (*model.User, error) {
	r.store.RLock()
	defer r.store.RUnlock()

	user, exists := r.store.Users[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// FindByVerifyToken 按邮箱验证 token 查找用户
func (r *UserRepo) FindByVerifyToken(token string) (*model.User, error) {
	r.store.RLock()
	defer r.store.RUnlock()

	email, exists := r.store.VerifyTokens[token]
	if !exists {
		return nil, ErrUserNotFound
	}
	user, exists := r.store.Users[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Update 更新用户信息
func (r *UserRepo) Update(user *model.User) error {
	r.store.Lock()
	defer r.store.Unlock()

	if _, exists := r.store.Users[user.Email]; !exists {
		return ErrUserNotFound
	}

	// 清除旧的验证 token
	if old, ok := r.store.Users[user.Email]; ok && old.VerifyToken != "" {
		delete(r.store.VerifyTokens, old.VerifyToken)
	}

	r.store.Users[user.Email] = user
	return nil
}

// FindAll 返回所有已验证邮箱的用户
func (r *UserRepo) FindAll() []*model.User {
	r.store.RLock()
	defer r.store.RUnlock()

	var result []*model.User
	for _, u := range r.store.Users {
		if u.EmailVerified {
			result = append(result, u)
		}
	}
	return result
}
