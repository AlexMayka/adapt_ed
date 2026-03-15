package storage

import (
	"backend/internal/storage/interfaces"
	"backend/internal/storage/minio"
	"context"
	"errors"
	"fmt"
)

// ErrTypeS3 returned when an unsupported S3 storage type is requested.
var ErrTypeS3 = errors.New("type s3 error")

// InitS3 creates an S3-compatible object storage client selected by typeS3.
func InitS3(ctx context.Context, host string, port int, user, password, bucket, region string, objectLocking, useSSL bool, typeS3 interfaces.S3Type) (interfaces.S3Storage, error) {
	switch typeS3 {
	case interfaces.Minio:
		return minio.Init(ctx, host, port, user, password, bucket, region, objectLocking, useSSL)
	}

	return nil, fmt.Errorf("%w: %s", ErrTypeS3, typeS3)
}
