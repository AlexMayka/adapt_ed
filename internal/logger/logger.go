package logger

import (
	"backend/internal/logger/interfaces"
	"backend/internal/logger/slog"
)

func Init(appVersion, instance, envType, logLevel, appService string, isLogger bool, logType interfaces.LoggerType) (interfaces.Logger, error) {
	switch logType {
	case interfaces.Slog:
		return slog.Init(appVersion, instance, envType, logLevel, appService, isLogger), nil
	}

	return nil, nil
}
