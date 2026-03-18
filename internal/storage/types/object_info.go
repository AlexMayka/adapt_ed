package types

import "time"

// ObjectInfo содержит метаданные объекта без привязки к типам MinIO SDK.
type ObjectInfo struct {
	Name         string
	Size         int64
	ContentType  string
	LastModified time.Time
}
