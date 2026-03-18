package postgres

import (
	appErr "backend/internal/errors"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// PoolPsg оборачивает pgxpool.Pool с таймаутом запросов.
type PoolPsg struct {
	Pool         *pgxpool.Pool
	QueryTimeout time.Duration
}

// NewPool создаёт пул соединений PostgreSQL и проверяет доступность через ping.
func NewPool(ctx context.Context, host, user, password, name string, port int, maxConns, minConns int32,
	connLifetime, connIdleTime, healthCheckTime, queryTimeout, pingTimeout time.Duration, sslMode string) (*PoolPsg, error) {

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", user, password, host, port, name, sslMode)
	poolCnf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", appErr.ErrPgParseConfig, err)
	}

	poolCnf.MaxConns = maxConns
	poolCnf.MinConns = minConns
	poolCnf.MaxConnLifetime = connLifetime
	poolCnf.MaxConnIdleTime = connIdleTime
	poolCnf.HealthCheckPeriod = healthCheckTime

	pool, err := pgxpool.NewWithConfig(ctx, poolCnf)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", appErr.ErrPgCreatePool, err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%w: %w", appErr.ErrPgPing, err)
	}

	return &PoolPsg{Pool: pool, QueryTimeout: queryTimeout}, nil
}

// Ping проверяет доступность базы данных.
func (p *PoolPsg) Ping(ctx context.Context) error {
	return p.Pool.Ping(ctx)
}

// Close освобождает все соединения пула.
func (p *PoolPsg) Close() error {
	if p.Pool != nil {
		p.Pool.Close()
	}

	return nil
}
