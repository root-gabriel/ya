package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/lionslon/go-yapmetrics/internal/models"
	"github.com/lionslon/go-yapmetrics/internal/storage"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

type storageUpdater interface {
	UpdateCounter(string, int64)
	UpdateGauge(string, float64)
	GetValue(string, string) (string, int)
	AllMetrics() string
	GetCounterValue(string) int64
	GetGaugeValue(string) float64
	StoreBatch([]models.Metrics)
}

type handler struct {
	store storageUpdater
}

func New(stor *storage.MemStorage) *handler {
	return &handler{
		store: stor,
	}
}

// UpdateMetrics handles POST requests to update metric values
func (h *handler) UpdateMetrics() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var m models.Metrics
		if err := json.NewDecoder(ctx.Request().Body).Decode(&m); err != nil {
			return ctx.JSON(http.StatusBadRequest, "Invalid request payload")
		}
		
		if m.MType == "counter" {
			value, err := strconv.ParseInt(*m.Delta, 10, 64)
			if err != nil {
				return ctx.JSON(http.StatusBadRequest, "Invalid counter value")
			}
			h.store.UpdateCounter(m.ID, value)
		} else if m.MType == "gauge" {
			value, err := strconv.ParseFloat(*m.Value, 64)
			if err != nil {
				return ctx.JSON(http.StatusBadRequest, "Invalid gauge value")
			}
			h.store.UpdateGauge(m.ID, value)
		} else {
			return ctx.JSON(http.StatusBadRequest, "Invalid metric type")
		}

		return ctx.JSON(http.StatusOK, m)
	}
}

// MetricsValue handles POST requests to get metric values
func (h *handler) MetricsValue() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var m models.Metrics
		if err := json.NewDecoder(ctx.Request().Body).Decode(&m); err != nil {
			return ctx.JSON(http.StatusBadRequest, "Invalid request payload")
		}

		var value string
		var status int
		if m.MType == "counter" {
			value, status = h.store.GetValue("counter", m.ID)
		} else if m.MType == "gauge" {
			value, status = h.store.GetValue("gauge", m.ID)
		} else {
			return ctx.JSON(http.StatusBadRequest, "Invalid metric type")
		}

		if status != http.StatusOK {
			return ctx.JSON(status, "Metric not found")
		}

		return ctx.JSON(http.StatusOK, value)
	}
}

// AllMetrics handles GET requests to retrieve all metrics
func (h *handler) AllMetrics() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		allMetrics := h.store.AllMetrics()
		ctx.Response().Header().Set("Content-Type", "application/json")
		return ctx.String(http.StatusOK, allMetrics)
	}
}

