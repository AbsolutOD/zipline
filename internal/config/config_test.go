package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AbsolutOD/zipline/internal/config"
)

func TestDirUsesXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/xdg")
	if got, want := config.Dir(), filepath.Join("/custom/xdg", "zipline"); got != want {
		t.Errorf("Dir() = %q, want %q", got, want)
	}
}

func TestDirDefaultsToHomeConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	os.Unsetenv("XDG_CONFIG_HOME")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no home dir: %v", err)
	}
	if got, want := config.Dir(), filepath.Join(home, ".config", "zipline"); got != want {
		t.Errorf("Dir() = %q, want %q", got, want)
	}
}

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if want := filepath.Join(dir, "zipline", "zipline.db"); cfg.DBPath != want {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, want)
	}
	if cfg.Cmd != "zl" {
		t.Errorf("Cmd = %q, want %q", cfg.Cmd, "zl")
	}
}

func TestLoadReadsConfigFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	zdir := filepath.Join(dir, "zipline")
	if err := os.MkdirAll(zdir, 0o700); err != nil {
		t.Fatal(err)
	}
	toml := "db_path = \"/elsewhere/z.db\"\ncmd = \"zip\"\n"
	if err := os.WriteFile(filepath.Join(zdir, "config.toml"), []byte(toml), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DBPath != "/elsewhere/z.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "/elsewhere/z.db")
	}
	if cfg.Cmd != "zip" {
		t.Errorf("Cmd = %q, want %q", cfg.Cmd, "zip")
	}
}
