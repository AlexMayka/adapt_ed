package utils

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrValidationFailed indicates that one or more semantic checks failed.
	ErrValidationFailed = errors.New("validation failed")
	// ErrEmptinessParam indicates that a required string value is empty.
	ErrEmptinessParam = errors.New("emptiness param")
	// ErrCheckPort indicates that a port is outside allowed range.
	ErrCheckPort = errors.New("invalid port")
	// ErrCheckMore indicates that a value is not greater than the required bound.
	ErrCheckMore = errors.New("the parameter is less than the required value")
	// ErrCheckLevel indicates that log level is not in the allowed set.
	ErrCheckLevel = errors.New("invalid level")
)

// SupportedParamMore lists numeric types supported by ValidateParamMore.
type SupportedParamMore interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}

var logLevel = map[string]struct{}{
	"debug": {},
	"info":  {},
	"warn":  {},
	"error": {},
}

// ValidatePort checks that port is in range 1..65535.
func ValidatePort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("%w: %d (expected 1..65535)", ErrCheckPort, port)
	}

	return nil
}

// ValidateEmptinessParam checks that a required string is not blank.
func ValidateEmptinessParam(name, value string) error {
	if len(strings.TrimSpace(value)) == 0 {
		return fmt.Errorf("%w: %s", ErrEmptinessParam, name)
	}

	return nil
}

// ValidateParamMore checks that param is strictly greater than check.
func ValidateParamMore[T SupportedParamMore](name string, param T, check T) error {
	if param <= check {
		return fmt.Errorf("%w: %s=%v <= %v", ErrCheckMore, name, param, check)
	}

	return nil
}

// ValidateLogLevel checks that log level is one of debug/info/warn/error.
func ValidateLogLevel(level string) error {
	level = strings.ToLower(strings.TrimSpace(level))
	if _, ok := logLevel[level]; !ok {
		return fmt.Errorf("%w: %q. Must be debug | info | warn | error", ErrCheckLevel, level)
	}

	return nil
}
