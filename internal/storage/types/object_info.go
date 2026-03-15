package types

import "time"

// ObjectInfo contains metadata about a stored object without exposing MinIO SDK types.
type ObjectInfo struct {
	Name         string
	Size         int64
	ContentType  string
	LastModified time.Time
}
