package testhelper

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PgConnInfo struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type MinioConnInfo struct {
	Host     string
	Port     int
	User     string
	Password string
}

type RedisConnInfo struct {
	Host     string
	Port     int
	Password string
}

func StartPostgres(ctx context.Context) (testcontainers.Container, PgConnInfo, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test_adapt_ed",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, PgConnInfo{}, fmt.Errorf("start postgres container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, PgConnInfo{}, fmt.Errorf("get postgres host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, PgConnInfo{}, fmt.Errorf("get postgres port: %w", err)
	}

	return container, PgConnInfo{
		Host:     host,
		Port:     mappedPort.Int(),
		User:     "test",
		Password: "test",
		Database: "test_adapt_ed",
	}, nil
}

func StartMinio(ctx context.Context) (testcontainers.Container, MinioConnInfo, error) {
	req := testcontainers.ContainerRequest{
		Image:        "minio/minio:latest",
		ExposedPorts: []string{"9000/tcp"},
		Env: map[string]string{
			"MINIO_ROOT_USER":     "minioadmin",
			"MINIO_ROOT_PASSWORD": "minioadmin",
		},
		Cmd: []string{"server", "/data"},
		WaitingFor: wait.ForHTTP("/minio/health/live").
			WithPort("9000/tcp").
			WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, MinioConnInfo{}, fmt.Errorf("start minio container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, MinioConnInfo{}, fmt.Errorf("get minio host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "9000")
	if err != nil {
		return nil, MinioConnInfo{}, fmt.Errorf("get minio port: %w", err)
	}

	return container, MinioConnInfo{
		Host:     host,
		Port:     mappedPort.Int(),
		User:     "minioadmin",
		Password: "minioadmin",
	}, nil
}

func StartRedis(ctx context.Context) (testcontainers.Container, RedisConnInfo, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7",
		ExposedPorts: []string{"6379/tcp"},
		Cmd:          []string{"redis-server", "--requirepass", "test"},
		WaitingFor: wait.ForLog("Ready to accept connections").
			WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, RedisConnInfo{}, fmt.Errorf("start redis container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, RedisConnInfo{}, fmt.Errorf("get redis host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return nil, RedisConnInfo{}, fmt.Errorf("get redis port: %w", err)
	}

	return container, RedisConnInfo{
		Host:     host,
		Port:     mappedPort.Int(),
		Password: "test",
	}, nil
}
