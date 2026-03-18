package logger

import (
	appErr "backend/internal/errors"
	"backend/internal/logger/interfaces"
	"backend/internal/logger/slog"
	"fmt"
)

// Init создаёт логгер по заданному типу.
func Init(appVersion, instance, envType, logLevel, appService string, isLogger bool, logType interfaces.LoggerType) (interfaces.Logger, error) {
	switch logType {
	case interfaces.Slog:
		return slog.Init(appVersion, instance, envType, logLevel, appService, isLogger), nil
	}

	return nil, fmt.Errorf("%w: %d", appErr.ErrInitLogger, logType)
}
