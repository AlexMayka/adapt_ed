package interfaces

import "context"

// CacheType identifies a cache/key-value storage implementation.
type CacheType string

// Redis selects Redis as the CacheStorage backend.
const Redis CacheType = "redis"

// CacheStorage defines operations for cache/key-value storage.
type CacheStorage interface {
	Close() error
	Ping(ctx context.Context) error
}
