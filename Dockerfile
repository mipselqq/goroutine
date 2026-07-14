FROM golang:1.26.5-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build-bin

FROM gcr.io/distroless/static-debian12@sha256:61b7ccecebc7c474a531717de80a94709d20547cdcdaf740c25876f2a8e38b44

COPY --from=builder /bin/app /app
COPY --from=builder /bin/ping /ping
COPY --from=builder /app/migrations /migrations

ENTRYPOINT ["/app"]
