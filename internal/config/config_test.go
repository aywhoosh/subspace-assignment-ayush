package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate_RejectsNonLocalBaseURL(t *testing.T) {
	cfg := Default()
	cfg.Mocknet.BaseURL = "https://example.com"
	if err := Validate(cfg); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestLoad_EnvOverridesConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	contents := []byte(`mocknet:
  base_url: "http://localhost:7777"
storage:
  sqlite_path: "` + filepath.ToSlash(filepath.Join(dir, "db.sqlite")) + `"
logging:
  level: "info"
  json: true
run:
  timeout: 10s
`)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	// Override via convenience env var.
	t.Setenv("SUBSPACE_MOCKNET_BASE_URL", "http://localhost:8888")
	// Override via nested env var.
	t.Setenv("SUBSPACE_LOGGING__LEVEL", "debug")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Mocknet.BaseURL != "http://localhost:8888" {
		t.Fatalf("expected base_url override, got %q", cfg.Mocknet.BaseURL)
	}
	if cfg.Logging.Level != "debug" {
		t.Fatalf("expected logging.level override, got %q", cfg.Logging.Level)
	}
	if cfg.Run.Timeout.String() != "10s" {
		t.Fatalf("expected timeout=10s, got %s", cfg.Run.Timeout)
	}
}

func TestLoad_ConvenienceAuthEnvVars(t *testing.T) {
	t.Setenv("SUBSPACE_AUTH_USERNAME", "u")
	t.Setenv("SUBSPACE_AUTH_PASSWORD", "p")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Auth.Username != "u" || cfg.Auth.Password != "p" {
		t.Fatalf("expected auth creds from env, got username=%q password=%q", cfg.Auth.Username, cfg.Auth.Password)
	}
}
