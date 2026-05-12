# Testing

Default tests use `httptest.Server`; they do not call live Jupiter APIs.

```bash
go test ./...
go vet ./...
go test -race ./...
```

Live integration tests should use `//go:build integration` and require `JUPITER_API_KEY`; RPC-backed tests should also require `SOLANA_RPC_URL`.

