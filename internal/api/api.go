package api

import (
    "github.com/labstack/echo/v4"
    "github.com/root-gabriel/ya/internal/handlers"
    "net/http"
)

type Server struct {
    echo *echo.Echo
}

func NewServer() *Server {
    server := &Server{
        echo: echo.New(),
    }
    server.routes()
    return server
}

func (s *Server) Start() error {
    return s.echo.Start(":8080")
}

func (s *Server) routes() {
    s.echo.GET("/value/counter/:name", handlers.GetCounter)
}

