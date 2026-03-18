package storage

import (
	appErr "backend/internal/errors"
	"backend/internal/storage/interfaces"
	"backend/internal/storage/minio"
	"context"
	"fmt"
)

// InitS3 создаёт клиент S3-совместимого хранилища по типу typeS3.
func InitS3(ctx context.Context, host string, port int, user, password, bucket, region string, objectLocking, useSSL bool, typeS3 interfaces.S3Type) (interfaces.S3Storage, error) {
	switch typeS3 {
	case interfaces.Minio:
		return minio.Init(ctx, host, port, user, password, bucket, region, objectLocking, useSSL)
	}

	return nil, fmt.Errorf("%w: %s", appErr.ErrTypeS3, typeS3)
}
