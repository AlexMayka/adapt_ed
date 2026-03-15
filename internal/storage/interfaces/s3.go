package _interface

import (
	typesSg "backend/internal/storage/types"
	"context"
	"io"
)

type DbS3 string

var Minio DbS3 = "minio"

type S3Storage interface {
	PutObject(ctx context.Context, objectName string, reader io.Reader, size int64, extraMeta map[string]string) (*typesSg.UploadInfo, error)
	GetObject(ctx context.Context, name string) (io.ReadCloser, error)
	GetObjectRange(ctx context.Context, name string, start, end int64) (io.ReadCloser, error)
	StatObject(ctx context.Context, name string) (*typesSg.ObjectInfo, error)
	RemoveObject(ctx context.Context, name string) error
	ListObjects(ctx context.Context, prefix string, recursive bool) ([]typesSg.ObjectInfo, error)
}
