FROM golang:1.27.2-alpine@sha256:3b15ba438c60ab8fabe5fee3b11025c0049e358b31feff384648f11fc8e335f8 AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build-bin

FROM gcr.io/distroless/static-debian12@sha256:61b7ccecebc7c474a531717de80a94709d20547cdcdaf740c25876f2a8e38b44

COPY --from=builder /bin/app /app
COPY --from=builder /bin/notify /notify
COPY --from=builder /bin/ping /ping
COPY --from=builder /bin/migrate /migrate
COPY --from=builder /app/migrations /migrations

ENTRYPOINT ["/app"]
