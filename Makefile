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
