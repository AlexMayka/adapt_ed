package config

import (
	"backend/internal/utils"
	"errors"
	"fmt"
	"time"
)

// EnvConfig содержит настройки окружения.
type EnvConfig struct {
	Version  string
	Instance string
	Type     string
}

// AppConfig содержит настройки приложения.
type AppConfig struct {
	Service string
	Host    string
	Port    int
	Secret  string
}

// AuthConfig содержит настройки JWT и refresh-токенов.
type AuthConfig struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// LogConfig содержит настройки логирования.
type LogConfig struct {
	IsLogging bool
	LogLevel  string
}

// MinioConfig содержит настройки S3-совместимого объектного хранилища.
type MinioConfig struct {
	Host     string
	User     string
	Password string
	Bucket   string
	ApiPort  int
	UseSSL   bool

	RegionName    string
	ObjectLocking bool
	ForceCreate   bool
}

// DBConfig содержит настройки подключения и пула PostgreSQL.
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string

	MaxConns          int32
	MinConns          int32
	ConnLifeTime      time.Duration
	ConnIdleTime      time.Duration
	QueryTimeout      time.Duration
	HealthCheckPeriod time.Duration
	PingTimeout       time.Duration
}

// HTTPConfig содержит настройки таймаутов HTTP-сервера.
type HTTPConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// RedisConfig содержит настройки подключения к Redis.
type RedisConfig struct {
	Host       string
	Port       int
	Password   string
	DB         int
	UseSSL     bool
	MaxRetries int
	Timeout    time.Duration
}

// Config объединяет все секции конфигурации приложения.
type Config struct {
	Env   *EnvConfig
	App   *AppConfig
	Auth  *AuthConfig
	Log   *LogConfig
	HTTP  *HTTPConfig
	DB    *DBConfig
	Minio *MinioConfig
	Redis *RedisConfig
}

// appendErr добавляет ошибку с контекстом ключа в список ошибок.
func appendErr(errs []error, key string, err error) []error {
	if err != nil {
		return append(errs, fmt.Errorf("%s: %w", key, err))
	}
	return errs
}

// loadEnv считывает переменные окружения в поля Config с применением значений по умолчанию.
func loadEnv() (*Config, error) {
	var errs []error
	var err error

	// настройки окружения
	appVersion, err := utils.GetEnv[string]("APP_VERSION")
	errs = appendErr(errs, "APP_VERSION", err)

	appInstance, err := utils.GetEnvDefault[string]("APP_INSTANCE", "local")
	errs = appendErr(errs, "APP_INSTANCE", err)

	envType, err := utils.GetEnvDefault[string]("APP_TYPE", "dev")
	errs = appendErr(errs, "APP_TYPE", err)

	// настройки приложения
	appService, err := utils.GetEnvDefault[string]("APP_SERVICE", "backend_adapt_ed")
	errs = appendErr(errs, "APP_SERVICE", err)

	appHost, err := utils.GetEnvDefault[string]("APP_HOST", "localhost")
	errs = appendErr(errs, "APP_HOST", err)

	appPort, err := utils.GetEnvDefault[int]("APP_PORT", 8000)
	errs = appendErr(errs, "APP_PORT", err)

	appSecret, err := utils.GetEnv[string]("APP_SECRET")
	errs = appendErr(errs, "APP_SECRET", err)

	// настройки авторизации
	authAccessTTL, err := utils.GetDurationEnvDefault("AUTH_ACCESS_TTL", 15*time.Minute)
	errs = appendErr(errs, "AUTH_ACCESS_TTL", err)

	authRefreshTTL, err := utils.GetDurationEnvDefault("AUTH_REFRESH_TTL", 30*24*time.Hour)
	errs = appendErr(errs, "AUTH_REFRESH_TTL", err)

	// логирование
	appIsLogging, err := utils.GetEnvDefault[bool]("APP_IS_LOGGING", true)
	errs = appendErr(errs, "APP_IS_LOGGING", err)

	appLogLevel, err := utils.GetEnvDefault[string]("APP_LOG_LEVEL", "info")
	errs = appendErr(errs, "APP_LOG_LEVEL", err)

	// настройки HTTP сервера
	httpReadTimeout, err := utils.GetDurationEnvDefault("APP_HTTP_READ_TIMEOUT", 10*time.Second)
	errs = appendErr(errs, "APP_HTTP_READ_TIMEOUT", err)

	httpWriteTimeout, err := utils.GetDurationEnvDefault("APP_HTTP_WRITE_TIMEOUT", 10*time.Second)
	errs = appendErr(errs, "APP_HTTP_WRITE_TIMEOUT", err)

	httpIdleTimeout, err := utils.GetDurationEnvDefault("APP_HTTP_IDLE_TIMEOUT", 120*time.Second)
	errs = appendErr(errs, "APP_HTTP_IDLE_TIMEOUT", err)

	// настройки БД
	dbUser, err := utils.GetEnvDefault[string]("PG_USER", "postgres_root")
	errs = appendErr(errs, "PG_USER", err)

	dbPassword, err := utils.GetEnv[string]("PG_PASSWORD")
	errs = appendErr(errs, "PG_PASSWORD", err)

	dbHost, err := utils.GetEnvDefault[string]("PG_HOST", "localhost")
	errs = appendErr(errs, "PG_HOST", err)

	dbPort, err := utils.GetEnvDefault[int]("PG_PORT", 5433)
	errs = appendErr(errs, "PG_PORT", err)

	dbName, err := utils.GetEnvDefault[string]("PG_DB", "adapt_ed")
	errs = appendErr(errs, "PG_DB", err)

	dbMaxConns, err := utils.GetEnvDefault[int32]("PG_MAX_CONNS", 20)
	errs = appendErr(errs, "PG_MAX_CONNS", err)

	dbMinConns, err := utils.GetEnvDefault[int32]("PG_MIN_CONNS", 20)
	errs = appendErr(errs, "PG_MIN_CONNS", err)

	dbConnLifetime, err := utils.GetDurationEnvDefault("PG_CONN_LIFETIME", time.Second*60)
	errs = appendErr(errs, "PG_CONN_LIFETIME", err)

	dbConnIdleTime, err := utils.GetDurationEnvDefault("PG_CONN_IDLE_TIME", time.Second*60)
	errs = appendErr(errs, "PG_CONN_IDLE_TIME", err)

	dbQueryTimeout, err := utils.GetDurationEnvDefault("PG_QUERY_TIMEOUT", time.Second*60)
	errs = appendErr(errs, "PG_QUERY_TIMEOUT", err)

	dbHealthCheck, err := utils.GetDurationEnvDefault("PG_HEALTH_CHECK_PERIOD", time.Second*30)
	errs = appendErr(errs, "PG_HEALTH_CHECK_PERIOD", err)

	dbPingTimeout, err := utils.GetDurationEnvDefault("PG_PING_TIMEOUT", time.Second*5)
	errs = appendErr(errs, "PG_PING_TIMEOUT", err)

	dbSSLMode, err := utils.GetEnvDefault[string]("PG_SSL_MODE", "disable")
	errs = appendErr(errs, "PG_SSL_MODE", err)

	// настройки MinIO
	mnUser, err := utils.GetEnvDefault[string]("MINIO_USER", "minio_root")
	errs = appendErr(errs, "MINIO_USER", err)

	mnPassword, err := utils.GetEnv[string]("MINIO_PASSWORD")
	errs = appendErr(errs, "MINIO_PASSWORD", err)

	mnHost, err := utils.GetEnvDefault[string]("MINIO_HOST", "localhost")
	errs = appendErr(errs, "MINIO_HOST", err)

	mnPortAPI, err := utils.GetEnvDefault[int]("MINIO_PORT_API", 9000)
	errs = appendErr(errs, "MINIO_PORT_API", err)

	mnBucket, err := utils.GetEnvDefault[string]("MINIO_BUCKET", "adapt_ed")
	errs = appendErr(errs, "MINIO_BUCKET", err)

	mnRegion, err := utils.GetEnvDefault[string]("MINIO_REGION", "us-east-1")
	errs = appendErr(errs, "MINIO_REGION", err)

	mnObjectLocking, err := utils.GetEnvDefault[bool]("MINIO_OBJECT_LOCKING", false)
	errs = appendErr(errs, "MINIO_OBJECT_LOCKING", err)

	mnForceCreate, err := utils.GetEnvDefault[bool]("MINIO_FORCE_CREATE", false)
	errs = appendErr(errs, "MINIO_FORCE_CREATE", err)

	mnUseSSL, err := utils.GetEnvDefault[bool]("MINIO_USE_SSL", false)
	errs = appendErr(errs, "MINIO_USE_SSL", err)

	// настройки Redis
	rdHost, err := utils.GetEnvDefault[string]("REDIS_HOST", "localhost")
	errs = appendErr(errs, "REDIS_HOST", err)

	rdPort, err := utils.GetEnvDefault[int]("REDIS_PORT", 6379)
	errs = appendErr(errs, "REDIS_PORT", err)

	rdPassword, err := utils.GetEnv[string]("REDIS_PASSWORD")
	errs = appendErr(errs, "REDIS_PASSWORD", err)

	rdDB, err := utils.GetEnvDefault[int]("REDIS_DB", 0)
	errs = appendErr(errs, "REDIS_DB", err)

	rdUseSSL, err := utils.GetEnvDefault[bool]("REDIS_USE_SSL", false)
	errs = appendErr(errs, "REDIS_USE_SSL", err)

	rdMaxRetries, err := utils.GetEnvDefault[int]("REDIS_MAX_RETRIES", 3)
	errs = appendErr(errs, "REDIS_MAX_RETRIES", err)

	rdTimeout, err := utils.GetDurationEnvDefault("REDIS_TIMEOUT", 10*time.Second)
	errs = appendErr(errs, "REDIS_TIMEOUT", err)

	if len(errs) > 0 {
		return nil, fmt.Errorf("config initialization failed: %w", errors.Join(errs...))
	}

	return &Config{
		Env: &EnvConfig{
			Version:  appVersion,
			Instance: appInstance,
			Type:     envType,
		},
		App: &AppConfig{
			Service: appService,
			Host:    appHost,
			Port:    appPort,
			Secret:  appSecret,
		},
		Auth: &AuthConfig{
			AccessTTL:  authAccessTTL,
			RefreshTTL: authRefreshTTL,
		},
		Log: &LogConfig{
			IsLogging: appIsLogging,
			LogLevel:  appLogLevel,
		},
		HTTP: &HTTPConfig{
			ReadTimeout:  httpReadTimeout,
			WriteTimeout: httpWriteTimeout,
			IdleTimeout:  httpIdleTimeout,
		},
		DB: &DBConfig{
			Host:     dbHost,
			Password: dbPassword,
			User:     dbUser,
			Database: dbName,
			Port:     dbPort,
			SSLMode:  dbSSLMode,

			MaxConns:          dbMaxConns,
			MinConns:          dbMinConns,
			ConnLifeTime:      dbConnLifetime,
			ConnIdleTime:      dbConnIdleTime,
			QueryTimeout:      dbQueryTimeout,
			HealthCheckPeriod: dbHealthCheck,
			PingTimeout:       dbPingTimeout,
		},
		Minio: &MinioConfig{
			Host:     mnHost,
			User:     mnUser,
			Password: mnPassword,
			Bucket:   mnBucket,
			ApiPort:  mnPortAPI,
			UseSSL:   mnUseSSL,

			RegionName:    mnRegion,
			ObjectLocking: mnObjectLocking,
			ForceCreate:   mnForceCreate,
		},
		Redis: &RedisConfig{
			Host:       rdHost,
			Port:       rdPort,
			Password:   rdPassword,
			DB:         rdDB,
			UseSSL:     rdUseSSL,
			MaxRetries: rdMaxRetries,
			Timeout:    rdTimeout,
		},
	}, nil
}

// Load загружает конфигурацию из окружения и валидирует результат.
func Load() (Config, error) {

	cfg, err := loadEnv()
	if err != nil {
		return Config{}, err
	}

	if err = cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("config validation failed: %w", err)
	}

	return *cfg, nil
}
