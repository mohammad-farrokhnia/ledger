package config

import (
	"fmt"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Audit    AuditConfig
	Log      LogConfig
}

type ServerConfig struct {
	GRPCPort    string `koanf:"grpc_port"`
	HTTPPort    string `koanf:"http_port"`
	MetricsPort string `koanf:"metrics_port"`
}

type DatabaseConfig struct {
	DSN string `koanf:"dsn"`
}

type AuditConfig struct {
	Mode    string `koanf:"mode"`
	HookURL string `koanf:"hook_url"`
}

type LogConfig struct {
	Level string `koanf:"level"`
}

func Load(configPath string) (*Config, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("config: load file %s: %w", configPath, err)
	}

	envMap := map[string]string{
		"DB_DSN":        "database.dsn",
		"GRPC_PORT":     "server.grpc_port",
		"HTTP_PORT":     "server.http_port",
		"METRICS_PORT":  "server.metrics_port",
		"AUDIT_MODE":    "audit.mode",
		"AUDIT_HOOK_URL": "audit.hook_url",
		"LOG_LEVEL":     "log.level",
	}

	if err := k.Load(env.ProviderWithValue("", ".", func(s, v string) (string, interface{}) {
		key, ok := envMap[s]
		if !ok {
			return "", nil
		}
		return key, v
	}), nil); err != nil {
		return nil, fmt.Errorf("config: load env: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Database.DSN == "" {
		return fmt.Errorf("config: database.dsn is required (set DB_DSN env var)")
	}
	if cfg.Server.GRPCPort == "" {
		cfg.Server.GRPCPort = "9090"
	}
	if cfg.Server.HTTPPort == "" {
		cfg.Server.HTTPPort = "8080"
	}
	if cfg.Server.MetricsPort == "" {
		cfg.Server.MetricsPort = "9091"
	}
	if cfg.Audit.Mode == "" {
		cfg.Audit.Mode = "async"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	return nil
}
