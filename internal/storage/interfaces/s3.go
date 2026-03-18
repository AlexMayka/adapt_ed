package interfaces

import (
	typesSg "backend/internal/storage/types"
	"context"
	"io"
)

// S3Type определяет тип реализации S3-хранилища.
type S3Type string

const Minio S3Type = "minio"

// S3Storage описывает операции S3-совместимого объектного хранилища.
type S3Storage interface {
	PutObject(ctx context.Context, objectName string, reader io.Reader, size int64, extraMeta map[string]string) (*typesSg.UploadInfo, error)
	GetObject(ctx context.Context, name string) (io.ReadCloser, error)
	GetObjectRange(ctx context.Context, name string, start, end int64) (io.ReadCloser, error)
	StatObject(ctx context.Context, name string) (*typesSg.ObjectInfo, error)
	RemoveObject(ctx context.Context, name string) error
	ListObjects(ctx context.Context, prefix string, recursive bool) ([]typesSg.ObjectInfo, error)
	Close() error
}
