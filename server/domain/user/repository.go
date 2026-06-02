package user

// UserRepository 定义用户持久化接口。
// 接口定义在领域层，实现在基础设施层 —— 依赖反转的核心体现。
type UserRepository interface {
	// Create 创建新用户。如果邮箱已存在，返回 ErrEmailAlreadyExists。
	Create(user *User) error

	// FindByEmail 按邮箱查找用户。找不到返回 ErrUserNotFound。
	FindByEmail(email string) (*User, error)

	// FindByVerifyToken 按邮箱验证 token 查找用户。
	FindByVerifyToken(token string) (*User, error)

	// Update 更新用户信息（所有字段全量更新）。
	Update(user *User) error

	// FindAll 返回所有已验证邮箱的用户。
	FindAll() []*User

	// Delete 删除用户及其关联的空闲时间和预约。
	// 如果用户不存在，返回 ErrUserNotFound。
	Delete(email string) error
}
