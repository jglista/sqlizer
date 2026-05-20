package config_test

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/jglista/sqlizer/internal/config"
)

func TestConfigPath_WithinUserConfigDir(t *testing.T) {
	base, err := os.UserConfigDir()
	if err != nil {
		t.Skip("os.UserConfigDir() unavailable:", err)
	}
	path, err := config.ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() error: %v", err)
	}
	if !strings.HasPrefix(path, base) {
		t.Errorf("ConfigPath() = %q, want prefix %q", path, base)
	}
}

func TestConfigPath_CreatesDirectory(t *testing.T) {
	base, err := os.UserConfigDir()
	if err != nil {
		t.Skip("os.UserConfigDir() unavailable:", err)
	}
	_, err = config.ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() error: %v", err)
	}
	info, err := os.Stat(base + "/sqlizer")
	if err != nil {
		t.Fatalf("sqlizer dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("sqlizer path is not a directory")
	}
}

func TestConfigPath_DirectoryPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits not applicable on Windows")
	}
	base, err := os.UserConfigDir()
	if err != nil {
		t.Skip("os.UserConfigDir() unavailable:", err)
	}
	_, err = config.ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() error: %v", err)
	}
	info, err := os.Stat(base + "/sqlizer")
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	got := info.Mode().Perm()
	if got != 0700 {
		t.Errorf("directory permissions = %04o, want 0700", got)
	}
}
