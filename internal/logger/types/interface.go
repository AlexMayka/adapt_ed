package types

type LoggerType int

const (
	Slog LoggerType = iota
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)

	With(args ...any) Logger
	WithGroup(name string) Logger
}
