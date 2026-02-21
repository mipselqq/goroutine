package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/health" {
				t.Errorf("expected path /health, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		err := run(u.Hostname(), u.Port(), 1*time.Second)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("non-200 status", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		err := run(u.Hostname(), u.Port(), 1*time.Second)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("timeout", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		err := run(u.Hostname(), u.Port(), 10*time.Millisecond)
		if err == nil {
			t.Fatal("expected timeout error, got nil")
		}
	})

	t.Run("unreachable host", func(t *testing.T) {
		err := run("127.0.0.1", "0", 1*time.Second)
		if err == nil {
			t.Fatal("expected error for unreachable host, got nil")
		}
	})
}
