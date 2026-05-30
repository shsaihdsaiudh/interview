package application

import (
	"testing"
	"time"

	"interview-server/domain/appointment"
	"interview-server/domain/user"
)

// newTestAppointmentService 创建测试用的 AppointmentService 实例。
func newTestAppointmentService(ar *mockApptRepo, ur *mockUserRepo) *AppointmentService {
	return &AppointmentService{apptRepo: ar, userRepo: ur}
}

// =============================================================================
// GetMyAvailability
// =============================================================================

func TestAppointmentService_GetMyAvailability(t *testing.T) {
	t.Run("returns empty when no slots", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		slots := svc.GetMyAvailability("user@school.edu")
		if len(slots) != 0 {
			t.Errorf("expected empty slice, got %d items", len(slots))
		}
	})

	t.Run("returns user slots", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		ar.availabilities["s1"] = &appointment.Availability{ID: "s1", UserID: "mentor@school.edu", Date: "2025-07-01", StartTime: "14:00", EndTime: "15:00"}
		ar.availabilities["s2"] = &appointment.Availability{ID: "s2", UserID: "mentor@school.edu", Date: "2025-07-02", StartTime: "10:00", EndTime: "11:00"}
		ar.availabilities["s3"] = &appointment.Availability{ID: "s3", UserID: "other@school.edu", Date: "2025-07-03", StartTime: "09:00", EndTime: "10:00"}

		slots := svc.GetMyAvailability("mentor@school.edu")
		if len(slots) != 2 {
			t.Errorf("expected 2 slots, got %d", len(slots))
		}
	})
}

// =============================================================================
// AddAvailability
// =============================================================================

func TestAppointmentService_AddAvailability(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		slot, err := svc.AddAvailability("mentor@school.edu", appointment.AddAvailabilityRequest{
			Date:      "2099-12-31",
			StartTime: "14:00",
			EndTime:   "15:00",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if slot.UserID != "mentor@school.edu" {
			t.Errorf("UserID = %q", slot.UserID)
		}
		if slot.Date != "2099-12-31" {
			t.Errorf("Date = %q", slot.Date)
		}
	})

	t.Run("end time must be after start time", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		_, err := svc.AddAvailability("m@school.edu", appointment.AddAvailabilityRequest{
			Date:      "2099-12-31",
			StartTime: "15:00",
			EndTime:   "14:00",
		})
		if err != appointment.ErrTimeConflict {
			t.Errorf("expected ErrTimeConflict, got %v", err)
		}
	})

	t.Run("same time rejected", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		_, err := svc.AddAvailability("m@school.edu", appointment.AddAvailabilityRequest{
			Date:      "2099-12-31",
			StartTime: "14:00",
			EndTime:   "14:00",
		})
		if err != appointment.ErrTimeConflict {
			t.Errorf("expected ErrTimeConflict, got %v", err)
		}
	})

	t.Run("invalid start time format", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		_, err := svc.AddAvailability("m@school.edu", appointment.AddAvailabilityRequest{
			Date:      "2099-12-31",
			StartTime: "25:00",
			EndTime:   "26:00",
		})
		if err != appointment.ErrInvalidTimeFormat {
			t.Errorf("expected ErrInvalidTimeFormat, got %v", err)
		}
	})

	t.Run("past date rejected", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		_, err := svc.AddAvailability("m@school.edu", appointment.AddAvailabilityRequest{
			Date:      "2020-01-01",
			StartTime: "14:00",
			EndTime:   "15:00",
		})
		if err != appointment.ErrPastDate {
			t.Errorf("expected ErrPastDate, got %v", err)
		}
	})

	t.Run("bad date format", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		_, err := svc.AddAvailability("m@school.edu", appointment.AddAvailabilityRequest{
			Date:      "not-a-date",
			StartTime: "14:00",
			EndTime:   "15:00",
		})
		if err == nil {
			t.Error("expected error for bad date format")
		}
	})
}

// =============================================================================
// DeleteAvailability
// =============================================================================

func TestAppointmentService_DeleteAvailability(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		ar.availabilities["s1"] = &appointment.Availability{ID: "s1", UserID: "mentor@school.edu", Date: "2099-07-01", StartTime: "14:00", EndTime: "15:00"}

		err := svc.DeleteAvailability("mentor@school.edu", "s1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, exists := ar.availabilities["s1"]; exists {
			t.Error("slot should be deleted")
		}
	})

	t.Run("slot not found", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		err := svc.DeleteAvailability("mentor@school.edu", "nonexistent")
		if err != appointment.ErrSlotNotFound {
			t.Errorf("expected ErrSlotNotFound, got %v", err)
		}
	})

	t.Run("not owner", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		ar.availabilities["s1"] = &appointment.Availability{ID: "s1", UserID: "alice@school.edu", Date: "2099-07-01", StartTime: "14:00", EndTime: "15:00"}

		err := svc.DeleteAvailability("bob@school.edu", "s1")
		if err != appointment.ErrSlotNotOwned {
			t.Errorf("expected ErrSlotNotOwned, got %v", err)
		}
	})
}

// =============================================================================
// CreateAppointment
// =============================================================================

func TestAppointmentService_CreateAppointment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		// 设置 mentor
		ur.usersByEmail["mentor@school.edu"] = &user.User{Email: "mentor@school.edu", Nickname: "Mentor"}
		// 设置空闲时间
		ar.availabilities["slot-1"] = &appointment.Availability{
			ID: "slot-1", UserID: "mentor@school.edu",
			Date: "2099-12-31", StartTime: "14:00", EndTime: "15:00",
		}

		appt, err := svc.CreateAppointment("student@school.edu", appointment.CreateAppointmentRequest{
			TimeSlotID: "slot-1",
			Message:    "请教问题",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if appt.Status != appointment.StatusPending {
			t.Errorf("Status = %q, want pending", appt.Status)
		}
		if appt.StudentID != "student@school.edu" {
			t.Errorf("StudentID = %q", appt.StudentID)
		}
		if appt.MentorID != "mentor@school.edu" {
			t.Errorf("MentorID = %q", appt.MentorID)
		}
	})

	t.Run("slot not found", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		_, err := svc.CreateAppointment("student@school.edu", appointment.CreateAppointmentRequest{
			TimeSlotID: "nonexistent",
			Message:    "hi",
		})
		if err != appointment.ErrSlotNotFound {
			t.Errorf("expected ErrSlotNotFound, got %v", err)
		}
	})

	t.Run("cannot book own slot", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		ur.usersByEmail["mentor@school.edu"] = &user.User{Email: "mentor@school.edu", Nickname: "Mentor"}
		ar.availabilities["slot-1"] = &appointment.Availability{
			ID: "slot-1", UserID: "mentor@school.edu",
			Date: "2099-12-31", StartTime: "14:00", EndTime: "15:00",
		}

		_, err := svc.CreateAppointment("mentor@school.edu", appointment.CreateAppointmentRequest{
			TimeSlotID: "slot-1",
			Message:    "自己约自己",
		})
		if err != appointment.ErrCannotBookOwnSlot {
			t.Errorf("expected ErrCannotBookOwnSlot, got %v", err)
		}
	})

	t.Run("mentor not found", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		// 不创建 mentor 用户
		ar.availabilities["slot-1"] = &appointment.Availability{
			ID: "slot-1", UserID: "nonexistent@school.edu",
			Date: "2099-12-31", StartTime: "14:00", EndTime: "15:00",
		}

		_, err := svc.CreateAppointment("student@school.edu", appointment.CreateAppointmentRequest{
			TimeSlotID: "slot-1",
			Message:    "hi",
		})
		if err != appointment.ErrMentorNotFound {
			t.Errorf("expected ErrMentorNotFound, got %v", err)
		}
	})

	t.Run("slot already booked", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		ur.usersByEmail["mentor@school.edu"] = &user.User{Email: "mentor@school.edu", Nickname: "Mentor"}
		ar.availabilities["slot-1"] = &appointment.Availability{
			ID: "slot-1", UserID: "mentor@school.edu",
			Date: "2099-12-31", StartTime: "14:00", EndTime: "15:00",
		}
		// 先创建一个 pending 预约
		ar.appointments["existing"] = &appointment.Appointment{
			ID: "existing", TimeSlotID: "slot-1", Status: appointment.StatusPending,
		}

		_, err := svc.CreateAppointment("student@school.edu", appointment.CreateAppointmentRequest{
			TimeSlotID: "slot-1",
			Message:    "hi",
		})
		if err != appointment.ErrSlotAlreadyBooked {
			t.Errorf("expected ErrSlotAlreadyBooked, got %v", err)
		}
	})

	t.Run("past date slot", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		ur.usersByEmail["mentor@school.edu"] = &user.User{Email: "mentor@school.edu", Nickname: "Mentor"}
		ar.availabilities["slot-1"] = &appointment.Availability{
			ID: "slot-1", UserID: "mentor@school.edu",
			Date: "2020-01-01", StartTime: "14:00", EndTime: "15:00",
		}

		_, err := svc.CreateAppointment("student@school.edu", appointment.CreateAppointmentRequest{
			TimeSlotID: "slot-1",
			Message:    "hi",
		})
		if err != appointment.ErrPastDate {
			t.Errorf("expected ErrPastDate, got %v", err)
		}
	})
}

// =============================================================================
// GetMyAppointments
// =============================================================================

func TestAppointmentService_GetMyAppointments(t *testing.T) {
	t.Run("mentor role returns received appointments", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		ar.appointments["a1"] = &appointment.Appointment{
			ID: "a1", MentorID: "mentor@school.edu", StudentID: "s1@school.edu", Status: appointment.StatusPending,
		}
		ar.appointments["a2"] = &appointment.Appointment{
			ID: "a2", MentorID: "mentor@school.edu", StudentID: "s2@school.edu", Status: appointment.StatusAccepted,
		}
		ar.appointments["a3"] = &appointment.Appointment{
			ID: "a3", MentorID: "other@school.edu", StudentID: "mentor@school.edu", Status: appointment.StatusPending,
		}

		apps, err := svc.GetMyAppointments("mentor@school.edu", "mentor")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(apps) != 2 {
			t.Errorf("expected 2 appointments as mentor, got %d", len(apps))
		}
	})

	t.Run("student role returns sent appointments", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		ar.appointments["a1"] = &appointment.Appointment{
			ID: "a1", MentorID: "m1@school.edu", StudentID: "student@school.edu", Status: appointment.StatusPending,
		}

		apps, err := svc.GetMyAppointments("student@school.edu", "student")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(apps) != 1 {
			t.Errorf("expected 1 appointment as student, got %d", len(apps))
		}
	})

	t.Run("empty result returns empty slice", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		apps, err := svc.GetMyAppointments("nobody@school.edu", "mentor")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(apps) != 0 {
			t.Errorf("expected empty slice, got %d", len(apps))
		}
	})
}

// =============================================================================
// AcceptAppointment
// =============================================================================

func TestAppointmentService_AcceptAppointment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		ar.appointments["a1"] = &appointment.Appointment{
			ID: "a1", MentorID: "mentor@school.edu", StudentID: "s@school.edu", Status: appointment.StatusPending,
		}

		appt, err := svc.AcceptAppointment("mentor@school.edu", "a1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if appt.Status != appointment.StatusAccepted {
			t.Errorf("Status = %q, want accepted", appt.Status)
		}
	})

	t.Run("appointment not found", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		_, err := svc.AcceptAppointment("mentor@school.edu", "nonexistent")
		if err != appointment.ErrAppointmentNotFound {
			t.Errorf("expected ErrAppointmentNotFound, got %v", err)
		}
	})

	t.Run("not mentor", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		ar.appointments["a1"] = &appointment.Appointment{
			ID: "a1", MentorID: "mentor@school.edu", StudentID: "s@school.edu", Status: appointment.StatusPending,
		}

		_, err := svc.AcceptAppointment("someone_else@school.edu", "a1")
		if err != appointment.ErrNotMentor {
			t.Errorf("expected ErrNotMentor, got %v", err)
		}
	})

	t.Run("not pending", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		ar.appointments["a1"] = &appointment.Appointment{
			ID: "a1", MentorID: "mentor@school.edu", StudentID: "s@school.edu", Status: appointment.StatusRejected,
		}

		_, err := svc.AcceptAppointment("mentor@school.edu", "a1")
		if err != appointment.ErrNotPending {
			t.Errorf("expected ErrNotPending, got %v", err)
		}
	})
}

// =============================================================================
// RejectAppointment
// =============================================================================

func TestAppointmentService_RejectAppointment(t *testing.T) {
	t.Run("success with reason", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		ar.appointments["a1"] = &appointment.Appointment{
			ID: "a1", MentorID: "mentor@school.edu", StudentID: "s@school.edu", Status: appointment.StatusPending,
		}

		appt, err := svc.RejectAppointment("mentor@school.edu", "a1", "时间不合适")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if appt.Status != appointment.StatusRejected {
			t.Errorf("Status = %q, want rejected", appt.Status)
		}
		if appt.RejectReason != "时间不合适" {
			t.Errorf("RejectReason = %q", appt.RejectReason)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		_, err := svc.RejectAppointment("mentor@school.edu", "nonexistent", "")
		if err != appointment.ErrAppointmentNotFound {
			t.Errorf("expected ErrAppointmentNotFound, got %v", err)
		}
	})

	t.Run("not mentor", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		ar.appointments["a1"] = &appointment.Appointment{
			ID: "a1", MentorID: "mentor@school.edu", StudentID: "s@school.edu", Status: appointment.StatusPending,
		}

		_, err := svc.RejectAppointment("student@school.edu", "a1", "不要了")
		if err != appointment.ErrNotMentor {
			t.Errorf("expected ErrNotMentor, got %v", err)
		}
	})
}

// =============================================================================
// ResolveUsers
// =============================================================================

func TestAppointmentService_ResolveUsers(t *testing.T) {
	t.Run("success with pending status (no contacts)", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		ur.usersByEmail["mentor@school.edu"] = &user.User{
			Email: "mentor@school.edu", Nickname: "Mentor", ContactInfo: "m@private.com",
		}
		ur.usersByEmail["student@school.edu"] = &user.User{
			Email: "student@school.edu", Nickname: "Student", ContactInfo: "s@private.com",
		}
		ar.availabilities["slot-1"] = &appointment.Availability{
			ID: "slot-1", UserID: "mentor@school.edu", Date: "2099-07-01", StartTime: "14:00", EndTime: "15:00",
		}

		appt := &appointment.Appointment{
			ID: "a1", MentorID: "mentor@school.edu", StudentID: "student@school.edu",
			TimeSlotID: "slot-1", Message: "hi", Status: appointment.StatusPending,
			CreatedAt: time.Now(),
		}

		resp, err := svc.ResolveUsers(appt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// 类型断言：resp.Mentor 和 resp.Student 是 user.UserResponse
		mentorResp, ok := resp.Mentor.(user.UserResponse)
		if !ok {
			t.Fatal("Mentor is not UserResponse")
		}
		studentResp, ok := resp.Student.(user.UserResponse)
		if !ok {
			t.Fatal("Student is not UserResponse")
		}

		// pending 状态下不展示联系方式
		if mentorResp.ContactInfo != "" {
			t.Errorf("mentor contact should be hidden for pending, got %q", mentorResp.ContactInfo)
		}
		if studentResp.ContactInfo != "" {
			t.Errorf("student contact should be hidden for pending, got %q", studentResp.ContactInfo)
		}
	})

	t.Run("accepted reveals contacts", func(t *testing.T) {
		ar := newMockApptRepo()
		ur := newMockUserRepo()
		svc := newTestAppointmentService(ar, ur)

		ur.usersByEmail["mentor@school.edu"] = &user.User{
			Email: "mentor@school.edu", Nickname: "Mentor", ContactInfo: "m@private.com",
		}
		ur.usersByEmail["student@school.edu"] = &user.User{
			Email: "student@school.edu", Nickname: "Student", ContactInfo: "s@private.com",
		}
		ar.availabilities["slot-1"] = &appointment.Availability{
			ID: "slot-1", UserID: "mentor@school.edu", Date: "2099-07-01", StartTime: "14:00", EndTime: "15:00",
		}

		appt := &appointment.Appointment{
			ID: "a1", MentorID: "mentor@school.edu", StudentID: "student@school.edu",
			TimeSlotID: "slot-1", Message: "hi", Status: appointment.StatusAccepted,
		}

		resp, err := svc.ResolveUsers(appt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		mentorResp := resp.Mentor.(user.UserResponse)
		studentResp := resp.Student.(user.UserResponse)

		if mentorResp.ContactInfo != "m@private.com" {
			t.Errorf("mentor contact should be visible, got %q", mentorResp.ContactInfo)
		}
		if studentResp.ContactInfo != "s@private.com" {
			t.Errorf("student contact should be visible, got %q", studentResp.ContactInfo)
		}
	})
}
