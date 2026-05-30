package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"interview-server/internal/model"
)

var ErrAppointmentNotFound = errors.New("预约不存在")

// AppointmentRepo 预约仓库（PostgreSQL）
type AppointmentRepo struct {
	pool *pgxpool.Pool
}

// NewAppointmentRepo 创建预约仓库
func NewAppointmentRepo(pool *pgxpool.Pool) *AppointmentRepo {
	return &AppointmentRepo{pool: pool}
}

// Create 创建预约
func (r *AppointmentRepo) Create(a *model.Appointment) error {
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO appointments (id, mentor_id, student_id, time_slot_id,
		 message, status, reject_reason, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.MentorID, a.StudentID, a.TimeSlotID,
		a.Message, a.Status, a.RejectReason, a.CreatedAt,
	)
	return err
}

// Update 更新预约
func (r *AppointmentRepo) Update(a *model.Appointment) error {
	tag, err := r.pool.Exec(context.Background(),
		`UPDATE appointments SET status=$1, reject_reason=$2 WHERE id=$3`,
		a.Status, a.RejectReason, a.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAppointmentNotFound
	}
	return nil
}

// FindByID 按 ID 查找
func (r *AppointmentRepo) FindByID(id string) (*model.Appointment, error) {
	a := &model.Appointment{}
	err := r.pool.QueryRow(context.Background(),
		`SELECT id, mentor_id, student_id, time_slot_id,
		        message, status, reject_reason, created_at
		 FROM appointments WHERE id = $1`, id,
	).Scan(
		&a.ID, &a.MentorID, &a.StudentID, &a.TimeSlotID,
		&a.Message, &a.Status, &a.RejectReason, &a.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAppointmentNotFound
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

// FindByMentorID 查找 mentor 收到的所有预约
func (r *AppointmentRepo) FindByMentorID(mentorID string) []*model.Appointment {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id, mentor_id, student_id, time_slot_id,
		        message, status, reject_reason, created_at
		 FROM appointments WHERE mentor_id = $1
		 ORDER BY created_at DESC`, mentorID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	return scanAppointments(rows)
}

// FindByStudentID 查找 student 发出的所有预约
func (r *AppointmentRepo) FindByStudentID(studentID string) []*model.Appointment {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id, mentor_id, student_id, time_slot_id,
		        message, status, reject_reason, created_at
		 FROM appointments WHERE student_id = $1
		 ORDER BY created_at DESC`, studentID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	return scanAppointments(rows)
}

// FindByTimeSlotID 查找某时间段的所有预约
func (r *AppointmentRepo) FindByTimeSlotID(timeSlotID string) []*model.Appointment {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id, mentor_id, student_id, time_slot_id,
		        message, status, reject_reason, created_at
		 FROM appointments WHERE time_slot_id = $1`, timeSlotID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	return scanAppointments(rows)
}

// HasActiveAppointment 检查时间段是否有活跃预约
func (r *AppointmentRepo) HasActiveAppointment(timeSlotID string) bool {
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

// scanAppointments 通用行扫描
func scanAppointments(rows pgx.Rows) []*model.Appointment {
	var result []*model.Appointment
	for rows.Next() {
		a := &model.Appointment{}
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
