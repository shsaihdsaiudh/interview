package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"interview-server/internal/model"
)

var ErrAvailabilityNotFound = errors.New("空闲时间段不存在")

// AvailabilityRepo 空闲时间仓库（PostgreSQL）
type AvailabilityRepo struct {
	pool *pgxpool.Pool
}

// NewAvailabilityRepo 创建空闲时间仓库
func NewAvailabilityRepo(pool *pgxpool.Pool) *AvailabilityRepo {
	return &AvailabilityRepo{pool: pool}
}

// Create 添加空闲时间
func (r *AvailabilityRepo) Create(a *model.Availability) error {
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO availabilities (id, user_id, date, start_time, end_time)
		 VALUES ($1,$2,$3,$4,$5)`,
		a.ID, a.UserID, a.Date, a.StartTime, a.EndTime,
	)
	return err
}

// Delete 删除空闲时间
func (r *AvailabilityRepo) Delete(id string) error {
	tag, err := r.pool.Exec(context.Background(),
		`DELETE FROM availabilities WHERE id = $1`, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAvailabilityNotFound
	}
	return nil
}

// FindByID 按 ID 查找
func (r *AvailabilityRepo) FindByID(id string) (*model.Availability, error) {
	a := &model.Availability{}
	err := r.pool.QueryRow(context.Background(),
		`SELECT id, user_id, date::text, start_time::text, end_time::text
		 FROM availabilities WHERE id = $1`, id,
	).Scan(&a.ID, &a.UserID, &a.Date, &a.StartTime, &a.EndTime)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAvailabilityNotFound
	}
	if err != nil {
		return nil, err
	}
	// 截取时间部分 HH:MM
	if len(a.StartTime) >= 5 {
		a.StartTime = a.StartTime[:5]
	}
	if len(a.EndTime) >= 5 {
		a.EndTime = a.EndTime[:5]
	}
	return a, nil
}

// FindByUserID 查找某用户的所有空闲时间
func (r *AvailabilityRepo) FindByUserID(userID string) []*model.Availability {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id, user_id, date::text, start_time::text, end_time::text
		 FROM availabilities WHERE user_id = $1
		 ORDER BY date, start_time`, userID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []*model.Availability
	for rows.Next() {
		a := &model.Availability{}
		if err := rows.Scan(&a.ID, &a.UserID, &a.Date, &a.StartTime, &a.EndTime); err != nil {
			continue
		}
		if len(a.StartTime) >= 5 {
			a.StartTime = a.StartTime[:5]
		}
		if len(a.EndTime) >= 5 {
			a.EndTime = a.EndTime[:5]
		}
		result = append(result, a)
	}
	return result
}
