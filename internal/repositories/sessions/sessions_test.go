package sessions

import (
	appErr "backend/internal/errors"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

// mockCache — простой мок CacheStorage для unit-тестов.
type mockCache struct {
	store map[string]string
}

func newMockCache() *mockCache {
	return &mockCache{store: make(map[string]string)}
}

func (m *mockCache) Close() error                    { return nil }
func (m *mockCache) Ping(_ context.Context) error    { return nil }

func (m *mockCache) Set(_ context.Context, key, value string, _ time.Duration) error {
	m.store[key] = value
	return nil
}

func (m *mockCache) Get(_ context.Context, key string) (string, error) {
	v, ok := m.store[key]
	if !ok {
		return "", nil
	}
	return v, nil
}

func (m *mockCache) Del(_ context.Context, key string) error {
	delete(m.store, key)
	return nil
}

func TestSetSessionVersion(t *testing.T) {
	cache := newMockCache()
	repo := NewSessionRepository(cache)
	ctx := context.Background()
	uid := uuid.New()

	if err := repo.SetSessionVersion(ctx, uid, 5, time.Minute); err != nil {
		t.Fatalf("SetSessionVersion() failed: %v", err)
	}

	got, err := repo.GetSessionVersion(ctx, uid)
	if err != nil {
		t.Fatalf("GetSessionVersion() failed: %v", err)
	}
	if got != 5 {
		t.Fatalf("GetSessionVersion() = %d, want 5", got)
	}
}

func TestGetSessionVersion_CacheMiss(t *testing.T) {
	cache := newMockCache()
	repo := NewSessionRepository(cache)
	ctx := context.Background()

	_, err := repo.GetSessionVersion(ctx, uuid.New())
	if !errors.Is(err, appErr.ErrCacheMiss) {
		t.Fatalf("GetSessionVersion() on miss error = %v, want ErrCacheMiss", err)
	}
}

func TestSetRefreshTokenHash(t *testing.T) {
	cache := newMockCache()
	repo := NewSessionRepository(cache)
	ctx := context.Background()
	uid := uuid.New()
	hash := "bcrypt_hash_here"

	if err := repo.SetRefreshTokenHash(ctx, uid, hash, time.Hour); err != nil {
		t.Fatalf("SetRefreshTokenHash() failed: %v", err)
	}

	got, err := repo.GetRefreshTokenHash(ctx, uid)
	if err != nil {
		t.Fatalf("GetRefreshTokenHash() failed: %v", err)
	}
	if got != hash {
		t.Fatalf("GetRefreshTokenHash() = %q, want %q", got, hash)
	}
}

func TestGetRefreshTokenHash_CacheMiss(t *testing.T) {
	cache := newMockCache()
	repo := NewSessionRepository(cache)
	ctx := context.Background()

	got, err := repo.GetRefreshTokenHash(ctx, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Fatalf("GetRefreshTokenHash() on miss = %q, want empty", got)
	}
}

func TestDelRefreshTokenHash(t *testing.T) {
	cache := newMockCache()
	repo := NewSessionRepository(cache)
	ctx := context.Background()
	uid := uuid.New()

	repo.SetRefreshTokenHash(ctx, uid, "hash", time.Minute)

	if err := repo.DelRefreshTokenHash(ctx, uid); err != nil {
		t.Fatalf("DelRefreshTokenHash() failed: %v", err)
	}

	tok, _ := repo.GetRefreshTokenHash(ctx, uid)
	if tok != "" {
		t.Fatalf("refresh token after Del = %q, want empty", tok)
	}
}

func TestDelSession(t *testing.T) {
	cache := newMockCache()
	repo := NewSessionRepository(cache)
	ctx := context.Background()
	uid := uuid.New()

	repo.SetSessionVersion(ctx, uid, 3, time.Minute)
	repo.SetRefreshTokenHash(ctx, uid, "hash", time.Minute)

	if err := repo.DelSession(ctx, uid); err != nil {
		t.Fatalf("DelSession() failed: %v", err)
	}

	_, err := repo.GetSessionVersion(ctx, uid)
	if !errors.Is(err, appErr.ErrCacheMiss) {
		t.Fatalf("session version after Del error = %v, want ErrCacheMiss", err)
	}

	tok, _ := repo.GetRefreshTokenHash(ctx, uid)
	if tok != "" {
		t.Fatalf("refresh token after Del = %q, want empty", tok)
	}
}
