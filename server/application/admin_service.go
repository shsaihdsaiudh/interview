package application

import (
	"time"

	"interview-server/domain/appointment"
	"interview-server/domain/recruitment"
	"interview-server/domain/user"
)

// AdminService 管理员后台服务。
type AdminService struct {
	userRepo  user.UserRepository
	apptRepo  appointment.AppointmentRepository
	cardRepo  recruitment.RecruitmentCardRepository
}

// NewAdminService 创建管理员服务。
func NewAdminService(
	userRepo user.UserRepository,
	apptRepo appointment.AppointmentRepository,
	cardRepo recruitment.RecruitmentCardRepository,
) *AdminService {
	return &AdminService{
		userRepo: userRepo,
		apptRepo: apptRepo,
		cardRepo: cardRepo,
	}
}

// ── DTO ──

// AdminStats 仪表盘统计。
type AdminStats struct {
	TotalUsers        int `json:"total_users"`
	TotalCards        int `json:"total_cards"`
	TotalAppointments int `json:"total_appointments"`
	NewUsersToday     int `json:"new_users_today"`
}

// AdminUserListResponse 管理员用户列表。
type AdminUserListResponse struct {
	Users    []user.UserResponse `json:"users"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// AdminCardListResponse 管理员名片列表。
type AdminCardListResponse struct {
	Cards    []*recruitment.RecruitmentCard `json:"cards"`
	Total    int                            `json:"total"`
	Page     int                            `json:"page"`
	PageSize int                            `json:"page_size"`
}

// AdminAppointmentInfo 管理员预约详情（含昵称和时间信息）。
type AdminAppointmentInfo struct {
	ID             string `json:"id"`
	MentorID       string `json:"mentor_id"`
	MentorNickname string `json:"mentor_nickname"`
	StudentID      string `json:"student_id"`
	StudentNickname string `json:"student_nickname"`
	TimeSlotDate   string `json:"time_slot_date"`
	TimeSlotStart  string `json:"time_slot_start"`
	TimeSlotEnd    string `json:"time_slot_end"`
	Message        string `json:"message"`
	Status         string `json:"status"`
	RejectReason   string `json:"reject_reason"`
	CreatedAt      string `json:"created_at"`
}

// AdminApptListResponse 管理员预约列表。
type AdminApptListResponse struct {
	Appointments []*AdminAppointmentInfo `json:"appointments"`
	Total        int                     `json:"total"`
	Page         int                     `json:"page"`
	PageSize     int                     `json:"page_size"`
}

// UpdateRoleRequest 修改用户角色请求。
type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// ── 仪表盘 ──

// GetStats 获取仪表盘统计数据。
func (s *AdminService) GetStats() (*AdminStats, error) {
	// 用户总数（取 total 字段，无需实际数据）
	_, totalUsers, err := s.userRepo.FindAllAdmin("", 1, 1)
	if err != nil {
		return nil, err
	}

	// 名片总数
	_, totalCards, err := s.cardRepo.ListAllAdmin("", 1, 1)
	if err != nil {
		return nil, err
	}

	// 预约总数
	_, totalAppts, err := s.apptRepo.FindAllAppointments(1, 1)
	if err != nil {
		return nil, err
	}

	// 今日新增用户
	today := time.Now().Format("2006-01-02")
	newToday, err := s.userRepo.CountByDate(today)
	if err != nil {
		return nil, err
	}

	return &AdminStats{
		TotalUsers:        totalUsers,
		TotalCards:        totalCards,
		TotalAppointments: totalAppts,
		NewUsersToday:     newToday,
	}, nil
}

// ── 用户管理 ──

// ListUsers 管理员查询用户列表。
func (s *AdminService) ListUsers(keyword string, page, pageSize int) (*AdminUserListResponse, error) {
	users, total, err := s.userRepo.FindAllAdmin(keyword, page, pageSize)
	if err != nil {
		return nil, err
	}

	result := make([]user.UserResponse, 0, len(users))
	for _, u := range users {
		result = append(result, u.ToResponse())
	}

	return &AdminUserListResponse{
		Users:    result,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// UpdateUserRole 修改用户角色。
func (s *AdminService) UpdateUserRole(email, role string) (*user.UserResponse, error) {
	if role != user.RoleUser && role != user.RoleAdmin {
		return nil, user.ErrInvalidRole
	}

	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	u.Role = role
	if err := s.userRepo.Update(u); err != nil {
		return nil, err
	}

	resp := u.ToResponse()
	return &resp, nil
}

// BanUser 封禁用户。
func (s *AdminService) BanUser(email string) (*user.UserResponse, error) {
	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	u.Ban()
	if err := s.userRepo.Update(u); err != nil {
		return nil, err
	}

	resp := u.ToResponse()
	return &resp, nil
}

// UnbanUser 解封用户。
func (s *AdminService) UnbanUser(email string) (*user.UserResponse, error) {
	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	u.Unban()
	if err := s.userRepo.Update(u); err != nil {
		return nil, err
	}

	resp := u.ToResponse()
	return &resp, nil
}

// ── 名片管理 ──

// ListCards 管理员查询名片列表。
func (s *AdminService) ListCards(keyword string, page, pageSize int) (*AdminCardListResponse, error) {
	cards, total, err := s.cardRepo.ListAllAdmin(keyword, page, pageSize)
	if err != nil {
		return nil, err
	}

	if cards == nil {
		cards = []*recruitment.RecruitmentCard{}
	}

	return &AdminCardListResponse{
		Cards:    cards,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// DeleteCard 删除名片。
func (s *AdminService) DeleteCard(id string) error {
	return s.cardRepo.DeleteByID(id)
}

// ── 预约管理 ──

// ListAppointments 管理员查询预约列表（含昵称和时段信息）。
func (s *AdminService) ListAppointments(page, pageSize int) (*AdminApptListResponse, error) {
	rows, total, err := s.apptRepo.FindAllAppointmentsAdmin(page, pageSize)
	if err != nil {
		return nil, err
	}

	result := make([]*AdminAppointmentInfo, 0, len(rows))
	for _, row := range rows {
		result = append(result, &AdminAppointmentInfo{
			ID:              row.ID,
			MentorID:        row.MentorID,
			MentorNickname:  row.MentorNickname,
			StudentID:       row.StudentID,
			StudentNickname: row.StudentNickname,
			TimeSlotDate:    row.TimeSlotDate,
			TimeSlotStart:   row.TimeSlotStart,
			TimeSlotEnd:     row.TimeSlotEnd,
			Message:         row.Message,
			Status:          row.Status,
			RejectReason:    row.RejectReason,
			CreatedAt:       row.CreatedAt,
		})
	}

	return &AdminApptListResponse{
		Appointments: result,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
	}, nil
}

// DeleteAppointment 删除预约。
func (s *AdminService) DeleteAppointment(id string) error {
	return s.apptRepo.DeleteAppointmentByID(id)
}
