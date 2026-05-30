// Package appointment 定义预约领域的聚合根、实体和值对象。
// 本包零外部依赖，只使用 Go 标准库。
package appointment

import "time"

// ── 预约状态常量 ──

const (
	StatusPending  = "pending"
	StatusAccepted = "accepted"
	StatusRejected = "rejected"
)

// ── Appointment 聚合根 ──

// Appointment 是预约聚合根。
// 只有 mentor（时间所有者）可以变更预约状态。
type Appointment struct {
	ID           string
	MentorID     string // 被预约的人（时间所有者）
	StudentID    string // 发起预约的人
	TimeSlotID   string // 选中的空闲时间 ID
	Message      string // 附言
	Status       string // pending / accepted / rejected
	RejectReason string // 拒绝原因
	CreatedAt    time.Time
}

// ── 聚合根行为方法（封装状态变更，外部不能直接改 Status 字段）──

// Accept 接受预约。只有 pending 状态的预约可以被接受。
func (a *Appointment) Accept() error {
	if a.Status != StatusPending {
		return ErrNotPending
	}
	a.Status = StatusAccepted
	return nil
}

// Reject 拒绝预约。只有 pending 状态的预约可以被拒绝。
func (a *Appointment) Reject(reason string) error {
	if a.Status != StatusPending {
		return ErrNotPending
	}
	a.Status = StatusRejected
	a.RejectReason = reason
	return nil
}

// CanBeOperatedBy 检查用户是否有权操作此预约。
// 只有 mentor 可以接受/拒绝预约。
func (a *Appointment) CanBeOperatedBy(userID string) bool {
	return a.MentorID == userID
}

// IsAccepted 检查预约是否已被接受。
func (a *Appointment) IsAccepted() bool {
	return a.Status == StatusAccepted
}

// ── Availability 实体 ──

// Availability 表示用户的一段空闲时间。
// 它是一个实体（有自己的 ID 和生命周期），但不是聚合根。
type Availability struct {
	ID        string
	UserID    string
	Date      string // "2006-01-02" 格式
	StartTime string // "14:00" 格式
	EndTime   string // "15:00" 格式
}

// ── 请求/响应 DTO ──

// CreateAppointmentRequest 发起预约请求
type CreateAppointmentRequest struct {
	TimeSlotID string `json:"time_slot_id" binding:"required"`
	Message    string `json:"message" binding:"required"`
}

// RejectAppointmentRequest 拒绝预约请求
type RejectAppointmentRequest struct {
	Reason string `json:"reason"`
}

// AddAvailabilityRequest 添加空闲时间请求
type AddAvailabilityRequest struct {
	Date      string `json:"date" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

// AppointmentResponse 预约响应（含关联用户信息）
type AppointmentResponse struct {
	ID           string    `json:"id"`
	MentorID     string    `json:"mentor_id"`
	StudentID    string    `json:"student_id"`
	TimeSlotID   string    `json:"time_slot_id"`
	Message      string    `json:"message"`
	Status       string    `json:"status"`
	RejectReason string    `json:"reject_reason"`
	CreatedAt    time.Time `json:"created_at"`
	// 填充字段
	Mentor   interface{} `json:"mentor"`
	Student  interface{} `json:"student"`
	TimeSlot interface{} `json:"time_slot"`
}
