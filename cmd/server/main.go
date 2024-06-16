package main

import (
	"log"
	"github.com/labstack/echo/v4"
	"github.com/root-gabriel/ya/internal/handlers"
	"github.com/root-gabriel/ya/internal/storage"
)

func main() {
	store := storage.NewMem()
	handler := handlers.New(store)

	e := echo.New()
	e.POST("/update/:type/:name/:value", handler.UpdateMetrics())
	e.GET("/value/:type/:name", handler.MetricsValue())

	log.Fatal(e.Start(":8080"))
}

