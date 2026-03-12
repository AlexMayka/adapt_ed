package logger

import (
	"backend/internal/logger/slog"
	"backend/internal/logger/types"
)

func Init(appVersion, instance, envType, logLevel, appService string, isLogger bool, logType types.LoggerType) types.Logger {
	switch logType {
	case types.Slog:
		return slog.Init(appVersion, instance, envType, logLevel, appService, isLogger)
	}

	return nil
}
