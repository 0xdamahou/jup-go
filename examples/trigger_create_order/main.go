package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/0xdamahou/jup-go/pkg/jupiter"
	"github.com/0xdamahou/jup-go/pkg/trigger"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = jupiter.NewClient(jupiter.ConfigFromEnv())
	req := trigger.CreateOrderRequest{
		WalletPubkey:     os.Getenv("WALLET"),
		InputMint:        os.Getenv("INPUT_MINT"),
		OutputMint:       os.Getenv("OUTPUT_MINT"),
		MakingAmount:     os.Getenv("RAW_AMOUNT"),
		TriggerPrice:     os.Getenv("TRIGGER_PRICE"),
		TriggerCondition: trigger.TriggerCondition(os.Getenv("TRIGGER_CONDITION")),
		OrderType:        trigger.Single,
	}
	if os.Getenv("DRY_RUN") != "0" {
		if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"dryRun": true, "request": req}); err != nil {
			log.Fatal(err)
		}
		return
	}
	_ = ctx
	log.Fatal("live trigger creation requires wallet auth, deposit crafting, local signing, and signed deposit transaction submission")
}
