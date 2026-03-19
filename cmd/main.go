package main

import (
	"backend/internal/config"
	appErr "backend/internal/errors"
	"backend/internal/logger"
	logInf "backend/internal/logger/interfaces"
	"backend/internal/routers"
	"backend/internal/storage"
	stgInf "backend/internal/storage/interfaces"
	"backend/internal/storage/postgres"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	docs "backend/docs"
)

// gracefulShutdown завершает обработку текущих запросов и закрывает хранилища
// в обратном порядке инициализации: S3 → Cache → DB.
func gracefulShutdown(srv *http.Server, db *postgres.PoolPsg, cache stgInf.CacheStorage, s3 stgInf.S3Storage) []error {
	var errs []error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("%w: %w", appErr.ErrShutdownHTTP, err))
	}

	if err := s3.Close(); err != nil {
		errs = append(errs, fmt.Errorf("%w: %w", appErr.ErrCloseS3, err))
	}

	if err := cache.Close(); err != nil {
		errs = append(errs, fmt.Errorf("%w: %w", appErr.ErrCloseCache, err))
	}

	if err := db.Close(); err != nil {
		errs = append(errs, fmt.Errorf("%w: %w", appErr.ErrCloseDB, err))
	}

	return errs
}

func setSwagger(appName, host, version, envType, instance string) {
	docs.SwaggerInfo.Title = appName
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Schemes = []string{"http"}
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Host = host

	docs.SwaggerInfo.Description = fmt.Sprintf("Бэкенд API приложения для сервиса Adapt Education.\n"+
		" Env: %s, Instance: %s", envType, instance)
}

// @title API
// @version 1.0
// @description API description
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer {access_token}

// @tag.name infra
// @tag.description Служебные эндпоинты приложения: доступность, готовность и метрики мониторинга.
// @tag.name auth
// @tag.description Эндпоинты для работы с авторизацией пользователей
// @tag.name schools
// @tag.description Управление школами: создание, обновление, удаление и поиск.
// @tag.name classes
// @tag.description Управление классами внутри школы: создание, обновление, удаление и поиск.
// @tag.name users
// @tag.description Управление пользователями: профиль, пароль, список, активация, удаление.
func main() {
	cnf, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("%w: %s", appErr.ErrInitConfig, err))
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
		panic(fmt.Errorf("%w: %s", appErr.ErrInitLogger, err))
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

	addr := fmt.Sprintf("%s:%d", cnf.App.Host, cnf.App.Port)
	setSwagger(cnf.App.Service, addr, cnf.Env.Version, cnf.Env.Type, cnf.Env.Instance)

	router := routers.NewRouter(routers.Deps{
		Logger: log,
		Config: cnf,
		DB:     psg,
		Cache:  cache,
		S3:     s3,
	}, cnf.Env.Type)
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
