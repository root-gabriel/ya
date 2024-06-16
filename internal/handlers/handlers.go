package handlers

import (
	"net/http"
	"strconv"
	"github.com/labstack/echo/v4"
	"github.com/root-gabriel/ya/internal/storage"
)

type Handler struct {
	store *storage.MemStorage
}

func New(store *storage.MemStorage) *Handler {
	return &Handler{store: store}
}

func (h *Handler) UpdateMetrics() echo.HandlerFunc {
	return func(c echo.Context) error {
		metricType := c.Param("type")
		metricName := c.Param("name")
		metricValue := c.Param("value")

		switch metricType {
		case "counter":
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				return c.String(http.StatusBadRequest, err.Error())
			}
			h.store.UpdateCounter(metricName, value)
		case "gauge":
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				return c.String(http.StatusBadRequest, err.Error())
			}
			h.store.UpdateGauge(metricName, value)
		default:
			return c.String(http.StatusBadRequest, "invalid metric type")
		}

		return c.String(http.StatusOK, "success")
	}
}

func (h *Handler) MetricsValue() echo.HandlerFunc {
	return func(c echo.Context) error {
		metricType := c.Param("type")
		metricName := c.Param("name")

		value, status := h.store.GetValue(metricType, metricName)
		return c.String(status, value)
	}
}

