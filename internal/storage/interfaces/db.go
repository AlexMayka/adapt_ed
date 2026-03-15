package interfaces

import "context"

// DbType identifies a database implementation.
type DbType string

// Postgres selects PostgreSQL as the DbStorage backend.
const Postgres DbType = "postgres"

// DbStorage defines operations for database connection pool management.
type DbStorage interface {
	Close() error
	Ping(ctx context.Context) error
}
