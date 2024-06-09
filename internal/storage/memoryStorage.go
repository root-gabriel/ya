package storage

import (
	"fmt"
	"github.com/lionslon/go-yapmetrics/internal/models"
	"net/http"
)

type gauge float64
type counter int64

type MemStorage struct {
	GaugeData   map[string]gauge   `json:"gauge"`
	CounterData map[string]counter `json:"counter"`
}

//type AllMetrics struct {
//	Gauge   map[string]gauge   `json:"gauge"`
//	Counter map[string]counter `json:"counter"`
//}

func NewMem() *MemStorage {
	storage := MemStorage{
		GaugeData:   make(map[string]gauge),
		CounterData: make(map[string]counter),
	}

	return &storage
}

func (s *MemStorage) UpdateCounter(n string, v int64) {
	s.CounterData[n] += counter(v)
}

func (s *MemStorage) UpdateGauge(n string, v float64) {
	s.GaugeData[n] = gauge(v)
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

func (s *MemStorage) AllMetrics() string {
	var result string
	result += "Gauge metrics:\n"
	for n, v := range s.GaugeData {
		result += fmt.Sprintf("- %s = %f\n", n, v)
	}

	result += "Counter metrics:\n"
	for n, v := range s.CounterData {
		result += fmt.Sprintf("- %s = %d\n", n, v)
	}

	return result
}

func (s *MemStorage) GetCounterValue(id string) int64 {
	return int64(s.CounterData[id])
}

func (s *MemStorage) GetGaugeValue(id string) float64 {
	return float64(s.GaugeData[id])
}

func (s *MemStorage) GetCounterData() map[string]counter {
	return s.CounterData
}

func (s *MemStorage) GetGaugeData() map[string]gauge {
	return s.GaugeData
}

func (s *MemStorage) UpdateGaugeData(gaugeData map[string]gauge) {
	s.GaugeData = gaugeData
}

func (s *MemStorage) UpdateCounterData(counterData map[string]counter) {
	s.CounterData = counterData
}

func (s *MemStorage) StoreBatch(metrics []models.Metrics) {
	for _, m := range metrics {
		switch m.MType {
		case "counter":
			s.UpdateCounter(m.ID, *m.Delta)
		case "gauge":
			s.UpdateGauge(m.ID, *m.Value)
		}

	}
}
