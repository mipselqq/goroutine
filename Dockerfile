FROM golang:1.25.7-alpine AS builder

RUN apk add --no-cache "git<3" "make<5"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build-bin

FROM gcr.io/distroless/static-debian12@sha256:cd64bec9cec257044ce3a8dd3620cf83b387920100332f2b041f19c4d2febf93

COPY --from=builder /bin/app /app
COPY --from=builder /app/migrations /migrations

ENTRYPOINT ["/app"]
