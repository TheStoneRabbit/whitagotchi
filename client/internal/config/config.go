package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Server   string `json:"server"`
	Username string `json:"username,omitempty"`
	Token    string `json:"token,omitempty"`
}

func path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "whitagotchi", "config.json"), nil
}

func Load() (*Config, error) {
	cfg := &Config{Server: "http://localhost:8080"}
	if s := os.Getenv("WHITAGOTCHI_SERVER"); s != "" {
		cfg.Server = s
	}
	p, err := path()
	if err != nil {
		return cfg, nil
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return cfg, nil
	}
	_ = json.Unmarshal(b, cfg)
	if s := os.Getenv("WHITAGOTCHI_SERVER"); s != "" {
		cfg.Server = s
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o600)
}
