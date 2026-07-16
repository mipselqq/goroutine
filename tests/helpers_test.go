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
	"github.com/redis/go-redis/v9"
)

type PreludeResult struct {
	HTTPClient   *http.Client
	Server       *httptest.Server
	Pool         *pgxpool.Pool
	RedisClient  *redis.Client
	MockTelegram *testutil.MockTelegramAPI
}

func Prelude(t *testing.T) PreludeResult {
	t.Helper()

	pool := testutil.SetupPostgres(t, "../migrations")

	if os.Getenv("ENV") != "prod" {
		err := godotenv.Load("../.env.dev")
		if err != nil {
			t.Fatalf("godotenv.Load() error = %v", err)
		}
	} else {
		t.Fatalf("ENV = %q, want non-prod", os.Getenv("ENV"))
	}

	mockTelegram := testutil.NewMockTelegramAPI(t, http.StatusOK)
	t.Setenv("TELEGRAM_API_BASE_URL", mockTelegram.URL)

	logger := testutil.NewLogger(t)
	cfg := config.NewAppFromEnv(logger)
	telegramCfg, err := config.NewTelegramFromEnv(logger)
	if err != nil {
		t.Fatalf("NewTelegramFromEnv() error = %v", err)
	}
	logger.Info("App config", slog.Any("config", cfg))

	redisClient := testutil.SetupRedis(t)
	a := app.New(logger, pool, redisClient, &cfg, &telegramCfg, prometheus.NewRegistry())

	ts := httptest.NewServer(a.Router)
	t.Cleanup(func() {
		ts.Close()
		mockTelegram.Close()
		pool.Close()
		_ = redisClient.Close()
	})

	return PreludeResult{
		HTTPClient:   ts.Client(),
		Server:       ts,
		Pool:         pool,
		RedisClient:  redisClient,
		MockTelegram: mockTelegram,
	}
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
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
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

func CreateUserAndAuthenticateClient(t *testing.T, client *http.Client, baseURL string) *AuthenticatedClient {
	t.Helper()

	email := fmt.Sprintf("e2e-%s@example.com", uuid.NewString())
	password := testutil.ValidPassword().String()
	token := E2ERegisterAndLogin(t, client, baseURL, email, password)

	return &AuthenticatedClient{
		Client:  client,
		BaseURL: baseURL,
		Token:   token,
	}
}

func LinkTelegram(t *testing.T, ac *AuthenticatedClient) {
	t.Helper()

	linkResp := ac.Do(t, http.MethodPost, "/v1/users/me/telegram/link", nil)
	defer func() {
		_ = linkResp.Body.Close()
	}()

	if linkResp.StatusCode != http.StatusOK {
		t.Fatalf("LinkTelegram() link status = %d, want %d", linkResp.StatusCode, http.StatusOK)
	}

	var linkBody struct {
		Token string `json:"token"`
	}
	err := json.NewDecoder(linkResp.Body).Decode(&linkBody)
	if err != nil {
		t.Fatalf("LinkTelegram() link response Decode() error = %v", err)
	}

	webhookBody := map[string]any{
		"message": map[string]any{
			"text": "/start " + linkBody.Token,
			"chat": map[string]any{
				"id":       int64(123456789),
				"username": "testuser",
			},
		},
	}
	body, err := json.Marshal(webhookBody)
	if err != nil {
		t.Fatalf("LinkTelegram() json.Marshal() error = %v", err)
	}

	webhookResp, err := ac.Client.Post(ac.BaseURL+"/webhook/telegram", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("LinkTelegram() webhook Post() error = %v", err)
	}
	defer func() {
		_ = webhookResp.Body.Close()
	}()

	if webhookResp.StatusCode != http.StatusOK {
		t.Fatalf("LinkTelegram() webhook status = %d, want %d", webhookResp.StatusCode, http.StatusOK)
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
