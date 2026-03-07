package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPrefersEnvironmentOverYAML(t *testing.T) {
	t.Setenv("APP_PORT", "7070")

	tempDir := t.TempDir()
	configYAML := []byte("APP_PORT: \"9090\"\nAPP_BASE_URL: \"http://yaml.example\"\n")
	if err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), configYAML, 0644); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Port != "7070" {
		t.Fatalf("expected APP_PORT from env, got %q", cfg.App.Port)
	}
	if cfg.App.BaseURL != "http://yaml.example" {
		t.Fatalf("expected APP_BASE_URL from yaml, got %q", cfg.App.BaseURL)
	}
}
