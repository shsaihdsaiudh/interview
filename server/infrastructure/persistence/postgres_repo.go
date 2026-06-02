// Package persistence 提供 PostgreSQL 实现，满足 domain 层定义的 Repository 接口。
// 这是依赖反转的关键：domain 定义接口，infrastructure 提供实现。
package persistence

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"interview-server/domain/appointment"
	"interview-server/domain/user"
)

// PostgresRepo 同时实现 user.UserRepository 和 appointment.AppointmentRepository。
// 使用组合方式满足编译时接口检查。
type PostgresRepo struct {
	pool *pgxpool.Pool
}

// 编译时接口检查：确保 PostgresRepo 实现了两个领域接口
var (
	_ user.UserRepository                 = (*PostgresRepo)(nil)
	_ appointment.AppointmentRepository   = (*PostgresRepo)(nil)
)

// NewPostgresRepo 创建 PostgreSQL 仓库。
func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{pool: pool}
}

// =============================================================================
// user.UserRepository 接口实现
// =============================================================================

// Create 创建新用户。
func (r *PostgresRepo) Create(u *user.User) error {
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO users (email, password_hash, nickname, student_id, department,
		 tags, avatar, contact_info, email_verified, verify_token, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		u.Email, u.PasswordHash, u.Nickname, u.StudentID,
		u.Department, u.Tags, u.Avatar, u.ContactInfo,
		u.EmailVerified, u.VerifyToken, u.CreatedAt,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return user.ErrEmailAlreadyExists
		}
		return err
	}
	return nil
}

// FindByEmail 按邮箱查找用户。
func (r *PostgresRepo) FindByEmail(email string) (*user.User, error) {
	u := &user.User{}
	err := r.pool.QueryRow(context.Background(),
		`SELECT email, password_hash, nickname, student_id, department,
		        tags, avatar, contact_info, email_verified, verify_token, created_at
		 FROM users WHERE email = $1`, email,
	).Scan(
		&u.Email, &u.PasswordHash, &u.Nickname, &u.StudentID,
		&u.Department, &u.Tags, &u.Avatar, &u.ContactInfo,
		&u.EmailVerified, &u.VerifyToken, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, user.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// FindByVerifyToken 按验证 token 查找用户。
func (r *PostgresRepo) FindByVerifyToken(token string) (*user.User, error) {
	u := &user.User{}
	err := r.pool.QueryRow(context.Background(),
		`SELECT email, password_hash, nickname, student_id, department,
		        tags, avatar, contact_info, email_verified, verify_token, created_at
		 FROM users WHERE verify_token = $1`, token,
	).Scan(
		&u.Email, &u.PasswordHash, &u.Nickname, &u.StudentID,
		&u.Department, &u.Tags, &u.Avatar, &u.ContactInfo,
		&u.EmailVerified, &u.VerifyToken, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, user.ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Update 更新用户全部字段。
func (r *PostgresRepo) Update(u *user.User) error {
	tag, err := r.pool.Exec(context.Background(),
		`UPDATE users SET password_hash=$1, nickname=$2, student_id=$3,
		 department=$4, tags=$5, avatar=$6, contact_info=$7,
		 email_verified=$8, verify_token=$9, created_at=$10
		 WHERE email=$11`,
		u.PasswordHash, u.Nickname, u.StudentID,
		u.Department, u.Tags, u.Avatar, u.ContactInfo,
		u.EmailVerified, u.VerifyToken, u.CreatedAt,
		u.Email,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return user.ErrUserNotFound
	}
	return nil
}

// FindAll 返回分页的已验证邮箱用户列表。
// page 从 1 开始。
func (r *PostgresRepo) FindAll(page, pageSize int) ([]*user.User, int, error) {
	// 先查询总数
	var total int
	err := r.pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM users WHERE email_verified = true`,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*user.User{}, 0, nil
	}

	// 计算 offset
	offset := (page - 1) * pageSize

	rows, err := r.pool.Query(context.Background(),
		`SELECT email, password_hash, nickname, student_id, department,
		        tags, avatar, contact_info, email_verified, verify_token, created_at
		 FROM users WHERE email_verified = true
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`, pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		u := &user.User{}
		if err := rows.Scan(
			&u.Email, &u.PasswordHash, &u.Nickname, &u.StudentID,
			&u.Department, &u.Tags, &u.Avatar, &u.ContactInfo,
			&u.EmailVerified, &u.VerifyToken, &u.CreatedAt,
		); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users, total, nil
}

// Delete 删除用户及其关联数据。
// 先删除用户的预约记录（含 mentor_id / student_id 外键），再删除用户本身；
// 空闲时间通过 ON DELETE CASCADE 自动级联删除。
func (r *PostgresRepo) Delete(email string) error {
	// 删除该用户作为 mentor 或 student 的所有预约
	_, err := r.pool.Exec(context.Background(),
		`DELETE FROM appointments WHERE mentor_id = $1 OR student_id = $1`, email)
	if err != nil {
		return err
	}

	// 删除用户（availabilities 通过 ON DELETE CASCADE 自动删除）
	tag, err := r.pool.Exec(context.Background(),
		`DELETE FROM users WHERE email = $1`, email)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return user.ErrUserNotFound
	}
	return nil
}

// =============================================================================
// appointment.AppointmentRepository 接口实现
// =============================================================================

// ── 预约操作 ──

// CreateAppointment 创建预约记录。
func (r *PostgresRepo) CreateAppointment(a *appointment.Appointment) error {
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO appointments (id, mentor_id, student_id, time_slot_id,
		 message, status, reject_reason, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.MentorID, a.StudentID, a.TimeSlotID,
		a.Message, a.Status, a.RejectReason, a.CreatedAt,
	)
	return err
}

// UpdateAppointment 更新预约（主要是状态变更）。
func (r *PostgresRepo) UpdateAppointment(a *appointment.Appointment) error {
	tag, err := r.pool.Exec(context.Background(),
		`UPDATE appointments SET status=$1, reject_reason=$2 WHERE id=$3`,
		a.Status, a.RejectReason, a.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return appointment.ErrAppointmentNotFound
	}
	return nil
}

// FindAppointmentByID 按 ID 查找预约。
func (r *PostgresRepo) FindAppointmentByID(id string) (*appointment.Appointment, error) {
	a := &appointment.Appointment{}
	err := r.pool.QueryRow(context.Background(),
		`SELECT id, mentor_id, student_id, time_slot_id,
		        message, status, reject_reason, created_at
		 FROM appointments WHERE id = $1`, id,
	).Scan(
		&a.ID, &a.MentorID, &a.StudentID, &a.TimeSlotID,
		&a.Message, &a.Status, &a.RejectReason, &a.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appointment.ErrAppointmentNotFound
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

// FindAppointmentsByMentorID 查找 mentor 收到的所有预约。
func (r *PostgresRepo) FindAppointmentsByMentorID(mentorID string) []*appointment.Appointment {
	return r.queryAppointments(
		`SELECT id, mentor_id, student_id, time_slot_id,
		        message, status, reject_reason, created_at
		 FROM appointments WHERE mentor_id = $1
		 ORDER BY created_at DESC`, mentorID,
	)
}

// FindAppointmentsByStudentID 查找 student 发出的所有预约。
func (r *PostgresRepo) FindAppointmentsByStudentID(studentID string) []*appointment.Appointment {
	return r.queryAppointments(
		`SELECT id, mentor_id, student_id, time_slot_id,
		        message, status, reject_reason, created_at
		 FROM appointments WHERE student_id = $1
		 ORDER BY created_at DESC`, studentID,
	)
}

// FindAppointmentsByTimeSlotID 查找某时间段的所有预约。
func (r *PostgresRepo) FindAppointmentsByTimeSlotID(timeSlotID string) []*appointment.Appointment {
	return r.queryAppointments(
		`SELECT id, mentor_id, student_id, time_slot_id,
		        message, status, reject_reason, created_at
		 FROM appointments WHERE time_slot_id = $1`, timeSlotID,
	)
}

// HasActiveAppointment 检查时间段是否有活跃预约（pending 或 accepted）。
func (r *PostgresRepo) HasActiveAppointment(timeSlotID string) bool {
	var count int
	err := r.pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM appointments
		 WHERE time_slot_id = $1 AND status IN ('pending','accepted')`,
		timeSlotID,
	).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// queryAppointments 通用预约查询辅助方法。
func (r *PostgresRepo) queryAppointments(sql string, args ...interface{}) []*appointment.Appointment {
	rows, err := r.pool.Query(context.Background(), sql, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []*appointment.Appointment
	for rows.Next() {
		a := &appointment.Appointment{}
		if err := rows.Scan(
			&a.ID, &a.MentorID, &a.StudentID, &a.TimeSlotID,
			&a.Message, &a.Status, &a.RejectReason, &a.CreatedAt,
		); err != nil {
			continue
		}
		result = append(result, a)
	}
	return result
}

// ── 空闲时间操作 ──

// CreateAvailability 添加空闲时间。
func (r *PostgresRepo) CreateAvailability(a *appointment.Availability) error {
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO availabilities (id, user_id, date, start_time, end_time)
		 VALUES ($1,$2,$3,$4,$5)`,
		a.ID, a.UserID, a.Date, a.StartTime, a.EndTime,
	)
	return err
}

// DeleteAvailability 删除空闲时间。
func (r *PostgresRepo) DeleteAvailability(id string) error {
	tag, err := r.pool.Exec(context.Background(),
		`DELETE FROM availabilities WHERE id = $1`, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return appointment.ErrSlotNotFound
	}
	return nil
}

// FindAvailabilityByID 按 ID 查找空闲时间。
func (r *PostgresRepo) FindAvailabilityByID(id string) (*appointment.Availability, error) {
	a := &appointment.Availability{}
	err := r.pool.QueryRow(context.Background(),
		`SELECT id, user_id, date::text, start_time::text, end_time::text
		 FROM availabilities WHERE id = $1`, id,
	).Scan(&a.ID, &a.UserID, &a.Date, &a.StartTime, &a.EndTime)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appointment.ErrSlotNotFound
	}
	if err != nil {
		return nil, err
	}
	// 截取时间部分 HH:MM
	if len(a.StartTime) >= 5 {
		a.StartTime = a.StartTime[:5]
	}
	if len(a.EndTime) >= 5 {
		a.EndTime = a.EndTime[:5]
	}
	return a, nil
}

// FindAvailabilitiesByUserID 查找用户的所有空闲时间。
func (r *PostgresRepo) FindAvailabilitiesByUserID(userID string) []*appointment.Availability {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id, user_id, date::text, start_time::text, end_time::text
		 FROM availabilities WHERE user_id = $1
		 ORDER BY date, start_time`, userID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []*appointment.Availability
	for rows.Next() {
		a := &appointment.Availability{}
		if err := rows.Scan(&a.ID, &a.UserID, &a.Date, &a.StartTime, &a.EndTime); err != nil {
			continue
		}
		if len(a.StartTime) >= 5 {
			a.StartTime = a.StartTime[:5]
		}
		if len(a.EndTime) >= 5 {
			a.EndTime = a.EndTime[:5]
		}
		result = append(result, a)
	}
	return result
}

// =============================================================================
// 辅助函数
// =============================================================================

// isDuplicateKey 判断是否为 PostgreSQL 唯一约束冲突（错误码 23505）。
func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
