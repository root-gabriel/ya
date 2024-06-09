package main

import (
	"github.com/lionslon/go-yapmetrics/internal/api"
)

func main() {
	s := api.New()
	if err := s.Start(); err != nil {
		panic(err)
	}
}
