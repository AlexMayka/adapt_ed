//go:build integration

package redis

import (
	"backend/internal/storage/testhelper"
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

var (
	rdContainer testcontainers.Container
	rdInfo      testhelper.RedisConnInfo
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	rdContainer, rdInfo, err = testhelper.StartRedis(ctx)
	if err != nil {
		log.Fatalf("failed to start redis container: %v", err)
	}

	code := m.Run()

	if err := rdContainer.Terminate(ctx); err != nil {
		log.Printf("failed to terminate redis container: %v", err)
	}
	os.Exit(code)
}

func newTestConnect(t *testing.T) *Connect {
	t.Helper()
	ctx := context.Background()
	iface, err := Init(ctx, rdInfo.Host, rdInfo.Port, 0, rdInfo.Password, false, 3, 10*time.Second)
	if err != nil {
		t.Fatalf("Init() unexpected error: %v", err)
	}
	con := iface.(*Connect)
	t.Cleanup(func() { con.Close() })
	return con
}

func TestInit_Success(t *testing.T) {
	ctx := context.Background()
	con := newTestConnect(t)

	if err := con.Ping(ctx); err != nil {
		t.Fatalf("Ping() failed: %v", err)
	}
}

func TestInit_WrongPassword(t *testing.T) {
	ctx := context.Background()
	_, err := Init(ctx, rdInfo.Host, rdInfo.Port, 0, "wrong_password", false, 1, 5*time.Second)
	if err == nil {
		t.Fatal("Init() expected error for wrong password, got nil")
	}
}

func TestSetGet_Roundtrip(t *testing.T) {
	ctx := context.Background()
	con := newTestConnect(t)

	key := "test:greeting"
	val := "hello, redis"

	if err := con.Client().Set(ctx, key, val, time.Minute).Err(); err != nil {
		t.Fatalf("SET failed: %v", err)
	}

	got, err := con.Client().Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if got != val {
		t.Fatalf("value mismatch: want %q, got %q", val, got)
	}
}

func TestSetGet_Expiration(t *testing.T) {
	ctx := context.Background()
	con := newTestConnect(t)

	key := "test:expiring"
	con.Client().Set(ctx, key, "temp", 100*time.Millisecond)

	time.Sleep(200 * time.Millisecond)

	exists, err := con.Client().Exists(ctx, key).Result()
	if err != nil {
		t.Fatalf("EXISTS failed: %v", err)
	}
	if exists != 0 {
		t.Fatal("expected key to be expired, but it still exists")
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	con := newTestConnect(t)

	key := "test:delete_me"
	con.Client().Set(ctx, key, "bye", 0)

	deleted, err := con.Client().Del(ctx, key).Result()
	if err != nil {
		t.Fatalf("DEL failed: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted key, got %d", deleted)
	}

	exists, _ := con.Client().Exists(ctx, key).Result()
	if exists != 0 {
		t.Fatal("key still exists after DEL")
	}
}

func TestCacheStorage_SetGetDel(t *testing.T) {
	ctx := context.Background()
	con := newTestConnect(t)

	t.Run("set and get", func(t *testing.T) {
		if err := con.Set(ctx, "cs:key1", "value1", time.Minute); err != nil {
			t.Fatalf("Set() failed: %v", err)
		}
		got, err := con.Get(ctx, "cs:key1")
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		if got != "value1" {
			t.Fatalf("Get() = %q, want %q", got, "value1")
		}
	})

	t.Run("get missing key returns empty string", func(t *testing.T) {
		got, err := con.Get(ctx, "cs:nonexistent")
		if err != nil {
			t.Fatalf("Get() unexpected error: %v", err)
		}
		if got != "" {
			t.Fatalf("Get() = %q, want empty string", got)
		}
	})

	t.Run("del removes key", func(t *testing.T) {
		con.Set(ctx, "cs:delme", "temp", time.Minute)

		if err := con.Del(ctx, "cs:delme"); err != nil {
			t.Fatalf("Del() failed: %v", err)
		}
		got, _ := con.Get(ctx, "cs:delme")
		if got != "" {
			t.Fatalf("Get() after Del() = %q, want empty", got)
		}
	})

	t.Run("set with ttl expires", func(t *testing.T) {
		con.Set(ctx, "cs:expiring", "temp", 100*time.Millisecond)
		time.Sleep(200 * time.Millisecond)

		got, err := con.Get(ctx, "cs:expiring")
		if err != nil {
			t.Fatalf("Get() unexpected error: %v", err)
		}
		if got != "" {
			t.Fatalf("Get() = %q, expected empty after TTL", got)
		}
	})

	t.Run("del nonexistent key no error", func(t *testing.T) {
		if err := con.Del(ctx, "cs:never_existed"); err != nil {
			t.Fatalf("Del() on missing key failed: %v", err)
		}
	})
}

func TestDifferentDBs(t *testing.T) {
	ctx := context.Background()

	con0 := newTestConnect(t) // db=0

	con1, err := Init(ctx, rdInfo.Host, rdInfo.Port, 1, rdInfo.Password, false, 3, 10*time.Second)
	if err != nil {
		t.Fatalf("Init(db=1) failed: %v", err)
	}
	t.Cleanup(func() { con1.Close() })

	key := "test:db_isolation"
	con0.Client().Set(ctx, key, "in_db0", 0)

	exists, _ := con1.Client().Exists(ctx, key).Result()
	if exists != 0 {
		t.Fatal("key from db=0 should not be visible in db=1")
	}
}
