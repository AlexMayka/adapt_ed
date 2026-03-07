package config

import "testing"

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
				cfg.App.LogLevel = "verbose"
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
			name:    "nil config",
			cfg:     nil,
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

func validConfig() *Config {
	return &Config{
		App: &AppConfig{
			Host:     "localhost",
			Port:     8000,
			Secret:   "secret",
			Logging:  true,
			LogLevel: "info",
		},
		DB: &DBConfig{
			Host:         "localhost",
			Port:         5433,
			User:         "postgres_root",
			Password:     "123",
			Database:     "SALES_RADAR",
			MaxConns:     20,
			MinConns:     5,
			ConnLifeTime: 60,
			ConnIdleTime: 60,
			QueryTimeout: 60,
		},
		Minio: &MinioConfig{
			Host:     "localhost",
			User:     "minio_root",
			Password: "123",
			Bucket:   "Sales_Radar",
			ApiPort:  9000,
		},
	}
}
