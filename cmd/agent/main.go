package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	for {
		_, err := http.Get("http://localhost:8080/update/counter/testCounter/1")
		if err != nil {
			log.Println("Error updating counter:", err)
		}
		time.Sleep(10 * time.Second)
	}
}
