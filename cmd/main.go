package main

import (
	"backend/internal/config"
	"backend/internal/logger"
	logInf "backend/internal/logger/interfaces"
	"backend/internal/storage"
	stgInf "backend/internal/storage/interfaces"
	"context"
	"errors"
	"fmt"
)

var (
	ErrInitConfig = errors.New("init config error")
	ErrInitLogger = errors.New("init logger error")
)

// main loads runtime configuration and prints it for local sanity-check runs.
func main() {
	cnf, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("%w: %s", ErrInitConfig, err))
	}

	log, err := logger.Init(
		cnf.Env.Version,
		cnf.Env.Instance,
		cnf.Env.Type,
		cnf.Log.LogLevel,
		cnf.App.Service,
		cnf.Log.IsLogging,
		logInf.Slog,
	)

	if err != nil {
		panic(fmt.Errorf("%w: %s", ErrInitLogger, err))
	}

	log.Info("Начало работы программы",
		"env", cnf.Env.Instance,
		"type", cnf.Env.Type,
		"version", cnf.Env.Version)

	log.Info("Конфиг инициализировался: ✅", "cnf", fmt.Sprintf("%+v\n", cnf))
	log.Info("Логгер инициализировался: ✅", "log", fmt.Sprintf("%+v", log))

	fmt.Printf("%+v", cnf)

	ctx := context.Background()

	psg, err := storage.InitDb(
		ctx,
		cnf.DB.Host,
		cnf.DB.User,
		cnf.DB.Password,
		cnf.DB.Database,
		cnf.DB.Port,
		cnf.DB.MaxConns,
		cnf.DB.MinConns,
		cnf.DB.ConnLifeTime,
		cnf.DB.ConnIdleTime,
		cnf.DB.HealthCheckPeriod,
		cnf.DB.QueryTimeout,
		cnf.DB.PingTimeout,
		cnf.DB.SSLMode,
		stgInf.Postgres,
	)

	if err != nil {
		log.Error("БД не подключена: ❌. Остановка приложения", "app", cnf.App.Service, "err", err)
		panic(err)
	}

	log.Info("Бд подключена: ✅", "db", fmt.Sprintf("%+v", psg))

	cache, err := storage.InitCache(
		ctx, cnf.Redis.Host,
		cnf.Redis.Port,
		cnf.Redis.DB,
		cnf.Redis.Password,
		cnf.Redis.UseSSL,
		cnf.Redis.MaxRetries,
		cnf.Redis.Timeout,
		stgInf.Redis,
	)

	if err != nil {
		log.Error("Кэш не подключен: ❌. Остановка приложения", "app", cnf.App.Service, "err", err)
		panic(err)
	}

	log.Info("Кэш подключен: ✅", "cache", fmt.Sprintf("%+v", cache))

	s3, err := storage.InitS3(
		ctx,
		cnf.Minio.Host,
		cnf.Minio.ApiPort,
		cnf.Minio.User,
		cnf.Minio.Password,
		cnf.Minio.Bucket,
		cnf.Minio.RegionName,
		cnf.Minio.ObjectLocking,
		cnf.Minio.UseSSL,
		stgInf.Minio,
	)

	if err != nil {
		log.Error("S3 не подключена: ❌", "app", cnf.App.Service, "err", err)
		panic(err)
	}

	log.Info("S3 подключена: ✅", "s3", fmt.Sprintf("%+v", s3))
}
