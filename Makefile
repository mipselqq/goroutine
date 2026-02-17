.PHONY: dev test test-integration build-bin lint fmt swag migrate-up migrate-down migrate-status tools try-fetch-tags vuln

VERSION := $(shell git describe --tags --always --dirty || "")

dev:
	go run -ldflags "-X main.version=$(VERSION)" ./cmd/server/main.go
dev-env:
	docker compose --env-file .env.dev up -d

test:
	go test ./...
prepare-test-env:
	test -f .env.dev || cp .env.example .env.dev
	docker compose --env-file .env.dev up -d --wait

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
test-some-race: test-race test-integration-race

build-bin: try-fetch-tags
	CGO_ENABLED=0 \
	GOOS=linux \
	go build \
	-ldflags "-X main.version=$(VERSION)" \
	-o /bin/app ./cmd/server/main.go

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
