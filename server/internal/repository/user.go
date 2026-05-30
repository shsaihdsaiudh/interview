package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"interview-server/internal/model"
)

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserAlreadyExists = errors.New("用户已存在")
)

// UserRepo 用户数据仓库（PostgreSQL）
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo 创建新的用户仓库
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// Create 创建用户
func (r *UserRepo) Create(user *model.User) error {
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO users (email, password_hash, nickname, student_id, department,
		 tags, avatar, contact_info, email_verified, verify_token, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		user.Email, user.PasswordHash, user.Nickname, user.StudentID,
		user.Department, user.Tags, user.Avatar, user.ContactInfo,
		user.EmailVerified, user.VerifyToken, user.CreatedAt,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

// FindByEmail 按邮箱查找用户
func (r *UserRepo) FindByEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := r.pool.QueryRow(context.Background(),
		`SELECT email, password_hash, nickname, student_id, department,
		        tags, avatar, contact_info, email_verified, verify_token, created_at
		 FROM users WHERE email = $1`, email,
	).Scan(
		&user.Email, &user.PasswordHash, &user.Nickname, &user.StudentID,
		&user.Department, &user.Tags, &user.Avatar, &user.ContactInfo,
		&user.EmailVerified, &user.VerifyToken, &user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByVerifyToken 按验证 token 查找用户
func (r *UserRepo) FindByVerifyToken(token string) (*model.User, error) {
	user := &model.User{}
	err := r.pool.QueryRow(context.Background(),
		`SELECT email, password_hash, nickname, student_id, department,
		        tags, avatar, contact_info, email_verified, verify_token, created_at
		 FROM users WHERE verify_token = $1`, token,
	).Scan(
		&user.Email, &user.PasswordHash, &user.Nickname, &user.StudentID,
		&user.Department, &user.Tags, &user.Avatar, &user.ContactInfo,
		&user.EmailVerified, &user.VerifyToken, &user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update 更新用户信息
func (r *UserRepo) Update(user *model.User) error {
	tag, err := r.pool.Exec(context.Background(),
		`UPDATE users SET password_hash=$1, nickname=$2, student_id=$3,
		 department=$4, tags=$5, avatar=$6, contact_info=$7,
		 email_verified=$8, verify_token=$9, created_at=$10
		 WHERE email=$11`,
		user.PasswordHash, user.Nickname, user.StudentID,
		user.Department, user.Tags, user.Avatar, user.ContactInfo,
		user.EmailVerified, user.VerifyToken, user.CreatedAt,
		user.Email,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// FindAll 返回所有已验证邮箱的用户
func (r *UserRepo) FindAll() []*model.User {
	rows, err := r.pool.Query(context.Background(),
		`SELECT email, password_hash, nickname, student_id, department,
		        tags, avatar, contact_info, email_verified, verify_token, created_at
		 FROM users WHERE email_verified = true ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		if err := rows.Scan(
			&u.Email, &u.PasswordHash, &u.Nickname, &u.StudentID,
			&u.Department, &u.Tags, &u.Avatar, &u.ContactInfo,
			&u.EmailVerified, &u.VerifyToken, &u.CreatedAt,
		); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users
}

// isDuplicateKey 判断是否为唯一约束冲突
func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
