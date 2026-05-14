package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/0xdamahou/jup-go/pkg/jupiter"
	"github.com/0xdamahou/jup-go/pkg/swap"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := jupiter.NewClient(jupiter.ConfigFromEnv())
	out, err := client.Swap.GetOrder(ctx, swap.GetOrderRequest{
		InputMint:  os.Getenv("INPUT_MINT"),
		OutputMint: os.Getenv("OUTPUT_MINT"),
		Amount:     os.Getenv("RAW_AMOUNT"),
		Taker:      os.Getenv("TAKER"),
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		log.Fatal(err)
	}
}
