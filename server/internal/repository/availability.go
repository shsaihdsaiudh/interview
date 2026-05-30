package repository

import (
	"errors"

	"interview-server/internal/model"
)

var ErrAvailabilityNotFound = errors.New("空闲时间段不存在")

// AvailabilityRepo 空闲时间仓库
type AvailabilityRepo struct {
	store *Store
}

// NewAvailabilityRepo 创建空闲时间仓库
func NewAvailabilityRepo(store *Store) *AvailabilityRepo {
	return &AvailabilityRepo{store: store}
}

// Create 添加空闲时间
func (r *AvailabilityRepo) Create(a *model.Availability) error {
	r.store.Lock()
	defer r.store.Unlock()

	r.store.Availabilities[a.ID] = a
	return nil
}

// Delete 删除空闲时间
func (r *AvailabilityRepo) Delete(id string) error {
	r.store.Lock()
	defer r.store.Unlock()

	if _, exists := r.store.Availabilities[id]; !exists {
		return ErrAvailabilityNotFound
	}
	delete(r.store.Availabilities, id)
	return nil
}

// FindByID 按 ID 查找
func (r *AvailabilityRepo) FindByID(id string) (*model.Availability, error) {
	r.store.RLock()
	defer r.store.RUnlock()

	a, exists := r.store.Availabilities[id]
	if !exists {
		return nil, ErrAvailabilityNotFound
	}
	return a, nil
}

// FindByUserID 查找某用户的所有空闲时间
func (r *AvailabilityRepo) FindByUserID(userID string) []*model.Availability {
	r.store.RLock()
	defer r.store.RUnlock()

	var result []*model.Availability
	for _, a := range r.store.Availabilities {
		if a.UserID == userID {
			result = append(result, a)
		}
	}
	return result
}
