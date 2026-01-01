FROM golang:1.25.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./cmd/server/main.go

FROM gcr.io/distroless/static-debian12

COPY --from=builder /main /main

ENTRYPOINT ["/main"]
