package storage

import (
    "encoding/json"
    "sync"
    "github.com/root-gabriel/ya/internal/models"
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

func (m *MemStorage) UpdateCounter(key string, value int64) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.counters[key] += value
}

func (m *MemStorage) UpdateGauge(key string, value float64) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.gauges[key] = value
}

func (m *MemStorage) GetCounterValue(key string) (int64, int) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    value, ok := m.counters[key]
    if !ok {
        return 0, http.StatusNotFound
    }
    return value, http.StatusOK
}

func (m *MemStorage) GetGaugeValue(key string) (float64, int) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    value, ok := m.gauges[key]
    if !ok {
        return 0, http.StatusNotFound
    }
    return value, http.StatusOK
}

func (m *MemStorage) AllMetrics() string {
    m.mu.RLock()
    defer m.mu.RUnlock()

    allMetrics := make([]models.Metrics, 0, len(m.counters)+len(m.gauges))

    for key, value := range m.counters {
        delta := value
        allMetrics = append(allMetrics, models.Metrics{
            ID:    key,
            MType: "counter",
            Delta: &delta,
        })
    }

    for key, value := range m.gauges {
        val := value
        allMetrics = append(allMetrics, models.Metrics{
            ID:    key,
            MType: "gauge",
            Value: &val,
        })
    }

    jsonMetrics, _ := json.Marshal(allMetrics)
    return string(jsonMetrics)
}

