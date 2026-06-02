package application

import (
	"errors"
	"testing"
	"time"

	"interview-server/domain/appointment"
	"interview-server/domain/user"
)

// =============================================================================
// Mock 实现
// =============================================================================

// mockUserRepo 实现 user.UserRepository，用于单元测试。
type mockUserRepo struct {
	usersByEmail   map[string]*user.User
	usersByToken   map[string]*user.User
	createErr      error
	findByEmailErr error
	findByTokenErr error
	updateErr      error
	deleteErr      error
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

func (m *mockUserRepo) FindAll(page, pageSize int) ([]*user.User, int, error) {
	var all []*user.User
	for _, u := range m.usersByEmail {
		if u.EmailVerified {
			all = append(all, u)
		}
	}
	total := len(all)

	// 简单排序：按创建时间降序（保持与真实实现一致的语义）
	// mock 中用户可能没有 CreatedAt，这里做简单切片实现即可

	// 分页
	start := (page - 1) * pageSize
	if start >= total {
		return []*user.User{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func (m *mockUserRepo) Delete(email string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.usersByEmail[email]; !ok {
		return user.ErrUserNotFound
	}
	delete(m.usersByEmail, email)
	return nil
}

// mockApptRepo 实现 appointment.AppointmentRepository，用于单元测试。
type mockApptRepo struct {
	appointments   map[string]*appointment.Appointment
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
	return &UserService{
		userRepo:          ur,
		apptRepo:          ar,
		verificationCodes: make(map[string]verificationCode),
	}
}

// registerHelper 完整注册流程：发送验证码 → 注册。
func registerHelper(svc *UserService, email, password, nickname, studentID string) (*user.AuthResponse, error) {
	if err := svc.SendCode(email); err != nil {
		return nil, err
	}
	code := svc.VerificationCodeForTest(email)
	return svc.Register(user.RegisterRequest{
		Email:     email,
		Code:      code,
		Password:  password,
		Nickname:  nickname,
		StudentID: studentID,
	})
}

// =============================================================================
// SendCode
// =============================================================================

func TestUserService_SendCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		err := svc.SendCode("alice@std.uestc.edu.cn")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		code := svc.VerificationCodeForTest("alice@std.uestc.edu.cn")
		if len(code) != 6 {
			t.Errorf("expected 6-digit code, got %q", code)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		err := svc.SendCode("alice@gmail.com")
		if err != user.ErrInvalidEmail {
			t.Errorf("expected ErrInvalidEmail, got %v", err)
		}
	})

	t.Run("already registered", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		registerHelper(svc, "alice@std.uestc.edu.cn", "123456", "Alice", "S001")

		err := svc.SendCode("alice@std.uestc.edu.cn")
		if err != user.ErrEmailAlreadyExists {
			t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)
		dbErr := errors.New("database unavailable")
		ur.findByEmailErr = dbErr

		err := svc.SendCode("alice@std.uestc.edu.cn")
		if err != dbErr {
			t.Errorf("expected repository error, got %v", err)
		}
		if code := svc.VerificationCodeForTest("alice@std.uestc.edu.cn"); code != "" {
			t.Errorf("verification code should not be stored on repo error, got %q", code)
		}
	})
}

// =============================================================================
// Register
// =============================================================================

func TestUserService_Register(t *testing.T) {
	t.Run("success with code", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		resp, err := registerHelper(svc, "2024010914026@std.uestc.edu.cn", "123456", "Alice", "S001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Token == "" {
			t.Error("JWT token should not be empty")
		}
		if resp.User.Email != "2024010914026@std.uestc.edu.cn" {
			t.Errorf("Email = %q", resp.User.Email)
		}
		if !resp.User.EmailVerified {
			t.Error("new user should be verified")
		}
		// 验证码一次性使用
		if _, ok := svc.verificationCodes["2024010914026@std.uestc.edu.cn"]; ok {
			t.Error("verification code should be cleared after use")
		}
	})

	t.Run("normalizes email before storing user", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		if err := svc.SendCode(" MixedCase@STD.UESTC.EDU.CN "); err != nil {
			t.Fatalf("unexpected send code error: %v", err)
		}
		code := svc.VerificationCodeForTest("mixedcase@std.uestc.edu.cn")
		resp, err := svc.Register(user.RegisterRequest{
			Email:    " MixedCase@STD.UESTC.EDU.CN ",
			Code:     code,
			Password: "123456",
		})
		if err != nil {
			t.Fatalf("unexpected register error: %v", err)
		}
		if resp.User.Email != "mixedcase@std.uestc.edu.cn" {
			t.Errorf("email = %q, want normalized", resp.User.Email)
		}
	})

	t.Run("register without nickname defaults to email", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		svc.SendCode("minimal@std.uestc.edu.cn")
		code := svc.VerificationCodeForTest("minimal@std.uestc.edu.cn")
		resp, err := svc.Register(user.RegisterRequest{
			Email:    "minimal@std.uestc.edu.cn",
			Code:     code,
			Password: "123456",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.User.Nickname != "minimal@std.uestc.edu.cn" {
			t.Errorf("expected default nickname, got %q", resp.User.Nickname)
		}
	})

	t.Run("non-uestc email rejected", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		_, err := svc.Register(user.RegisterRequest{
			Email:    "alice@gmail.com",
			Code:     "000000",
			Password: "123456",
		})
		if err != user.ErrInvalidEmail {
			t.Errorf("expected ErrInvalidEmail, got %v", err)
		}
	})

	t.Run("wrong verification code", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		svc.SendCode("alice@std.uestc.edu.cn")
		// 用错误的验证码
		_, err := svc.Register(user.RegisterRequest{
			Email:    "alice@std.uestc.edu.cn",
			Code:     "999999",
			Password: "123456",
		})
		if err != user.ErrInvalidCode {
			t.Errorf("expected ErrInvalidCode, got %v", err)
		}
		if svc.VerificationCodeForTest("alice@std.uestc.edu.cn") == "" {
			t.Error("wrong code should not consume the stored verification code")
		}
	})

	t.Run("expired verification code", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		email := "expired@std.uestc.edu.cn"
		if err := svc.SendCode(email); err != nil {
			t.Fatalf("unexpected send code error: %v", err)
		}
		code := svc.VerificationCodeForTest(email)
		svc.verificationCodes[email] = verificationCode{
			Code:      code,
			ExpiresAt: time.Now().Add(-time.Minute),
		}
		_, err := svc.Register(user.RegisterRequest{
			Email:    email,
			Code:     code,
			Password: "123456",
		})
		if err != user.ErrInvalidCode {
			t.Errorf("expected ErrInvalidCode, got %v", err)
		}
		if svc.VerificationCodeForTest(email) != "" {
			t.Error("expired code should be cleared")
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		registerHelper(svc, "bob@std.uestc.edu.cn", "123456", "Bob", "S002")

		err := svc.SendCode("bob@std.uestc.edu.cn")
		if err != user.ErrEmailAlreadyExists {
			t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
		}
	})

	t.Run("repository error while checking duplicate", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)
		dbErr := errors.New("database unavailable")
		email := "repoerr@std.uestc.edu.cn"
		svc.verificationCodes[email] = verificationCode{
			Code:      "123456",
			ExpiresAt: time.Now().Add(time.Minute),
		}
		ur.findByEmailErr = dbErr

		_, err := svc.Register(user.RegisterRequest{
			Email:    email,
			Code:     "123456",
			Password: "123456",
		})
		if err != dbErr {
			t.Errorf("expected repository error, got %v", err)
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

		registerHelper(svc, "alice@std.uestc.edu.cn", "password123", "Alice", "S001")

		authResp, err := svc.Login(user.LoginRequest{
			Email:    "alice@std.uestc.edu.cn",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if authResp.Token == "" {
			t.Error("JWT token should not be empty")
		}
		if authResp.User.Email != "alice@std.uestc.edu.cn" {
			t.Errorf("User.Email = %q", authResp.User.Email)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		_, err := svc.Login(user.LoginRequest{
			Email:    "nobody@std.uestc.edu.cn",
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

		registerHelper(svc, "alice@std.uestc.edu.cn", "correct", "Alice", "S001")

		_, err := svc.Login(user.LoginRequest{
			Email:    "alice@std.uestc.edu.cn",
			Password: "wrongpassword",
		})
		if err != user.ErrWrongPassword {
			t.Errorf("expected ErrWrongPassword, got %v", err)
		}
	})
}

// =============================================================================
// ChangePassword
// =============================================================================

func TestUserService_ChangePassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		registerHelper(svc, "alice@std.uestc.edu.cn", "oldpassword", "Alice", "S001")

		// 修改密码
		err := svc.ChangePassword("alice@std.uestc.edu.cn", user.ChangePasswordRequest{
			OldPassword: "oldpassword",
			NewPassword: "newpassword",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// 用旧密码登录应失败
		_, err = svc.Login(user.LoginRequest{
			Email:    "alice@std.uestc.edu.cn",
			Password: "oldpassword",
		})
		if err != user.ErrWrongPassword {
			t.Errorf("expected ErrWrongPassword for old password, got %v", err)
		}

		// 用新密码登录应成功
		authResp, err := svc.Login(user.LoginRequest{
			Email:    "alice@std.uestc.edu.cn",
			Password: "newpassword",
		})
		if err != nil {
			t.Fatalf("unexpected login error with new password: %v", err)
		}
		if authResp.Token == "" {
			t.Error("JWT token should not be empty")
		}
	})

	t.Run("wrong old password", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		registerHelper(svc, "bob@std.uestc.edu.cn", "correctpassword", "Bob", "S002")

		err := svc.ChangePassword("bob@std.uestc.edu.cn", user.ChangePasswordRequest{
			OldPassword: "wrongpassword",
			NewPassword: "newpassword",
		})
		if err != user.ErrWrongOldPassword {
			t.Errorf("expected ErrWrongOldPassword, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		err := svc.ChangePassword("nobody@std.uestc.edu.cn", user.ChangePasswordRequest{
			OldPassword: "old",
			NewPassword: "new",
		})
		if err != user.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("normalizes email", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		registerHelper(svc, "case@std.uestc.edu.cn", "oldpassword", "Case", "C001")

		err := svc.ChangePassword(" CASE@STD.UESTC.EDU.CN ", user.ChangePasswordRequest{
			OldPassword: "oldpassword",
			NewPassword: "newpassword",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
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

		registerHelper(svc, "alice@std.uestc.edu.cn", "123456", "Alice", "S001")

		resp, err := svc.GetMe("alice@std.uestc.edu.cn")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Email != "alice@std.uestc.edu.cn" {
			t.Errorf("Email = %q", resp.Email)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		_, err := svc.GetMe("nobody@std.uestc.edu.cn")
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

		registerHelper(svc, "alice@std.uestc.edu.cn", "123456", "Alice", "S001")

		resp, err := svc.UpdateProfile("alice@std.uestc.edu.cn", user.UpdateProfileRequest{
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

		_, err := svc.UpdateProfile("nobody@std.uestc.edu.cn", user.UpdateProfileRequest{})
		if err != user.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

// =============================================================================
// GetProfile
// =============================================================================

func TestUserService_GetProfile(t *testing.T) {
	ur := newMockUserRepo()
	ar := newMockApptRepo()
	svc := newTestUserService(ur, ar)

	registerHelper(svc, "profile@std.uestc.edu.cn", "123456", "Profile", "P001")
	u, _ := ur.FindByEmail("profile@std.uestc.edu.cn")
	u.ContactInfo = "wechat-profile"
	ur.Update(u)
	ar.availabilities["slot-1"] = &appointment.Availability{
		ID:        "slot-1",
		UserID:    "profile@std.uestc.edu.cn",
		Date:      "2099-12-31",
		StartTime: "09:00",
		EndTime:   "10:00",
	}

	detail, err := svc.GetProfile("profile@std.uestc.edu.cn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.User.ContactInfo != "wechat-profile" {
		t.Errorf("contact_info = %q, want own contact", detail.User.ContactInfo)
	}
	if len(detail.Availabilities) != 1 {
		t.Errorf("availabilities len = %d, want 1", len(detail.Availabilities))
	}
}

// =============================================================================
// GetAllUsers
// =============================================================================

func TestUserService_GetAllUsers(t *testing.T) {
	ur := newMockUserRepo()
	ar := newMockApptRepo()
	svc := newTestUserService(ur, ar)

	// 注册两个用户（新流程自动验证）
	registerHelper(svc, "alice@std.uestc.edu.cn", "123456", "Alice", "S001")
	registerHelper(svc, "bob@std.uestc.edu.cn", "123456", "Bob", "S002")

	// 都已验证，列表应有 2 个
	resp, err := svc.GetAllUsers(1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("expected total 2, got %d", resp.Total)
	}
	if len(resp.Users) != 2 {
		t.Errorf("expected 2 users, got %d", len(resp.Users))
	}
	if resp.Page != 1 {
		t.Errorf("expected page 1, got %d", resp.Page)
	}
	if resp.PageSize != 20 {
		t.Errorf("expected page_size 20, got %d", resp.PageSize)
	}
}

func TestUserService_GetAllUsers_Pagination(t *testing.T) {
	ur := newMockUserRepo()
	ar := newMockApptRepo()
	svc := newTestUserService(ur, ar)

	// 注册 5 个用户
	for i := 1; i <= 5; i++ {
		email := "user" + string(rune('0'+i)) + "@std.uestc.edu.cn"
		registerHelper(svc, email, "123456", "User"+string(rune('0'+i)), "S00"+string(rune('0'+i)))
	}

	t.Run("first page with page_size=2", func(t *testing.T) {
		resp, err := svc.GetAllUsers(1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Total != 5 {
			t.Errorf("expected total 5, got %d", resp.Total)
		}
		if len(resp.Users) != 2 {
			t.Errorf("expected 2 users on page 1, got %d", len(resp.Users))
		}
	})

	t.Run("last page with page_size=2", func(t *testing.T) {
		resp, err := svc.GetAllUsers(3, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Total != 5 {
			t.Errorf("expected total 5, got %d", resp.Total)
		}
		if len(resp.Users) != 1 {
			t.Errorf("expected 1 user on page 3, got %d", len(resp.Users))
		}
	})

	t.Run("empty page beyond total", func(t *testing.T) {
		resp, err := svc.GetAllUsers(10, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Users) != 0 {
			t.Errorf("expected 0 users on page 10, got %d", len(resp.Users))
		}
	})
}

// =============================================================================
// GetUserDetail
// =============================================================================

func TestUserService_GetUserDetail(t *testing.T) {
	t.Run("success without contact (no accepted appointment)", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		registerHelper(svc, "alice@std.uestc.edu.cn", "123456", "Alice", "S001")
		// 设置联系方式
		u, _ := ur.FindByEmail("alice@std.uestc.edu.cn")
		u.ContactInfo = "alice@private.com"
		ur.Update(u)

		detail, err := svc.GetUserDetail("alice@std.uestc.edu.cn", "")
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

		registerHelper(svc, "mentor@std.uestc.edu.cn", "123456", "Mentor", "M001")
		registerHelper(svc, "student@std.uestc.edu.cn", "123456", "Student", "S001")

		u, _ := ur.FindByEmail("mentor@std.uestc.edu.cn")
		u.ContactInfo = "mentor@private.com"
		ur.Update(u)

		// 创建一个 accepted 的预约
		ar.appointments["appt-1"] = &appointment.Appointment{
			ID:        "appt-1",
			MentorID:  "mentor@std.uestc.edu.cn",
			StudentID: "student@std.uestc.edu.cn",
			Status:    appointment.StatusAccepted,
		}

		detail, err := svc.GetUserDetail("mentor@std.uestc.edu.cn", "student@std.uestc.edu.cn")
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

		_, err := svc.GetUserDetail("nobody@std.uestc.edu.cn", "")
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
		if svc.hasAcceptedAppointment("a@std.uestc.edu.cn", "b@std.uestc.edu.cn") {
			t.Error("should be false with no appointments")
		}
	})

	t.Run("true when student has accepted with mentor", func(t *testing.T) {
		ar.appointments["appt-1"] = &appointment.Appointment{
			MentorID:  "mentor@std.uestc.edu.cn",
			StudentID: "student@std.uestc.edu.cn",
			Status:    appointment.StatusAccepted,
		}
		if !svc.hasAcceptedAppointment("student@std.uestc.edu.cn", "mentor@std.uestc.edu.cn") {
			t.Error("should be true")
		}
	})

	t.Run("false when pending", func(t *testing.T) {
		ar.appointments["appt-2"] = &appointment.Appointment{
			MentorID:  "mentor@std.uestc.edu.cn",
			StudentID: "student@std.uestc.edu.cn",
			Status:    appointment.StatusPending,
		}
		// 清掉之前的 accepted
		delete(ar.appointments, "appt-1")
		if svc.hasAcceptedAppointment("student@std.uestc.edu.cn", "mentor@std.uestc.edu.cn") {
			t.Error("should be false for pending")
		}
	})

	t.Run("true when mentor has accepted from student", func(t *testing.T) {
		ar.appointments["appt-3"] = &appointment.Appointment{
			MentorID:  "mentor@std.uestc.edu.cn",
			StudentID: "student2@std.uestc.edu.cn",
			Status:    appointment.StatusAccepted,
		}
		if !svc.hasAcceptedAppointment("mentor@std.uestc.edu.cn", "student2@std.uestc.edu.cn") {
			t.Error("should be true when mentor has accepted student")
		}
	})
}

// =============================================================================
// DeleteAccount
// =============================================================================

func TestUserService_DeleteAccount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		registerHelper(svc, "alice@std.uestc.edu.cn", "password123", "Alice", "S001")

		err := svc.DeleteAccount("alice@std.uestc.edu.cn", "password123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// 注销后用户不应存在
		_, err = ur.FindByEmail("alice@std.uestc.edu.cn")
		if err != user.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound after delete, got %v", err)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		registerHelper(svc, "bob@std.uestc.edu.cn", "password123", "Bob", "S002")

		err := svc.DeleteAccount("bob@std.uestc.edu.cn", "wrongpassword")
		if err != user.ErrWrongPassword {
			t.Errorf("expected ErrWrongPassword, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		err := svc.DeleteAccount("nobody@std.uestc.edu.cn", "password")
		if err != user.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("normalizes email", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		registerHelper(svc, "case@std.uestc.edu.cn", "password123", "Case", "C001")

		err := svc.DeleteAccount(" CASE@STD.UESTC.EDU.CN ", "password123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("has active appointment", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		registerHelper(svc, "mentor@std.uestc.edu.cn", "password123", "Mentor", "M001")

		// 创建 pending 预约
		ar.appointments["appt-1"] = &appointment.Appointment{
			MentorID:  "mentor@std.uestc.edu.cn",
			StudentID: "student@std.uestc.edu.cn",
			Status:    appointment.StatusPending,
		}

		err := svc.DeleteAccount("mentor@std.uestc.edu.cn", "password123")
		if err != user.ErrCannotDeleteWithActiveAppointments {
			t.Errorf("expected ErrCannotDeleteWithActiveAppointments, got %v", err)
		}
	})

	t.Run("has accepted appointment", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		registerHelper(svc, "mentor2@std.uestc.edu.cn", "password123", "Mentor2", "M002")

		ar.appointments["appt-2"] = &appointment.Appointment{
			MentorID:  "mentor2@std.uestc.edu.cn",
			StudentID: "student2@std.uestc.edu.cn",
			Status:    appointment.StatusAccepted,
		}

		err := svc.DeleteAccount("mentor2@std.uestc.edu.cn", "password123")
		if err != user.ErrCannotDeleteWithActiveAppointments {
			t.Errorf("expected ErrCannotDeleteWithActiveAppointments, got %v", err)
		}
	})

	t.Run("only rejected appointments allows delete", func(t *testing.T) {
		ur := newMockUserRepo()
		ar := newMockApptRepo()
		svc := newTestUserService(ur, ar)

		registerHelper(svc, "mentor3@std.uestc.edu.cn", "password123", "Mentor3", "M003")

		// 创建 rejected 预约
		ar.appointments["appt-3"] = &appointment.Appointment{
			MentorID:  "mentor3@std.uestc.edu.cn",
			StudentID: "student3@std.uestc.edu.cn",
			Status:    appointment.StatusRejected,
		}

		err := svc.DeleteAccount("mentor3@std.uestc.edu.cn", "password123")
		if err != nil {
			t.Errorf("expected no error for rejected-only appointments, got %v", err)
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
