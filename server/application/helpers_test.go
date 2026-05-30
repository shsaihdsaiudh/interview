package application

import (
	"strings"
	"testing"

	"interview-server/domain/appointment"
)

// =============================================================================
// validateTime
// =============================================================================

func TestValidateTime(t *testing.T) {
	t.Run("valid time", func(t *testing.T) {
		if err := validateTime("14:00"); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		if err := validateTime("00:00"); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		if err := validateTime("23:59"); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("wrong length", func(t *testing.T) {
		err := validateTime("14:000")
		if err != appointment.ErrInvalidTimeFormat {
			t.Errorf("expected ErrInvalidTimeFormat, got %v", err)
		}
	})

	t.Run("missing colon", func(t *testing.T) {
		err := validateTime("1400")
		if err != appointment.ErrInvalidTimeFormat {
			t.Errorf("expected ErrInvalidTimeFormat, got %v", err)
		}
	})

	t.Run("invalid hour", func(t *testing.T) {
		err := validateTime("25:00")
		if err != appointment.ErrInvalidTimeFormat {
			t.Errorf("expected ErrInvalidTimeFormat, got %v", err)
		}
	})

	t.Run("invalid minute", func(t *testing.T) {
		err := validateTime("14:60")
		if err != appointment.ErrInvalidTimeFormat {
			t.Errorf("expected ErrInvalidTimeFormat, got %v", err)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		err := validateTime("")
		if err != appointment.ErrInvalidTimeFormat {
			t.Errorf("expected ErrInvalidTimeFormat, got %v", err)
		}
	})

	t.Run("non-numeric", func(t *testing.T) {
		err := validateTime("ab:cd")
		if err != appointment.ErrInvalidTimeFormat {
			t.Errorf("expected ErrInvalidTimeFormat, got %v", err)
		}
	})
}

// =============================================================================
// generateRandomHex
// =============================================================================

func TestGenerateRandomHex(t *testing.T) {
	t.Run("correct length", func(t *testing.T) {
		s, err := generateRandomHex(16)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// 16 bytes → 32 hex chars
		if len(s) != 32 {
			t.Errorf("length = %d, want 32", len(s))
		}
	})

	t.Run("different bytes produce different lengths", func(t *testing.T) {
		s16, _ := generateRandomHex(16)
		s8, _ := generateRandomHex(8)
		if len(s16) != 32 {
			t.Errorf("16 bytes → %d chars", len(s16))
		}
		if len(s8) != 16 {
			t.Errorf("8 bytes → %d chars", len(s8))
		}
	})

	t.Run("two calls produce different results", func(t *testing.T) {
		s1, _ := generateRandomHex(32)
		s2, _ := generateRandomHex(32)
		if s1 == s2 {
			t.Error("two random strings should not be equal")
		}
	})

	t.Run("only hex characters", func(t *testing.T) {
		s, _ := generateRandomHex(64)
		for _, c := range s {
			if !strings.ContainsRune("0123456789abcdef", c) {
				t.Errorf("non-hex character found: %c", c)
				break
			}
		}
	})

	t.Run("zero bytes", func(t *testing.T) {
		s, err := generateRandomHex(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "" {
			t.Errorf("expected empty string, got %q", s)
		}
	})
}
