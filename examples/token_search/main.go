package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/0xdamahou/jup-go/pkg/jupiter"
	"github.com/0xdamahou/jup-go/pkg/token"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := jupiter.NewClient(jupiter.ConfigFromEnv()).Token.Search(ctx, token.SearchRequest{Query: os.Getenv("TOKEN_QUERY")})
	if err != nil {
		log.Fatal(err)
	}
	_ = json.NewEncoder(os.Stdout).Encode(out)
}
