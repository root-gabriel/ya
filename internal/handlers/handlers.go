package handlers

import (
    "github.com/labstack/echo/v4"
    "net/http"
)

func GetCounter(c echo.Context) error {
    name := c.Param("name")
    // Для простоты возвращаем фиксированное значение
    return c.String(http.StatusOK, "42")
}

