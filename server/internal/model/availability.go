package model

// Availability 用户空闲时间段
type Availability struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`    // 用户 email
	Date      string `json:"date"`       // "2006-01-02" 格式
	StartTime string `json:"start_time"` // "14:00"
	EndTime   string `json:"end_time"`   // "15:00"
}

// AddAvailabilityRequest 添加空闲时间请求
type AddAvailabilityRequest struct {
	Date      string `json:"date" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}
