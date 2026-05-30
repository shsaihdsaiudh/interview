package appointment

import (
	"testing"
	"time"
)

// =============================================================================
// Accept
// =============================================================================

func TestAppointment_Accept(t *testing.T) {
	t.Run("pending to accepted", func(t *testing.T) {
		a := &Appointment{Status: StatusPending}
		err := a.Accept()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.Status != StatusAccepted {
			t.Errorf("Status = %q, want %q", a.Status, StatusAccepted)
		}
	})

	t.Run("already accepted", func(t *testing.T) {
		a := &Appointment{Status: StatusAccepted}
		err := a.Accept()
		if err != ErrNotPending {
			t.Errorf("expected ErrNotPending, got %v", err)
		}
	})

	t.Run("already rejected", func(t *testing.T) {
		a := &Appointment{Status: StatusRejected}
		err := a.Accept()
		if err != ErrNotPending {
			t.Errorf("expected ErrNotPending, got %v", err)
		}
	})
}

// =============================================================================
// Reject
// =============================================================================

func TestAppointment_Reject(t *testing.T) {
	t.Run("pending to rejected with reason", func(t *testing.T) {
		a := &Appointment{Status: StatusPending}
		err := a.Reject("时间冲突")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.Status != StatusRejected {
			t.Errorf("Status = %q, want %q", a.Status, StatusRejected)
		}
		if a.RejectReason != "时间冲突" {
			t.Errorf("RejectReason = %q, want %q", a.RejectReason, "时间冲突")
		}
	})

	t.Run("reject with empty reason", func(t *testing.T) {
		a := &Appointment{Status: StatusPending}
		err := a.Reject("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.RejectReason != "" {
			t.Errorf("RejectReason = %q, want empty", a.RejectReason)
		}
	})

	t.Run("already rejected", func(t *testing.T) {
		a := &Appointment{Status: StatusRejected}
		err := a.Reject("又一次")
		if err != ErrNotPending {
			t.Errorf("expected ErrNotPending, got %v", err)
		}
	})

	t.Run("already accepted", func(t *testing.T) {
		a := &Appointment{Status: StatusAccepted}
		err := a.Reject("冲突")
		if err != ErrNotPending {
			t.Errorf("expected ErrNotPending, got %v", err)
		}
	})
}

// =============================================================================
// CanBeOperatedBy
// =============================================================================

func TestAppointment_CanBeOperatedBy(t *testing.T) {
	a := &Appointment{MentorID: "mentor@school.edu"}

	t.Run("mentor can operate", func(t *testing.T) {
		if !a.CanBeOperatedBy("mentor@school.edu") {
			t.Error("mentor should be able to operate")
		}
	})

	t.Run("student cannot operate", func(t *testing.T) {
		if a.CanBeOperatedBy("student@school.edu") {
			t.Error("student should NOT be able to operate")
		}
	})

	t.Run("random user cannot operate", func(t *testing.T) {
		if a.CanBeOperatedBy("other@school.edu") {
			t.Error("random user should NOT be able to operate")
		}
	})

	t.Run("empty userID", func(t *testing.T) {
		if a.CanBeOperatedBy("") {
			t.Error("empty userID should not match mentor")
		}
	})
}

// =============================================================================
// IsAccepted
// =============================================================================

func TestAppointment_IsAccepted(t *testing.T) {
	t.Run("pending returns false", func(t *testing.T) {
		a := &Appointment{Status: StatusPending}
		if a.IsAccepted() {
			t.Error("pending appointment should not be accepted")
		}
	})

	t.Run("accepted returns true", func(t *testing.T) {
		a := &Appointment{Status: StatusAccepted}
		if !a.IsAccepted() {
			t.Error("accepted appointment should return true")
		}
	})

	t.Run("rejected returns false", func(t *testing.T) {
		a := &Appointment{Status: StatusRejected}
		if a.IsAccepted() {
			t.Error("rejected appointment should not be accepted")
		}
	})
}

// =============================================================================
// Status constants
// =============================================================================

func TestStatusConstants(t *testing.T) {
	if StatusPending != "pending" {
		t.Errorf("StatusPending = %q", StatusPending)
	}
	if StatusAccepted != "accepted" {
		t.Errorf("StatusAccepted = %q", StatusAccepted)
	}
	if StatusRejected != "rejected" {
		t.Errorf("StatusRejected = %q", StatusRejected)
	}
}

// =============================================================================
// Appointment struct
// =============================================================================

func TestAppointment_Struct(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	a := &Appointment{
		ID:           "appt-1",
		MentorID:     "mentor@school.edu",
		StudentID:    "student@school.edu",
		TimeSlotID:   "slot-1",
		Message:      "想请教一些问题",
		Status:       StatusPending,
		RejectReason: "",
		CreatedAt:    now,
	}

	if a.ID != "appt-1" {
		t.Errorf("ID = %q", a.ID)
	}
	if a.Message != "想请教一些问题" {
		t.Errorf("Message = %q", a.Message)
	}
	if !a.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt mismatch")
	}
}

// =============================================================================
// Availability struct
// =============================================================================

func TestAvailability_Struct(t *testing.T) {
	a := &Availability{
		ID:        "avail-1",
		UserID:    "mentor@school.edu",
		Date:      "2025-06-15",
		StartTime: "14:00",
		EndTime:   "15:00",
	}

	if a.ID != "avail-1" {
		t.Errorf("ID = %q", a.ID)
	}
	if a.UserID != "mentor@school.edu" {
		t.Errorf("UserID = %q", a.UserID)
	}
	if a.Date != "2025-06-15" {
		t.Errorf("Date = %q", a.Date)
	}
}
