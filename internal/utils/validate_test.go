package utils

import (
	"errors"
	"testing"

	appErr "backend/internal/errors"
)

// TestValidatePort checks accepted and rejected port bounds.
func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{name: "valid lower bound", port: 1, wantErr: false},
		{name: "valid upper bound", port: 65535, wantErr: false},
		{name: "zero", port: 0, wantErr: true},
		{name: "negative", port: -1, wantErr: true},
		{name: "too big", port: 70000, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidatePort() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, appErr.ErrCheckPort) {
				t.Fatalf("expected ErrCheckPort, got %v", err)
			}
		})
	}
}

// TestValidateEmptinessParam checks blank/non-blank string validation.
func TestValidateEmptinessParam(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "non-empty", value: "value", wantErr: false},
		{name: "spaces only", value: "   ", wantErr: true},
		{name: "empty", value: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmptinessParam("TEST_KEY", tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateEmptinessParam() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, appErr.ErrEmptinessParam) {
				t.Fatalf("expected ErrEmptinessParam, got %v", err)
			}
		})
	}
}

// TestValidateParamMore checks strictly-greater-than numeric validation.
func TestValidateParamMore(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if err := ValidateParamMore("MAX", 10, 0); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("equal", func(t *testing.T) {
		err := ValidateParamMore("MAX", 10, 10)
		if !errors.Is(err, appErr.ErrCheckMore) {
			t.Fatalf("expected ErrCheckMore, got %v", err)
		}
	})

	t.Run("less", func(t *testing.T) {
		err := ValidateParamMore("MAX", -1, 0)
		if !errors.Is(err, appErr.ErrCheckMore) {
			t.Fatalf("expected ErrCheckMore, got %v", err)
		}
	})
}

// TestValidateLogLevel checks allowed and rejected log level values.
func TestValidateLogLevel(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "error", " INFO "}
	for _, level := range validLevels {
		t.Run("valid_"+level, func(t *testing.T) {
			if err := ValidateLogLevel(level); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

	err := ValidateLogLevel("verbose")
	if !errors.Is(err, appErr.ErrCheckLevel) {
		t.Fatalf("expected ErrCheckLevel, got %v", err)
	}
}
