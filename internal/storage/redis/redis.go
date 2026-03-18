package redis

import (
	appErr "backend/internal/errors"
	"backend/internal/storage/interfaces"
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Connect оборачивает go-redis клиент с управлением жизненным циклом соединения.
type Connect struct {
	client *redis.Client
}

// Init создаёт Redis-клиент, проверяет соединение через Ping и возвращает готовый Connect.
func Init(ctx context.Context, host string, port, db int, password string, useSSL bool, maxRetries int, timeout time.Duration) (interfaces.CacheStorage, error) {
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
		return nil, fmt.Errorf("%w: %v", appErr.ErrRedisConnectionFailed, err)
	}

	return con, nil
}

// Close закрывает соединение с Redis.
func (c *Connect) Close() error {
	return c.client.Close()
}

// Ping проверяет доступность Redis через команду PING.
func (c *Connect) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Set сохраняет пару ключ-значение с заданным TTL.
func (c *Connect) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

// Get возвращает значение по ключу. Пустая строка и nil при отсутствии ключа.
func (c *Connect) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// Del удаляет ключ.
func (c *Connect) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Client возвращает базовый go-redis клиент для прямого доступа к командам.
func (c *Connect) Client() *redis.Client {
	return c.client
}
