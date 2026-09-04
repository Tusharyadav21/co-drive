# Multi-stage Dockerfile for Go PostgreSQL Backend
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/migrate ./cmd/migrate

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S codrive && adduser -S codrive -G codrive

COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrate /app/migrate
COPY static /app/static

USER codrive

ENV PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/health || exit 1

CMD ["/app/server"]
