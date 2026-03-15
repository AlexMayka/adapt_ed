package slog

import (
	"backend/internal/logger/interfaces"
	"backend/internal/utils"
	"log/slog"
	"os"
	"time"
)

type SlogLogger struct {
	isLogger bool
	logger   *slog.Logger
}

var loggerLevels = map[string]slog.Level{
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
	"debug": slog.LevelDebug,
}

func getSlogLevel(level string) slog.Level {
	lv, ok := loggerLevels[level]
	if !ok {
		panic("No search level")
	}

	return lv
}

func Init(appVersion, instance, envType, logLevel, appService string, isLogger bool) interfaces.Logger {
	slogLevel := getSlogLevel(logLevel)
	opts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: false,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				t := a.Value.Time().UTC().Format(time.RFC3339)
				return slog.String("time", t)
			case slog.LevelKey:
				return slog.String("level", a.Value.String())
			case slog.MessageKey:
				return slog.String("msg", a.Value.String())
			default:
				return a
			}
		},
	}

	h := slog.NewJSONHandler(os.Stdout, opts)

	base := slog.New(h).With(
		"service", appService,
		"uuid_work", utils.GetUniqUUID(),
		"env", envType,
		"version", appVersion,
		"instance", instance,
	)

	return &SlogLogger{logger: base, isLogger: isLogger}
}

func (sl *SlogLogger) Debug(msg string, args ...any) {
	if sl.isLogger {
		sl.logger.Debug(msg, args...)
	}
}

func (sl *SlogLogger) Info(msg string, args ...any) {
	if sl.isLogger {
		sl.logger.Info(msg, args...)
	}
}

func (sl *SlogLogger) Warn(msg string, args ...any) {
	if sl.isLogger {
		sl.logger.Warn(msg, args...)
	}
}

func (sl *SlogLogger) Error(msg string, args ...any) {
	if sl.isLogger {
		sl.logger.Error(msg, args...)
	}
}

func (sl *SlogLogger) With(args ...any) interfaces.Logger {
	return &SlogLogger{
		logger:   sl.logger.With(args...),
		isLogger: sl.isLogger,
	}
}

func (sl *SlogLogger) WithGroup(name string) interfaces.Logger {
	return &SlogLogger{
		logger:   sl.logger.WithGroup(name),
		isLogger: sl.isLogger,
	}
}
