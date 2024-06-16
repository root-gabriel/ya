package main

import (
    "log"
    "net/http"
    "metrics/internal/storage"
)

func main() {
    http.HandleFunc("/update", storage.UpdateMetricHandler) // Обработчик для обновления метрик
    http.HandleFunc("/value", storage.GetMetricHandler)     // Обработчик для получения значения метрики

    log.Println("Server is starting on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

