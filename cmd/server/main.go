package main

import (
    "github.com/root-gabriel/ya/internal/api"
    "log"
)

func main() {
    server := api.NewServer()
    if err := server.Start(); err != nil {
        log.Fatal("Failed to start server: ", err)
    }
}

