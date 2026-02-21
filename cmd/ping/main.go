package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("HOST")
	if host == "" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}

	if err := run(host, port, 5*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK")
}

func run(host, port string, timeout time.Duration) error {
	url := fmt.Sprintf("http://%s:%s/health", host, port)
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck failed with status: %s", resp.Status)
	}

	return nil
}
