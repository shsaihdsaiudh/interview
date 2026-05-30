package appointment

// AppointmentRepository 定义预约和空闲时间的持久化接口。
// 接口定义在领域层，实现在基础设施层。
type AppointmentRepository interface {
	// ── 预约操作 ──

	// CreateAppointment 创建预约记录。
	CreateAppointment(a *Appointment) error

	// UpdateAppointment 更新预约（状态变更后持久化）。
	UpdateAppointment(a *Appointment) error

	// FindAppointmentByID 按 ID 查找预约。
	FindAppointmentByID(id string) (*Appointment, error)

	// FindAppointmentsByMentorID 查找 mentor 收到的所有预约。
	FindAppointmentsByMentorID(mentorID string) []*Appointment

	// FindAppointmentsByStudentID 查找 student 发出的所有预约。
	FindAppointmentsByStudentID(studentID string) []*Appointment

	// FindAppointmentsByTimeSlotID 查找某时间段的所有预约。
	FindAppointmentsByTimeSlotID(timeSlotID string) []*Appointment

	// HasActiveAppointment 检查时间段是否有活跃预约（pending/accepted）。
	HasActiveAppointment(timeSlotID string) bool

	// ── 空闲时间操作 ──

	// CreateAvailability 添加空闲时间。
	CreateAvailability(a *Availability) error

	// DeleteAvailability 删除空闲时间。
	DeleteAvailability(id string) error

	// FindAvailabilityByID 按 ID 查找空闲时间。
	FindAvailabilityByID(id string) (*Availability, error)

	// FindAvailabilitiesByUserID 查找用户的所有空闲时间。
	FindAvailabilitiesByUserID(userID string) []*Availability
}
