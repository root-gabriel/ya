package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lionslon/go-yapmetrics/internal/api"
)

func main() {
	e := echo.New()

	// Включение middleware для gzip
	e.Use(middleware.Gzip())

	// Инициализация API сервера
	s := api.New()
	s.RegisterRoutes(e)

	// Запуск сервера
	if err := e.Start(s.Cfg.Addr); err != nil {
		panic(err)
	}
}

//func main() {
//	s := api.New()
//	if err := s.Start(); err != nil {
//		panic(err)
//	}
//}
