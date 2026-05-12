package main

import "testing"

func TestSwapExecuteDryRun(t *testing.T) {
	err := run([]string{"--json", "--dry-run", "swap", "execute", "--signed-transaction", "abc", "--request-id", "req"})
	if err != nil {
		t.Fatal(err)
	}
}
