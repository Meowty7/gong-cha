// Package main is a tiny container health probe for the API.
// It GETs HEALTH_CHECK_URL (default http://localhost:8080/health/live)
// and exits 0 on HTTP 200, 1 otherwise.
package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	url := os.Getenv("HEALTH_CHECK_URL")
	if url == "" {
		url = "http://localhost:8080/health/live"
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "health check failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "health check status %d\n", resp.StatusCode)
		os.Exit(1)
	}
}
