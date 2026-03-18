package logger

import (
	"backend/internal/logger/interfaces"
	"testing"
)

func TestInit_Slog(t *testing.T) {
	log, err := Init("1.0.0", "local", "dev", "info", "test_service", false, interfaces.Slog)
	if err != nil {
		t.Fatalf("Init(Slog) unexpected error: %v", err)
	}
	if log == nil {
		t.Fatal("Init(Slog) returned nil logger")
	}
}

func TestInit_UnknownType(t *testing.T) {
	_, err := Init("1.0.0", "local", "dev", "info", "test_service", false, interfaces.LoggerType(99))
	if err == nil {
		t.Fatal("Init(unknown type) expected error, got nil")
	}
}
