package interfaces

import (
	typesSg "backend/internal/storage/types"
	"context"
	"io"
)

// S3Type identifies an S3-compatible object storage implementation.
type S3Type string

// Minio selects MinIO as the S3Storage backend.
const Minio S3Type = "minio"

// S3Storage defines operations for S3-compatible object storage.
type S3Storage interface {
	PutObject(ctx context.Context, objectName string, reader io.Reader, size int64, extraMeta map[string]string) (*typesSg.UploadInfo, error)
	GetObject(ctx context.Context, name string) (io.ReadCloser, error)
	GetObjectRange(ctx context.Context, name string, start, end int64) (io.ReadCloser, error)
	StatObject(ctx context.Context, name string) (*typesSg.ObjectInfo, error)
	RemoveObject(ctx context.Context, name string) error
	ListObjects(ctx context.Context, prefix string, recursive bool) ([]typesSg.ObjectInfo, error)
	Close() error
}
