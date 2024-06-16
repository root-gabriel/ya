package storage

import (
	"net/http"
	"strconv"
)

type gauge float64
type counter int64

type MemStorage struct {
	Gauges   map[string]gauge
	Counters map[string]counter
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]gauge),
		Counters: make(map[string]counter),
	}
}

func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.Counters[name] += counter(value)
}

func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.Gauges[name] = gauge(value)
}

func (s *MemStorage) GetValue(metricType, name string) (string, int) {
	switch metricType {
	case "counter":
		if val, ok := s.Counters[name]; ok {
			return strconv.FormatInt(int64(val), 10), http.StatusOK
		}
	case "gauge":
		if val, ok := s.Gauges[name]; ok {
			return strconv.FormatFloat(float64(val), 'f', -1, 64), http.StatusOK
		}
	}
	return "", http.StatusNotFound
}

