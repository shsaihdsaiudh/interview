package model

import "time"

// Appointment 预约记录
type Appointment struct {
	ID           string    `json:"id"`
	MentorID     string    `json:"mentor_id"`     // 被预约的人（时间所有者）
	StudentID    string    `json:"student_id"`    // 发起预约的人
	TimeSlotID   string    `json:"time_slot_id"`  // 选中的空闲时间 ID
	Message      string    `json:"message"`       // 附言
	Status       string    `json:"status"`        // pending / accepted / rejected
	RejectReason string    `json:"reject_reason"` // 拒绝原因
	CreatedAt    time.Time `json:"created_at"`
}

// 预约状态常量
const (
	AppointmentPending  = "pending"
	AppointmentAccepted = "accepted"
	AppointmentRejected = "rejected"
)

// CreateAppointmentRequest 发起预约请求
type CreateAppointmentRequest struct {
	TimeSlotID string `json:"time_slot_id" binding:"required"`
	Message    string `json:"message" binding:"required"`
}

// RejectAppointmentRequest 拒绝预约请求
type RejectAppointmentRequest struct {
	Reason string `json:"reason"`
}

// AppointmentResponse 预约响应（含关联用户信息）
type AppointmentResponse struct {
	Appointment
	Mentor     *UserResponse `json:"mentor"`
	Student    *UserResponse `json:"student"`
	TimeSlot   *Availability `json:"time_slot"`
}
