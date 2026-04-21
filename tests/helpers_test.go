//go:build e2e

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"goroutine/internal/app"
	"goroutine/internal/config"
	"goroutine/internal/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
)

func Prelude(t *testing.T) (*http.Client, *httptest.Server, *pgxpool.Pool) {
	t.Helper()

	pool := testutil.SetupTestDB(t, "../migrations")

	if os.Getenv("ENV") != "prod" {
		err := godotenv.Load("../.env.dev")
		if err != nil {
			t.Fatalf("godotenv.Load() error = %v", err)
		}
	} else {
		t.Fatalf("ENV = %q, want non-prod", os.Getenv("ENV"))
	}

	logger := testutil.NewTestLogger(t)
	cfg := config.NewAppConfigFromEnv(logger)
	logger.Info("App config", slog.Any("config", cfg))

	application := app.New(logger, pool, &cfg, prometheus.NewRegistry())

	ts := httptest.NewServer(application.Router)
	t.Cleanup(func() {
		ts.Close()
		pool.Close()
	})

	client := ts.Client()

	return client, ts, pool
}

func Register(t *testing.T, c *http.Client, baseURL, email, password string) {
	t.Helper()

	body, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		t.Fatalf("Register() json.Marshal() error = %v", err)
	}

	resp, err := c.Post(baseURL+"/v1/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Register() Post() error = %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Register() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func Login(t *testing.T, c *http.Client, baseURL, email, password string) string {
	t.Helper()

	body, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		t.Fatalf("Login() json.Marshal() error = %v", err)
	}

	resp, err := c.Post(baseURL+"/v1/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Login() Post() error = %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("Login response Decode() error = %v", err)
	}
	if out.Token == "" {
		t.Fatalf("got token %q, want non-empty", out.Token)
	}

	return out.Token
}

func E2ERegisterAndLogin(t *testing.T, c *http.Client, baseURL, email, password string) string {
	t.Helper()
	Register(t, c, baseURL, email, password)

	return Login(t, c, baseURL, email, password)
}

type AuthenticatedClient struct {
	Client  *http.Client
	BaseURL string
	Token   string
}

func CreateUserAndAuthenticateClient(t *testing.T, httpClient *http.Client, baseURL string) *AuthenticatedClient {
	t.Helper()

	email := fmt.Sprintf("e2e-%s@example.com", uuid.NewString())
	password := testutil.ValidPassword().String()
	token := E2ERegisterAndLogin(t, httpClient, baseURL, email, password)

	return &AuthenticatedClient{
		Client:  httpClient,
		BaseURL: baseURL,
		Token:   token,
	}
}

func (c *AuthenticatedClient) Do(t *testing.T, method, path string, body any) *http.Response {
	t.Helper()

	var rdr io.Reader = http.NoBody
	if body != nil {
		switch b := body.(type) {
		case []byte:
			rdr = bytes.NewReader(b)
		default:
			data, err := json.Marshal(b)
			if err != nil {
				t.Fatalf("AuthenticatedClient.Do() json.Marshal() error = %v", err)
			}
			rdr = bytes.NewReader(data)
		}
	}

	req, err := http.NewRequest(method, c.BaseURL+path, rdr)
	if err != nil {
		t.Fatalf("AuthenticatedClient.Do() NewRequest() error = %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		t.Fatalf("AuthenticatedClient.Do(%s %s) error = %v", method, path, err)
	}

	return resp
}

func WaitForTimestampTicker(t *testing.T) {
	t.Helper()
	time.Sleep(5 * time.Millisecond)
}
