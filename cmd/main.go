package main

import (
	"backend/internal/config"
	"backend/internal/logger"
	logInf "backend/internal/logger/interfaces"
	"backend/internal/routers"
	"backend/internal/storage"
	stgInf "backend/internal/storage/interfaces"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	ErrInitConfig = errors.New("init config error")
	ErrInitLogger = errors.New("init logger error")

	ErrCloseDB      = errors.New("close db error")
	ErrCloseCache   = errors.New("close redis error")
	ErrCloseS3      = errors.New("close s3 error")
	ErrShutdownHTTP = errors.New("shutdown http server error")
)

// gracefulShutdown drains in-flight HTTP requests, then closes storage connections
// in reverse initialization order: S3 → Cache → DB.
func gracefulShutdown(srv *http.Server, db stgInf.DbStorage, cache stgInf.CacheStorage, s3 stgInf.S3Storage) []error {
	var errs []error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("%w: %w", ErrShutdownHTTP, err))
	}

	if err := s3.Close(); err != nil {
		errs = append(errs, fmt.Errorf("%w: %w", ErrCloseS3, err))
	}

	if err := cache.Close(); err != nil {
		errs = append(errs, fmt.Errorf("%w: %w", ErrCloseCache, err))
	}

	if err := db.Close(); err != nil {
		errs = append(errs, fmt.Errorf("%w: %w", ErrCloseDB, err))
	}

	return errs
}

// main loads runtime configuration and starts the HTTP server.
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
		os.Exit(1)
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
		os.Exit(1)
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
		os.Exit(1)
	}

	log.Info("S3 подключена: ✅", "s3", fmt.Sprintf("%+v", s3))

	// HTTP server
	router := routers.NewRouter(routers.Deps{
		Logger: log,
		DB:     psg,
		Cache:  cache,
		S3:     s3,
	}, cnf.Env.Type)

	addr := fmt.Sprintf("%s:%d", cnf.App.Host, cnf.App.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cnf.HTTP.ReadTimeout,
		WriteTimeout: cnf.HTTP.WriteTimeout,
		IdleTimeout:  cnf.HTTP.IdleTimeout,
	}

	go func() {
		log.Info("HTTP сервер запущен", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP сервер остановлен с ошибкой: ❌", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("Остановка сервера")
	errs := gracefulShutdown(srv, psg, cache, s3)

	if len(errs) > 0 {
		log.Error("Ошибки при завершении работы ❌", "app", cnf.App.Service, "err", errs)
		os.Exit(1)
	}

	log.Info("Успешное завершение приложения ✅")
}
