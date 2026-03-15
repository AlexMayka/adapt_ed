package storage

import (
	"backend/internal/storage/interfaces"
	"backend/internal/storage/redis"
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrTypeCache returned when an unsupported cache type is requested.
var ErrTypeCache = errors.New("type cache error")

// InitCache creates a cache/key-value storage client selected by typeCache.
func InitCache(ctx context.Context, host string, port, db int,
	password string, useSSL bool, maxRetries int, timeout time.Duration, typeCache interfaces.CacheType) (interfaces.Cache, error) {

	switch typeCache {
	case interfaces.Redis:
		return redis.Init(ctx, host, port, db, password, useSSL, maxRetries, timeout)
	}

	return nil, fmt.Errorf("%w: %s", ErrTypeCache, typeCache)
}
