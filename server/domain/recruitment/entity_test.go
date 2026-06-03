package recruitment

import (
	"testing"
	"time"
)

// =============================================================================
// Role constants
// =============================================================================

func TestRoleConstants(t *testing.T) {
	if RoleInterviewee != "interviewee" {
		t.Errorf("RoleInterviewee = %q", RoleInterviewee)
	}
	if RoleInterviewer != "interviewer" {
		t.Errorf("RoleInterviewer = %q", RoleInterviewer)
	}
	if RoleBoth != "both" {
		t.Errorf("RoleBoth = %q", RoleBoth)
	}
}

// =============================================================================
// Activate / Deactivate
// =============================================================================

func TestRecruitmentCard_ActivateDeactivate(t *testing.T) {
	t.Run("activate sets is_active to true", func(t *testing.T) {
		c := &RecruitmentCard{IsActive: false}
		c.Activate()
		if !c.IsActive {
			t.Error("IsActive should be true after Activate")
		}
	})

	t.Run("deactivate sets is_active to false", func(t *testing.T) {
		c := &RecruitmentCard{IsActive: true}
		c.Deactivate()
		if c.IsActive {
			t.Error("IsActive should be false after Deactivate")
		}
	})
}

// =============================================================================
// CanBeManagedBy
// =============================================================================

func TestRecruitmentCard_CanBeManagedBy(t *testing.T) {
	c := &RecruitmentCard{UserID: "alice@std.uestc.edu.cn"}

	t.Run("owner can manage", func(t *testing.T) {
		if !c.CanBeManagedBy("alice@std.uestc.edu.cn") {
			t.Error("owner should be able to manage")
		}
	})

	t.Run("other cannot manage", func(t *testing.T) {
		if c.CanBeManagedBy("bob@std.uestc.edu.cn") {
			t.Error("other user should NOT be able to manage")
		}
	})

	t.Run("empty userID", func(t *testing.T) {
		if c.CanBeManagedBy("") {
			t.Error("empty userID should not match")
		}
	})
}

// =============================================================================
// ToResponse
// =============================================================================

func TestRecruitmentCard_ToResponse(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	c := &RecruitmentCard{
		ID:              "card-1",
		UserID:          "alice@std.uestc.edu.cn",
		Skills:          []string{"React", "TypeScript"},
		TargetCompanies: []string{"字节", "腾讯"},
		Role:            RoleBoth,
		ExperienceYears: 3,
		Bio:             "资深前端",
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	r := c.ToResponse()
	if r.ID != "card-1" {
		t.Errorf("ID = %q", r.ID)
	}
	if r.UserID != "alice@std.uestc.edu.cn" {
		t.Errorf("UserID = %q", r.UserID)
	}
	if len(r.Skills) != 2 {
		t.Errorf("Skills len = %d", len(r.Skills))
	}
	if len(r.TargetCompanies) != 2 {
		t.Errorf("TargetCompanies len = %d", len(r.TargetCompanies))
	}
	if r.Role != RoleBoth {
		t.Errorf("Role = %q", r.Role)
	}
	if r.ExperienceYears != 3 {
		t.Errorf("ExperienceYears = %d", r.ExperienceYears)
	}
	if r.Bio != "资深前端" {
		t.Errorf("Bio = %q", r.Bio)
	}
	if !r.IsActive {
		t.Error("IsActive should be true")
	}
	if !r.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt mismatch")
	}
	if !r.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt mismatch")
	}
}

func TestRecruitmentCard_ToResponse_NilSlices(t *testing.T) {
	c := &RecruitmentCard{
		ID:              "card-2",
		UserID:          "bob@std.uestc.edu.cn",
		Skills:          nil,
		TargetCompanies: nil,
		Role:            RoleInterviewee,
		IsActive:        true,
	}

	r := c.ToResponse()
	if r.Skills == nil {
		t.Error("Skills should be empty slice, not nil")
	}
	if len(r.Skills) != 0 {
		t.Errorf("Skills len = %d, want 0", len(r.Skills))
	}
	if r.TargetCompanies == nil {
		t.Error("TargetCompanies should be empty slice, not nil")
	}
}

// =============================================================================
// CreateOrUpdateCardRequest
// =============================================================================

func TestCreateOrUpdateCardRequest(t *testing.T) {
	isActive := true
	req := CreateOrUpdateCardRequest{
		Skills:          []string{"Go", "PostgreSQL"},
		TargetCompanies: []string{"字节"},
		Role:            RoleInterviewer,
		ExperienceYears: 5,
		Bio:             "后端老鸟",
		IsActive:        &isActive,
	}

	if req.Role != RoleInterviewer {
		t.Errorf("Role = %q", req.Role)
	}
	if req.ExperienceYears != 5 {
		t.Errorf("ExperienceYears = %d", req.ExperienceYears)
	}
	if req.IsActive == nil || !*req.IsActive {
		t.Error("IsActive should be true")
	}

	// nil IsActive
	req2 := CreateOrUpdateCardRequest{
		Role: RoleInterviewee,
	}
	if req2.IsActive != nil {
		t.Error("IsActive should be nil when omitted")
	}
}

// =============================================================================
// safeSlice
// =============================================================================

func TestSafeSlice(t *testing.T) {
	t.Run("nil returns empty", func(t *testing.T) {
		s := safeSlice(nil)
		if s == nil {
			t.Error("safeSlice(nil) should return empty slice, not nil")
		}
		if len(s) != 0 {
			t.Errorf("len = %d, want 0", len(s))
		}
	})

	t.Run("empty returns empty", func(t *testing.T) {
		s := safeSlice([]string{})
		if s == nil {
			t.Error("should return empty slice")
		}
		if len(s) != 0 {
			t.Errorf("len = %d, want 0", len(s))
		}
	})

	t.Run("non-empty returns same", func(t *testing.T) {
		input := []string{"a", "b"}
		s := safeSlice(input)
		if len(s) != 2 {
			t.Errorf("len = %d, want 2", len(s))
		}
		if s[0] != "a" || s[1] != "b" {
			t.Errorf("content mismatch: %v", s)
		}
	})
}
