package application

import (
	"testing"

	"interview-server/domain/recruitment"
	"interview-server/domain/user"
)

// =============================================================================
// Mock
// =============================================================================

type mockRecruitmentCardRepo struct {
	cards map[string]*recruitment.RecruitmentCard
}

func newMockRecruitmentCardRepo() *mockRecruitmentCardRepo {
	return &mockRecruitmentCardRepo{cards: make(map[string]*recruitment.RecruitmentCard)}
}

func (m *mockRecruitmentCardRepo) Upsert(card *recruitment.RecruitmentCard) error {
	m.cards[card.UserID] = card
	return nil
}

func (m *mockRecruitmentCardRepo) FindByUserID(userID string) (*recruitment.RecruitmentCard, error) {
	card, ok := m.cards[userID]
	if !ok {
		return nil, recruitment.ErrCardNotFound
	}
	return card, nil
}

func (m *mockRecruitmentCardRepo) List(filter recruitment.ListCardsFilter) ([]*recruitment.RecruitmentCard, int, error) {
	var matched []*recruitment.RecruitmentCard
	for _, c := range m.cards {
		if !c.IsActive {
			continue
		}
		if filter.Skill != "" {
			found := false
			for _, s := range c.Skills {
				if s == filter.Skill {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if filter.Company != "" {
			found := false
			for _, comp := range c.TargetCompanies {
				if comp == filter.Company {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if filter.Role != "" && c.Role != filter.Role {
			continue
		}
		matched = append(matched, c)
	}

	total := len(matched)
	page := filter.Page
	if page < 1 {
		page = 1
	}
	size := filter.Size
	if size < 1 {
		size = 20
	}
	start := (page - 1) * size
	if start >= total {
		return []*recruitment.RecruitmentCard{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

var _ recruitment.RecruitmentCardRepository = (*mockRecruitmentCardRepo)(nil)

func newTestRecruitmentService(cr *mockRecruitmentCardRepo, ur *mockUserRepo) *RecruitmentService {
	return &RecruitmentService{cardRepo: cr, userRepo: ur}
}

// =============================================================================
// CreateOrUpdateCard
// =============================================================================

func TestRecruitmentService_CreateOrUpdateCard(t *testing.T) {
	t.Run("create new card", func(t *testing.T) {
		cr := newMockRecruitmentCardRepo()
		ur := newMockUserRepo()
		svc := newTestRecruitmentService(cr, ur)

		ur.usersByEmail["alice@std.uestc.edu.cn"] = &user.User{
			Email: "alice@std.uestc.edu.cn", Nickname: "Alice",
		}

		req := recruitment.CreateOrUpdateCardRequest{
			Skills:          []string{"React", "TypeScript"},
			TargetCompanies: []string{"字节"},
			Role:            recruitment.RoleBoth,
			ExperienceYears: 3,
			Bio:             "全栈工程师",
		}

		resp, err := svc.CreateOrUpdateCard("alice@std.uestc.edu.cn", req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.UserID != "alice@std.uestc.edu.cn" {
			t.Errorf("UserID = %q", resp.UserID)
		}
		if resp.Role != recruitment.RoleBoth {
			t.Errorf("Role = %q", resp.Role)
		}
		if len(resp.Skills) != 2 {
			t.Errorf("Skills len = %d", len(resp.Skills))
		}
		if !resp.IsActive {
			t.Error("new card should be active by default")
		}
	})

	t.Run("update existing card", func(t *testing.T) {
		cr := newMockRecruitmentCardRepo()
		ur := newMockUserRepo()
		svc := newTestRecruitmentService(cr, ur)

		ur.usersByEmail["alice@std.uestc.edu.cn"] = &user.User{
			Email: "alice@std.uestc.edu.cn", Nickname: "Alice",
		}

		_, _ = svc.CreateOrUpdateCard("alice@std.uestc.edu.cn", recruitment.CreateOrUpdateCardRequest{
			Skills:          []string{"React"},
			TargetCompanies: []string{"字节"},
			Role:            recruitment.RoleInterviewee,
			ExperienceYears: 2,
		})

		resp, err := svc.CreateOrUpdateCard("alice@std.uestc.edu.cn", recruitment.CreateOrUpdateCardRequest{
			Skills:          []string{"Go", "PostgreSQL"},
			TargetCompanies: []string{"腾讯"},
			Role:            recruitment.RoleInterviewer,
			ExperienceYears: 5,
			Bio:             "更新版",
		})
		if err != nil {
			t.Fatalf("unexpected error on update: %v", err)
		}
		if len(resp.Skills) != 2 {
			t.Errorf("Skills len = %d, want 2", len(resp.Skills))
		}
		if resp.Role != recruitment.RoleInterviewer {
			t.Errorf("Role = %q", resp.Role)
		}
		if resp.Bio != "更新版" {
			t.Errorf("Bio = %q", resp.Bio)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		cr := newMockRecruitmentCardRepo()
		ur := newMockUserRepo()
		svc := newTestRecruitmentService(cr, ur)

		_, err := svc.CreateOrUpdateCard("nobody@std.uestc.edu.cn", recruitment.CreateOrUpdateCardRequest{
			Role: recruitment.RoleInterviewee,
		})
		if err != recruitment.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("invalid role", func(t *testing.T) {
		cr := newMockRecruitmentCardRepo()
		ur := newMockUserRepo()
		svc := newTestRecruitmentService(cr, ur)

		ur.usersByEmail["alice@std.uestc.edu.cn"] = &user.User{
			Email: "alice@std.uestc.edu.cn", Nickname: "Alice",
		}

		_, err := svc.CreateOrUpdateCard("alice@std.uestc.edu.cn", recruitment.CreateOrUpdateCardRequest{
			Role: "invalid_role",
		})
		if err != recruitment.ErrInvalidRole {
			t.Errorf("expected ErrInvalidRole, got %v", err)
		}
	})

	t.Run("explicit is_active false", func(t *testing.T) {
		cr := newMockRecruitmentCardRepo()
		ur := newMockUserRepo()
		svc := newTestRecruitmentService(cr, ur)

		ur.usersByEmail["alice@std.uestc.edu.cn"] = &user.User{
			Email: "alice@std.uestc.edu.cn", Nickname: "Alice",
		}

		isActive := false
		resp, err := svc.CreateOrUpdateCard("alice@std.uestc.edu.cn", recruitment.CreateOrUpdateCardRequest{
			Role:     recruitment.RoleInterviewee,
			IsActive: &isActive,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.IsActive {
			t.Error("IsActive should be false when explicitly set")
		}
	})
}

// =============================================================================
// GetCardByUserID
// =============================================================================

func TestRecruitmentService_GetCardByUserID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cr := newMockRecruitmentCardRepo()
		ur := newMockUserRepo()
		svc := newTestRecruitmentService(cr, ur)

		ur.usersByEmail["alice@std.uestc.edu.cn"] = &user.User{
			Email: "alice@std.uestc.edu.cn", Nickname: "Alice",
		}
		_, _ = svc.CreateOrUpdateCard("alice@std.uestc.edu.cn", recruitment.CreateOrUpdateCardRequest{
			Skills: []string{"Go"},
			Role:   recruitment.RoleInterviewee,
		})

		resp, err := svc.GetCardByUserID("alice@std.uestc.edu.cn")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.UserID != "alice@std.uestc.edu.cn" {
			t.Errorf("UserID = %q", resp.UserID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		cr := newMockRecruitmentCardRepo()
		ur := newMockUserRepo()
		svc := newTestRecruitmentService(cr, ur)

		_, err := svc.GetCardByUserID("nobody@std.uestc.edu.cn")
		if err != recruitment.ErrCardNotFound {
			t.Errorf("expected ErrCardNotFound, got %v", err)
		}
	})
}

// =============================================================================
// ListCards
// =============================================================================

func TestRecruitmentService_ListCards(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		cr := newMockRecruitmentCardRepo()
		ur := newMockUserRepo()
		svc := newTestRecruitmentService(cr, ur)

		resp, err := svc.ListCards(recruitment.ListCardsFilter{Page: 1, Size: 20})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Total != 0 {
			t.Errorf("Total = %d, want 0", resp.Total)
		}
		if len(resp.Cards) != 0 {
			t.Errorf("Cards len = %d, want 0", len(resp.Cards))
		}
	})

	t.Run("list with skill filter", func(t *testing.T) {
		cr := newMockRecruitmentCardRepo()
		ur := newMockUserRepo()
		svc := newTestRecruitmentService(cr, ur)

		ur.usersByEmail["alice@std.uestc.edu.cn"] = &user.User{Email: "alice@std.uestc.edu.cn", Nickname: "Alice"}
		ur.usersByEmail["bob@std.uestc.edu.cn"] = &user.User{Email: "bob@std.uestc.edu.cn", Nickname: "Bob"}

		_, _ = svc.CreateOrUpdateCard("alice@std.uestc.edu.cn", recruitment.CreateOrUpdateCardRequest{
			Skills: []string{"React"}, TargetCompanies: []string{"字节"}, Role: recruitment.RoleInterviewee,
		})
		_, _ = svc.CreateOrUpdateCard("bob@std.uestc.edu.cn", recruitment.CreateOrUpdateCardRequest{
			Skills: []string{"Go"}, TargetCompanies: []string{"腾讯"}, Role: recruitment.RoleInterviewer,
		})

		resp, err := svc.ListCards(recruitment.ListCardsFilter{Skill: "React", Page: 1, Size: 20})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Total != 1 {
			t.Errorf("Total = %d, want 1", resp.Total)
		}
		if resp.Cards[0].UserID != "alice@std.uestc.edu.cn" {
			t.Errorf("UserID = %q", resp.Cards[0].UserID)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		cr := newMockRecruitmentCardRepo()
		ur := newMockUserRepo()
		svc := newTestRecruitmentService(cr, ur)

		uids := []string{"u1@std.uestc.edu.cn", "u2@std.uestc.edu.cn", "u3@std.uestc.edu.cn"}
		for _, uid := range uids {
			ur.usersByEmail[uid] = &user.User{Email: uid, Nickname: uid}
			_, _ = svc.CreateOrUpdateCard(uid, recruitment.CreateOrUpdateCardRequest{
				Skills: []string{"Go"}, Role: recruitment.RoleBoth,
			})
		}

		resp, err := svc.ListCards(recruitment.ListCardsFilter{Page: 1, Size: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Total != 3 {
			t.Errorf("Total = %d, want 3", resp.Total)
		}
		if len(resp.Cards) != 2 {
			t.Errorf("Cards len = %d, want 2", len(resp.Cards))
		}
	})

	t.Run("inactive cards not listed", func(t *testing.T) {
		cr := newMockRecruitmentCardRepo()
		ur := newMockUserRepo()
		svc := newTestRecruitmentService(cr, ur)

		ur.usersByEmail["a@std.uestc.edu.cn"] = &user.User{Email: "a@std.uestc.edu.cn", Nickname: "A"}
		ur.usersByEmail["b@std.uestc.edu.cn"] = &user.User{Email: "b@std.uestc.edu.cn", Nickname: "B"}

		isActive := false
		_, _ = svc.CreateOrUpdateCard("a@std.uestc.edu.cn", recruitment.CreateOrUpdateCardRequest{
			Skills: []string{"Go"}, Role: recruitment.RoleBoth, IsActive: &isActive,
		})
		_, _ = svc.CreateOrUpdateCard("b@std.uestc.edu.cn", recruitment.CreateOrUpdateCardRequest{
			Skills: []string{"React"}, Role: recruitment.RoleBoth,
		})

		resp, err := svc.ListCards(recruitment.ListCardsFilter{Page: 1, Size: 20})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Total != 1 {
			t.Errorf("Total = %d, want 1 (only active)", resp.Total)
		}
	})
}
