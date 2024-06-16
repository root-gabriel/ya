package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	for {
		resp, err := http.Post(
			"http://localhost:8080/update/counter/testCounter/1",
			"text/plain",
			nil,
		)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			resp.Body.Close()
		}

		time.Sleep(10 * time.Second)
	}
}

