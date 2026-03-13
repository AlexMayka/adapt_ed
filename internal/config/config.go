package config

import (
	"backend/internal/utils"
	"errors"
	"fmt"
	"time"
)

// EnvConfig contains settings for environment
type EnvConfig struct {
	Version  string
	Instance string
	Type     string
}

// AppConfig contains application-level runtime settings.
type AppConfig struct {
	Service string
	Host    string
	Port    int
	Secret  string
}

// LogConfig contains settings for logging
type LogConfig struct {
	IsLogging bool
	LogLevel  string
}

// MinioConfig contains settings for S3-compatible object storage.
type MinioConfig struct {
	Host     string
	User     string
	Password string
	Bucket   string
	ApiPort  int

	RegionName    string
	ObjectLocking bool
	ForceCreate   bool
}

// DBConfig contains PostgreSQL connection and pool settings.
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string

	MaxConns          int32
	MinConns          int32
	ConnLifeTime      time.Duration
	ConnIdleTime      time.Duration
	QueryTimeout      time.Duration
	HealthCheckPeriod time.Duration
	PingTimeout       time.Duration
}

// Config groups all runtime configuration sections.
type Config struct {
	Env   *EnvConfig
	App   *AppConfig
	Log   *LogConfig
	DB    *DBConfig
	Minio *MinioConfig
}

// appendErr keeps the env key context while collecting multiple errors.
func appendErr(errs []error, key string, err error) []error {
	if err != nil {
		return append(errs, fmt.Errorf("%s: %w", key, err))
	}
	return errs
}

// loadEnv parses env vars into Config fields and applies default values.
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

	// логирование
	appIsLogging, err := utils.GetEnvDefault[bool]("APP_IS_LOGGING", true)
	errs = appendErr(errs, "APP_IS_LOGGING", err)

	appLogLevel, err := utils.GetEnvDefault[string]("APP_LOG_LEVEL", "info")
	errs = appendErr(errs, "APP_LOG_LEVEL", err)

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
		Log: &LogConfig{
			IsLogging: appIsLogging,
			LogLevel:  appLogLevel,
		},
		DB: &DBConfig{
			Host:     dbHost,
			Password: dbPassword,
			User:     dbUser,
			Database: dbName,
			Port:     dbPort,

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

			RegionName:    mnRegion,
			ObjectLocking: mnObjectLocking,
			ForceCreate:   mnForceCreate,
		},
	}, nil
}

// Load reads environment variables, applies defaults, and validates the result.
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
