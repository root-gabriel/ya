package main

import (
	"github.com/root-gabriel/ya/internal/api"
)

func main() {
	s := api.New()
	if err := s.Start(); err != nil {
		panic(err)
	}
}
