package config_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/jglista/sqlizer/internal/config"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "config.yaml")
}

var devConn = config.Connection{Host: "localhost", Port: "1433", User: "sa", Pass: "devpass"}
var prodConn = config.Connection{Host: "prod.example.com", Port: "1433", User: "sa", Pass: "secret"}

func TestAdd_HappyPath(t *testing.T) {
	p := tempPath(t)
	if err := config.Add(p, "dev", devConn); err != nil {
		t.Fatalf("Add: %v", err)
	}
	got, err := config.Get(p, "dev")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != devConn {
		t.Errorf("got %+v, want %+v", got, devConn)
	}
}

func TestAdd_AlreadyExists(t *testing.T) {
	p := tempPath(t)
	config.Add(p, "dev", devConn)
	err := config.Add(p, "dev", prodConn)
	if !errors.Is(err, config.ErrAlreadyExists) {
		t.Errorf("want ErrAlreadyExists, got %v", err)
	}
}

func TestOverwrite_Creates(t *testing.T) {
	p := tempPath(t)
	if err := config.Overwrite(p, "dev", devConn); err != nil {
		t.Fatalf("Overwrite: %v", err)
	}
	got, _ := config.Get(p, "dev")
	if got != devConn {
		t.Errorf("got %+v, want %+v", got, devConn)
	}
}

func TestOverwrite_Updates(t *testing.T) {
	p := tempPath(t)
	config.Add(p, "dev", devConn)
	config.Overwrite(p, "dev", prodConn)
	got, _ := config.Get(p, "dev")
	if got != prodConn {
		t.Errorf("got %+v, want %+v", got, prodConn)
	}
}

func TestRemove_DeletesEntry(t *testing.T) {
	p := tempPath(t)
	config.Add(p, "dev", devConn)
	config.Remove(p, "dev")
	_, err := config.Get(p, "dev")
	if !errors.Is(err, config.ErrNotFound) {
		t.Errorf("want ErrNotFound after remove, got %v", err)
	}
}

func TestRemove_ClearsDefault(t *testing.T) {
	p := tempPath(t)
	config.Add(p, "dev", devConn)
	config.SetDefault(p, "dev")
	config.Remove(p, "dev")
	def, _ := config.GetDefault(p)
	if def != "" {
		t.Errorf("default should be cleared, got %q", def)
	}
}

func TestRemove_MissingNameIsNoOp(t *testing.T) {
	p := tempPath(t)
	if err := config.Remove(p, "nonexistent"); err != nil {
		t.Errorf("Remove on missing name should not error, got %v", err)
	}
}

func TestList_Empty(t *testing.T) {
	p := tempPath(t)
	conns, err := config.List(p)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(conns) != 0 {
		t.Errorf("expected empty map, got %v", conns)
	}
}

func TestList_ReturnsAll(t *testing.T) {
	p := tempPath(t)
	config.Add(p, "dev", devConn)
	config.Add(p, "prod", prodConn)
	conns, err := config.List(p)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(conns) != 2 {
		t.Errorf("expected 2 connections, got %d", len(conns))
	}
}

func TestGet_NotFound(t *testing.T) {
	p := tempPath(t)
	_, err := config.Get(p, "missing")
	if !errors.Is(err, config.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestGetDefault_EmptyOnFreshStore(t *testing.T) {
	p := tempPath(t)
	def, err := config.GetDefault(p)
	if err != nil {
		t.Fatalf("GetDefault: %v", err)
	}
	if def != "" {
		t.Errorf("expected empty default, got %q", def)
	}
}

func TestSetDefault_HappyPath(t *testing.T) {
	p := tempPath(t)
	config.Add(p, "dev", devConn)
	if err := config.SetDefault(p, "dev"); err != nil {
		t.Fatalf("SetDefault: %v", err)
	}
	def, _ := config.GetDefault(p)
	if def != "dev" {
		t.Errorf("expected default=dev, got %q", def)
	}
}

func TestSetDefault_NotFound(t *testing.T) {
	p := tempPath(t)
	err := config.SetDefault(p, "nonexistent")
	if !errors.Is(err, config.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}
