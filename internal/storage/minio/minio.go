package minio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	ErrMinioConnect   = errors.New("failed to connect to minio")
	ErrBucketCheck    = errors.New("failed to check bucket existence")
	ErrBucketCreate   = errors.New("failed to create bucket")
	ErrSetRangeOffset = errors.New("failed to set range offset")
	ErrGetObject      = errors.New("failed to get object from minio")
	ErrStatObject     = errors.New("failed to get object info")
	ErrRemoveObject   = errors.New("failed to remove object")
	ErrListObjects    = errors.New("failed to list objects")
)

type ObjectInfo struct {
	Name         string
	Size         int64
	ContentType  string
	LastModified time.Time
}

type Storage struct {
	client *minio.Client
	bucket string
}

func Init(ctx context.Context, host string, port int, user, password, bucket, region string, objectLocking, useSSL bool) (*Storage, error) {
	client, err := minio.New(fmt.Sprintf("%s:%d", host, port), &minio.Options{
		Creds:  credentials.NewStaticV4(user, password, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMinioConnect, err)
	}

	stg := &Storage{client: client, bucket: bucket}

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBucketCheck, err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{
			Region:        region,
			ObjectLocking: objectLocking,
		})
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrBucketCreate, err)
		}
	}

	return stg, nil
}

func (s *Storage) Close() error {
	s.client = nil
	return nil
}

func (s *Storage) PutObject(ctx context.Context, objectName string, reader io.Reader, size int64, extraMeta map[string]string) (minio.UploadInfo, error) {
	ct := mime.TypeByExtension(filepath.Ext(objectName))
	if ct == "" {
		ct = "application/octet-stream"
	}

	opts := minio.PutObjectOptions{
		ContentType:  ct,
		UserMetadata: extraMeta,
	}

	return s.client.PutObject(ctx, s.bucket, objectName, reader, size, opts)
}

func (s *Storage) GetObject(ctx context.Context, name string) (io.ReadCloser, error) {
	reader, err := s.client.GetObject(ctx, s.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetObject, err)
	}
	return reader, nil
}

func (s *Storage) GetObjectRange(ctx context.Context, name string, start, end int64) (io.ReadCloser, error) {
	opts := minio.GetObjectOptions{}
	if err := opts.SetRange(start, end); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSetRangeOffset, err)
	}

	reader, err := s.client.GetObject(ctx, s.bucket, name, opts)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetObject, err)
	}
	return reader, nil
}

func (s *Storage) StatObject(ctx context.Context, name string) (minio.ObjectInfo, error) {
	info, err := s.client.StatObject(ctx, s.bucket, name, minio.StatObjectOptions{})
	if err != nil {
		return minio.ObjectInfo{}, fmt.Errorf("%w: %v", ErrStatObject, err)
	}

	return info, nil
}

func (s *Storage) RemoveObject(ctx context.Context, name string) error {
	err := s.client.RemoveObject(ctx, s.bucket, name, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRemoveObject, err)
	}
	return nil
}

func (s *Storage) ListObjects(ctx context.Context, prefix string, recursive bool) ([]ObjectInfo, error) {
	var objects []ObjectInfo
	for obj := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	}) {
		if obj.Err != nil {
			return nil, fmt.Errorf("%w: %v", ErrListObjects, obj.Err)
		}
		objects = append(objects, ObjectInfo{
			Name:         obj.Key,
			Size:         obj.Size,
			ContentType:  obj.ContentType,
			LastModified: obj.LastModified,
		})
	}

	return objects, nil
}
