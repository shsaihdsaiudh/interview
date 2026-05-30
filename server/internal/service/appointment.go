package service

import (
	"errors"
	"time"

	"interview-server/internal/model"
	"interview-server/internal/repository"
)

var (
	ErrAppointmentNotFound   = errors.New("预约不存在")
	ErrNotMentor             = errors.New("只有 mentor 才能操作此预约")
	ErrSlotAlreadyBooked     = errors.New("该时间段已被预约")
	ErrCannotBookOwnSlot     = errors.New("不能预约自己的空闲时间")
	ErrInvalidStatus         = errors.New("只能处理待确认的预约")
	ErrMentorNotFound        = errors.New("mentor 不存在")
)

// AppointmentService 预约业务逻辑
type AppointmentService struct {
	apptRepo  *repository.AppointmentRepo
	availRepo *repository.AvailabilityRepo
	userRepo  *repository.UserRepo
}

// NewAppointmentService 创建预约服务
func NewAppointmentService(
	apptRepo *repository.AppointmentRepo,
	availRepo *repository.AvailabilityRepo,
	userRepo *repository.UserRepo,
) *AppointmentService {
	return &AppointmentService{
		apptRepo:  apptRepo,
		availRepo: availRepo,
		userRepo:  userRepo,
	}
}

// CreateAppointment 发起预约
func (s *AppointmentService) CreateAppointment(studentID string, req model.CreateAppointmentRequest) (*model.Appointment, error) {
	// 查找空闲时间段
	slot, err := s.availRepo.FindByID(req.TimeSlotID)
	if err != nil {
		return nil, ErrSlotNotFound
	}

	// 不能预约自己的时间段
	if slot.UserID == studentID {
		return nil, ErrCannotBookOwnSlot
	}

	// 检查 mentor 是否存在
	mentor, err := s.userRepo.FindByEmail(slot.UserID)
	if err != nil {
		return nil, ErrMentorNotFound
	}
	_ = mentor // mentor 存在即可

	// 检查时间段是否被占用
	if s.apptRepo.HasActiveAppointment(req.TimeSlotID) {
		return nil, ErrSlotAlreadyBooked
	}

	// 检查时间段日期是否已过
	slotDate, err := time.Parse("2006-01-02", slot.Date)
	if err != nil {
		return nil, errors.New("时间段时间格式不正确")
	}
	if slotDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, ErrPastDate
	}

	id, _ := generateID()
	appt := &model.Appointment{
		ID:         id,
		MentorID:   slot.UserID,
		StudentID:  studentID,
		TimeSlotID: req.TimeSlotID,
		Message:    req.Message,
		Status:     model.AppointmentPending,
		CreatedAt:  time.Now(),
	}

	if err := s.apptRepo.Create(appt); err != nil {
		return nil, err
	}
	return appt, nil
}

// GetMyAppointments 获取我的预约列表
// role: "mentor" = 收到的预约, "student" = 发出的预约
func (s *AppointmentService) GetMyAppointments(userID, role string) ([]*model.Appointment, error) {
	if role == "mentor" {
		apps := s.apptRepo.FindByMentorID(userID)
		if apps == nil {
			return []*model.Appointment{}, nil
		}
		return apps, nil
	}

	apps := s.apptRepo.FindByStudentID(userID)
	if apps == nil {
		return []*model.Appointment{}, nil
	}
	return apps, nil
}

// AcceptAppointment 接受预约（只有 mentor 可以）
func (s *AppointmentService) AcceptAppointment(userID, appointmentID string) (*model.Appointment, error) {
	appt, err := s.apptRepo.FindByID(appointmentID)
	if err != nil {
		return nil, ErrAppointmentNotFound
	}

	if appt.MentorID != userID {
		return nil, ErrNotMentor
	}

	if appt.Status != model.AppointmentPending {
		return nil, ErrInvalidStatus
	}

	appt.Status = model.AppointmentAccepted
	if err := s.apptRepo.Update(appt); err != nil {
		return nil, err
	}
	return appt, nil
}

// RejectAppointment 拒绝预约（只有 mentor 可以）
func (s *AppointmentService) RejectAppointment(userID, appointmentID, reason string) (*model.Appointment, error) {
	appt, err := s.apptRepo.FindByID(appointmentID)
	if err != nil {
		return nil, ErrAppointmentNotFound
	}

	if appt.MentorID != userID {
		return nil, ErrNotMentor
	}

	if appt.Status != model.AppointmentPending {
		return nil, ErrInvalidStatus
	}

	appt.Status = model.AppointmentRejected
	appt.RejectReason = reason
	if err := s.apptRepo.Update(appt); err != nil {
		return nil, err
	}
	return appt, nil
}

// ResolveUsers 填充预约关联的用户和时间段信息
func (s *AppointmentService) ResolveUsers(appt *model.Appointment) (*model.AppointmentResponse, error) {
	mentor, err := s.userRepo.FindByEmail(appt.MentorID)
	if err != nil {
		return nil, err
	}
	student, err := s.userRepo.FindByEmail(appt.StudentID)
	if err != nil {
		return nil, err
	}
	slot, err := s.availRepo.FindByID(appt.TimeSlotID)
	if err != nil {
		return nil, err
	}

	mentorResp := mentor.ToResponse()
	studentResp := student.ToResponse()

	// 只有 accepted 状态才展示联系方式
	if appt.Status == model.AppointmentAccepted {
		mentorResp = mentor.ToResponseWithContact()
		studentResp = student.ToResponseWithContact()
	}

	return &model.AppointmentResponse{
		Appointment: *appt,
		Mentor:     &mentorResp,
		Student:    &studentResp,
		TimeSlot:   slot,
	}, nil
}
