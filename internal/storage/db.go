package storage

import (
	"backend/internal/storage/interfaces"
	"backend/internal/storage/postgres"
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrTypeDb returned when an unsupported database type is requested.
var ErrTypeDb = errors.New("type db error")

// InitDb creates a database connection pool selected by dbType.
func InitDb(ctx context.Context, host, user, password, name string, port int, maxConns, minConns int32,
	connLifetime, connIdleTime, healthCheckTime, queryTimeout, pingTimeout time.Duration, sslMode string, dbType interfaces.DbType) (interfaces.DbStorage, error) {

	switch dbType {
	case interfaces.Postgres:
		return postgres.NewPool(ctx, host, user, password, name, port, maxConns, minConns,
			connLifetime, connIdleTime, healthCheckTime, queryTimeout, pingTimeout, sslMode)
	}

	return nil, fmt.Errorf("%w: %s", ErrTypeDb, dbType)
}
