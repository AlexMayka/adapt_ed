package storage

import (
	appErr "backend/internal/errors"
	"backend/internal/storage/interfaces"
	"backend/internal/storage/redis"
	"context"
	"fmt"
	"time"
)

// InitCache создаёт клиент кэш-хранилища по типу typeCache.
func InitCache(ctx context.Context, host string, port, db int,
	password string, useSSL bool, maxRetries int, timeout time.Duration, typeCache interfaces.CacheType) (interfaces.CacheStorage, error) {

	switch typeCache {
	case interfaces.Redis:
		return redis.Init(ctx, host, port, db, password, useSSL, maxRetries, timeout)
	}

	return nil, fmt.Errorf("%w: %s", appErr.ErrTypeCache, typeCache)
}
