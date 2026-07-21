FROM golang:1.26.5-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build-bin

FROM gcr.io/distroless/static-debian12@sha256:a9fcaedd4c9b59e12dd65d954f0b5044f19b0647a8a3712e77205df9e7b102cd

COPY --from=builder /bin/app /app
COPY --from=builder /bin/ping /ping
COPY --from=builder /app/migrations /migrations

ENTRYPOINT ["/app"]
