package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"interview-server/internal/model"
	"interview-server/internal/repository"
)

var (
	ErrSlotNotFound       = errors.New("空闲时间段不存在")
	ErrSlotNotOwned       = errors.New("只能删除自己的空闲时间")
	ErrTimeConflict       = errors.New("结束时间必须晚于开始时间")
	ErrPastDate           = errors.New("不能设置过去的时间段")
	ErrInvalidTimeFormat  = errors.New("时间格式不正确，请使用 HH:MM 格式")
)

// AvailabilityService 空闲时间业务逻辑
type AvailabilityService struct {
	availRepo *repository.AvailabilityRepo
}

// NewAvailabilityService 创建空闲时间服务
func NewAvailabilityService(availRepo *repository.AvailabilityRepo) *AvailabilityService {
	return &AvailabilityService{availRepo: availRepo}
}

// GetMyAvailability 获取自己的空闲时间列表
func (s *AvailabilityService) GetMyAvailability(userID string) []*model.Availability {
	slots := s.availRepo.FindByUserID(userID)
	if slots == nil {
		return []*model.Availability{}
	}
	return slots
}

// AddAvailability 添加空闲时间
func (s *AvailabilityService) AddAvailability(userID string, req model.AddAvailabilityRequest) (*model.Availability, error) {
	// 校验时间格式
	if err := validateTime(req.StartTime); err != nil {
		return nil, err
	}
	if err := validateTime(req.EndTime); err != nil {
		return nil, err
	}

	// 结束时间必须晚于开始时间
	if req.EndTime <= req.StartTime {
		return nil, ErrTimeConflict
	}

	// 日期不能是过去
	slotDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("日期格式不正确，请使用 YYYY-MM-DD 格式")
	}
	today := time.Now().Truncate(24 * time.Hour)
	if slotDate.Before(today) {
		return nil, ErrPastDate
	}

	id, _ := generateID()
	slot := &model.Availability{
		ID:        id,
		UserID:    userID,
		Date:      req.Date,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	if err := s.availRepo.Create(slot); err != nil {
		return nil, err
	}
	return slot, nil
}

// DeleteAvailability 删除空闲时间（需验证所有权）
func (s *AvailabilityService) DeleteAvailability(userID, slotID string) error {
	slot, err := s.availRepo.FindByID(slotID)
	if err != nil {
		return ErrSlotNotFound
	}
	if slot.UserID != userID {
		return ErrSlotNotOwned
	}
	return s.availRepo.Delete(slotID)
}

// FindByID 按 ID 查找空闲时间
func (s *AvailabilityService) FindByID(id string) (*model.Availability, error) {
	return s.availRepo.FindByID(id)
}

// validateTime 校验时间格式 HH:MM
func validateTime(t string) error {
	if len(t) != 5 || t[2] != ':' {
		return ErrInvalidTimeFormat
	}
	_, err := time.Parse("15:04", t)
	if err != nil {
		return ErrInvalidTimeFormat
	}
	return nil
}

// generateID 生成随机 hex ID
func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
