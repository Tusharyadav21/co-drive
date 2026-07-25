# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o codrive ./cmd/server

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary and static assets from builder
COPY --from=builder /app/codrive .
COPY --from=builder /app/static ./static

# Create non-root user
RUN adduser -D -g '' appuser && chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["./codrive"]