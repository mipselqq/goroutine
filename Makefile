.PHONY: dev

dev:
	go run ./cmd/server/main.go
test:
	go test ./...
build:
	go build -o ./bin/app main.go

lint:
	golangci-lint run
fmt:
	gofumpt	-l -w .
