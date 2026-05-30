package service

import (
	"errors"

	"interview-server/internal/model"
	"interview-server/internal/repository"
)

// UserService 用户业务逻辑
type UserService struct {
	userRepo  *repository.UserRepo
	availRepo *repository.AvailabilityRepo
	apptRepo  *repository.AppointmentRepo
}

// NewUserService 创建用户服务
func NewUserService(
	userRepo *repository.UserRepo,
	availRepo *repository.AvailabilityRepo,
	apptRepo *repository.AppointmentRepo,
) *UserService {
	return &UserService{
		userRepo:  userRepo,
		availRepo: availRepo,
		apptRepo:  apptRepo,
	}
}

// UpdateProfile 更新个人资料
func (s *UserService) UpdateProfile(email string, req model.UpdateProfileRequest) (*model.UserResponse, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.StudentID != "" {
		user.StudentID = req.StudentID
	}
	user.Department = req.Department
	user.Tags = req.Tags
	user.Avatar = req.Avatar
	user.ContactInfo = req.ContactInfo

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	resp := user.ToResponse()
	return &resp, nil
}

// GetAllUsers 获取所有已验证用户（含空闲时间摘要）
func (s *UserService) GetAllUsers() []model.UserResponse {
	users := s.userRepo.FindAll()
	result := make([]model.UserResponse, 0, len(users))
	for _, u := range users {
		resp := u.ToResponse()
		result = append(result, resp)
	}
	return result
}

// GetUserDetail 获取用户详情（含空闲时间）
func (s *UserService) GetUserDetail(email string, requesterEmail string) (*UserDetailResponse, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	resp := user.ToResponse()
	availabilities := s.availRepo.FindByUserID(email)
	if availabilities == nil {
		availabilities = []*model.Availability{}
	}

	// 检查 requester 是否与该用户有 accepted 的预约
	if requesterEmail != "" && s.hasAcceptedAppointment(requesterEmail, email) {
		resp.ContactInfo = user.ContactInfo
	}

	return &UserDetailResponse{
		User:          resp,
		Availabilities: availabilities,
	}, nil
}

// hasAcceptedAppointment 检查两个用户之间是否有已接受的预约
func (s *UserService) hasAcceptedAppointment(user1, user2 string) bool {
	// 检查 user1 作为 student，user2 作为 mentor
	apps := s.apptRepo.FindByStudentID(user1)
	for _, a := range apps {
		if a.MentorID == user2 && a.Status == model.AppointmentAccepted {
			return true
		}
	}
	// 检查 user1 作为 mentor，user2 作为 student
	apps = s.apptRepo.FindByMentorID(user1)
	for _, a := range apps {
		if a.StudentID == user2 && a.Status == model.AppointmentAccepted {
			return true
		}
	}
	return false
}

// UserDetailResponse 用户详情响应
type UserDetailResponse struct {
	User          model.UserResponse   `json:"user"`
	Availabilities []*model.Availability `json:"availabilities"`
}
