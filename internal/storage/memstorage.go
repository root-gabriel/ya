package storage

import (
	"fmt"
	"net/http"
)

type gauge float64
type counter int64

type MemStorage struct {
	GaugeData   map[string]gauge
	CounterData map[string]counter
}

func NewMem() *MemStorage {
	return &MemStorage{
		GaugeData:   make(map[string]gauge),
		CounterData: make(map[string]counter),
	}
}

func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.CounterData[name] += counter(value)
}

func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.GaugeData[name] = gauge(value)
}

func (s *MemStorage) GetValue(t string, n string) (string, int) {
	var v string
	statusCode := http.StatusOK
	if val, ok := s.GaugeData[n]; ok && t == "gauge" {
		v = fmt.Sprint(val)
	} else if val, ok := s.CounterData[n]; ok && t == "counter" {
		v = fmt.Sprint(val)
	} else {
		statusCode = http.StatusNotFound
	}
	return v, statusCode
}

