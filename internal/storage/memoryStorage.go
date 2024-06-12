package storage

import (
    "encoding/json"
    "sync"
    "github.com/root-gabriel/ya/internal/models"
    "net/http"
)

type Storage interface {
    UpdateCounter(string, int64)
    UpdateGauge(string, float64)
    GetCounterValue(string) (int64, int)
    GetGaugeValue(string) (float64, int)
    AllMetrics() string
    GetCounterData() map[string]int64
    GetGaugeData() map[string]float64
}

type MemoryStorage struct {
    mu       sync.RWMutex
    counters map[string]int64
    gauges   map[string]float64
}

func NewMemoryStorage() *MemoryStorage {
    return &MemoryStorage{
        counters: make(map[string]int64),
        gauges:   make(map[string]float64),
    }
}

func (m *MemoryStorage) UpdateCounter(key string, value int64) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.counters[key] += value
}

func (m *MemoryStorage) UpdateGauge(key string, value float64) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.gauges[key] = value
}

func (m *MemoryStorage) GetCounterValue(key string) (int64, int) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    value, ok := m.counters[key]
    if !ok {
        return 0, http.StatusNotFound
    }
    return value, http.StatusOK
}

func (m *MemoryStorage) GetGaugeValue(key string) (float64, int) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    value, ok := m.gauges[key]
    if !ok {
        return 0, http.StatusNotFound
    }
    return value, http.StatusOK
}

func (m *MemoryStorage) AllMetrics() string {
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

func (m *MemoryStorage) GetCounterData() map[string]int64 {
    m.mu.RLock()
    defer m.mu.RUnlock()
    data := make(map[string]int64, len(m.counters))
    for k, v := range m.counters {
        data[k] = v
    }
    return data
}

func (m *MemoryStorage) GetGaugeData() map[string]float64 {
    m.mu.RLock()
    defer m.mu.RUnlock()
    data := make(map[string]float64, len(m.gauges))
    for k, v := range m.gauges {
        data[k] = v
    }
    return data
}

