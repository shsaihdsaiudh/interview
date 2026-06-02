package persistence

import (
	"context"
	"testing"
	"time"

	"interview-server/domain/appointment"
	"interview-server/domain/user"
	"interview-server/testutil"
)

// =============================================================================
// 集成测试夹具 — 包初始化时启动一次 PG 容器，包内所有测试复用
// =============================================================================

var testPool = func() *testutil.PostgresContainer {
	ctx := context.Background()
	c, err := testutil.StartPostgres(ctx)
	if err != nil {
		panic("启动测试 PostgreSQL 失败: " + err.Error())
	}
	return c
}()

// newUser 创建带默认值的测试用户，避免 Tags 为 nil 违反 NOT NULL 约束。
func newUser(email, nickname string) *user.User {
	return &user.User{
		Email:     email,
		PasswordHash: "test-hash",
		Nickname:  nickname,
		Tags:      []string{},
		CreatedAt: time.Now(),
	}
}

// mustCreateUser 创建用户，失败则 panic（用于测试 setup）。
func mustCreateUser(t *testing.T, repo *PostgresRepo, u *user.User) {
	t.Helper()
	if err := repo.Create(u); err != nil {
		t.Fatalf("Create user %s failed: %v", u.Email, err)
	}
}

// mustCreateAvailability 创建空闲时间，失败则 panic。
func mustCreateAvailability(t *testing.T, repo *PostgresRepo, a *appointment.Availability) {
	t.Helper()
	if err := repo.CreateAvailability(a); err != nil {
		t.Fatalf("CreateAvailability %s failed: %v", a.ID, err)
	}
}

// mustCreateAppointment 创建预约，失败则 panic。
func mustCreateAppointment(t *testing.T, repo *PostgresRepo, a *appointment.Appointment) {
	t.Helper()
	if err := repo.CreateAppointment(a); err != nil {
		t.Fatalf("CreateAppointment %s failed: %v", a.ID, err)
	}
}

// =============================================================================
// User 持久化测试
// =============================================================================

func TestPostgresRepo_CreateAndFindUser(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	u := &user.User{
		Email:        "alice@school.edu",
		PasswordHash: "hash123",
		Nickname:     "Alice",
		StudentID:    "S001",
		Department:   "CS",
		Tags:         []string{"go", "rust"},
		Avatar:       "a.png",
		ContactInfo:  "a@private.com",
		CreatedAt:    time.Now().Truncate(time.Second),
	}
	mustCreateUser(t, repo, u)

	found, err := repo.FindByEmail("alice@school.edu")
	if err != nil {
		t.Fatalf("FindByEmail failed: %v", err)
	}
	if found.Nickname != "Alice" {
		t.Errorf("Nickname = %q", found.Nickname)
	}
	if len(found.Tags) != 2 || found.Tags[0] != "go" {
		t.Errorf("Tags = %v", found.Tags)
	}
}

func TestPostgresRepo_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	mustCreateUser(t, repo, newUser("bob@school.edu", "Bob"))
	err := repo.Create(newUser("bob@school.edu", "Bob2"))
	if err != user.ErrEmailAlreadyExists {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestPostgresRepo_FindByEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	_, err := repo.FindByEmail("nobody@school.edu")
	if err != user.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestPostgresRepo_UpdateUser(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	u := newUser("charlie@school.edu", "Charlie")
	mustCreateUser(t, repo, u)

	u.Nickname = "Charlie Updated"
	u.Department = "Math"
	if err := repo.Update(u); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.FindByEmail("charlie@school.edu")
	if found.Nickname != "Charlie Updated" {
		t.Errorf("Nickname = %q after update", found.Nickname)
	}
	if found.Department != "Math" {
		t.Errorf("Department = %q after update", found.Department)
	}
}

func TestPostgresRepo_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	err := repo.Update(newUser("nonexistent@school.edu", "X"))
	if err != user.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestPostgresRepo_FindAll(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	u1 := newUser("e1@school.edu", "E1")
	u1.EmailVerified = true
	u2 := newUser("e2@school.edu", "E2")
	u2.EmailVerified = true
	u3 := newUser("e3@school.edu", "E3") // not verified

	mustCreateUser(t, repo, u1)
	mustCreateUser(t, repo, u2)
	mustCreateUser(t, repo, u3)

	all, total, err := repo.FindAll(1, 20)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 verified users, got %d (total=%d)", len(all), total)
	}
}

// =============================================================================
// Availability 持久化测试
// =============================================================================

func TestPostgresRepo_CreateAndFindAvailability(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	mustCreateUser(t, repo, newUser("mentor@school.edu", "Mentor"))

	a := &appointment.Availability{
		ID: "avail-1", UserID: "mentor@school.edu",
		Date: "2099-07-01", StartTime: "14:00", EndTime: "15:00",
	}
	mustCreateAvailability(t, repo, a)

	found, err := repo.FindAvailabilityByID("avail-1")
	if err != nil {
		t.Fatalf("FindAvailabilityByID failed: %v", err)
	}
	if found.UserID != "mentor@school.edu" {
		t.Errorf("UserID = %q", found.UserID)
	}
}

func TestPostgresRepo_FindAvailabilityByID_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	_, err := repo.FindAvailabilityByID("nonexistent")
	if err != appointment.ErrSlotNotFound {
		t.Errorf("expected ErrSlotNotFound, got %v", err)
	}
}

func TestPostgresRepo_FindAvailabilitiesByUserID(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	mustCreateUser(t, repo, newUser("m1@school.edu", "M1"))
	mustCreateAvailability(t, repo, &appointment.Availability{ID: "a1", UserID: "m1@school.edu", Date: "2099-07-01", StartTime: "10:00", EndTime: "11:00"})
	mustCreateAvailability(t, repo, &appointment.Availability{ID: "a2", UserID: "m1@school.edu", Date: "2099-07-02", StartTime: "14:00", EndTime: "15:00"})

	slots := repo.FindAvailabilitiesByUserID("m1@school.edu")
	if len(slots) != 2 {
		t.Errorf("expected 2 slots, got %d", len(slots))
	}
}

func TestPostgresRepo_DeleteAvailability(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	mustCreateUser(t, repo, newUser("m2@school.edu", "M2"))
	mustCreateAvailability(t, repo, &appointment.Availability{ID: "to-delete", UserID: "m2@school.edu", Date: "2099-07-01", StartTime: "10:00", EndTime: "11:00"})

	if err := repo.DeleteAvailability("to-delete"); err != nil {
		t.Fatalf("DeleteAvailability failed: %v", err)
	}

	_, err := repo.FindAvailabilityByID("to-delete")
	if err != appointment.ErrSlotNotFound {
		t.Errorf("expected ErrSlotNotFound after delete, got %v", err)
	}
}

func TestPostgresRepo_DeleteAvailability_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	err := repo.DeleteAvailability("nonexistent")
	if err != appointment.ErrSlotNotFound {
		t.Errorf("expected ErrSlotNotFound, got %v", err)
	}
}

func TestPostgresRepo_TimeRangeConstraint(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	mustCreateUser(t, repo, newUser("m3@school.edu", "M3"))

	err := repo.CreateAvailability(&appointment.Availability{
		ID: "bad-range", UserID: "m3@school.edu",
		Date: "2099-07-01", StartTime: "15:00", EndTime: "14:00",
	})
	if err == nil {
		t.Error("expected error for time range constraint violation")
	}
}

// =============================================================================
// Appointment 持久化测试
// =============================================================================

func TestPostgresRepo_CreateAndFindAppointment(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	mustCreateUser(t, repo, newUser("mentor2@school.edu", "Mentor2"))
	mustCreateUser(t, repo, newUser("student2@school.edu", "Student2"))
	mustCreateAvailability(t, repo, &appointment.Availability{ID: "slot-2", UserID: "mentor2@school.edu", Date: "2099-07-01", StartTime: "14:00", EndTime: "15:00"})

	appt := &appointment.Appointment{
		ID: "appt-2", MentorID: "mentor2@school.edu", StudentID: "student2@school.edu",
		TimeSlotID: "slot-2", Message: "请教问题", Status: appointment.StatusPending,
		CreatedAt: time.Now(),
	}
	mustCreateAppointment(t, repo, appt)

	found, err := repo.FindAppointmentByID("appt-2")
	if err != nil {
		t.Fatalf("FindAppointmentByID failed: %v", err)
	}
	if found.Status != appointment.StatusPending {
		t.Errorf("Status = %q", found.Status)
	}
}

func TestPostgresRepo_FindAppointmentByID_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	_, err := repo.FindAppointmentByID("nonexistent")
	if err != appointment.ErrAppointmentNotFound {
		t.Errorf("expected ErrAppointmentNotFound, got %v", err)
	}
}

func TestPostgresRepo_UpdateAppointment(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	mustCreateUser(t, repo, newUser("m4@school.edu", "M4"))
	mustCreateUser(t, repo, newUser("s4@school.edu", "S4"))
	mustCreateAvailability(t, repo, &appointment.Availability{ID: "s4", UserID: "m4@school.edu", Date: "2099-07-01", StartTime: "14:00", EndTime: "15:00"})

	appt := &appointment.Appointment{
		ID: "appt-upd", MentorID: "m4@school.edu", StudentID: "s4@school.edu",
		TimeSlotID: "s4", Status: appointment.StatusPending, CreatedAt: time.Now(),
	}
	mustCreateAppointment(t, repo, appt)

	appt.Status = appointment.StatusAccepted
	if err := repo.UpdateAppointment(appt); err != nil {
		t.Fatalf("UpdateAppointment failed: %v", err)
	}

	found, _ := repo.FindAppointmentByID("appt-upd")
	if found.Status != appointment.StatusAccepted {
		t.Errorf("Status = %q after update", found.Status)
	}
}

func TestPostgresRepo_FindAppointmentsByMentorID(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	mustCreateUser(t, repo, newUser("m5@school.edu", "M5"))
	mustCreateUser(t, repo, newUser("s5a@school.edu", "S5A"))
	mustCreateUser(t, repo, newUser("s5b@school.edu", "S5B"))
	mustCreateAvailability(t, repo, &appointment.Availability{ID: "s5a", UserID: "m5@school.edu", Date: "2099-07-01", StartTime: "10:00", EndTime: "11:00"})
	mustCreateAvailability(t, repo, &appointment.Availability{ID: "s5b", UserID: "m5@school.edu", Date: "2099-07-01", StartTime: "14:00", EndTime: "15:00"})

	mustCreateAppointment(t, repo, &appointment.Appointment{ID: "a5a", MentorID: "m5@school.edu", StudentID: "s5a@school.edu", TimeSlotID: "s5a", Status: appointment.StatusPending, CreatedAt: time.Now()})
	mustCreateAppointment(t, repo, &appointment.Appointment{ID: "a5b", MentorID: "m5@school.edu", StudentID: "s5b@school.edu", TimeSlotID: "s5b", Status: appointment.StatusAccepted, CreatedAt: time.Now()})

	if len(repo.FindAppointmentsByMentorID("m5@school.edu")) != 2 {
		t.Error("expected 2 appointments for mentor")
	}
}

func TestPostgresRepo_FindAppointmentsByStudentID(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	mustCreateUser(t, repo, newUser("m6@school.edu", "M6"))
	mustCreateUser(t, repo, newUser("s6@school.edu", "S6"))
	mustCreateAvailability(t, repo, &appointment.Availability{ID: "s6", UserID: "m6@school.edu", Date: "2099-07-01", StartTime: "14:00", EndTime: "15:00"})
	mustCreateAppointment(t, repo, &appointment.Appointment{ID: "a6", MentorID: "m6@school.edu", StudentID: "s6@school.edu", TimeSlotID: "s6", Status: appointment.StatusPending, CreatedAt: time.Now()})

	if len(repo.FindAppointmentsByStudentID("s6@school.edu")) != 1 {
		t.Error("expected 1 appointment for student")
	}
}

func TestPostgresRepo_HasActiveAppointment(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	mustCreateUser(t, repo, newUser("m7@school.edu", "M7"))
	mustCreateUser(t, repo, newUser("s7@school.edu", "S7"))
	mustCreateAvailability(t, repo, &appointment.Availability{ID: "s7p", UserID: "m7@school.edu", Date: "2099-07-01", StartTime: "14:00", EndTime: "15:00"})
	mustCreateAvailability(t, repo, &appointment.Availability{ID: "s7a", UserID: "m7@school.edu", Date: "2099-07-02", StartTime: "10:00", EndTime: "11:00"})
	mustCreateAvailability(t, repo, &appointment.Availability{ID: "s7r", UserID: "m7@school.edu", Date: "2099-07-03", StartTime: "10:00", EndTime: "11:00"})

	mustCreateAppointment(t, repo, &appointment.Appointment{ID: "a7p", MentorID: "m7@school.edu", StudentID: "s7@school.edu", TimeSlotID: "s7p", Status: appointment.StatusPending, CreatedAt: time.Now()})
	mustCreateAppointment(t, repo, &appointment.Appointment{ID: "a7a", MentorID: "m7@school.edu", StudentID: "s7@school.edu", TimeSlotID: "s7a", Status: appointment.StatusAccepted, CreatedAt: time.Now()})
	mustCreateAppointment(t, repo, &appointment.Appointment{ID: "a7r", MentorID: "m7@school.edu", StudentID: "s7@school.edu", TimeSlotID: "s7r", Status: appointment.StatusRejected, CreatedAt: time.Now()})

	if !repo.HasActiveAppointment("s7p") {
		t.Error("pending should be active")
	}
	if !repo.HasActiveAppointment("s7a") {
		t.Error("accepted should be active")
	}
	if repo.HasActiveAppointment("s7r") {
		t.Error("rejected should NOT be active")
	}
}

func TestPostgresRepo_Appointment_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	err := repo.UpdateAppointment(&appointment.Appointment{ID: "nonexistent"})
	if err != appointment.ErrAppointmentNotFound {
		t.Errorf("expected ErrAppointmentNotFound, got %v", err)
	}
}

func TestPostgresRepo_ForeignKeyConstraint(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	err := repo.CreateAppointment(&appointment.Appointment{
		ID: "fk-fail", MentorID: "no-such@school.edu", StudentID: "no-such@school.edu",
		TimeSlotID: "no-such", Status: appointment.StatusPending, CreatedAt: time.Now(),
	})
	if err == nil {
		t.Error("expected foreign key constraint error")
	}
}

func TestPostgresRepo_CascadeDelete(t *testing.T) {
	ctx := context.Background()
	pool := testPool.NewPool(ctx)
	defer pool.Close()
	repo := NewPostgresRepo(pool)

	mustCreateUser(t, repo, newUser("cascade@school.edu", "Cascade"))
	mustCreateAvailability(t, repo, &appointment.Availability{ID: "c-avail", UserID: "cascade@school.edu", Date: "2099-07-01", StartTime: "10:00", EndTime: "11:00"})

	_, err := pool.Exec(ctx, `DELETE FROM users WHERE email = $1`, "cascade@school.edu")
	if err != nil {
		t.Fatalf("delete user failed: %v", err)
	}

	_, err = repo.FindAvailabilityByID("c-avail")
	if err != appointment.ErrSlotNotFound {
		t.Errorf("expected cascade delete, got %v", err)
	}
}
