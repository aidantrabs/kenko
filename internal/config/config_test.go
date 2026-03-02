package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeConfig(t, `
port: 8080
check_interval: 10s
check_timeout: 3s
redis_addr: localhost:6379
targets:
  - name: example
    url: https://example.com
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("port = %d, want 8080", cfg.Port)
	}
	if len(cfg.Targets) != 1 {
		t.Errorf("targets = %d, want 1", len(cfg.Targets))
	}
	if cfg.Targets[0].Name != "example" {
		t.Errorf("target name = %q, want %q", cfg.Targets[0].Name, "example")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	path := writeConfig(t, `
port: 99999
check_interval: 10s
check_timeout: 3s
redis_addr: localhost:6379
targets:
  - name: test
    url: https://example.com
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestLoad_ZeroPort(t *testing.T) {
	path := writeConfig(t, `
port: 0
check_interval: 10s
check_timeout: 3s
redis_addr: localhost:6379
targets:
  - name: test
    url: https://example.com
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for zero port")
	}
}

func TestLoad_NoTargets(t *testing.T) {
	path := writeConfig(t, `
port: 8080
check_interval: 10s
check_timeout: 3s
redis_addr: localhost:6379
targets: []
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty targets")
	}
}

func TestLoad_InvalidURL(t *testing.T) {
	path := writeConfig(t, `
port: 8080
check_interval: 10s
check_timeout: 3s
redis_addr: localhost:6379
targets:
  - name: bad
    url: not-a-url
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid url")
	}
}

func TestLoad_EmptyRedisAddr(t *testing.T) {
	path := writeConfig(t, `
port: 8080
check_interval: 10s
check_timeout: 3s
redis_addr: ""
targets:
  - name: test
    url: https://example.com
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty redis_addr")
	}
}

func TestLoad_ZeroInterval(t *testing.T) {
	path := writeConfig(t, `
port: 8080
check_interval: 0s
check_timeout: 3s
redis_addr: localhost:6379
targets:
  - name: test
    url: https://example.com
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for zero interval")
	}
}

func TestLoad_EnvExpansion(t *testing.T) {
	t.Setenv("TEST_REDIS_PASS", "secret123")

	path := writeConfig(t, `
port: 8080
check_interval: 10s
check_timeout: 3s
redis_addr: localhost:6379
redis_password: ${TEST_REDIS_PASS}
targets:
  - name: test
    url: https://example.com
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RedisPassword != "secret123" {
		t.Errorf("redis_password = %q, want %q", cfg.RedisPassword, "secret123")
	}
}
