package main

import (
	"github.com/labstack/echo/v4"
	"github.com/root-gabriel/ya/internal/handlers"
	"github.com/root-gabriel/ya/internal/storage"
)

func main() {
	e := echo.New()
	memStorage := storage.NewMemStorage()
	handler := handlers.NewHandler(memStorage)

	e.POST("/update/:typeM/:nameM/:valueM", handler.UpdateMetric)
	e.GET("/value/:typeM/:nameM", handler.GetMetric)

	e.Start(":8080")
}
