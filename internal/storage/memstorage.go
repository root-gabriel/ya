package storage

import (
	"sync"
)

// MemStorage - структура для хранения метрик в памяти
type MemStorage struct {
	mu       sync.RWMutex
	counters map[string]int64
}

// NewMem создает новое хранилище в памяти
func NewMem() *MemStorage {
	return &MemStorage{
		counters: make(map[string]int64),
	}
}

// UpdateCounter обновляет значение счетчика
func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[name] += value
}

// GetCounterValue возвращает значение счетчика
func (s *MemStorage) GetCounterValue(name string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.counters[name]
}

