.PHONY: dev

dev:
	go run ./cmd/server/main.go
test:
	go test ./...
test-integration:
	sudo docker-compose up -d postgres
	sleep 2
	go test -v -tags=integration ./internal/repository/...
build:
	go build -o ./bin/app main.go

lint:
	golangci-lint run
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

quickfuzz:
	go test -fuzz=. -fuzztime=20s
fuzz:
	go test -fuzz=. -fuzztime=2m
