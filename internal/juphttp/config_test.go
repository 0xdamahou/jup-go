package juphttp

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfigFileJSONAndRedaction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jupiter.json")
	if err := os.WriteFile(path, []byte(`{"apiKey":"abcdefghijklmnopqrstuvwxyz","timeout":5000000000,"maxRetries":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfigFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey == "" || cfg.Timeout != 5*time.Second || cfg.MaxRetries != 1 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if got := RedactSecret(cfg.APIKey); got != "abcd...wxyz" {
		t.Fatalf("redaction = %q", got)
	}
}

func TestLoadConfigFileFlatYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jupiter.yaml")
	raw := "baseURL: http://example.test\nmaxRetries: 3\nretryBackoff: 10ms\n"
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfigFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "http://example.test" || cfg.MaxRetries != 3 || cfg.RetryBackoff != 10*time.Millisecond {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}
