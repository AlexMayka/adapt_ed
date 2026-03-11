package config

import (
	"backend/internal/utils"
	"errors"
	"fmt"
	"time"
)

// AppConfig contains application-level runtime settings.
type AppConfig struct {
	Host   string
	Port   int
	Secret string

	Logging  bool
	LogLevel string
}

// MinioConfig contains settings for S3-compatible object storage.
type MinioConfig struct {
	Host     string
	User     string
	Password string
	Bucket   string
	ApiPort  int
}

// DBConfig contains PostgreSQL connection and pool settings.
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string

	MaxConns     int
	MinConns     int
	ConnLifeTime time.Duration
	ConnIdleTime time.Duration
	QueryTimeout time.Duration
}

// Config groups all runtime configuration sections.
type Config struct {
	App   *AppConfig
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

	// настройки приложения
	appHost, err := utils.GetEnvDefault[string]("SR_AP_HOST", "localhost")
	errs = appendErr(errs, "SR_AP_HOST", err)

	appPort, err := utils.GetEnvDefault[int]("SR_AP_PORT", 8000)
	errs = appendErr(errs, "SR_AP_PORT", err)

	appSecret, err := utils.GetEnv[string]("SR_AP_SECRET")
	errs = appendErr(errs, "SR_AP_SECRET", err)

	appLogging, err := utils.GetEnvDefault[bool]("SR_AP_LOGGING", true)
	errs = appendErr(errs, "SR_AP_LOGGING", err)

	appLogLevel, err := utils.GetEnvDefault[string]("SR_AP_LOG_LEVEL", "info")
	errs = appendErr(errs, "SR_AP_LOG_LEVEL", err)

	// настройки БД
	dbUser, err := utils.GetEnvDefault[string]("PG_USER", "postgres_root")
	errs = appendErr(errs, "PG_USER", err)

	dbPassword, err := utils.GetEnv[string]("PG_PASSWORD")
	errs = appendErr(errs, "PG_PASSWORD", err)

	dbHost, err := utils.GetEnvDefault[string]("PG_HOST", "localhost")
	errs = appendErr(errs, "PG_HOST", err)

	dbPort, err := utils.GetEnvDefault[int]("PG_PORT", 5433)
	errs = appendErr(errs, "PG_PORT", err)

	dbName, err := utils.GetEnvDefault[string]("PG_DB", "SALES_RADAR")
	errs = appendErr(errs, "PG_DB", err)

	dbMaxConns, err := utils.GetEnvDefault[int]("PG_MAX_CONNS", 20)
	errs = appendErr(errs, "PG_MAX_CONNS", err)

	dbMinConns, err := utils.GetEnvDefault[int]("PG_MIN_CONNS", 20)
	errs = appendErr(errs, "PG_MIN_CONNS", err)

	dbConnLifetime, err := utils.GetDurationEnvDefault("PG_CONN_LIFETIME", time.Second*60)
	errs = appendErr(errs, "PG_CONN_LIFETIME", err)

	dbConnIdleTime, err := utils.GetDurationEnvDefault("PG_CONN_IDLE_TIME", time.Second*60)
	errs = appendErr(errs, "PG_CONN_IDLE_TIME", err)

	dbQueryTimeout, err := utils.GetDurationEnvDefault("PG_QUERY_TIMEOUT", time.Second*60)
	errs = appendErr(errs, "PG_QUERY_TIMEOUT", err)

	// настройки MinIO
	mnUser, err := utils.GetEnvDefault[string]("SR_MN_USER", "minio_root")
	errs = appendErr(errs, "SR_MN_USER", err)

	mnPassword, err := utils.GetEnv[string]("SR_MN_PASSWORD")
	errs = appendErr(errs, "SR_MN_PASSWORD", err)

	mnHost, err := utils.GetEnvDefault[string]("SR_MN_HOST", "localhost")
	errs = appendErr(errs, "SR_MN_HOST", err)

	mnPortAPI, err := utils.GetEnvDefault[int]("SR_MN_PORT_API", 9000)
	errs = appendErr(errs, "SR_MN_PORT_API", err)

	mnBucket, err := utils.GetEnvDefault[string]("SR_MN_BUCKET", "Sales_Radar")
	errs = appendErr(errs, "SR_MN_BUCKET", err)

	if len(errs) > 0 {
		return nil, fmt.Errorf("config initialization failed: %w", errors.Join(errs...))
	}

	return &Config{
		App: &AppConfig{
			Host:     appHost,
			Port:     appPort,
			Secret:   appSecret,
			Logging:  appLogging,
			LogLevel: appLogLevel,
		},
		DB: &DBConfig{
			Host:     dbHost,
			Password: dbPassword,
			User:     dbUser,
			Database: dbName,
			Port:     dbPort,

			MaxConns:     dbMaxConns,
			MinConns:     dbMinConns,
			ConnLifeTime: dbConnLifetime,
			ConnIdleTime: dbConnIdleTime,
			QueryTimeout: dbQueryTimeout,
		},
		Minio: &MinioConfig{
			Host:     mnHost,
			User:     mnUser,
			Password: mnPassword,
			Bucket:   mnBucket,
			ApiPort:  mnPortAPI,
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
