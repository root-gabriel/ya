package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/root-gabriel/ya/internal/storage"
)

// UpdateCounterHandler обновляет значение счетчика
func UpdateCounterHandler(storage *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		valueStr := vars["value"]
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid value", http.StatusBadRequest)
			return
		}
		storage.UpdateCounter(name, value)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}
}

// GetCounterValueHandler получает значение счетчика
func GetCounterValueHandler(storage *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		value := storage.GetCounterValue(name)
		json.NewEncoder(w).Encode(map[string]int64{"value": value})
	}
}

// PingHandler проверяет доступность сервера
func PingHandler(storage *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Pong"))
	}
}

