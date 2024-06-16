package api

import (
	"github.com/labstack/echo/v4"
	"github.com/root-gabriel/ya/internal/handlers"
	"github.com/root-gabriel/ya/internal/storage"
)

type Server struct {
	handler *handlers.Handler
}

func NewServer() *Server {
	storage := storage.NewMemStorage()
	handler := handlers.NewHandler(storage)
	return &Server{handler: handler}
}

func (s *Server) RegisterRoutes(e *echo.Echo) {
	e.POST("/update/:typeM/:nameM/:valueM", s.handler.UpdateMetric)
	e.GET("/value/:typeM/:nameM", s.handler.GetMetric)
}

