package redis

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrConnectionFailed = errors.New("redis connection failed")
)

type Connect struct {
	client *redis.Client
}

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

func (c *Connect) Close() error {
	return c.client.Close()
}

func (c *Connect) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Connect) Client() *redis.Client {
	return c.client
}
