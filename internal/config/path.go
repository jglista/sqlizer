package config

import (
	"os"
	"path/filepath"
)

// ConfigPath returns <UserConfigDir>/sqlizer/config.yaml,
// creating the sqlizer/ directory with 0700 permissions on first call.
func ConfigPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "sqlizer")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}
