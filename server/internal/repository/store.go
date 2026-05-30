package repository

import (
	"sync"

	"interview-server/internal/model"
)

// Store 共享内存存储，所有仓库共用一把锁
type Store struct {
	mu sync.RWMutex

	Users         map[string]*model.User         // key = email
	Availabilities map[string]*model.Availability // key = id
	Appointments  map[string]*model.Appointment  // key = id
	VerifyTokens  map[string]string              // key = token, value = email（快速查找）
}

// NewStore 创建共享存储
func NewStore() *Store {
	return &Store{
		Users:         make(map[string]*model.User),
		Availabilities: make(map[string]*model.Availability),
		Appointments:  make(map[string]*model.Appointment),
		VerifyTokens:  make(map[string]string),
	}
}

// Lock 写锁
func (s *Store) Lock()   { s.mu.Lock() }

// Unlock 解写锁
func (s *Store) Unlock() { s.mu.Unlock() }

// RLock 读锁
func (s *Store) RLock()   { s.mu.RLock() }

// RUnlock 解读锁
func (s *Store) RUnlock() { s.mu.RUnlock() }
