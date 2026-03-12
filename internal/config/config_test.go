package config

import (
	"github.com/google/go-cmp/cmp"
	"github.com/joho/godotenv"
	"os"
	"testing"
	"time"
)

const (
	paramsScriptPath1 = "./test/.env.script_1"
	paramsScriptPath2 = "./test/.env.script_2"
	paramsScriptPath3 = "./test/.env.script_3"
	paramsScriptPath4 = "./test/.env.script_4"
	paramsScriptPath5 = "./test/.env.script_5"
	paramsScriptPath6 = "./test/.env.script_6"
)

var wantConfig = Config{
	Env: &EnvConfig{
		Version:  "1.0.0",
		Instance: "local",
		Type:     "dev",
	},
	App: &AppConfig{
		Service: "backend_adapt_ed",
		Host:    "localhost",
		Port:    8000,
		Secret:  "secret",
	},
	Log: &LogConfig{
		IsLogging: true,
		LogLevel:  "info",
	},
	Minio: &MinioConfig{
		User:     "adapt_ed",
		Password: "123",
		Host:     "localhost",
		ApiPort:  9000,
		Bucket:   "adapt_ed",
	},
	DB: &DBConfig{
		User:     "adapt_ed",
		Password: "123",
		Host:     "localhost",
		Port:     5433,
		Database: "adapt_ed",

		MaxConns:     20,
		MinConns:     5,
		ConnLifeTime: time.Duration(time.Second * 60),
		ConnIdleTime: time.Duration(time.Second * 60),
		QueryTimeout: time.Duration(time.Second * 60),
	},
}

// TestLoad checks env parsing and validation outcomes for positive and negative scenarios.
func TestLoad(t *testing.T) {
	tests := []struct {
		name   string
		params string
		want   Config
		is_err bool
	}{
		{name: "positive", params: paramsScriptPath1, want: wantConfig, is_err: false},
		{name: "negative1", params: paramsScriptPath2, want: Config{}, is_err: true},
		{name: "negative2", params: paramsScriptPath3, want: Config{}, is_err: true},
		{name: "negative3", params: paramsScriptPath4, want: Config{}, is_err: true},
		{name: "negative4", params: paramsScriptPath5, want: Config{}, is_err: true},
		{name: "negative5", params: paramsScriptPath6, want: Config{}, is_err: true},
	}

	for index, tt := range tests {
		err := godotenv.Load(tt.params)
		if err != nil {
			t.Errorf("Error loading .env file at index %d: %v", index, err)
		}

		t.Run(tt.name, func(t *testing.T) {
			cnf, err := Load()
			if err != nil && tt.is_err {
				return
			}

			if (err != nil) != tt.is_err {
				t.Errorf("%v: Load() error = %v, wantErr %v", err, index, tt.is_err)
			}

			if diff := cmp.Diff(tt.want, cnf); diff != "" {
				t.Fatalf("report mismatch (-want +got):\n%s", diff)
			}
		})
		os.Clearenv()
	}
}
