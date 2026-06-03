package application

import (
	"time"

	"interview-server/domain/recruitment"
	"interview-server/domain/user"
)

// RecruitmentService 招募卡片用例编排。
type RecruitmentService struct {
	cardRepo recruitment.RecruitmentCardRepository
	userRepo user.UserRepository
}

// NewRecruitmentService 创建招募卡片服务。
func NewRecruitmentService(cardRepo recruitment.RecruitmentCardRepository, userRepo user.UserRepository) *RecruitmentService {
	return &RecruitmentService{cardRepo: cardRepo, userRepo: userRepo}
}

// ── 有效的角色值集合 ──

var validRoles = map[string]bool{
	recruitment.RoleInterviewee: true,
	recruitment.RoleInterviewer: true,
	recruitment.RoleBoth:        true,
}

// ── 名片管理 ──

// CreateOrUpdateCard 创建或更新用户的招募卡片。
func (s *RecruitmentService) CreateOrUpdateCard(userID string, req recruitment.CreateOrUpdateCardRequest) (*recruitment.CardResponse, error) {
	// 校验用户是否存在
	if _, err := s.userRepo.FindByEmail(userID); err != nil {
		return nil, recruitment.ErrUserNotFound
	}

	// 校验角色
	if !validRoles[req.Role] {
		return nil, recruitment.ErrInvalidRole
	}

	// 确定是否活跃
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	now := time.Now()

	// 查找已有卡片（决定是 create 还是 update）
	existing, findErr := s.cardRepo.FindByUserID(userID)
	var card *recruitment.RecruitmentCard

	if findErr != nil {
		// 不存在 → 创建
		id, _ := generateRandomHex(16)
		card = &recruitment.RecruitmentCard{
			ID:              id,
			UserID:          userID,
			Skills:          safeStringSlice(req.Skills),
			TargetCompanies: safeStringSlice(req.TargetCompanies),
			Role:            req.Role,
			ExperienceYears: req.ExperienceYears,
			Bio:             req.Bio,
			IsActive:        isActive,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
	} else {
		// 已存在 → 更新
		existing.Skills = safeStringSlice(req.Skills)
		existing.TargetCompanies = safeStringSlice(req.TargetCompanies)
		existing.Role = req.Role
		existing.ExperienceYears = req.ExperienceYears
		existing.Bio = req.Bio
		existing.IsActive = isActive
		existing.UpdatedAt = now
		card = existing
	}

	if err := s.cardRepo.Upsert(card); err != nil {
		return nil, err
	}

	resp := card.ToResponse()
	return &resp, nil
}

// GetCardByUserID 获取指定用户的招募卡片。
func (s *RecruitmentService) GetCardByUserID(userID string) (*recruitment.CardResponse, error) {
	card, err := s.cardRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	resp := card.ToResponse()
	return &resp, nil
}

// ListCards 列表查询，支持多条件筛选和分页。
func (s *RecruitmentService) ListCards(filter recruitment.ListCardsFilter) (*recruitment.CardListResponse, error) {
	// 设置默认分页
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Size < 1 || filter.Size > 100 {
		filter.Size = 20
	}

	cards, total, err := s.cardRepo.List(filter)
	if err != nil {
		return nil, err
	}

	if cards == nil {
		cards = []*recruitment.RecruitmentCard{}
	}

	responses := make([]recruitment.CardResponse, 0, len(cards))
	for _, c := range cards {
		responses = append(responses, c.ToResponse())
	}

	return &recruitment.CardListResponse{
		Cards:    responses,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.Size,
	}, nil
}

// ── 辅助函数 ──

// safeStringSlice 确保 nil slice 返回空数组而不是 null。
func safeStringSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
