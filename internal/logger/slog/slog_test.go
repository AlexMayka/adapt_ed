package slog

import (
	"log/slog"
	"testing"
)

func TestGetSlogLevel_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := getSlogLevel(tt.input)
			if got != tt.want {
				t.Fatalf("getSlogLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetSlogLevel_Unknown_DefaultsToInfo(t *testing.T) {
	got := getSlogLevel("unknown_level")
	if got != slog.LevelInfo {
		t.Fatalf("getSlogLevel(unknown) = %v, want %v", got, slog.LevelInfo)
	}
}

func TestGetSlogLevel_Empty_DefaultsToInfo(t *testing.T) {
	got := getSlogLevel("")
	if got != slog.LevelInfo {
		t.Fatalf("getSlogLevel('') = %v, want %v", got, slog.LevelInfo)
	}
}

func TestInit_ReturnsLogger(t *testing.T) {
	log := Init("1.0.0", "local", "dev", "info", "test", true)
	if log == nil {
		t.Fatal("Init() returned nil")
	}
}

func TestInit_LoggingDisabled(t *testing.T) {
	log := Init("1.0.0", "local", "dev", "info", "test", false)
	// Не должен паниковать при вызове с isLogger=false
	log.Info("this should be silent")
	log.Error("this too")
}
