FROM golang:1.26.5-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build-bin

FROM gcr.io/distroless/static-debian12@sha256:22fd79fd75eab2372585b44517f8a094349938919dc613aafc37e4bdc9967c82

COPY --from=builder /bin/app /app
COPY --from=builder /bin/ping /ping
COPY --from=builder /app/migrations /migrations

ENTRYPOINT ["/app"]
