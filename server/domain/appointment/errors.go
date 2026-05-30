package appointment

import "errors"

// 领域错误
var (
	ErrAppointmentNotFound = errors.New("预约不存在")
	ErrSlotNotFound        = errors.New("空闲时间段不存在")
	ErrSlotNotOwned        = errors.New("只能删除自己的空闲时间")
	ErrSlotAlreadyBooked   = errors.New("该时间段已被预约")
	ErrCannotBookOwnSlot   = errors.New("不能预约自己的空闲时间")
	ErrNotPending          = errors.New("只能处理待确认的预约")
	ErrNotMentor           = errors.New("只有时间所有者才能操作此预约")
	ErrMentorNotFound      = errors.New("mentor 不存在")
	ErrTimeConflict        = errors.New("结束时间必须晚于开始时间")
	ErrPastDate            = errors.New("不能设置过去的时间段")
	ErrInvalidTimeFormat   = errors.New("时间格式不正确，请使用 HH:MM 格式")
)
