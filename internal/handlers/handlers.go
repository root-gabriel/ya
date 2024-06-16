package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/root-gabriel/ya/internal/storage"
	"net/http"
	"strconv"
)

type Handler struct {
	storage *storage.MemStorage
}

func NewHandler(storage *storage.MemStorage) *Handler {
	return &Handler{storage: storage}
}

func (h *Handler) UpdateMetric(ctx echo.Context) error {
	metricType := ctx.Param("typeM")
	metricName := ctx.Param("nameM")
	metricValue := ctx.Param("valueM")

	switch metricType {
	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid counter value"})
		}
		h.storage.UpdateCounter(metricName, value)
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid gauge value"})
		}
		h.storage.UpdateGauge(metricName, value)
	default:
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid metric type"})
	}

	return ctx.JSON(http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) GetMetric(ctx echo.Context) error {
	metricType := ctx.Param("typeM")
	metricName := ctx.Param("nameM")

	value, status := h.storage.GetValue(metricType, metricName)
	if status == http.StatusNotFound {
		return ctx.JSON(status, map[string]string{"error": "metric not found"})
	}

	return ctx.String(status, value)
}

