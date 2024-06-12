package api

import (
    "github.com/root-gabriel/ya/internal/handlers"
    "github.com/root-gabriel/ya/internal/storage"
    "github.com/labstack/echo/v4"
)

type API struct {
    handler *handlers.Handler
}

func NewAPI() *API {
    storage := storage.NewMemoryStorage()
    handler := handlers.NewHandler(storage)
    return &API{handler: handler}
}

func (api *API) RegisterRoutes(e *echo.Echo) {
    e.POST("/update", api.handler.UpdateMetrics)
    e.POST("/value", api.handler.GetMetricsValue)
    e.GET("/metrics", api.handler.AllMetrics)
}

