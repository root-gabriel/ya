package server

import (
    "encoding/json"
    "github.com/labstack/echo/v4"
    "github.com/root-gabriel/ya/internal/models"
    "net/http"
)

type handler struct {
    storage Storage
}

func NewHandler(storage Storage) *handler {
    return &handler{storage: storage}
}

func (h *handler) UpdateMetrics(c echo.Context) error {
    var m models.Metrics
    if err := json.NewDecoder(c.Request().Body).Decode(&m); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
    }

    switch m.MType {
    case "counter":
        if m.Delta == nil {
            return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid counter value"})
        }
        h.storage.UpdateCounter(m.ID, *m.Delta)
    case "gauge":
        if m.Value == nil {
            return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid gauge value"})
        }
        h.storage.UpdateGauge(m.ID, *m.Value)
    default:
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid metric type"})
    }

    return c.JSON(http.StatusOK, m)
}

func (h *handler) GetMetricsValue(c echo.Context) error {
    var m models.Metrics
    if err := json.NewDecoder(c.Request().Body).Decode(&m); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
    }

    var value interface{}
    var status int
    switch m.MType {
    case "counter":
        value, status = h.storage.GetCounterValue(m.ID)
    case "gauge":
        value, status = h.storage.GetGaugeValue(m.ID)
    default:
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid metric type"})
    }

    if status != http.StatusOK {
        return c.JSON(status, map[string]string{"error": "Metric not found"})
    }

    return c.JSON(http.StatusOK, map[string]interface{}{"value": value})
}

func (h *handler) AllMetrics(c echo.Context) error {
    allMetrics := h.storage.AllMetrics()
    c.Response().Header().Set("Content-Type", "application/json")
    return c.String(http.StatusOK, allMetrics)
}

