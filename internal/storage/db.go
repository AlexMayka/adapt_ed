package storage

import (
	"backend/internal/storage/postgres"
	"context"
	"time"
)

// InitDb создаёт пул соединений PostgreSQL.
func InitDb(ctx context.Context, host, user, password, name string, port int, maxConns, minConns int32,
	connLifetime, connIdleTime, healthCheckTime, queryTimeout, pingTimeout time.Duration, sslMode string) (*postgres.PoolPsg, error) {

	return postgres.NewPool(ctx, host, user, password, name, port, maxConns, minConns,
		connLifetime, connIdleTime, healthCheckTime, queryTimeout, pingTimeout, sslMode)
}
