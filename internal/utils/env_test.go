package utils

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestGetEnvTable(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		val    string
		is_err bool
		check  func(t *testing.T)
	}{
		{
			name: "string", key: "t1", val: "test_1",
			check: func(t *testing.T) {
				res, err := GetEnv[string]("t1")
				if err != nil || res != "test_1" {
					t.Errorf("got %v, err %v", res, err)
				}
			},
		},
		{
			name: "int", key: "t2", val: "123",
			check: func(t *testing.T) {
				res, err := GetEnv[int]("t2")
				if err != nil || res != 123 {
					t.Errorf("got %v, err %v", res, err)
				}
			},
		},
		{
			name: "float64", key: "t7", val: "123.45",
			check: func(t *testing.T) {
				res, err := GetEnv[float64]("t7")
				if err != nil || res != 123.45 {
					t.Errorf("got %v, err %v", res, err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.key, tt.val)
			tt.check(t)
		})
	}
}

func TestGetEnvDefault(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		val     string
		def     int
		want    int
		wantErr bool
	}{
		{
			name:    "missing key uses default",
			key:     "missing_env_int",
			def:     42,
			want:    42,
			wantErr: false,
		},
		{
			name:    "existing key overrides default",
			key:     "existing_env_int",
			val:     "12",
			def:     42,
			want:    12,
			wantErr: false,
		},
		{
			name:    "invalid value returns error",
			key:     "invalid_env_int",
			val:     "abc",
			def:     42,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.val == "" {
				_ = os.Unsetenv(tt.key)
			} else {
				t.Setenv(tt.key, tt.val)
			}

			got, err := GetEnvDefault[int](tt.key, tt.def)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetEnvDefault() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("GetEnvDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvErrorCases(t *testing.T) {
	t.Run("missing key", func(t *testing.T) {
		_ = os.Unsetenv("missing_key")
		_, err := GetEnv[int]("missing_key")
		if !errors.Is(err, ErrKeyNotFound) {
			t.Fatalf("expected ErrKeyNotFound, got %v", err)
		}
	})

	t.Run("invalid int parse", func(t *testing.T) {
		t.Setenv("bad_int", "qwe")
		_, err := GetEnv[int]("bad_int")
		if !errors.Is(err, ErrParseError) {
			t.Fatalf("expected ErrParseError, got %v", err)
		}
	})

	t.Run("bool parse", func(t *testing.T) {
		t.Setenv("bool_val", "true")
		got, err := GetEnv[bool]("bool_val")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Fatalf("expected true, got false")
		}
	})
}

func TestDurationEnv(t *testing.T) {
	t.Run("get duration", func(t *testing.T) {
		t.Setenv("dur_ok", "15s")
		got, err := GetDurationEnv("dur_ok")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 15*time.Second {
			t.Fatalf("expected 15s, got %v", got)
		}
	})

	t.Run("missing key default", func(t *testing.T) {
		_ = os.Unsetenv("dur_missing")
		got, err := GetDurationEnvDefault("dur_missing", 30*time.Second)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 30*time.Second {
			t.Fatalf("expected 30s, got %v", got)
		}
	})

	t.Run("invalid duration", func(t *testing.T) {
		t.Setenv("dur_bad", "abc")
		_, err := GetDurationEnv("dur_bad")
		if !errors.Is(err, ErrParseError) {
			t.Fatalf("expected ErrParseError, got %v", err)
		}
	})
}
