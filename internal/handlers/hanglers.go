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
    "strings"
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

func (h *handler) UpdateMetrics() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		metricsType := ctx.Param("typeM")
		metricsName := ctx.Param("nameM")
		metricsValue := ctx.Param("valueM")

		switch metricsType {
		case "counter":
			value, err := strconv.ParseInt(metricsValue, 10, 64)
			if err != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("%s cannot be converted to an integer", metricsValue)})
			}
			h.store.UpdateCounter(metricsName, value)
		case "gauge":
			value, err := strconv.ParseFloat(metricsValue, 64)
			if err != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("%s cannot be converted to a float", metricsValue)})
			}
			h.store.UpdateGauge(metricsName, value)
		default:
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid metric type. Can only be 'gauge' or 'counter'"})
		}

		acceptHeader := ctx.Request().Header.Get("Accept")
		if acceptHeader == "application/json" {
			return ctx.JSON(http.StatusOK, map[string]string{"status": "success"})
		}

		return ctx.String(http.StatusOK, "success")
	}
}

func (h *handler) MetricsValue() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		typeM := ctx.Param("typeM")
		nameM := ctx.Param("nameM")

		val, status := h.store.GetValue(typeM, nameM)
		if status != http.StatusOK {
			return ctx.JSON(status, map[string]string{"error": "Metric not found"})
		}

		acceptHeader := ctx.Request().Header.Get("Accept")
		if acceptHeader == "application/json" {
			if typeM == "counter" {
				delta, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse counter value"})
				}
				return ctx.JSON(status, map[string]interface{}{"id": nameM, "type": typeM, "delta": delta})
			}
			if typeM == "gauge" {
				value, err := strconv.ParseFloat(val, 64)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse gauge value"})
				}
				return ctx.JSON(status, map[string]interface{}{"id": nameM, "type": typeM, "value": value})
			}
		}

		return ctx.String(status, val)
	}
}

func (h *handler) AllMetricsValues() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		metrics := h.store.AllMetrics()

		acceptHeader := ctx.Request().Header.Get("Accept")
		if acceptHeader == "application/json" {
			metricsMap := make(map[string]interface{})
			
			metricsList := strings.Split(metrics, "\n")
			for _, metric := range metricsList {
				if metric == "" {
					continue
				}
				parts := strings.Split(metric, " = ")
				if len(parts) != 2 {
					continue
				}
				metricName := strings.TrimSpace(parts[0])
				metricValue := strings.TrimSpace(parts[1])
				if strings.HasPrefix(metricName, "- ") {
					metricName = metricName[2:]
				}
				if strings.HasPrefix(metricName, "gauge") {
					value, err := strconv.ParseFloat(metricValue, 64)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse gauge value"})
					}
					metricsMap[metricName] = value
				} else if strings.HasPrefix(metricName, "counter") {
					value, err := strconv.ParseInt(metricValue, 10, 64)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse counter value"})
					}
					metricsMap[metricName] = value
				} else {
					metricsMap[metricName] = metricValue
				}
			}

			return ctx.JSON(http.StatusOK, metricsMap)
		}

		ctx.Response().Header().Set("Content-Type", "text/html")
		return ctx.String(http.StatusOK, metrics)
	}
}

func (h *handler) UpdateJSON() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var metric models.Metrics

		err := json.NewDecoder(ctx.Request().Body).Decode(&metric)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Error in JSON decode: %s", err)})
		}

		switch metric.MType {
		case "counter":
			h.store.UpdateCounter(metric.ID, *metric.Delta)
		case "gauge":
			h.store.UpdateGauge(metric.ID, *metric.Value)
		default:
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Invalid metric type. Can only be 'gauge' or 'counter'"})
		}

		ctx.Response().Header().Set("Content-Type", "application/json")
		return ctx.JSON(http.StatusOK, metric)
	}
}

func (h *handler) GetValueJSON() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ctx.Response().Header().Set("Content-Type", "application/json")
		var metric models.Metrics
		err := json.NewDecoder(ctx.Request().Body).Decode(&metric)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Error in JSON decode: %s", err)})
		}

		switch metric.MType {
		case "counter":
			value := h.store.GetCounterValue(metric.ID)
			metric.Delta = &value
		case "gauge":
			value := h.store.GetGaugeValue(metric.ID)
			metric.Value = &value
		default:
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Invalid metric type. Can only be 'gauge' or 'counter'"})
		}

		return ctx.JSON(http.StatusOK, metric)
	}
}

func (h *handler) PingDB(sw storage.StorageWorker) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		err := sw.Check()
		ctx.Response().Header().Set("Content-Type", "text/html")
		if err == nil {
			return ctx.String(http.StatusOK, "Connection database is OK")
		} else {
			zap.S().Error("Connection database is NOT OK")
			return ctx.String(http.StatusInternalServerError, "Connection database is NOT OK")
		}
	}
}

func (h *handler) UpdatesJSON() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		metrics := make([]models.Metrics, 0)
		err := json.NewDecoder(ctx.Request().Body).Decode(&metrics)
		if err != nil && !errors.Is(err, io.EOF) {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Error in JSON decode: %s", err)})
		}
		h.store.StoreBatch(metrics)
		ctx.Response().Header().Set("Content-Type", "application/json")

		return ctx.JSON(http.StatusOK, map[string]string{"status": "success"})
	}
}

