package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/labstack/echo/v4"
    "github.com/root-gabriel/ya/internal/models"
    "github.com/root-gabriel/ya/internal/storage"
)

type Handler struct {
    Storage storage.Storage
}

func NewHandler(storage storage.Storage) *Handler {
    return &Handler{Storage: storage}
}

func (h *Handler) UpdateMetrics(c echo.Context) error {
    var m models.Metrics
    if err := json.NewDecoder(c.Request().Body).Decode(&m); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
    }

    switch m.MType {
    case "counter":
        if m.Delta == nil {
            return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid counter value"})
        }
        h.Storage.UpdateCounter(m.ID, *m.Delta)
    case "gauge":
        if m.Value == nil {
            return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid gauge value"})
        }
        h.Storage.UpdateGauge(m.ID, *m.Value)
    default:
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid metric type"})
    }

    return c.JSON(http.StatusOK, m)
}

func (h *Handler) GetMetricsValue(c echo.Context) error {
    var m models.Metrics
    if err := json.NewDecoder(c.Request().Body).Decode(&m); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
    }

    var value interface{}
    var status int
    switch m.MType {
    case "counter":
        value, status = h.Storage.GetCounterValue(m.ID)
    case "gauge":
        value, status = h.Storage.GetGaugeValue(m.ID)
    default:
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid metric type"})
    }

    if status != http.StatusOK {
        return c.JSON(status, map[string]string{"error": "Metric not found"})
    }

    return c.JSON(http.StatusOK, map[string]interface{}{"value": value})
}

func (h *Handler) AllMetrics(c echo.Context) error {
    allMetrics := h.Storage.AllMetrics()
    c.Response().Header().Set("Content-Type", "application/json")
    return c.String(http.StatusOK, allMetrics)
}

