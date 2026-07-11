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

// TestLoadReturnsErrorOnMalformedConfig ensures a present-but-unparseable
// config.toml surfaces an error instead of silently falling back to
// defaults (which would redirect to a different database unnoticed).
func TestLoadReturnsErrorOnMalformedConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	zdir := filepath.Join(dir, "zipline")
	if err := os.MkdirAll(zdir, 0o700); err != nil {
		t.Fatal(err)
	}
	// Not valid TOML: unterminated string value.
	bad := "db_path = \"unterminated\n"
	if err := os.WriteFile(filepath.Join(zdir, "config.toml"), []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(); err == nil {
		t.Error("Load() with malformed config.toml = nil error, want error")
	}
}

// TestLoadReturnsErrorOnUnreadableConfig ensures a present-but-permission-
// denied config.toml surfaces an error instead of silently falling back
// to defaults. Skipped only when running as a user that bypasses file
// permissions (e.g. root), since chmod 000 wouldn't actually block reads.
func TestLoadReturnsErrorOnUnreadableConfig(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: file permissions are not enforced")
	}
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	zdir := filepath.Join(dir, "zipline")
	if err := os.MkdirAll(zdir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(zdir, "config.toml")
	if err := os.WriteFile(path, []byte("cmd = \"zip\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(path, 0o600)
	if _, err := config.Load(); err == nil {
		t.Error("Load() with unreadable config.toml = nil error, want error")
	}
}
