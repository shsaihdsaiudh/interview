package user

import (
	"testing"
	"time"
)

// =============================================================================
// IsVerified
// =============================================================================

func TestUser_IsVerified(t *testing.T) {
	t.Run("initial false", func(t *testing.T) {
		u := &User{EmailVerified: false}
		if u.IsVerified() {
			t.Error("expected IsVerified to be false for new user")
		}
	})

	t.Run("after mark verified", func(t *testing.T) {
		u := &User{EmailVerified: true}
		if !u.IsVerified() {
			t.Error("expected IsVerified to be true after MarkVerified")
		}
	})
}

// =============================================================================
// ClearVerifyToken
// =============================================================================

func TestUser_ClearVerifyToken(t *testing.T) {
	u := &User{VerifyToken: "abc123"}
	u.ClearVerifyToken()
	if u.VerifyToken != "" {
		t.Errorf("expected VerifyToken to be empty after ClearVerifyToken, got %q", u.VerifyToken)
	}
}

// =============================================================================
// MarkVerified
// =============================================================================

func TestUser_MarkVerified(t *testing.T) {
	u := &User{
		EmailVerified: false,
		VerifyToken:   "token123",
	}
	u.MarkVerified()

	if !u.EmailVerified {
		t.Error("expected EmailVerified to be true after MarkVerified")
	}
	if u.VerifyToken != "" {
		t.Errorf("expected VerifyToken to be cleared after MarkVerified, got %q", u.VerifyToken)
	}
}

// =============================================================================
// UpdateProfile
// =============================================================================

func TestUser_UpdateProfile(t *testing.T) {
	t.Run("update all fields", func(t *testing.T) {
		u := &User{}
		u.UpdateProfile("Alice", "S001", "CS", []string{"go", "rust"}, "avatar.png", "alice@example.com")

		if u.Nickname != "Alice" {
			t.Errorf("Nickname = %q, want %q", u.Nickname, "Alice")
		}
		if u.StudentID != "S001" {
			t.Errorf("StudentID = %q, want %q", u.StudentID, "S001")
		}
		if u.Department != "CS" {
			t.Errorf("Department = %q, want %q", u.Department, "CS")
		}
		if len(u.Tags) != 2 || u.Tags[0] != "go" {
			t.Errorf("Tags = %v, want [go rust]", u.Tags)
		}
		if u.Avatar != "avatar.png" {
			t.Errorf("Avatar = %q, want %q", u.Avatar, "avatar.png")
		}
		if u.ContactInfo != "alice@example.com" {
			t.Errorf("ContactInfo = %q, want %q", u.ContactInfo, "alice@example.com")
		}
	})

	t.Run("empty nickname does not overwrite", func(t *testing.T) {
		u := &User{Nickname: "OldName"}
		u.UpdateProfile("", "", "", nil, "", "")
		if u.Nickname != "OldName" {
			t.Errorf("Nickname = %q, want %q (should not overwrite with empty)", u.Nickname, "OldName")
		}
	})

	t.Run("empty studentID does not overwrite", func(t *testing.T) {
		u := &User{StudentID: "S999"}
		u.UpdateProfile("", "", "", nil, "", "")
		if u.StudentID != "S999" {
			t.Errorf("StudentID = %q, want %q (should not overwrite with empty)", u.StudentID, "S999")
		}
	})

	t.Run("department overwrites with empty", func(t *testing.T) {
		u := &User{Department: "Math"}
		u.UpdateProfile("", "", "", nil, "", "")
		if u.Department != "" {
			t.Errorf("Department = %q, want empty (department always overwrites)", u.Department)
		}
	})

	t.Run("nil tags does not overwrite", func(t *testing.T) {
		u := &User{Tags: []string{"python"}}
		u.UpdateProfile("", "", "", nil, "", "")
		if len(u.Tags) != 1 || u.Tags[0] != "python" {
			t.Errorf("Tags = %v, want [python] (nil should not overwrite)", u.Tags)
		}
	})

	t.Run("empty tags overwrites", func(t *testing.T) {
		u := &User{Tags: []string{"python"}}
		u.UpdateProfile("", "", "", []string{}, "", "")
		if len(u.Tags) != 0 {
			t.Errorf("Tags = %v, want [] (empty slice should overwrite)", u.Tags)
		}
	})

	t.Run("avatar overwrites with empty", func(t *testing.T) {
		u := &User{Avatar: "old.png"}
		u.UpdateProfile("", "", "", nil, "", "")
		if u.Avatar != "" {
			t.Errorf("Avatar = %q, want empty (avatar always overwrites)", u.Avatar)
		}
	})
}

// =============================================================================
// ToResponse
// =============================================================================

func TestUser_ToResponse(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	u := &User{
		Email:         "alice@school.edu",
		Nickname:      "Alice",
		StudentID:     "S001",
		Department:    "CS",
		Tags:          []string{"go"},
		Avatar:        "a.png",
		ContactInfo:   "secret@email.com",
		EmailVerified: true,
		CreatedAt:     now,
	}

	resp := u.ToResponse()

	if resp.ContactInfo != "" {
		t.Errorf("ToResponse should NOT include ContactInfo, got %q", resp.ContactInfo)
	}
	if resp.Email != "alice@school.edu" {
		t.Errorf("Email = %q", resp.Email)
	}
	if resp.EmailVerified != true {
		t.Error("EmailVerified should be true")
	}
	if !resp.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", resp.CreatedAt, now)
	}
}

// =============================================================================
// ToResponseWithContact
// =============================================================================

func TestUser_ToResponseWithContact(t *testing.T) {
	u := &User{
		Email:       "bob@school.edu",
		Nickname:    "Bob",
		ContactInfo: "bob@real.com",
	}

	resp := u.ToResponseWithContact()

	if resp.ContactInfo != "bob@real.com" {
		t.Errorf("ToResponseWithContact should include ContactInfo, got %q", resp.ContactInfo)
	}
}
