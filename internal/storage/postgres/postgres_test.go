//go:build integration

package postgres

import (
	appErr "backend/internal/errors"
	"backend/internal/storage/testhelper"
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
)

var (
	pgContainer testcontainers.Container
	pgInfo      testhelper.PgConnInfo
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	pgContainer, pgInfo, err = testhelper.StartPostgres(ctx)
	if err != nil {
		log.Fatalf("failed to start postgres container: %v", err)
	}

	code := m.Run()

	if err := pgContainer.Terminate(ctx); err != nil {
		log.Printf("failed to terminate postgres container: %v", err)
	}
	os.Exit(code)
}

func migrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations")
}

func newTestPool(t *testing.T) *PoolPsg {
	t.Helper()
	ctx := context.Background()
	pool, err := NewPool(ctx,
		pgInfo.Host, pgInfo.User, pgInfo.Password, pgInfo.Database, pgInfo.Port,
		5, 1,
		time.Minute, time.Minute, 30*time.Second,
		10*time.Second, 5*time.Second, "disable",
	)
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}
	t.Cleanup(func() { pool.Pool.Close() })
	return pool
}

func applyMigrations(t *testing.T, pool *PoolPsg) *sql.DB {
	t.Helper()
	db := stdlib.OpenDBFromPool(pool.Pool)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose.SetDialect() failed: %v", err)
	}
	if err := goose.Up(db, migrationsDir()); err != nil {
		t.Fatalf("goose.Up() failed: %v", err)
	}
	return db
}

func TestNewPool_Success(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t)

	if err := pool.Pool.Ping(ctx); err != nil {
		t.Fatalf("pool.Ping() failed: %v", err)
	}
}

func TestNewPool_WrongCredentials(t *testing.T) {
	ctx := context.Background()
	_, err := NewPool(ctx,
		pgInfo.Host, "wrong_user", "wrong_password", pgInfo.Database, pgInfo.Port,
		5, 1,
		time.Minute, time.Minute, 30*time.Second,
		10*time.Second, 5*time.Second, "disable",
	)
	if err == nil {
		t.Fatal("NewPool() expected error for wrong credentials, got nil")
	}
	if !errors.Is(err, appErr.ErrPgPing) && !errors.Is(err, appErr.ErrPgCreatePool) {
		t.Fatalf("expected ErrPgPing or ErrPgCreatePool, got: %v", err)
	}
}

func TestMigrations_ApplyCleanly(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t)
	db := applyMigrations(t, pool)
	defer db.Close()

	var exists bool
	err := db.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')").
		Scan(&exists)
	if err != nil {
		t.Fatalf("failed to check table existence: %v", err)
	}
	if !exists {
		t.Fatal("expected 'users' table to exist after migrations")
	}
}

func TestBasicQuery_InsertSelect(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t)
	db := applyMigrations(t, pool)
	defer db.Close()

	var id string
	err := pool.Pool.QueryRow(ctx,
		"INSERT INTO schools (name, city) VALUES ($1, $2) RETURNING id",
		"Test School", "Moscow",
	).Scan(&id)
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty id from INSERT")
	}

	var name, city string
	err = pool.Pool.QueryRow(ctx,
		"SELECT name, city FROM schools WHERE id = $1", id,
	).Scan(&name, &city)
	if err != nil {
		t.Fatalf("SELECT failed: %v", err)
	}

	if diff := cmp.Diff("Test School", name); diff != "" {
		t.Fatalf("name mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff("Moscow", city); diff != "" {
		t.Fatalf("city mismatch (-want +got):\n%s", diff)
	}
}
