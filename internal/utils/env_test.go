package utils

import (
	"errors"
	"os"
	"testing"
	"time"

	appErr "backend/internal/errors"
)

// TestGetEnvTable verifies typed parsing for representative scalar env values.
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
		{
			name: "int8", key: "t8", val: "12",
			check: func(t *testing.T) {
				res, err := GetEnv[int8]("t8")
				if err != nil || res != int8(12) {
					t.Errorf("got %v, err %v", res, err)
				}
			},
		},
		{
			name: "int16", key: "t9", val: "1234",
			check: func(t *testing.T) {
				res, err := GetEnv[int16]("t9")
				if err != nil || res != int16(1234) {
					t.Errorf("got %v, err %v", res, err)
				}
			},
		},
		{
			name: "int32", key: "t10", val: "123456",
			check: func(t *testing.T) {
				res, err := GetEnv[int32]("t10")
				if err != nil || res != int32(123456) {
					t.Errorf("got %v, err %v", res, err)
				}
			},
		},
		{
			name: "int64", key: "t11", val: "123456789",
			check: func(t *testing.T) {
				res, err := GetEnv[int64]("t11")
				if err != nil || res != int64(123456789) {
					t.Errorf("got %v, err %v", res, err)
				}
			},
		},
		{
			name: "float32", key: "t12", val: "12.5",
			check: func(t *testing.T) {
				res, err := GetEnv[float32]("t12")
				if err != nil || res != float32(12.5) {
					t.Errorf("got %v, err %v", res, err)
				}
			},
		},
		{
			name: "bool", key: "t13", val: "false",
			check: func(t *testing.T) {
				res, err := GetEnv[bool]("t13")
				if err != nil || res != false {
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

// TestGetEnvDefault verifies default fallback and parse error behavior.
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

// TestGetEnvErrorCases verifies key-not-found and parse errors from GetEnv.
func TestGetEnvErrorCases(t *testing.T) {
	t.Run("missing key", func(t *testing.T) {
		_ = os.Unsetenv("missing_key")
		_, err := GetEnv[int]("missing_key")
		if !errors.Is(err, appErr.ErrEnvKeyNotFound) {
			t.Fatalf("expected ErrKeyNotFound, got %v", err)
		}
	})

	t.Run("invalid int parse", func(t *testing.T) {
		t.Setenv("bad_int", "qwe")
		_, err := GetEnv[int]("bad_int")
		if !errors.Is(err, appErr.ErrEnvParseError) {
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

	t.Run("int8 parse error", func(t *testing.T) {
		t.Setenv("bad_int8", "200")
		_, err := GetEnv[int8]("bad_int8")
		if !errors.Is(err, appErr.ErrEnvParseError) {
			t.Fatalf("expected ErrParseError, got %v", err)
		}
	})

	t.Run("int16 parse error", func(t *testing.T) {
		t.Setenv("bad_int16", "999999")
		_, err := GetEnv[int16]("bad_int16")
		if !errors.Is(err, appErr.ErrEnvParseError) {
			t.Fatalf("expected ErrParseError, got %v", err)
		}
	})

	t.Run("int32 parse error", func(t *testing.T) {
		t.Setenv("bad_int32", "999999999999")
		_, err := GetEnv[int32]("bad_int32")
		if !errors.Is(err, appErr.ErrEnvParseError) {
			t.Fatalf("expected ErrParseError, got %v", err)
		}
	})

	t.Run("int64 parse error", func(t *testing.T) {
		t.Setenv("bad_int64", "not-an-int")
		_, err := GetEnv[int64]("bad_int64")
		if !errors.Is(err, appErr.ErrEnvParseError) {
			t.Fatalf("expected ErrParseError, got %v", err)
		}
	})

	t.Run("float32 parse error", func(t *testing.T) {
		t.Setenv("bad_float32", "abc")
		_, err := GetEnv[float32]("bad_float32")
		if !errors.Is(err, appErr.ErrEnvParseError) {
			t.Fatalf("expected ErrParseError, got %v", err)
		}
	})

	t.Run("float64 parse error", func(t *testing.T) {
		t.Setenv("bad_float64", "abc")
		_, err := GetEnv[float64]("bad_float64")
		if !errors.Is(err, appErr.ErrEnvParseError) {
			t.Fatalf("expected ErrParseError, got %v", err)
		}
	})

	t.Run("bool parse error", func(t *testing.T) {
		t.Setenv("bad_bool", "yes")
		_, err := GetEnv[bool]("bad_bool")
		if !errors.Is(err, appErr.ErrEnvParseError) {
			t.Fatalf("expected ErrParseError, got %v", err)
		}
	})
}

// TestDurationEnv verifies duration parsing and default fallback behavior.
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
		if !errors.Is(err, appErr.ErrEnvParseError) {
			t.Fatalf("expected ErrParseError, got %v", err)
		}
	})

	t.Run("invalid duration in default getter", func(t *testing.T) {
		t.Setenv("dur_bad_default", "abc")
		_, err := GetDurationEnvDefault("dur_bad_default", 30*time.Second)
		if !errors.Is(err, appErr.ErrEnvParseError) {
			t.Fatalf("expected ErrParseError, got %v", err)
		}
	})

	t.Run("existing duration in default getter", func(t *testing.T) {
		t.Setenv("dur_existing_default", "5m")
		got, err := GetDurationEnvDefault("dur_existing_default", 30*time.Second)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 5*time.Minute {
			t.Fatalf("expected 5m, got %v", got)
		}
	})
}
