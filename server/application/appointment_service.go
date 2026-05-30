package application

import (
	"errors"
	"time"

	"interview-server/domain/appointment"
	"interview-server/domain/user"
)

// AppointmentService 预约用例编排。
// 依赖领域接口，不依赖具体技术实现。
type AppointmentService struct {
	apptRepo appointment.AppointmentRepository
	userRepo user.UserRepository
}

// NewAppointmentService 创建预约服务。
func NewAppointmentService(apptRepo appointment.AppointmentRepository, userRepo user.UserRepository) *AppointmentService {
	return &AppointmentService{apptRepo: apptRepo, userRepo: userRepo}
}

// ── 空闲时间管理 ──

// GetMyAvailability 获取当前用户的空闲时间列表。
func (s *AppointmentService) GetMyAvailability(userID string) []*appointment.Availability {
	slots := s.apptRepo.FindAvailabilitiesByUserID(userID)
	if slots == nil {
		return []*appointment.Availability{}
	}
	return slots
}

// AddAvailability 添加空闲时间。
func (s *AppointmentService) AddAvailability(userID string, req appointment.AddAvailabilityRequest) (*appointment.Availability, error) {
	// 校验时间格式
	if err := validateTime(req.StartTime); err != nil {
		return nil, err
	}
	if err := validateTime(req.EndTime); err != nil {
		return nil, err
	}

	// 结束时间必须晚于开始时间
	if req.EndTime <= req.StartTime {
		return nil, appointment.ErrTimeConflict
	}

	// 日期不能是过去
	slotDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("日期格式不正确，请使用 YYYY-MM-DD 格式")
	}
	today := time.Now().Truncate(24 * time.Hour)
	if slotDate.Before(today) {
		return nil, appointment.ErrPastDate
	}

	id, _ := generateRandomHex(16)
	slot := &appointment.Availability{
		ID:        id,
		UserID:    userID,
		Date:      req.Date,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	if err := s.apptRepo.CreateAvailability(slot); err != nil {
		return nil, err
	}
	return slot, nil
}

// DeleteAvailability 删除空闲时间（需验证所有权）。
func (s *AppointmentService) DeleteAvailability(userID, slotID string) error {
	slot, err := s.apptRepo.FindAvailabilityByID(slotID)
	if err != nil {
		return appointment.ErrSlotNotFound
	}
	if slot.UserID != userID {
		return appointment.ErrSlotNotOwned
	}
	return s.apptRepo.DeleteAvailability(slotID)
}

// ── 预约管理 ──

// CreateAppointment 发起预约。
func (s *AppointmentService) CreateAppointment(studentID string, req appointment.CreateAppointmentRequest) (*appointment.Appointment, error) {
	// 查找空闲时间段
	slot, err := s.apptRepo.FindAvailabilityByID(req.TimeSlotID)
	if err != nil {
		return nil, appointment.ErrSlotNotFound
	}

	// 不能预约自己的时间段
	if slot.UserID == studentID {
		return nil, appointment.ErrCannotBookOwnSlot
	}

	// 检查 mentor 是否存在
	if _, err := s.userRepo.FindByEmail(slot.UserID); err != nil {
		return nil, appointment.ErrMentorNotFound
	}

	// 检查时间段是否被占用
	if s.apptRepo.HasActiveAppointment(req.TimeSlotID) {
		return nil, appointment.ErrSlotAlreadyBooked
	}

	// 检查时间段日期是否已过
	slotDate, err := time.Parse("2006-01-02", slot.Date)
	if err != nil {
		return nil, errors.New("时间段时间格式不正确")
	}
	if slotDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, appointment.ErrPastDate
	}

	id, _ := generateRandomHex(16)
	appt := &appointment.Appointment{
		ID:         id,
		MentorID:   slot.UserID,
		StudentID:  studentID,
		TimeSlotID: req.TimeSlotID,
		Message:    req.Message,
		Status:     appointment.StatusPending,
		CreatedAt:  time.Now(),
	}

	if err := s.apptRepo.CreateAppointment(appt); err != nil {
		return nil, err
	}
	return appt, nil
}

// GetMyAppointments 获取我的预约列表。
// role: "mentor" = 收到的预约, "student" = 发出的预约。
func (s *AppointmentService) GetMyAppointments(userID, role string) ([]*appointment.Appointment, error) {
	if role == "mentor" {
		apps := s.apptRepo.FindAppointmentsByMentorID(userID)
		if apps == nil {
			return []*appointment.Appointment{}, nil
		}
		return apps, nil
	}

	apps := s.apptRepo.FindAppointmentsByStudentID(userID)
	if apps == nil {
		return []*appointment.Appointment{}, nil
	}
	return apps, nil
}

// AcceptAppointment 接受预约（只有 mentor 可以）。
// 亮点：使用聚合根方法 a.Accept() 而非直接修改 a.Status。
func (s *AppointmentService) AcceptAppointment(userID, appointmentID string) (*appointment.Appointment, error) {
	appt, err := s.apptRepo.FindAppointmentByID(appointmentID)
	if err != nil {
		return nil, appointment.ErrAppointmentNotFound
	}

	// 权限检查：只有 mentor 可以操作
	if !appt.CanBeOperatedBy(userID) {
		return nil, appointment.ErrNotMentor
	}

	// 使用聚合根方法 — 封装了状态变更的业务规则
	if err := appt.Accept(); err != nil {
		return nil, err
	}

	if err := s.apptRepo.UpdateAppointment(appt); err != nil {
		return nil, err
	}
	return appt, nil
}

// RejectAppointment 拒绝预约（只有 mentor 可以）。
func (s *AppointmentService) RejectAppointment(userID, appointmentID, reason string) (*appointment.Appointment, error) {
	appt, err := s.apptRepo.FindAppointmentByID(appointmentID)
	if err != nil {
		return nil, appointment.ErrAppointmentNotFound
	}

	if !appt.CanBeOperatedBy(userID) {
		return nil, appointment.ErrNotMentor
	}

	// 使用聚合根方法
	if err := appt.Reject(reason); err != nil {
		return nil, err
	}

	if err := s.apptRepo.UpdateAppointment(appt); err != nil {
		return nil, err
	}
	return appt, nil
}

// ResolveUsers 填充预约关联的用户和时间段信息（用于 API 响应）。
func (s *AppointmentService) ResolveUsers(appt *appointment.Appointment) (*appointment.AppointmentResponse, error) {
	mentor, err := s.userRepo.FindByEmail(appt.MentorID)
	if err != nil {
		return nil, err
	}
	student, err := s.userRepo.FindByEmail(appt.StudentID)
	if err != nil {
		return nil, err
	}
	slot, err := s.apptRepo.FindAvailabilityByID(appt.TimeSlotID)
	if err != nil {
		return nil, err
	}

	mentorResp := mentor.ToResponse()
	studentResp := student.ToResponse()

	// 只有 accepted 状态才展示联系方式
	if appt.IsAccepted() {
		mentorResp = mentor.ToResponseWithContact()
		studentResp = student.ToResponseWithContact()
	}

	return &appointment.AppointmentResponse{
		ID:           appt.ID,
		MentorID:     appt.MentorID,
		StudentID:    appt.StudentID,
		TimeSlotID:   appt.TimeSlotID,
		Message:      appt.Message,
		Status:       appt.Status,
		RejectReason: appt.RejectReason,
		CreatedAt:    appt.CreatedAt,
		Mentor:       mentorResp,
		Student:      studentResp,
		TimeSlot:     slot,
	}, nil
}

// =============================================================================
// 辅助函数
// =============================================================================

// validateTime 校验时间格式 HH:MM。
func validateTime(t string) error {
	if len(t) != 5 || t[2] != ':' {
		return appointment.ErrInvalidTimeFormat
	}
	_, err := time.Parse("15:04", t)
	if err != nil {
		return appointment.ErrInvalidTimeFormat
	}
	return nil
}
