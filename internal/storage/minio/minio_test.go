//go:build integration

package minio

import (
	"backend/internal/storage/testhelper"
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/testcontainers/testcontainers-go"
)

var (
	minioContainer testcontainers.Container
	minioInfo      testhelper.MinioConnInfo
	storage        *Storage
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	minioContainer, minioInfo, err = testhelper.StartMinio(ctx)
	if err != nil {
		log.Fatalf("failed to start minio container: %v", err)
	}

	storage, err = Init(ctx,
		minioInfo.Host, minioInfo.Port,
		minioInfo.User, minioInfo.Password,
		"test-bucket", "us-east-1", false,
	)
	if err != nil {
		log.Fatalf("Init() failed: %v", err)
	}

	code := m.Run()

	if err := minioContainer.Terminate(ctx); err != nil {
		log.Printf("failed to terminate minio container: %v", err)
	}
	os.Exit(code)
}

func TestInit_CreatesBucket(t *testing.T) {
	ctx := context.Background()

	s, err := Init(ctx,
		minioInfo.Host, minioInfo.Port,
		minioInfo.User, minioInfo.Password,
		"another-bucket", "us-east-1", false,
	)
	if err != nil {
		t.Fatalf("Init() for new bucket failed: %v", err)
	}
	_ = s

	s2, err := Init(ctx,
		minioInfo.Host, minioInfo.Port,
		minioInfo.User, minioInfo.Password,
		"another-bucket", "us-east-1", false,
	)
	if err != nil {
		t.Fatalf("Init() for existing bucket failed: %v", err)
	}
	_ = s2
}

func TestPutGetObject_Roundtrip(t *testing.T) {
	ctx := context.Background()
	content := []byte("hello adapt_ed integration test")
	objectName := "test/roundtrip.txt"

	_, err := storage.PutObject(ctx, objectName,
		bytes.NewReader(content), int64(len(content)), nil)
	if err != nil {
		t.Fatalf("PutObject() failed: %v", err)
	}

	reader, err := storage.GetObject(ctx, objectName)
	if err != nil {
		t.Fatalf("GetObject() failed: %v", err)
	}
	defer reader.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() failed: %v", err)
	}

	if diff := cmp.Diff(content, got); diff != "" {
		t.Fatalf("content mismatch (-want +got):\n%s", diff)
	}
}

func TestStatObject_Metadata(t *testing.T) {
	ctx := context.Background()
	content := []byte("metadata test content 12345")
	objectName := "test/metadata.bin"

	_, err := storage.PutObject(ctx, objectName,
		bytes.NewReader(content), int64(len(content)), nil)
	if err != nil {
		t.Fatalf("PutObject() failed: %v", err)
	}

	info, err := storage.StatObject(ctx, objectName)
	if err != nil {
		t.Fatalf("StatObject() failed: %v", err)
	}

	if info.Size != int64(len(content)) {
		t.Fatalf("size mismatch: want %d, got %d", len(content), info.Size)
	}

	// .bin has no registered MIME type, so PutObject falls back to application/octet-stream
	wantCT := "application/octet-stream"
	if info.ContentType != wantCT {
		t.Fatalf("content type mismatch: want %q, got %q", wantCT, info.ContentType)
	}
}

func TestRemoveObject_Deletes(t *testing.T) {
	ctx := context.Background()
	objectName := "test/to-delete.txt"
	content := []byte("delete me")

	_, err := storage.PutObject(ctx, objectName,
		bytes.NewReader(content), int64(len(content)), nil)
	if err != nil {
		t.Fatalf("PutObject() failed: %v", err)
	}

	if err := storage.RemoveObject(ctx, objectName); err != nil {
		t.Fatalf("RemoveObject() failed: %v", err)
	}

	_, err = storage.StatObject(ctx, objectName)
	if err == nil {
		t.Fatal("StatObject() expected error after RemoveObject, got nil")
	}
}

func TestListObjects_WithPrefix(t *testing.T) {
	ctx := context.Background()

	objects := map[string][]byte{
		"list/a/file1.txt": []byte("a1"),
		"list/a/file2.txt": []byte("a2"),
		"list/b/file3.txt": []byte("b3"),
	}
	for name, content := range objects {
		_, err := storage.PutObject(ctx, name,
			bytes.NewReader(content), int64(len(content)), nil)
		if err != nil {
			t.Fatalf("PutObject(%s) failed: %v", name, err)
		}
	}

	result, err := storage.ListObjects(ctx, "list/a/", true)
	if err != nil {
		t.Fatalf("ListObjects() failed: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(result))
	}

	names := make([]string, len(result))
	for i, obj := range result {
		names[i] = obj.Name
	}
	sort.Strings(names)

	wantNames := []string{"list/a/file1.txt", "list/a/file2.txt"}
	if diff := cmp.Diff(wantNames, names); diff != "" {
		t.Fatalf("object names mismatch (-want +got):\n%s", diff)
	}
}

func TestGetObjectRange_PartialRead(t *testing.T) {
	ctx := context.Background()
	content := []byte("0123456789ABCDEF")
	objectName := "test/range.bin"

	_, err := storage.PutObject(ctx, objectName,
		bytes.NewReader(content), int64(len(content)), nil)
	if err != nil {
		t.Fatalf("PutObject() failed: %v", err)
	}

	// Read bytes 4..9 (inclusive in HTTP Range)
	reader, err := storage.GetObjectRange(ctx, objectName, 4, 9)
	if err != nil {
		t.Fatalf("GetObjectRange() failed: %v", err)
	}
	defer reader.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() failed: %v", err)
	}

	// HTTP Range is inclusive on both ends: bytes 4,5,6,7,8,9
	want := content[4:10]
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("range content mismatch (-want +got):\n%s", diff)
	}
}
