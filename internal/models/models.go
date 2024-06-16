package models

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type DBMetric struct {
	Metrictype string  `json:"metrictype"`
	Metricname string  `json:"metricname"`
	Counter    int64   `json:"counter"`
	Gauge      float64 `json:"gauge"`
	Email      string  `json:"email"`
}
