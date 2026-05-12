# Swap

Base: `https://api.jup.ag/swap/v2`.

Implemented:

- `GET /order` via `client.Swap.GetOrder`.
- `GET /build` via `client.Swap.GetBuild`.
- `POST /execute` via `client.Swap.Execute`.

Use `/order` plus `/execute` for managed landing. Use `/build` only when the caller needs custom transaction assembly and self-managed RPC submission. `/execute` is not automatically retried because it submits a signed transaction.

Always treat `Amount` as raw integer token units. Keep quote freshness checks in caller strategy; never treat a quote as executable forever.

