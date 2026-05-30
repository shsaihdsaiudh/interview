package user

import "errors"

// 领域错误 — 与具体实现无关的纯业务语义错误。
var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrEmailAlreadyExists = errors.New("邮箱已被注册")
	ErrInvalidEmail      = errors.New("仅支持 .edu 邮箱注册")
	ErrWrongPassword     = errors.New("密码错误")
	ErrEmailNotVerified  = errors.New("邮箱未验证，请先验证邮箱")
	ErrInvalidToken      = errors.New("无效的验证 token")
)
