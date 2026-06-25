package main

import (
	"net/http"
	"os"
)

func main() {
	resp, err := http.Get("http://127.0.0.1:8080")

	if err == nil && resp.StatusCode == 200 {
		os.Exit(0)
	}

	os.Exit(1)
}
