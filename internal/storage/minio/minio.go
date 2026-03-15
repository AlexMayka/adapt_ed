package minio

import (
	"backend/internal/storage/interfaces"
	typesSg "backend/internal/storage/types"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"path/filepath"

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
	ErrPutObject      = errors.New("failed to put object")
)

// Storage wraps a MinIO client scoped to a single bucket.
type Storage struct {
	client *minio.Client
	bucket string
}

// Init creates a MinIO client, checks if the target bucket exists and creates it if needed.
func Init(ctx context.Context, host string, port int, user, password, bucket, region string, objectLocking, useSSL bool) (interfaces.S3Storage, error) {
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

// Close releases the MinIO client reference.
func (s *Storage) Close() error {
	s.client = nil
	return nil
}

// PutObject uploads data from reader into the bucket under objectName.
// Content-Type is auto-detected from the file extension; size must match the reader length.
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
		return nil, fmt.Errorf("%w: %v", ErrPutObject, err)
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

// GetObject returns the full object content as an io.ReadCloser. Caller must close it.
func (s *Storage) GetObject(ctx context.Context, name string) (io.ReadCloser, error) {
	reader, err := s.client.GetObject(ctx, s.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetObject, err)
	}
	return reader, nil
}

// GetObjectRange returns a byte range [start, end] of the object. Caller must close the reader.
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

// StatObject returns metadata (size, content-type, etc.) for the named object.
func (s *Storage) StatObject(ctx context.Context, name string) (*typesSg.ObjectInfo, error) {
	infoMn, err := s.client.StatObject(ctx, s.bucket, name, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStatObject, err)
	}

	info := &typesSg.ObjectInfo{
		Name:         infoMn.Key,
		Size:         infoMn.Size,
		ContentType:  infoMn.ContentType,
		LastModified: infoMn.LastModified,
	}

	return info, nil
}

// RemoveObject deletes the named object from the bucket.
func (s *Storage) RemoveObject(ctx context.Context, name string) error {
	err := s.client.RemoveObject(ctx, s.bucket, name, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRemoveObject, err)
	}
	return nil
}

// ListObjects returns all objects matching the prefix. When recursive is false,
// only the top-level entries under the prefix are returned.
func (s *Storage) ListObjects(ctx context.Context, prefix string, recursive bool) ([]typesSg.ObjectInfo, error) {
	var objects []typesSg.ObjectInfo
	for obj := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	}) {
		if obj.Err != nil {
			return nil, fmt.Errorf("%w: %v", ErrListObjects, obj.Err)
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
