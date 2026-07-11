// Package config resolves zipline's file locations and optional
// user configuration from $XDG_CONFIG_HOME/zipline/config.toml.
package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds user-tunable settings with their defaults applied.
type Config struct {
	DBPath string // sqlite database location
	Cmd    string // short wrapper function name for shell hooks
}

// Dir returns the zipline config directory: $XDG_CONFIG_HOME/zipline,
// falling back to ~/.config/zipline.
func Dir() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "zipline")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", "zipline")
	}
	return filepath.Join(home, ".config", "zipline")
}

// Load returns defaults overridden by config.toml when it exists.
// A missing config file is not an error.
func Load() (*Config, error) {
	cfg := &Config{
		DBPath: filepath.Join(Dir(), "zipline.db"),
		Cmd:    "zl",
	}
	v := viper.New()
	v.SetConfigFile(filepath.Join(Dir(), "config.toml"))
	if err := v.ReadInConfig(); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return cfg, nil
		}
		return nil, err
	}
	if s := v.GetString("db_path"); s != "" {
		cfg.DBPath = s
	}
	if s := v.GetString("cmd"); s != "" {
		cfg.Cmd = s
	}
	return cfg, nil
}
