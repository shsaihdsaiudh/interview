package recruitment

import "errors"

// 领域错误
var (
	ErrCardNotFound    = errors.New("招募卡片不存在")
	ErrInvalidRole     = errors.New("角色必须为 interviewee、interviewer 或 both")
	ErrUserNotFound    = errors.New("用户不存在")
	ErrCardNotOwned    = errors.New("只能管理自己的卡片")
	ErrSkillEmpty      = errors.New("技能标签不能为空")
)
