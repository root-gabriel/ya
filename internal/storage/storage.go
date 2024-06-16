package storage

import (
    "fmt"
    "net/http"
    "sync"
)

var metrics = make(map[string]int)
var mutex = &sync.Mutex{}

// UpdateMetricHandler обновляет значение метрики
func UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
    // Получаем метрику и ее новое значение из запроса
    metric := r.URL.Query().Get("metric")
    value := r.URL.Query().Get("value")

    // Блокировка для синхронизации доступа к map
    mutex.Lock()
    metrics[metric] = value
    mutex.Unlock()

    fmt.Fprintf(w, "Updated metric %s to %s", metric, value)
}

// GetMetricHandler возвращает значение метрики
func GetMetricHandler(w http.ResponseWriter, r *http.Request) {
    metric := r.URL.Query().Get("metric")

    mutex.Lock()
    value, exists := metrics[metric]
    mutex.Unlock()

    if !exists {
        http.Error(w, "Metric not found", 404)
        return
    }

    fmt.Fprintf(w, "Value of metric %s is %d", metric, value)
}

