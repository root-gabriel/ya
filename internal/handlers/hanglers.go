package handlers

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/lionslon/go-yapmetrics/internal/models"
	"github.com/lionslon/go-yapmetrics/internal/storage"
	"net/http"
)

type handler struct {
	store *storage.MemStorage
}

func New(stor *storage.MemStorage) *handler {
	return &handler{
		store: stor,
	}
}

func (h *handler) UpdateMetrics() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var m models.Metrics
		if err := json.NewDecoder(ctx.Request().Body).Decode(&m); err != nil {
			return ctx.JSON(http.StatusBadRequest, "Invalid request payload")
		}

		if m.MType == "counter" {
			if m.Delta == nil {
				return ctx.JSON(http.StatusBadRequest, "Invalid counter value")
			}
			h.store.UpdateCounter(m.ID, *m.Delta)
		} else if m.MType == "gauge" {
			if m.Value == nil {
				return ctx.JSON(http.StatusBadRequest, "Invalid gauge value")
			}
			h.store.UpdateGauge(m.ID, *m.Value)
		} else {
			return ctx.JSON(http.StatusBadRequest, "Invalid metric type")
		}

		return ctx.JSON(http.StatusOK, m)
	}
}

func (h *handler) MetricsValue() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var m models.Metrics
		if err := json.NewDecoder(ctx.Request().Body).Decode(&m); err != nil {
			return ctx.JSON(http.StatusBadRequest, "Invalid request payload")
		}

		var value interface{}
		var status int
		if m.MType == "counter" {
			value, status = h.store.GetCounterValue(m.ID)
		} else if m.MType == "gauge" {
			value, status = h.store.GetGaugeValue(m.ID)
		} else {
			return ctx.JSON(http.StatusBadRequest, "Invalid metric type")
		}

		if status != http.StatusOK {
			return ctx.JSON(status, "Metric not found")
		}

		return ctx.JSON(http.StatusOK, map[string]interface{}{"value": value})
	}
}

func (h *handler) AllMetrics() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		allMetrics := h.store.AllMetrics()
		ctx.Response().Header().Set("Content-Type", "application/json")
		return ctx.String(http.StatusOK, allMetrics)
	}
}

