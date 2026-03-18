package interfaces

import (
	"context"
	"time"
)

// CacheType определяет тип реализации кэш-хранилища.
type CacheType string

const Redis CacheType = "redis"

// CacheStorage описывает операции кэш-хранилища (key-value).
type CacheStorage interface {
	Close() error
	Ping(ctx context.Context) error
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}
