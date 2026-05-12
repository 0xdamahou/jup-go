# jup-go

Production-oriented Go SDK and CLI foundation for Jupiter APIs on Solana.

The client uses `context.Context`, timeouts, structured API errors, bounded response bodies, retry/backoff for safe requests, and isolated transaction submission methods. It never handles private keys; signing stays in caller-owned code.

```go
client := jupiter.NewClient(jupiter.ConfigFromEnv())
order, err := client.Swap.GetOrder(ctx, swap.GetOrderRequest{
	InputMint:  "So11111111111111111111111111111111111111112",
	OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
	Amount:     "1000000",
	Taker:      wallet,
})
```

Covered packages:

- `pkg/swap`: Swap API V2 `/order`, `/build`, `/execute`.
- `pkg/token`: Tokens V2 search, tags, category, recent.
- `pkg/price`: Price V3 batch lookup.
- `pkg/trigger`: auth, price order create/cancel/history helpers.
- `pkg/recurring`: create, execute, cancel, query recurring orders.
- `pkg/lend`: earn markets, positions, deposit, withdraw; borrow/repay placeholders.
- `pkg/portfolio`: holdings and positions.
- `pkg/perps`: query helpers plus explicit unsupported mutation errors.
- `pkg/prediction`: market/outcome query helpers plus explicit unsupported mutation errors.

Run:

```bash
go test ./...
go vet ./...
go test -race ./...
```

CLI:

```bash
go run ./cmd/jupctl --json price get --id So11111111111111111111111111111111111111112
go run ./cmd/jupctl --json token search --query SOL
go run ./cmd/jupctl --dry-run swap execute --signed-transaction base64 --request-id req
```

Trading, lending, recurring, trigger, prediction, and perps flows carry financial and transaction risk. Quotes can become stale, liquidity can move, slippage can exceed expectations, transactions can fail to land, and partial execution can occur. Requote before execution and set explicit risk limits.

