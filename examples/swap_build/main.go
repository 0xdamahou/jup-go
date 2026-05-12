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
	out, err := client.Swap.GetBuild(ctx, swap.BuildRequest{InputMint: os.Getenv("INPUT_MINT"), OutputMint: os.Getenv("OUTPUT_MINT"), Amount: os.Getenv("RAW_AMOUNT"), Taker: os.Getenv("TAKER")})
	if err != nil {
		log.Fatal(err)
	}
	_ = json.NewEncoder(os.Stdout).Encode(out)
}
