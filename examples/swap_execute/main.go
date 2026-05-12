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
	if os.Getenv("DRY_RUN") == "1" {
		log.Println("dry run: not submitting signed transaction")
		return
	}
	client := jupiter.NewClient(jupiter.ConfigFromEnv())
	out, err := client.Swap.Execute(ctx, swap.ExecuteRequest{SignedTransaction: os.Getenv("SIGNED_TRANSACTION"), RequestID: os.Getenv("REQUEST_ID")})
	if err != nil {
		log.Fatal(err)
	}
	_ = json.NewEncoder(os.Stdout).Encode(out)
}
