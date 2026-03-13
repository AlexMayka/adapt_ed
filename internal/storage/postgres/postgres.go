package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

var (
	ErrParseConfig = errors.New("failed to parse config")
	ErrCreatePool  = errors.New("failed to create pool")
	ErrPing        = errors.New("failed to ping")
)

type PoolPsg struct {
	Pool         *pgxpool.Pool
	QueryTimeout time.Duration
}

func NewPool(ctx context.Context, host, user, password, name string, port int, maxConns, minConns int32,
	connLifetime, connIdleTime, healthCheckTime, queryTimeout, pingTimeout time.Duration, sslMode string) (*PoolPsg, error) {

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", user, password, host, port, name, sslMode)
	poolCnf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseConfig, err)
	}

	poolCnf.MaxConns = maxConns
	poolCnf.MinConns = minConns
	poolCnf.MaxConnLifetime = connLifetime
	poolCnf.MaxConnIdleTime = connIdleTime
	poolCnf.HealthCheckPeriod = healthCheckTime

	pool, err := pgxpool.NewWithConfig(ctx, poolCnf)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreatePool, err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%w: %w", ErrPing, err)
	}

	return &PoolPsg{Pool: pool, QueryTimeout: queryTimeout}, nil
}

func (p *PoolPsg) Close() error {
	if p.Pool != nil {
		p.Pool.Close()
	}

	return nil
}
