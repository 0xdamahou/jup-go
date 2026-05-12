package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/0xdamahou/jup-go/pkg/jupiter"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := jupiter.NewClient(jupiter.ConfigFromEnv()).Perps.Markets(ctx)
	if err != nil {
		log.Fatal(err)
	}
	_ = json.NewEncoder(os.Stdout).Encode(out)
}
