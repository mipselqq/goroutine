package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	// This code does not reuse the config package because
	// getEnvOrDefault is not exported and covered by tests.
	// It isn't worth refactoring the config package.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("HOST")
	if host == "" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}

	url := fmt.Sprintf("http://%s:%s/health", host, port)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck failed: %v\n", err)
		os.Exit(1)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "healthcheck failed with status: %s\n", resp.Status)
		os.Exit(1)
	}

	fmt.Println("OK")
}
