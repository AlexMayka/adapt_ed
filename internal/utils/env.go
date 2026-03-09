package utils

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// SupportedEnvType lists scalar types that can be parsed from environment values.
type SupportedEnvType interface {
	~string | ~int | ~bool | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}

// Env parsing errors.
var (
	ErrParseError     = fmt.Errorf("error parsing environment variable")
	ErrKeyNotFound    = fmt.Errorf("key not found in environment variable")
	ErrNotSupportType = fmt.Errorf("environment variable not supported")
)

// GetEnv returns an environment variable parsed to the requested type.
func GetEnv[T SupportedEnvType](key string) (T, error) {
	var zero T

	raw, ok := os.LookupEnv(key)
	if !ok {
		return zero, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	switch any(zero).(type) {
	case string:
		return any(raw).(T), nil

	case int:
		v, err := strconv.Atoi(raw)
		if err != nil {
			return zero, fmt.Errorf("%w: type: int, key: %v, err: %v", ErrParseError, key, err)
		}
		return any(v).(T), nil

	case bool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return zero, fmt.Errorf("%w: type: bool, key: %v, err: %v", ErrParseError, key, err)
		}
		return any(v).(T), nil

	case int64:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return zero, fmt.Errorf("%w: type: int64, key: %v, err: %v", ErrParseError, key, err)
		}
		return any(v).(T), nil

	case int8:
		v, err := strconv.ParseInt(raw, 10, 8)
		if err != nil {
			return zero, fmt.Errorf("%w: type: int8, key: %v, err: %v", ErrParseError, key, err)
		}
		return any(int8(v)).(T), nil

	case int16:
		v, err := strconv.ParseInt(raw, 10, 16)
		if err != nil {
			return zero, fmt.Errorf("%w: type: int16, key: %v, err: %v", ErrParseError, key, err)
		}
		return any(int16(v)).(T), nil

	case int32:
		v, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return zero, fmt.Errorf("%w: type: int32, key: %v, err: %v", ErrParseError, key, err)
		}
		return any(int32(v)).(T), nil

	case float32:
		v, err := strconv.ParseFloat(raw, 32)
		if err != nil {
			return zero, fmt.Errorf("%w: type: float32, key: %v, err: %v", ErrParseError, key, err)
		}
		return any(float32(v)).(T), nil

	case float64:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return zero, fmt.Errorf("%w: type: float64, key: %v, err: %v", ErrParseError, key, err)
		}
		return any(v).(T), nil

	default:
		return zero, fmt.Errorf("%w: %q", ErrNotSupportType, key)
	}
}

// GetDurationEnv returns an environment variable parsed as time.Duration.
func GetDurationEnv(key string) (time.Duration, error) {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	v, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%w: type: duration, key: %v, err: %v", ErrParseError, key, err)
	}

	return v, nil
}

// GetEnvDefault returns defaultValue when the key is missing, otherwise parsed env value.
func GetEnvDefault[T SupportedEnvType](key string, defaultValue T) (T, error) {
	value, err := GetEnv[T](key)
	if errors.Is(err, ErrKeyNotFound) {
		return defaultValue, nil
	}

	return value, err
}

// GetDurationEnvDefault returns defaultValue when the key is missing.
func GetDurationEnvDefault(key string, defaultValue time.Duration) (time.Duration, error) {
	value, err := GetDurationEnv(key)
	if errors.Is(err, ErrKeyNotFound) {
		return defaultValue, nil
	}

	return value, err
}
