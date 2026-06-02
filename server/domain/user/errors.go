package user

import "errors"

// 领域错误 — 与具体实现无关的纯业务语义错误。
var (
	ErrUserNotFound       = errors.New("用户不存在")
	ErrEmailAlreadyExists = errors.New("邮箱已被注册")
	ErrInvalidEmail       = errors.New("仅支持 @std.uestc.edu.cn 邮箱注册")
	ErrWrongPassword      = errors.New("密码错误")
	ErrWrongOldPassword   = errors.New("旧密码错误")
	ErrEmailNotVerified   = errors.New("邮箱未验证，请先验证邮箱")
	ErrInvalidCode                      = errors.New("验证码错误或已过期")
	ErrCannotDeleteWithActiveAppointments = errors.New("存在进行中的预约，无法注销账号")
)
