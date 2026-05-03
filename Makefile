.PHONY: dev test test-integration build-bin lint fmt swag migrate-up migrate-down \
migrate-status migrate-create tools try-fetch-tags vuln happy-load \
test-race test-integration-race test-k6-race test-some-race

VERSION := $(shell git describe --tags --always --dirty || "")

dev:
	go run -ldflags "-X main.version=$(VERSION)" ./cmd/server/main.go
dev-env:
	docker compose --env-file .env.dev up -d

test:
	go test ./...
prepare-test-env:
	test -f .env.dev || cp .env.example .env.dev
	docker compose --env-file .env.dev up -d --wait postgres

test-integration: prepare-test-env
	go test -tags=integration ./internal/repository/...
test-e2e: prepare-test-env
	go test -tags=e2e ./tests/...
test-cover:
	go test -tags=integration -cover ./...
test-all: test test-integration test-e2e

test-race:
	go test -race ./...
test-integration-race: prepare-test-env
	go test -race -tags=integration ./internal/repository/...
test-k6-race:
	k6 run ./k6/columns-create-race.ts
	k6 run ./k6/columns-delete-compaction-race.ts
	k6 run ./k6/tasks-create-race.ts
	k6 run ./k6/tasks-delete-compaction-race.ts

test-some-race: test-race test-integration-race

build-bin: try-fetch-tags
	CGO_ENABLED=0 \
	GOOS=linux \
	go build \
	-ldflags "-X main.version=$(VERSION)" \
	-o /bin/app ./cmd/server/main.go
	CGO_ENABLED=0 \
	GOOS=linux \
	go build \
	-o /bin/ping ./cmd/ping/main.go

try-fetch-tags:
	git fetch --tags || true

lint:
	golangci-lint run
vuln:
	govulncheck ./...
fmt:
	gofumpt	-l -w .
swag:
	swag init -g cmd/server/main.go

migrate-create:
	@if [ -z "$(name)" ]; then echo "Error: migration name is required. Use: make migrate-create name=migration_name"; exit 1; fi
	goose -dir migrations create $(name) sql

migrate-up migrate-down migrate-status: migrate-%:
	export $$(cat .env.dev | xargs) && goose -dir migrations postgres "user=$$POSTGRES_USER password=$$POSTGRES_PASSWORD dbname=$$POSTGRES_DB host=$$POSTGRES_HOST sslmode=disable" $*

tools:
	go install github.com/evilmartians/lefthook@latest
	lefthook install
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install go.k6.io/k6@latest
	npm install -D @types/k6

K6_PROMETHEUS_RW_SERVER_URL ?=
K6_PROMETHEUS_RW_TREND_STATS ?= p(90),p(95),p(99),min,max,avg,med,count,sum
K6_PROMETHEUS_RW_USERNAME ?=
K6_PROMETHEUS_RW_PASSWORD ?=
K6_TESTID ?= $(shell date +%s)
K6_EXTRA_ARGS ?= --tag testid=$(K6_TESTID)
K6_OUTPUT_ARGS := $(if $(strip $(K6_PROMETHEUS_RW_SERVER_URL)),-o experimental-prometheus-rw,)

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
