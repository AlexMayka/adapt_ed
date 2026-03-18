package config

import (
	appErr "backend/internal/errors"
	"backend/internal/utils"
	"errors"
	"fmt"
)

// Validate проверяет семантические правила для уже загруженной конфигурации.
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("%w: config is nil", appErr.ErrValidationFailed)
	}

	var errs []error

	errs = appendErr(errs, "app", validateApp(c.App))
	errs = appendErr(errs, "http", validateHTTP(c.HTTP))
	errs = appendErr(errs, "db", validateDB(c.DB))
	errs = appendErr(errs, "minio", validateMinio(c.Minio))
	errs = appendErr(errs, "redis", validateRedis(c.Redis))
	errs = appendErr(errs, "log", validateLog(c.Log))
	errs = appendErr(errs, "env", validateEnv(c.Env))

	if len(errs) > 0 {
		return fmt.Errorf("%w: %w", appErr.ErrValidationFailed, errors.Join(errs...))
	}

	return nil
}

// validateApp валидирует настройки приложения.
func validateApp(app *AppConfig) error {
	if app == nil {
		return fmt.Errorf("%w: app is nil", appErr.ErrValidationFailed)
	}

	var errs []error
	errs = appendErr(errs, "APP_SERVICE", utils.ValidateEmptinessParam("APP_SERVICE", app.Service))
	errs = appendErr(errs, "APP_HOST", utils.ValidateEmptinessParam("APP_HOST", app.Host))
	errs = appendErr(errs, "APP_PORT", utils.ValidatePort(app.Port))
	errs = appendErr(errs, "APP_SECRET", utils.ValidateEmptinessParam("APP_SECRET", app.Secret))

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// validateHTTP валидирует таймауты HTTP-сервера.
func validateHTTP(http *HTTPConfig) error {
	if http == nil {
		return fmt.Errorf("%w: http is nil", appErr.ErrValidationFailed)
	}

	var errs []error

	if http.ReadTimeout <= 0 {
		errs = append(errs, fmt.Errorf("APP_HTTP_READ_TIMEOUT: must be > 0"))
	}

	if http.WriteTimeout <= 0 {
		errs = append(errs, fmt.Errorf("APP_HTTP_WRITE_TIMEOUT: must be > 0"))
	}

	if http.IdleTimeout <= 0 {
		errs = append(errs, fmt.Errorf("APP_HTTP_IDLE_TIMEOUT: must be > 0"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// validateLog валидирует настройки логирования.
func validateLog(log *LogConfig) error {
	if log == nil {
		return fmt.Errorf("%w: log is nil", appErr.ErrValidationFailed)
	}

	var errs []error
	errs = appendErr(errs, "APP_LOG_LEVEL", utils.ValidateLogLevel(log.LogLevel))

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func validateEnv(env *EnvConfig) error {
	if env == nil {
		return fmt.Errorf("%w: env is nil", appErr.ErrValidationFailed)
	}

	var errs []error
	errs = appendErr(errs, "APP_VERSION", utils.ValidateVersion(env.Version))
	errs = appendErr(errs, "APP_INSTANCE", utils.ValidateInstance(env.Instance))
	errs = appendErr(errs, "ENV_TYPE", utils.ValidateEnv(env.Type))

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// validateDB валидирует настройки подключения и пула PostgreSQL.
func validateDB(db *DBConfig) error {
	if db == nil {
		return fmt.Errorf("%w: db is nil", appErr.ErrValidationFailed)
	}

	var errs []error
	errs = appendErr(errs, "PG_HOST", utils.ValidateEmptinessParam("PG_HOST", db.Host))
	errs = appendErr(errs, "PG_PORT", utils.ValidatePort(db.Port))
	errs = appendErr(errs, "PG_USER", utils.ValidateEmptinessParam("PG_USER", db.User))
	errs = appendErr(errs, "PG_PASSWORD", utils.ValidateEmptinessParam("PG_PASSWORD", db.Password))
	errs = appendErr(errs, "PG_DB", utils.ValidateEmptinessParam("PG_DB", db.Database))
	errs = appendErr(errs, "PG_MAX_CONNS", utils.ValidateParamMore("PG_MAX_CONNS", db.MaxConns, 0))
	errs = appendErr(errs, "PG_SSL_MODE", validateSSLMode(db.SSLMode))

	if db.MinConns < 0 {
		errs = append(errs, fmt.Errorf("PG_MIN_CONNS: must be >= 0"))
	}

	if db.MinConns > db.MaxConns {
		errs = append(errs, fmt.Errorf("PG_MIN_CONNS: %d > PG_MAX_CONNS: %d", db.MinConns, db.MaxConns))
	}

	if db.ConnLifeTime <= 0 {
		errs = append(errs, fmt.Errorf("PG_CONN_LIFETIME: must be > 0"))
	}

	if db.ConnIdleTime <= 0 {
		errs = append(errs, fmt.Errorf("PG_CONN_IDLE_TIME: must be > 0"))
	}

	if db.QueryTimeout <= 0 {
		errs = append(errs, fmt.Errorf("PG_QUERY_TIMEOUT: must be > 0"))
	}

	if db.HealthCheckPeriod <= 0 {
		errs = append(errs, fmt.Errorf("PG_HEALTH_CHECK_PERIOD: must be > 0"))
	}

	if db.PingTimeout <= 0 {
		errs = append(errs, fmt.Errorf("PG_PING_TIMEOUT: must be > 0"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// validateMinio валидирует настройки объектного хранилища.
func validateMinio(minio *MinioConfig) error {
	if minio == nil {
		return fmt.Errorf("%w: minio is nil", appErr.ErrValidationFailed)
	}

	var errs []error
	errs = appendErr(errs, "MINIO_HOST", utils.ValidateEmptinessParam("MINIO_HOST", minio.Host))
	errs = appendErr(errs, "MINIO_USER", utils.ValidateEmptinessParam("MINIO_USER", minio.User))
	errs = appendErr(errs, "MINIO_PASSWORD", utils.ValidateEmptinessParam("MINIO_PASSWORD", minio.Password))
	errs = appendErr(errs, "MINIO_BUCKET", utils.ValidateEmptinessParam("MINIO_BUCKET", minio.Bucket))
	errs = appendErr(errs, "MINIO_PORT_API", utils.ValidatePort(minio.ApiPort))
	errs = appendErr(errs, "MINIO_REGION_NAME", utils.ValidateEmptinessParam("MINIO_REGION_NAME", minio.RegionName))

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// validateRedis валидирует настройки подключения к Redis.
func validateRedis(redis *RedisConfig) error {
	if redis == nil {
		return fmt.Errorf("%w: redis is nil", appErr.ErrValidationFailed)
	}

	var errs []error
	errs = appendErr(errs, "REDIS_HOST", utils.ValidateEmptinessParam("REDIS_HOST", redis.Host))
	errs = appendErr(errs, "REDIS_PORT", utils.ValidatePort(redis.Port))
	errs = appendErr(errs, "REDIS_PASSWORD", utils.ValidateEmptinessParam("REDIS_PASSWORD", redis.Password))

	if redis.DB < 0 {
		errs = append(errs, fmt.Errorf("REDIS_DB: must be >= 0"))
	}

	if redis.MaxRetries < 0 {
		errs = append(errs, fmt.Errorf("REDIS_MAX_RETRIES: must be >= 0"))
	}

	if redis.Timeout <= 0 {
		errs = append(errs, fmt.Errorf("REDIS_TIMEOUT: must be > 0"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// validateSSLMode проверяет допустимость режима SSL для PostgreSQL.
func validateSSLMode(mode string) error {
	valid := map[string]bool{
		"disable":     true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if !valid[mode] {
		return fmt.Errorf("invalid PG_SSL_MODE: %q (allowed: disable, require, verify-ca, verify-full)", mode)
	}
	return nil
}
