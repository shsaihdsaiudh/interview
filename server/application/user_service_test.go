package application

import (
	"testing"

	"interview-server/domain/appointment"
	"interview-server/domain/user"
)

// =============================================================================
// Mock 实现
// =============================================================================

// mockUserRepo 实现 user.UserRepository，用于单元测试。
type mockUserRepo struct {
	usersByEmail      map[string]*user.User
	usersByToken      map[string]*user.User
	createErr         error
	findByEmailErr    error
	findByTokenErr    error
	updateErr         error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		usersByEmail: make(map[string]*user.User),
		usersByToken: make(map[string]*user.User),
	}
}

func (m *mockUserRepo) Create(u *user.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	if _, exists := m.usersByEmail[u.Email]; exists {
		return user.ErrEmailAlreadyExists
	}
	m.usersByEmail[u.Email] = u
	if u.VerifyToken != "" {
		m.usersByToken[u.VerifyToken] = u
	}
	return nil
}

func (m *mockUserRepo) FindByEmail(email string) (*user.User, error) {
	if m.findByEmailErr != nil {
		return nil, m.findByEmailErr
	}
	u, ok := m.usersByEmail[email]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) FindByVerifyToken(token string) (*user.User, error) {
	if m.findByTokenErr != nil {
		return nil, m.findByTokenErr
	}
	u, ok := m.usersByToken[token]
	if !ok {
		return nil, user.ErrInvalidToken
	}
	return u, nil
}

func (m *mockUserRepo) Update(u *user.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.usersByEmail[u.Email]; !ok {
		return user.ErrUserNotFound
	}
	m.usersByEmail[u.Email] = u
	return nil
}

func (m *mockUserRepo) FindAll() []*user.User {
	var result []*user.User
	for _, u := range m.usersByEmail {
		if u.EmailVerified {
			result = append(result, u)
		}
	}
	return result
}

// mockApptRepo 实现 appointment.AppointmentRepository，用于单元测试。
type mockApptRepo struct {
	appointments  map[string]*appointment.Appointment
	availabilities map[string]*appointment.Availability
}

func newMockApptRepo() *mockApptRepo {
	return &mockApptRepo{
		appointments:   make(map[string]*appointment.Appointment),
		availabilities: make(map[string]*appointment.Availability),
	}
}

func (m *mockApptRepo) CreateAppointment(a *appointment.Appointment) error {
	m.appointments[a.ID] = a
	return nil
}

func (m *mockApptRepo) UpdateAppointment(a *appointment.Appointment) error {
	m.appointments[a.ID] = a
	return nil
}

func (m *mockApptRepo) FindAppointmentByID(id string) (*appointment.Appointment, error) {
	a, ok := m.appointments[id]
	if !ok {
		return nil, appointment.ErrAppointmentNotFound
	}
	return a, nil
}

func (m *mockApptRepo) FindAppointmentsByMentorID(mentorID string) []*appointment.Appointment {
	var result []*appointment.Appointment
	for _, a := range m.appointments {
		if a.MentorID == mentorID {
			result = append(result, a)
		}
	}
	return result
}

func (m *mockApptRepo) FindAppointmentsByStudentID(studentID string) []*appointment.Appointment {
	var result []*appointment.Appointment
	for _, a := range m.appointments {
		if a.StudentID == studentID {
			result = append(result, a)
		}
	}
	return result
}

func (m *mockApptRepo) FindAppointmentsByTimeSlotID(timeSlotID string) []*appointment.Appointment {
	var result []*appointment.Appointment
	for _, a := range m.appointments {
		if a.TimeSlotID == timeSlotID {
			result = append(result, a)
		}
	}
	return result
}

func (m *mockApptRepo) HasActiveAppointment(timeSlotID string) bool {
	for _, a := range m.appointments {
		if a.TimeSlotID == timeSlotID && (a.Status == appointment.StatusPending || a.Status == appointment.StatusAccepted) {
			return true
		}
	}
	return false
}

func (m *mockApptRepo) CreateAvailability(a *appointment.Availability) error {
	m.availabilities[a.ID] = a
	return nil
}

func (m *mockApptRepo) DeleteAvailability(id string) error {
	delete(m.availabilities, id)
	return nil
}

func (m *mockApptRepo) FindAvailabilityByID(id string) (*appointment.Availability, error) {
	a, ok := m.availabilities[id]
	if !ok {
		return nil, appointment.ErrSlotNotFound
	}
	return a, nil
}

func (m *mockApptRepo) FindAvailabilitiesByUserID(userID string) []*appointment.Availability {
	var result []*appointment.Availability
	for _, a := range m.availabilities {
		if a.UserID == userID {
			result = append(result, a)
		}
	}
	return result
}

// newTestUserService 创建测试用的 UserService 实例。
func newTestUserService(ur *mockUserRepo, ar *mockApptRepo) *UserService {
	return &UserService{userRepo: ur, apptRepo: ar}
}

// =============================================================================
// Register
// =============================================================================

func TestUserService_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		resp, err := svc.Register(user.RegisterRequest{
			Email:     "alice@school.edu",
			Password:  "123456",
			Nickname:  "Alice",
			StudentID: "S001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Email != "alice@school.edu" {
			t.Errorf("Email = %q", resp.Email)
		}
		if resp.EmailVerified {
			t.Error("new user should not be verified")
		}
		// 验证 VerifyToken 被设置
		u, _ := ur.FindByEmail("alice@school.edu")
		if u.VerifyToken == "" {
			t.Error("VerifyToken should be set")
		}
	})

	t.Run("non-edu email rejected", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		_, err := svc.Register(user.RegisterRequest{
			Email:    "alice@gmail.com",
			Password: "123456",
			Nickname: "Alice",
		})
		if err != user.ErrInvalidEmail {
			t.Errorf("expected ErrInvalidEmail, got %v", err)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		// 先注册一个
		_, _ = svc.Register(user.RegisterRequest{
			Email:     "bob@school.edu",
			Password:  "123456",
			Nickname:  "Bob",
			StudentID: "S002",
		})

		// 再次注册同一邮箱
		_, err := svc.Register(user.RegisterRequest{
			Email:     "bob@school.edu",
			Password:  "123456",
			Nickname:  "Bob2",
			StudentID: "S003",
		})
		if err != user.ErrEmailAlreadyExists {
			t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
		}
	})
}

// =============================================================================
// VerifyEmail
// =============================================================================

func TestUserService_VerifyEmail(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		// 先注册
		svc.Register(user.RegisterRequest{
			Email:     "alice@school.edu",
			Password:  "123456",
			Nickname:  "Alice",
			StudentID: "S001",
		})

		// 获取 token
		u, _ := ur.FindByEmail("alice@school.edu")
		token := u.VerifyToken

		err := svc.VerifyEmail(token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		u, _ = ur.FindByEmail("alice@school.edu")
		if !u.EmailVerified {
			t.Error("expected EmailVerified to be true")
		}
		if u.VerifyToken != "" {
			t.Error("expected VerifyToken to be cleared")
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		err := svc.VerifyEmail("bad-token")
		if err != user.ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})
}

// =============================================================================
// Login
// =============================================================================

func TestUserService_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		// 注册并验证
		resp, _ := svc.Register(user.RegisterRequest{
			Email:     "alice@school.edu",
			Password:  "password123",
			Nickname:  "Alice",
			StudentID: "S001",
		})
		// 需要手动设密码 hash 因为我们 mock 不会存真实 bcrypt
		// 实际上 Register 已经调用了 bcrypt，mock Create 会存进去
		// 但要验证邮箱
		u, _ := ur.FindByEmail("alice@school.edu")
		u.MarkVerified()
		ur.Update(u)
		_ = resp

		authResp, err := svc.Login(user.LoginRequest{
			Email:    "alice@school.edu",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if authResp.Token == "" {
			t.Error("JWT token should not be empty")
		}
		if authResp.User.Email != "alice@school.edu" {
			t.Errorf("User.Email = %q", authResp.User.Email)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		_, err := svc.Login(user.LoginRequest{
			Email:    "nobody@school.edu",
			Password: "x",
		})
		if err != user.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		svc.Register(user.RegisterRequest{
			Email:     "alice@school.edu",
			Password:  "correct",
			Nickname:  "Alice",
			StudentID: "S001",
		})
		// 验证邮箱
		u, _ := ur.FindByEmail("alice@school.edu")
		u.MarkVerified()
		ur.Update(u)

		_, err := svc.Login(user.LoginRequest{
			Email:    "alice@school.edu",
			Password: "wrongpassword",
		})
		if err != user.ErrWrongPassword {
			t.Errorf("expected ErrWrongPassword, got %v", err)
		}
	})

	t.Run("email not verified", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		svc.Register(user.RegisterRequest{
			Email:     "alice@school.edu",
			Password:  "password123",
			Nickname:  "Alice",
			StudentID: "S001",
		})
		// 不验证邮箱

		_, err := svc.Login(user.LoginRequest{
			Email:    "alice@school.edu",
			Password: "password123",
		})
		if err != user.ErrEmailNotVerified {
			t.Errorf("expected ErrEmailNotVerified, got %v", err)
		}
	})
}

// =============================================================================
// GetMe
// =============================================================================

func TestUserService_GetMe(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		svc.Register(user.RegisterRequest{
			Email:     "alice@school.edu",
			Password:  "123456",
			Nickname:  "Alice",
			StudentID: "S001",
		})

		resp, err := svc.GetMe("alice@school.edu")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Email != "alice@school.edu" {
			t.Errorf("Email = %q", resp.Email)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		_, err := svc.GetMe("nobody@school.edu")
		if err != user.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

// =============================================================================
// UpdateProfile
// =============================================================================

func TestUserService_UpdateProfile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		svc.Register(user.RegisterRequest{
			Email:     "alice@school.edu",
			Password:  "123456",
			Nickname:  "Alice",
			StudentID: "S001",
		})

		resp, err := svc.UpdateProfile("alice@school.edu", user.UpdateProfileRequest{
			Nickname:  "Alice Updated",
			StudentID: "S002",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Nickname != "Alice Updated" {
			t.Errorf("Nickname = %q", resp.Nickname)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		_, err := svc.UpdateProfile("nobody@school.edu", user.UpdateProfileRequest{})
		if err != user.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

// =============================================================================
// GetAllUsers
// =============================================================================

func TestUserService_GetAllUsers(t *testing.T) {
	ur := newMockUserRepo()
	ar := newMockApptRepo()
	svc := newTestUserService(ur, ar)

	// 注册两个用户
	_, _ = svc.Register(user.RegisterRequest{
		Email:     "alice@school.edu",
		Password:  "123456",
		Nickname:  "Alice",
		StudentID: "S001",
	})
	_, _ = svc.Register(user.RegisterRequest{
		Email:     "bob@school.edu",
		Password:  "123456",
		Nickname:  "Bob",
		StudentID: "S002",
	})

	// 都未验证，列表应为空
	users := svc.GetAllUsers()
	if len(users) != 0 {
		t.Errorf("expected 0 verified users, got %d", len(users))
	}

	// 验证 Alice
	u, _ := ur.FindByEmail("alice@school.edu")
	u.MarkVerified()
	ur.Update(u)

	users = svc.GetAllUsers()
	if len(users) != 1 {
		t.Errorf("expected 1 verified user, got %d", len(users))
	}
	if users[0].Email != "alice@school.edu" {
		t.Errorf("Email = %q", users[0].Email)
	}
}

// =============================================================================
// GetUserDetail
// =============================================================================

func TestUserService_GetUserDetail(t *testing.T) {
	t.Run("success without contact (no accepted appointment)", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		svc.Register(user.RegisterRequest{
			Email:     "alice@school.edu",
			Password:  "123456",
			Nickname:  "Alice",
			StudentID: "S001",
		})
		// 设置联系方式
		u, _ := ur.FindByEmail("alice@school.edu")
		u.ContactInfo = "alice@private.com"
		ur.Update(u)

		detail, err := svc.GetUserDetail("alice@school.edu", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if detail.User.ContactInfo != "" {
			t.Errorf("contact should be hidden when no accepted appointment, got %q", detail.User.ContactInfo)
		}
	})

	t.Run("with contact (has accepted appointment)", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		svc.Register(user.RegisterRequest{
			Email:     "mentor@school.edu",
			Password:  "123456",
			Nickname:  "Mentor",
			StudentID: "M001",
		})
		svc.Register(user.RegisterRequest{
			Email:     "student@school.edu",
			Password:  "123456",
			Nickname:  "Student",
			StudentID: "S001",
		})

		u, _ := ur.FindByEmail("mentor@school.edu")
		u.ContactInfo = "mentor@private.com"
		ur.Update(u)

		// 创建一个 accepted 的预约
		ar.appointments["appt-1"] = &appointment.Appointment{
			ID:        "appt-1",
			MentorID:  "mentor@school.edu",
			StudentID: "student@school.edu",
			Status:    appointment.StatusAccepted,
		}

		detail, err := svc.GetUserDetail("mentor@school.edu", "student@school.edu")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if detail.User.ContactInfo != "mentor@private.com" {
			t.Errorf("contact should be visible, got %q", detail.User.ContactInfo)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		_, err := svc.GetUserDetail("nobody@school.edu", "")
		if err != user.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

// =============================================================================
// hasAcceptedAppointment
// =============================================================================

func TestUserService_hasAcceptedAppointment(t *testing.T) {
	ur := newMockUserRepo()
	ar := newMockApptRepo()
	svc := newTestUserService(ur, ar)

	t.Run("false when no appointments", func(t *testing.T) {
		if svc.hasAcceptedAppointment("a@school.edu", "b@school.edu") {
			t.Error("should be false with no appointments")
		}
	})

	t.Run("true when student has accepted with mentor", func(t *testing.T) {
		ar.appointments["appt-1"] = &appointment.Appointment{
			MentorID:  "mentor@school.edu",
			StudentID: "student@school.edu",
			Status:    appointment.StatusAccepted,
		}
		if !svc.hasAcceptedAppointment("student@school.edu", "mentor@school.edu") {
			t.Error("should be true")
		}
	})

	t.Run("false when pending", func(t *testing.T) {
		ar.appointments["appt-2"] = &appointment.Appointment{
			MentorID:  "mentor@school.edu",
			StudentID: "student@school.edu",
			Status:    appointment.StatusPending,
		}
		// 清掉之前的 accepted
		delete(ar.appointments, "appt-1")
		if svc.hasAcceptedAppointment("student@school.edu", "mentor@school.edu") {
			t.Error("should be false for pending")
		}
	})

	t.Run("true when mentor has accepted from student", func(t *testing.T) {
		ar.appointments["appt-3"] = &appointment.Appointment{
			MentorID:  "mentor@school.edu",
			StudentID: "student2@school.edu",
			Status:    appointment.StatusAccepted,
		}
		if !svc.hasAcceptedAppointment("mentor@school.edu", "student2@school.edu") {
			t.Error("should be true when mentor has accepted student")
		}
	})
}

// =============================================================================
// 确保 mock 编译时满足接口
// =============================================================================

var (
	_ user.UserRepository               = (*mockUserRepo)(nil)
	_ appointment.AppointmentRepository = (*mockApptRepo)(nil)
)


