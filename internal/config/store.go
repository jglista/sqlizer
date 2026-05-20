package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var ErrNotFound = errors.New("connection not found")
var ErrAlreadyExists = errors.New("connection already exists")

type Connection struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

type configFile struct {
	Default     string                `yaml:"default,omitempty"`
	Connections map[string]Connection `yaml:"connections"`
}

func load(path string) (*configFile, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &configFile{Connections: make(map[string]Connection)}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg configFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Connections == nil {
		cfg.Connections = make(map[string]Connection)
	}
	return &cfg, nil
}

func save(path string, cfg *configFile) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// Add writes a new connection. Returns ErrAlreadyExists if the name is taken.
func Add(path, name string, conn Connection) error {
	cfg, err := load(path)
	if err != nil {
		return err
	}
	if _, exists := cfg.Connections[name]; exists {
		return fmt.Errorf("%w: %s", ErrAlreadyExists, name)
	}
	cfg.Connections[name] = conn
	return save(path, cfg)
}

// Overwrite upserts a connection regardless of whether it already exists.
func Overwrite(path, name string, conn Connection) error {
	cfg, err := load(path)
	if err != nil {
		return err
	}
	cfg.Connections[name] = conn
	return save(path, cfg)
}

// Remove deletes a connection. If the removed connection was the default, the
// default is cleared. Removing a name that does not exist is a no-op.
func Remove(path, name string) error {
	cfg, err := load(path)
	if err != nil {
		return err
	}
	delete(cfg.Connections, name)
	if cfg.Default == name {
		cfg.Default = ""
	}
	return save(path, cfg)
}

// List returns all connections. The map is empty (not nil) when none exist.
func List(path string) (map[string]Connection, error) {
	cfg, err := load(path)
	if err != nil {
		return nil, err
	}
	return cfg.Connections, nil
}

// Get returns the named connection. Returns ErrNotFound if absent.
func Get(path, name string) (Connection, error) {
	cfg, err := load(path)
	if err != nil {
		return Connection{}, err
	}
	conn, ok := cfg.Connections[name]
	if !ok {
		return Connection{}, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return conn, nil
}

// GetDefault returns the name of the default connection, or "" if unset.
func GetDefault(path string) (string, error) {
	cfg, err := load(path)
	if err != nil {
		return "", err
	}
	return cfg.Default, nil
}

// SetDefault sets the named connection as the default. Returns ErrNotFound if
// the name does not exist in the connections map.
func SetDefault(path, name string) error {
	cfg, err := load(path)
	if err != nil {
		return err
	}
	if _, ok := cfg.Connections[name]; !ok {
		return fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	cfg.Default = name
	return save(path, cfg)
}
