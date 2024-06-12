package main

import (
    "github.com/labstack/echo/v4"
    "github.com/root-gabriel/ya/internal/api"
)

func main() {
    e := echo.New()
    api := api.NewAPI()
    api.RegisterRoutes(e)

    e.Logger.Fatal(e.Start(":8080"))
}

