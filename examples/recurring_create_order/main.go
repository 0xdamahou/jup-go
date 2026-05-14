package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/0xdamahou/jup-go/pkg/jupiter"
	"github.com/0xdamahou/jup-go/pkg/recurring"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = jupiter.NewClient(jupiter.ConfigFromEnv())
	n, _ := strconv.Atoi(os.Getenv("NUMBER_OF_ORDERS"))
	interval, _ := strconv.ParseInt(os.Getenv("INTERVAL_SECONDS"), 10, 64)
	req := recurring.CreateOrderRequest{
		User:            os.Getenv("WALLET"),
		InputMint:       os.Getenv("INPUT_MINT"),
		OutputMint:      os.Getenv("OUTPUT_MINT"),
		Amount:          os.Getenv("RAW_AMOUNT"),
		NumberOfOrders:  n,
		IntervalSeconds: interval,
	}
	if os.Getenv("DRY_RUN") != "0" {
		if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"dryRun": true, "request": req}); err != nil {
			log.Fatal(err)
		}
		return
	}
	_ = ctx
	log.Fatal("live recurring creation returns an unsigned transaction and requires local signing before execution")
}
