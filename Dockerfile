# ---- build ----
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app ./cmd

# ---- runtime ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app .

EXPOSE 8000

ENTRYPOINT ["./app"]
