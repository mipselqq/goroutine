FROM golang:1.25.7-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build-bin

FROM gcr.io/distroless/static-debian12@sha256:20bc6c0bc4d625a22a8fde3e55f6515709b32055ef8fb9cfbddaa06d1760f838

COPY --from=builder /bin/app /app
COPY --from=builder /bin/ping /ping
COPY --from=builder /app/migrations /migrations

ENTRYPOINT ["/app"]
