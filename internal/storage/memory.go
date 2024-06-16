package storage

import (
	"fmt"
	"sync"

	"github.com/DeneesK/metrics-alerting/internal/models"
	"go.uber.org/zap"
)

const (
	counterMetric string = "counter"
	gaugeMetric   string = "gauge"
)

type Result struct {
	Counter int64
	Gauge   float64
}

type counter struct {
	mx sync.Mutex
	c  map[string]int64
}

type gauge struct {
	mx sync.Mutex
	g  map[string]float64
}

func (c *counter) Load(key string) (int64, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()
	val, ok := c.c[key]
	return val, ok
}

func (c *counter) set(m map[string]int64) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.c = m
}

func (c *counter) LoadAll() map[string]int64 {
	c.mx.Lock()
	defer c.mx.Unlock()
	cCopy := make(map[string]int64)
	for k, v := range c.c {
		cCopy[k] = v
	}
	return cCopy
}

func (c *counter) Store(key string, value int64) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.c[key] += value
}

func (g *gauge) Load(key string) (float64, bool) {
	g.mx.Lock()
	defer g.mx.Unlock()
	val, ok := g.g[key]
	return val, ok
}

func (g *gauge) set(m map[string]float64) {
	g.mx.Lock()
	defer g.mx.Unlock()
	g.g = m
}

func (g *gauge) LoadAll() map[string]float64 {
	g.mx.Lock()
	defer g.mx.Unlock()
	gCopy := make(map[string]float64)
	for k, v := range g.g {
		gCopy[k] = v
	}
	return gCopy
}

func (g *gauge) Store(key string, value float64) {
	g.mx.Lock()
	defer g.mx.Unlock()
	g.g[key] = value
}

type MemStorage struct {
	gauge   gauge
	counter counter
	log     *zap.SugaredLogger
}

func NewMemStorage(log *zap.SugaredLogger) (*MemStorage, error) {
	ms := MemStorage{
		gauge:   gauge{g: make(map[string]float64)},
		counter: counter{c: make(map[string]int64)},
		log:     log,
	}
	return &ms, nil
}

func (storage *MemStorage) Store(metricType, name string, value interface{}) error {
	switch metricType {
	case counterMetric:
		v, ok := value.(int64)
		if !ok {
			return fmt.Errorf("value cannot be cast to a specific type")
		}
		storage.counter.Store(name, v)
		return nil
	case gaugeMetric:
		v, ok := value.(float64)
		if !ok {
			return fmt.Errorf("value cannot be cast to a specific type")
		}
		storage.gauge.Store(name, v)
		return nil
	default:
		return fmt.Errorf("metric type does not exist, given type: %v", metricType)
	}
}

func (storage *MemStorage) StoreBatch(metrics []models.Metrics) error {
	for _, m := range metrics {
		if m.MType == "gauge" && m.Value != nil {
			err := storage.Store(m.MType, m.ID, *m.Value)
			if err != nil {
				return fmt.Errorf("store batch to memory failed with error: %w", err)
			}
		} else if m.MType == "counter" && m.Delta != nil {
			err := storage.Store(m.MType, m.ID, *m.Delta)
			if err != nil {
				return fmt.Errorf("store batch to memory failed with error: %w", err)
			}
		} else {
			return fmt.Errorf("attempt to store batch to memory failed, value - %v, delta - %v, metrictype - %v", m.Value, m.ID, m.MType)
		}

	}
	return nil
}

func (storage *MemStorage) GetValue(metricType, name string) (Result, bool, error) {
	switch metricType {
	case counterMetric:
		v, ok := storage.counter.Load(name)
		if !ok {
			return Result{0, 0}, false, nil
		}
		return Result{Counter: v, Gauge: 0}, true, nil
	case gaugeMetric:
		v, ok := storage.gauge.Load(name)
		if !ok {
			return Result{0, 0}, false, nil
		}
		return Result{Counter: 0, Gauge: v}, true, nil
	}
	return Result{0, 0}, false, fmt.Errorf("metric type does not exist, given type: %v", metricType)
}

func (storage *MemStorage) GetCounterMetrics() (map[string]int64, error) {
	return storage.counter.LoadAll(), nil
}

func (storage *MemStorage) GetGaugeMetrics() (map[string]float64, error) {
	return storage.gauge.LoadAll(), nil
}

func (storage *MemStorage) Ping() error {
	return nil
}

func (storage *MemStorage) Close() error {
	return nil
}

func (storage *MemStorage) setMetrics(metrics *allMetrics) {
	storage.counter.set(metrics.Counter)
	storage.gauge.set(metrics.Gauge)
}
