.PHONY: dev test test-integration build lint fmt swag migrate-up migrate-down migrate-status tools quickfuzz fuzz fetch-tags vuln

VERSION := $(shell git describe --tags --always --dirty)

dev:
	go run -ldflags "-X main.version=$(VERSION)" ./cmd/server/main.go
test:
	go test ./...
test-integration:
	sudo docker-compose up -d
	sleep 2
	go test -tags=integration ./internal/repository/...
test-e2e:
	sudo docker-compose up -d
	sleep 2
	go test -tags=e2e ./tests/...
test-cover:
	go test -tags=integration -cover ./...
test-all: test test-integration test-e2e test-cover
build: fetch-tags
	go build -ldflags "-X main.version=$(VERSION)" -o ./bin/app ./cmd/server/main.go

fetch-tags:
	git fetch --tags

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
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install github.com/evilmartians/lefthook@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

quickfuzz:
	go test -fuzz=. -fuzztime=20s
fuzz:
	go test -fuzz=. -fuzztime=2m
