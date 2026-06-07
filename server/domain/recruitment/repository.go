package recruitment

// RecruitmentCardRepository 定义招募卡片持久化接口。
// 接口定义在领域层，实现在基础设施层。
type RecruitmentCardRepository interface {
	// Upsert 创建或更新卡片（user_id 唯一约束）。
	Upsert(card *RecruitmentCard) error

	// FindByUserID 按用户 ID 查找卡片。
	// 找不到返回 ErrCardNotFound。
	FindByUserID(userID string) (*RecruitmentCard, error)

	// List 列表查询，支持多条件筛选和分页。
	// 返回 (卡片列表, 总记录数, error)。
	List(filter ListCardsFilter) ([]*RecruitmentCard, int, error)

	// ListAllAdmin 管理员查询所有卡片（含非活跃），支持搜索和分页。
	ListAllAdmin(keyword string, page, pageSize int) ([]*RecruitmentCard, int, error)

	// DeleteByID 管理员删除卡片。
	DeleteByID(id string) error
}
