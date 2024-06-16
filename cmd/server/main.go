package main

import (
	"log"
	"net/http"

	"github.com/root-gabriel/ya/internal/api"
)

func main() {
	server := api.NewServer()
	log.Fatal(http.ListenAndServe(":8080", server))
}

