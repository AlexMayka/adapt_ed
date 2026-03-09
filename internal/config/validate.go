package config

import (
	"backend/internal/utils"
	"errors"
	"fmt"
)

// Validate checks semantic rules for already parsed configuration values.
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("%w: config is nil", utils.ErrValidationFailed)
	}

	var errs []error

	errs = appendErr(errs, "app", validateApp(c.App))
	errs = appendErr(errs, "db", validateDB(c.DB))
	errs = appendErr(errs, "minio", validateMinio(c.Minio))

	if len(errs) > 0 {
		return fmt.Errorf("%w: %w", utils.ErrValidationFailed, errors.Join(errs...))
	}

	return nil
}

// validateApp validates application-specific settings.
func validateApp(app *AppConfig) error {
	if app == nil {
		return errors.New("app config is nil")
	}

	var errs []error
	errs = appendErr(errs, "SR_AP_HOST", utils.ValidateEmptinessParam("SR_AP_HOST", app.Host))
	errs = appendErr(errs, "SR_AP_PORT", utils.ValidatePort(app.Port))
	errs = appendErr(errs, "SR_AP_SECRET", utils.ValidateEmptinessParam("SR_AP_SECRET", app.Secret))
	errs = appendErr(errs, "SR_AP_LOG_LEVEL", utils.ValidateLogLevel(app.LogLevel))

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// validateDB validates PostgreSQL connection and pool settings.
func validateDB(db *DBConfig) error {
	if db == nil {
		return errors.New("db config is nil")
	}

	var errs []error
	errs = appendErr(errs, "SR_PG_HOST", utils.ValidateEmptinessParam("SR_PG_HOST", db.Host))
	errs = appendErr(errs, "SR_PG_PORT", utils.ValidatePort(db.Port))
	errs = appendErr(errs, "SR_PG_USER", utils.ValidateEmptinessParam("SR_PG_USER", db.User))
	errs = appendErr(errs, "SR_PG_PASSWORD", utils.ValidateEmptinessParam("SR_PG_PASSWORD", db.Password))
	errs = appendErr(errs, "SR_PG_DB", utils.ValidateEmptinessParam("SR_PG_DB", db.Database))
	errs = appendErr(errs, "SR_PG_MAX_CONNS", utils.ValidateParamMore("SR_PG_MAX_CONNS", db.MaxConns, 0))
	if db.MinConns < 0 {
		errs = append(errs, fmt.Errorf("SR_PG_MIN_CONNS: must be >= 0"))
	}
	if db.MinConns > db.MaxConns {
		errs = append(errs, fmt.Errorf("SR_PG_MIN_CONNS: %d > SR_PG_MAX_CONNS: %d", db.MinConns, db.MaxConns))
	}
	if db.ConnLifeTime <= 0 {
		errs = append(errs, fmt.Errorf("SR_PG_CONN_LIFETIME: must be > 0"))
	}
	if db.ConnIdleTime <= 0 {
		errs = append(errs, fmt.Errorf("SR_PG_CONN_IDLE_TIME: must be > 0"))
	}
	if db.QueryTimeout <= 0 {
		errs = append(errs, fmt.Errorf("SR_PG_QUERY_TIMEOUT: must be > 0"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// validateMinio validates object storage connection settings.
func validateMinio(minio *MinioConfig) error {
	if minio == nil {
		return errors.New("minio config is nil")
	}

	var errs []error
	errs = appendErr(errs, "SR_MN_HOST", utils.ValidateEmptinessParam("SR_MN_HOST", minio.Host))
	errs = appendErr(errs, "SR_MN_USER", utils.ValidateEmptinessParam("SR_MN_USER", minio.User))
	errs = appendErr(errs, "SR_MN_PASSWORD", utils.ValidateEmptinessParam("SR_MN_PASSWORD", minio.Password))
	errs = appendErr(errs, "SR_MN_BUCKET", utils.ValidateEmptinessParam("SR_MN_BUCKET", minio.Bucket))
	errs = appendErr(errs, "SR_MN_PORT_API", utils.ValidatePort(minio.ApiPort))

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
