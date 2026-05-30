package repository

import (
	"errors"
	"sort"

	"interview-server/internal/model"
)

var ErrAppointmentNotFound = errors.New("预约不存在")

// AppointmentRepo 预约仓库
type AppointmentRepo struct {
	store *Store
}

// NewAppointmentRepo 创建预约仓库
func NewAppointmentRepo(store *Store) *AppointmentRepo {
	return &AppointmentRepo{store: store}
}

// Create 创建预约
func (r *AppointmentRepo) Create(a *model.Appointment) error {
	r.store.Lock()
	defer r.store.Unlock()

	r.store.Appointments[a.ID] = a
	return nil
}

// Update 更新预约状态
func (r *AppointmentRepo) Update(a *model.Appointment) error {
	r.store.Lock()
	defer r.store.Unlock()

	if _, exists := r.store.Appointments[a.ID]; !exists {
		return ErrAppointmentNotFound
	}
	r.store.Appointments[a.ID] = a
	return nil
}

// FindByID 按 ID 查找
func (r *AppointmentRepo) FindByID(id string) (*model.Appointment, error) {
	r.store.RLock()
	defer r.store.RUnlock()

	a, exists := r.store.Appointments[id]
	if !exists {
		return nil, ErrAppointmentNotFound
	}
	return a, nil
}

// FindByMentorID 查找 mentor 收到的所有预约
func (r *AppointmentRepo) FindByMentorID(mentorID string) []*model.Appointment {
	r.store.RLock()
	defer r.store.RUnlock()

	var result []*model.Appointment
	for _, a := range r.store.Appointments {
		if a.MentorID == mentorID {
			result = append(result, a)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result
}

// FindByStudentID 查找 student 发出的所有预约
func (r *AppointmentRepo) FindByStudentID(studentID string) []*model.Appointment {
	r.store.RLock()
	defer r.store.RUnlock()

	var result []*model.Appointment
	for _, a := range r.store.Appointments {
		if a.StudentID == studentID {
			result = append(result, a)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result
}

// FindByTimeSlotID 查找某时间段的所有预约（检查重复预约）
func (r *AppointmentRepo) FindByTimeSlotID(timeSlotID string) []*model.Appointment {
	r.store.RLock()
	defer r.store.RUnlock()

	var result []*model.Appointment
	for _, a := range r.store.Appointments {
		if a.TimeSlotID == timeSlotID {
			result = append(result, a)
		}
	}
	return result
}

// HasActiveAppointment 检查时间段是否有活跃的预约（pending 或 accepted）
func (r *AppointmentRepo) HasActiveAppointment(timeSlotID string) bool {
	apps := r.FindByTimeSlotID(timeSlotID)
	for _, a := range apps {
		if a.Status == model.AppointmentPending || a.Status == model.AppointmentAccepted {
			return true
		}
	}
	return false
}
