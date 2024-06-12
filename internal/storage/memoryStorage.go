package storage

import (
	"sync"
)

type MemStorage struct {
	mu       sync.RWMutex
	counters map[string]int64
	gauges   map[string]float64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

func (s *MemStorage) UpdateCounter(id string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[id] += value
}

func (s *MemStorage) UpdateGauge(id string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[id] = value
}

func (s *MemStorage) GetCounterValue(id string) (int64, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.counters[id]
	if !ok {
		return 0, http.StatusNotFound
	}
	return value, http.StatusOK
}

func (s *MemStorage) GetGaugeValue(id string) (float64, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.gauges[id]
	if !ok {
		return 0, http.StatusNotFound
	}
	return value, http.StatusOK
}

func (s *MemStorage) AllMetrics() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Возвращаем все метрики в виде JSON-строки
	// Реализация зависит от вашего формата
}

func (s *MemStorage) StoreBatch(metrics []models.Metrics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, m := range metrics {
		if m.MType == "counter" && m.Delta != nil {
			s.counters[m.ID] += *m.Delta
		} else if m.MType == "gauge" && m.Value != nil {
			s.gauges[m.ID] = *m.Value
		}
	}
}

