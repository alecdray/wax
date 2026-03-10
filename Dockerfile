# Build stage
FROM golang:1.25-bookworm AS builder

WORKDIR /build

# Install build dependencies for go-sqlite3 (requires CGO)
RUN apt-get update && apt-get install -y gcc libc6-dev && rm -rf /var/lib/apt/lists/*

# Download dependencies first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY src/ ./src/
RUN CGO_ENABLED=1 GOOS=linux go build -v -o ./bin/app ./src/cmd/app.go

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

# Install SQLite runtime library (needed by go-sqlite3 CGO binary)
RUN apt-get update && apt-get install -y libsqlite3-0 ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy binary
COPY --from=builder /build/bin/app ./bin/app

# Copy static assets and migrations (needed at runtime)
COPY static/public/ ./static/public/
COPY db/migrations/ ./db/migrations/

# Data directory for SQLite database (mount a volume here)
RUN mkdir -p /data

EXPOSE 4691

ENV ENV=production
ENV PORT=4691
ENV DB_PATH=/data/db.sql
ENV GOOSE_DRIVER=sqlite3
ENV GOOSE_DBSTRING=/data/db.sql
ENV GOOSE_MIGRATION_DIR=./db/migrations

ENTRYPOINT ["./bin/app"]
