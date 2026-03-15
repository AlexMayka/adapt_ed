# ---- build ----
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

# swag CLI для генерации swagger
RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# генерация swagger docs
RUN swag init -g cmd/main.go

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/backend ./cmd

# ---- runtime ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

# goose CLI для миграций
RUN wget -qO /usr/local/bin/goose \
    "https://github.com/pressly/goose/releases/download/v3.27.0/goose_linux_$(uname -m | sed 's/aarch64/arm64/' | sed 's/x86_64/amd64/')" \
    && chmod +x /usr/local/bin/goose

WORKDIR /app
COPY --from=builder /build/backend .
COPY migrations/ ./migrations/

EXPOSE 8000

CMD sh -c '\
  echo "=== Миграции ===" && \
  goose -dir /app/migrations postgres \
    "host=${PG_HOST} port=${PG_PORT} user=${PG_USER} password=${PG_PASSWORD} dbname=${PG_DB} sslmode=${PG_SSL_MODE}" \
    up && \
  echo "=== Миграции ✅ ===" && \
  exec /app/backend'
