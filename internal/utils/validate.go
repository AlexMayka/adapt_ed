package utils

import (
	appErr "backend/internal/errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// SupportedParamMore перечисляет числовые типы, поддерживаемые ValidateParamMore.
type SupportedParamMore interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}

var logLevel = map[string]struct{}{
	"debug": {},
	"info":  {},
	"warn":  {},
	"error": {},
}

var instances = map[string]struct{}{
	"local": {},
	"prod":  {},
}

var envType = map[string]struct{}{
	"dev":  {},
	"prod": {},
}

// ValidatePort проверяет, что порт в диапазоне 1..65535.
func ValidatePort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("%w: %d (expected 1..65535)", appErr.ErrCheckPort, port)
	}

	return nil
}

// ValidateEmptinessParam проверяет, что обязательная строка не пуста.
func ValidateEmptinessParam(name, value string) error {
	if len(strings.TrimSpace(value)) == 0 {
		return fmt.Errorf("%w: %s", appErr.ErrEmptinessParam, name)
	}

	return nil
}

// ValidateParamMore проверяет, что параметр строго больше заданного значения.
func ValidateParamMore[T SupportedParamMore](name string, param T, check T) error {
	if param <= check {
		return fmt.Errorf("%w: %s=%v <= %v", appErr.ErrCheckMore, name, param, check)
	}

	return nil
}

func ValidateVersion(value string) error {
	split := strings.Split(value, ".")

	if len(split) != 3 {
		return fmt.Errorf("%w: %s (example 20.10.10)", appErr.ErrLackVersion, value)
	}

	for index, v := range split {
		if _, err := strconv.Atoi(v); err != nil {
			return fmt.Errorf("%v - %w: %v (example 20.10.10)", index+1, appErr.ErrMustBeNumber, v)
		}
	}

	return nil
}

// ValidateInstance проверяет, что instance — одно из значений: local, prod.
func ValidateInstance(value string) error {
	level := strings.ToLower(strings.TrimSpace(value))
	if len(level) == 0 {
		return fmt.Errorf("%w: %s", appErr.ErrEmptinessInstance, value)
	}

	if _, ok := instances[level]; !ok {
		return fmt.Errorf("%w: %s", appErr.ErrNoSupportInstance, value)
	}

	return nil
}

// ValidateEnv проверяет, что тип окружения — одно из значений: dev, prod.
func ValidateEnv(value string) error {
	level := strings.ToLower(strings.TrimSpace(value))
	if len(level) == 0 {
		return fmt.Errorf("%w: %s", appErr.ErrEmptinessEnv, value)
	}

	if _, ok := envType[level]; !ok {
		return fmt.Errorf("%w: %s", appErr.ErrNoSupportEnv, value)
	}

	return nil
}

// ValidateLogLevel проверяет, что уровень логирования — одно из: debug, info, warn, error.
func ValidateLogLevel(level string) error {
	level = strings.ToLower(strings.TrimSpace(level))
	if _, ok := logLevel[level]; !ok {
		return fmt.Errorf("%w: %q. Must be debug | info | warn | error", appErr.ErrCheckLevel, level)
	}

	return nil
}

var special = []rune{'!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '_', '-', '+', '=', '?', '.', ',', ':', ';'}

func ValidatePassword(password string) []error {
	errors := make([]error, 0)

	countNumber := 0
	countUpper := 0
	countLower := 0
	countSpecial := 0
	countLetter := 0

	for _, char := range password {
		countLetter++

		if char >= 'A' && char <= 'Z' {
			countUpper++
			continue
		}

		if char >= 'a' && char <= 'z' {
			countLower++
			continue
		}

		if char >= '0' && char <= '9' {
			countNumber++
			continue
		}

		if slices.Contains(special, char) {
			countSpecial++
			continue
		}

		errors = append(errors, fmt.Errorf("%w: %s", appErr.ErrPassInvalidChar, string(char)))
	}

	if countLetter < 8 || countLetter > 64 {
		errors = append(errors, fmt.Errorf("%w: %v", appErr.ErrPassLen, countLetter))
	}

	if countUpper < 1 {
		errors = append(errors, fmt.Errorf("%w: %v", appErr.ErrPassCountUppers, countUpper))
	}

	if countLower < 1 {
		errors = append(errors, fmt.Errorf("%w: %v", appErr.ErrPassCountLowers, countLower))
	}

	if countNumber < 1 {
		errors = append(errors, fmt.Errorf("%w: %v", appErr.ErrPassCountNumbers, countNumber))
	}

	return errors
}
