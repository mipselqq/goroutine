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

# Migrations
migrate-up:
	export $$(cat .env.dev | xargs) && go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" up

migrate-down:
	export $$(cat .env.dev | xargs) && go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" down

migrate-status:
	export $$(cat .env.dev | xargs) && go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" status

