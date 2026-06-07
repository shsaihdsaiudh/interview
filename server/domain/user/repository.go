package user

// UserRepository 定义用户持久化接口。
// 接口定义在领域层，实现在基础设施层 —— 依赖反转的核心体现。
type UserRepository interface {
	// Create 创建新用户。如果邮箱已存在，返回 ErrEmailAlreadyExists。
	Create(user *User) error

	// FindByEmail 按邮箱查找用户。找不到返回 ErrUserNotFound。
	FindByEmail(email string) (*User, error)

	// Update 更新用户信息（所有字段全量更新）。
	Update(user *User) error

	// FindAll 返回分页的用户列表，仅包含已验证邮箱的用户。
	// page 从 1 开始；返回值 (用户列表, 总记录数, error)。
	FindAll(page, pageSize int) ([]*User, int, error)

	// FindAllAdmin 管理员查询所有用户（含未验证、已封禁），支持搜索和分页。
	FindAllAdmin(keyword string, page, pageSize int) ([]*User, int, error)

	// CountByDate 统计指定日期及之后注册的用户数。
	CountByDate(since string) (int, error)

	// Delete 删除用户及其关联的空闲时间和预约。
	// 如果用户不存在，返回 ErrUserNotFound。
	Delete(email string) error
}
