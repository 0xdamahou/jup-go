package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSwapExecuteDryRun(t *testing.T) {
	err := run([]string{"--json", "--dry-run", "swap", "execute", "--signed-transaction", "abc", "--request-id", "req"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestConfigAPIKeyNotOverwrittenByEmptyDefaultEnv(t *testing.T) {
	t.Setenv("JUPITER_API_KEY", "")
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"apiKey":"from-file","timeout":1000000000}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"--config", path, "--json", "--dry-run", "swap", "execute", "--signed-transaction", "abc"}); err != nil {
		t.Fatal(err)
	}
}

func TestTriggerAndRecurringDryRunCommands(t *testing.T) {
	cases := [][]string{
		{"--json", "--dry-run", "trigger", "create", "--wallet", "11111111111111111111111111111111", "--input-mint", "So11111111111111111111111111111111111111112", "--output-mint", "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", "--amount", "100", "--trigger-price", "1", "--condition", "above"},
		{"--json", "--dry-run", "trigger", "cancel", "--order-id", "order-1"},
		{"--json", "--dry-run", "recurring", "create", "--wallet", "11111111111111111111111111111111", "--input-mint", "So11111111111111111111111111111111111111112", "--output-mint", "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", "--amount", "100", "--orders", "2", "--interval-seconds", "60"},
		{"--json", "--dry-run", "recurring", "cancel", "--order-id", "order-1", "--wallet", "11111111111111111111111111111111"},
	}
	for _, tc := range cases {
		if err := run(tc); err != nil {
			t.Fatalf("%v: %v", tc, err)
		}
	}
}
