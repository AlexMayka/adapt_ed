package redis

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrConnectionFailed returned when Redis is unreachable or authentication fails.
var ErrConnectionFailed = errors.New("redis connection failed")

// Connect wraps a go-redis client providing connection lifecycle management.
type Connect struct {
	client *redis.Client
}

// Init creates a Redis client, verifies the connection with Ping and returns a ready Connect.
// timeout is used for dial, read and write deadlines.
func Init(ctx context.Context, host string, port, db int, password string, useSSL bool, maxRetries int, timeout time.Duration) (*Connect, error) {
	opts := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Password:     password,
		DB:           db,
		MaxRetries:   maxRetries,
		DialTimeout:  timeout,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	if useSSL {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client := redis.NewClient(opts)

	con := &Connect{client: client}
	if err := con.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	return con, nil
}

// Close gracefully shuts down the Redis connection.
func (c *Connect) Close() error {
	return c.client.Close()
}

// Ping sends a PING command to verify the connection is alive.
func (c *Connect) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Client returns the underlying go-redis client for direct command access.
func (c *Connect) Client() *redis.Client {
	return c.client
}
