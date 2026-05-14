package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/0xdamahou/jup-go/pkg/jupiter"
	"github.com/0xdamahou/jup-go/pkg/lend"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := jupiter.NewClient(jupiter.ConfigFromEnv())
	req := lend.EarnActionRequest{
		Owner:  os.Getenv("WALLET"),
		Mint:   os.Getenv("MINT"),
		Amount: os.Getenv("RAW_AMOUNT"),
	}
	if os.Getenv("DRY_RUN") != "0" {
		if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"dryRun": true, "request": req}); err != nil {
			log.Fatal(err)
		}
		return
	}
	out, err := client.Lend.Deposit(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		log.Fatal(err)
	}
}
