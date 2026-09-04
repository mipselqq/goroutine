.PHONY: dev app notify test test-integration build-bin lint fmt swag migrate-up migrate-down \
migrate-status migrate-create tools try-fetch-tags vuln happy-load \
test-race test-integration-race test-k6-race test-some-race

VERSION := $(shell git describe --tags --always --dirty || "")

# Run uncontainerized HTTP server and notification worker
dev:
	$(MAKE) --no-print-directory -j2 app notify

# Run uncontainerized HTTP server (app) or notification worker (notify)
app notify:
	bash -o pipefail -c 'go run -ldflags "-X main.version=$(VERSION)" ./cmd/$@/main.go 2>&1 | awk "{print \"[$@] \" $$0; fflush()}"'

# Run development environment with docker compose
dev-env:
	docker compose --env-file .env.dev up -d

# Run all tests
test:
	go test ./...

# Prepare test environment
prepare-test-env:
	test -f .env.dev || cp .env.example .env.dev
	docker compose --env-file .env.dev up -d --wait postgres redis

# Run integration tests
test-integration: prepare-test-env
	go test -tags=integration ./internal/repository/...

# Run E2E tests
test-e2e: prepare-test-env
	go test -tags=e2e ./tests/...

# Run coverage tests
test-cover:
	go test -tags=integration -cover ./...

# Run all tests
test-all: test test-integration test-e2e

# Run Go-level tests with race detection
test-race:
	go test -race ./...

# Run Go-level integration tests with race detection
test-integration-race: prepare-test-env
	go test -race -tags=integration ./internal/repository/...

# Run application-level k6 tests with race detection
test-k6-race:
	k6 run ./k6/columns-create-race.ts
	k6 run ./k6/columns-delete-compaction-race.ts
	k6 run ./k6/tasks-create-race.ts
	k6 run ./k6/tasks-delete-compaction-race.ts

# Run some tests with race detection
test-some-race: test-race test-integration-race

# Build binaries
build-bin: try-fetch-tags
	CGO_ENABLED=0 GOOS=linux go build \
    	-ldflags "-X main.version=$(VERSION)" \
    	-o /bin/app ./cmd/app/main.go

	CGO_ENABLED=0 GOOS=linux go build \
    	-ldflags "-X main.version=$(VERSION)" \
    	-o /bin/notify ./cmd/notify/main.go

	CGO_ENABLED=0 GOOS=linux go build \
    	-o /bin/ping ./cmd/ping/main.go

	CGO_ENABLED=0 GOOS=linux go build \
    	-o /bin/migrate ./cmd/migrate/main.go

# Try to fetch tags
try-fetch-tags:
	git fetch --tags || true

# Run golangci-lint
lint:
	golangci-lint run

# Run govulncheck
vuln:
	govulncheck ./...

# Run gofmt
fmt:
	gofumpt	-l -w .
swag:
	swag init -g cmd/app/main.go -o docs/openapi

# Create a new migration
migrate-create:
	@if [ -z "$(name)" ]; then echo "Error: migration name is required. Use: make migrate-create name=migration_name"; exit 1; fi
	goose -dir migrations create $(name) sql

# Run migrations
migrate-up migrate-down migrate-status: migrate-%:
	export $$(cat .env.dev | xargs) && goose -dir migrations postgres "user=$$POSTGRES_USER password=$$POSTGRES_PASSWORD dbname=$$POSTGRES_DB host=$$POSTGRES_HOST sslmode=disable" $*

migrate-up-go:
	go run ./cmd/migrate

# Install development tools
tools:
	go install github.com/evilmartians/lefthook@v1.12.2
	lefthook install
	go install github.com/pressly/goose/v3/cmd/goose@v3.27.3
	go install github.com/swaggo/swag/cmd/swag@v1.16.6
	go install mvdan.cc/gofumpt@v0.11.0
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2
	go install golang.org/x/vuln/cmd/govulncheck@v1.7.0
	go install go.k6.io/k6@v1.8.1
	npm install -D @types/k6@2.2.0

# Load testing configuration
K6_PROMETHEUS_RW_SERVER_URL ?=
K6_PROMETHEUS_RW_TREND_STATS ?= p(90),p(95),p(99),min,max,avg,med,count,sum
K6_PROMETHEUS_RW_USERNAME ?=
K6_PROMETHEUS_RW_PASSWORD ?=
K6_TESTID ?= $(shell date +%s)
K6_EXTRA_ARGS ?= --tag testid=$(K6_TESTID)
K6_OUTPUT_ARGS := $(if $(strip $(K6_PROMETHEUS_RW_SERVER_URL)),-o experimental-prometheus-rw,)

# Run happy path load test, simulating real user behavior
happy-load:
	K6_ROOT='$(K6_ROOT)' \
	K6_VUS_STEP='$(K6_VUS_STEP)' \
	K6_VUS_PLATEAU_DURATION='$(K6_VUS_PLATEAU_DURATION)' \
	K6_RAMP_DURATION='$(K6_RAMP_DURATION)' \
	K6_MAX_STAGES='$(K6_MAX_STAGES)' \
	K6_PROMETHEUS_RW_SERVER_URL='$(K6_PROMETHEUS_RW_SERVER_URL)' \
	K6_PROMETHEUS_RW_TREND_STATS='$(K6_PROMETHEUS_RW_TREND_STATS)' \
	K6_PROMETHEUS_RW_USERNAME='$(K6_PROMETHEUS_RW_USERNAME)' \
	K6_PROMETHEUS_RW_PASSWORD='$(K6_PROMETHEUS_RW_PASSWORD)' \
	k6 run $(K6_OUTPUT_ARGS) $(K6_EXTRA_ARGS) ./k6/realistic-happy-path-load.ts
