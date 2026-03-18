package minio

import (
	appErr "backend/internal/errors"
	"backend/internal/storage/interfaces"
	typesSg "backend/internal/storage/types"
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Storage оборачивает MinIO-клиент, привязанный к одному бакету.
type Storage struct {
	client *minio.Client
	bucket string
}

// Init создаёт MinIO-клиент, проверяет наличие бакета и создаёт его при необходимости.
func Init(ctx context.Context, host string, port int, user, password, bucket, region string, objectLocking, useSSL bool) (interfaces.S3Storage, error) {
	client, err := minio.New(fmt.Sprintf("%s:%d", host, port), &minio.Options{
		Creds:  credentials.NewStaticV4(user, password, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", appErr.ErrMinioConnect, err)
	}

	stg := &Storage{client: client, bucket: bucket}

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", appErr.ErrBucketCheck, err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{
			Region:        region,
			ObjectLocking: objectLocking,
		})
		if err != nil {
			return nil, fmt.Errorf("%w: %v", appErr.ErrBucketCreate, err)
		}
	}

	return stg, nil
}

// Close освобождает ссылку на MinIO-клиент.
func (s *Storage) Close() error {
	s.client = nil
	return nil
}

// PutObject загружает данные из reader в бакет. Content-Type определяется по расширению файла.
func (s *Storage) PutObject(ctx context.Context, objectName string, reader io.Reader, size int64, extraMeta map[string]string) (*typesSg.UploadInfo, error) {
	ct := mime.TypeByExtension(filepath.Ext(objectName))
	if ct == "" {
		ct = "application/octet-stream"
	}

	opts := minio.PutObjectOptions{
		ContentType:  ct,
		UserMetadata: extraMeta,
	}

	put, err := s.client.PutObject(ctx, s.bucket, objectName, reader, size, opts)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", appErr.ErrPutObject, err)
	}

	info := &typesSg.UploadInfo{
		Bucket:           put.Bucket,
		Key:              put.Key,
		Tags:             put.ETag,
		Size:             put.Size,
		LastModified:     put.LastModified,
		Location:         put.Location,
		VersionID:        put.VersionID,
		Expiration:       put.Expiration,
		ExpirationRuleID: put.ExpirationRuleID,
	}

	return info, nil
}

// GetObject возвращает полное содержимое объекта. Вызывающий обязан закрыть reader.
func (s *Storage) GetObject(ctx context.Context, name string) (io.ReadCloser, error) {
	reader, err := s.client.GetObject(ctx, s.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", appErr.ErrGetObject, err)
	}
	return reader, nil
}

// GetObjectRange возвращает диапазон байт [start, end] объекта. Вызывающий обязан закрыть reader.
func (s *Storage) GetObjectRange(ctx context.Context, name string, start, end int64) (io.ReadCloser, error) {
	opts := minio.GetObjectOptions{}
	if err := opts.SetRange(start, end); err != nil {
		return nil, fmt.Errorf("%w: %v", appErr.ErrSetRangeOffset, err)
	}

	reader, err := s.client.GetObject(ctx, s.bucket, name, opts)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", appErr.ErrGetObject, err)
	}
	return reader, nil
}

// StatObject возвращает метаданные объекта (размер, content-type и т.д.).
func (s *Storage) StatObject(ctx context.Context, name string) (*typesSg.ObjectInfo, error) {
	infoMn, err := s.client.StatObject(ctx, s.bucket, name, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", appErr.ErrStatObject, err)
	}

	info := &typesSg.ObjectInfo{
		Name:         infoMn.Key,
		Size:         infoMn.Size,
		ContentType:  infoMn.ContentType,
		LastModified: infoMn.LastModified,
	}

	return info, nil
}

// RemoveObject удаляет объект из бакета.
func (s *Storage) RemoveObject(ctx context.Context, name string) error {
	err := s.client.RemoveObject(ctx, s.bucket, name, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("%w: %v", appErr.ErrRemoveObject, err)
	}
	return nil
}

// ListObjects возвращает объекты по префиксу. При recursive=false — только верхний уровень.
func (s *Storage) ListObjects(ctx context.Context, prefix string, recursive bool) ([]typesSg.ObjectInfo, error) {
	var objects []typesSg.ObjectInfo
	for obj := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	}) {
		if obj.Err != nil {
			return nil, fmt.Errorf("%w: %v", appErr.ErrListObjects, obj.Err)
		}
		objects = append(objects, typesSg.ObjectInfo{
			Name:         obj.Key,
			Size:         obj.Size,
			ContentType:  obj.ContentType,
			LastModified: obj.LastModified,
		})
	}

	return objects, nil
}
