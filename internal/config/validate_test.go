package config

import "testing"

// TestValidate verifies semantic validation rules for different config shapes.
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "valid",
			cfg:     validConfig(),
			wantErr: false,
		},
		{
			name: "invalid app port",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.App.Port = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "invalid log level",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.Log.LogLevel = "verbose"
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "db min conns greater than max",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.DB.MinConns = 100
				cfg.DB.MaxConns = 20
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "empty minio host",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.Minio.Host = ""
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "nil app config",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.App = nil
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "nil db config",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.DB = nil
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "nil minio config",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.Minio = nil
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "db timeout is zero",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.DB.QueryTimeout = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "db health check period is zero",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.DB.HealthCheckPeriod = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "db ping timeout is zero",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.DB.PingTimeout = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "db min conns is negative",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.DB.MinConns = -1
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "empty app secret",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.App.Secret = "   "
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "invalid version format",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.Env.Version = "v1.0"
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "invalid instance",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.Env.Instance = "staging"
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "invalid env type",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.Env.Type = "testing"
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "empty app service",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.App.Service = ""
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "empty app host",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.App.Host = "   "
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "nil env config",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.Env = nil
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "nil log config",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.Log = nil
				return cfg
			}(),
			wantErr: true,
		},
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
		},
		{
			name: "nil redis config",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.Redis = nil
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "empty redis host",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.Redis.Host = ""
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "invalid redis port",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.Redis.Port = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "redis db negative",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.Redis.DB = -1
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "invalid pg ssl mode",
			cfg: func() *Config {
				cfg := validConfig()
				cfg.DB.SSLMode = "enable"
				return cfg
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// validConfig returns a known-good config used as a base fixture in tests.
func validConfig() *Config {
	return &Config{
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
		DB: &DBConfig{
			Host:              "localhost",
			Port:              5433,
			User:              "adapt_ed",
			Password:          "123",
			Database:          "adapt_ed",
			SSLMode:           "disable",
			MaxConns:          20,
			MinConns:          5,
			ConnLifeTime:      60,
			ConnIdleTime:      60,
			QueryTimeout:      60,
			HealthCheckPeriod: 30,
			PingTimeout:       5,
		},
		Minio: &MinioConfig{
			Host:       "localhost",
			User:       "adapt_ed",
			Password:   "123",
			Bucket:     "adapt_ed",
			ApiPort:    9000,
			RegionName: "us-east-1",
		},
		Redis: &RedisConfig{
			Host:       "localhost",
			Port:       6379,
			Password:   "123",
			MaxRetries: 3,
			Timeout:    10,
		},
	}
}
