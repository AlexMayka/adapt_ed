package interfaces

import "context"

// CacheType identifies a cache/key-value storage implementation.
type CacheType string

// Redis selects Redis as the Cache backend.
const Redis CacheType = "redis"

// Cache defines operations for cache/key-value storage.
type Cache interface {
	Close() error
	Ping(ctx context.Context) error
}
