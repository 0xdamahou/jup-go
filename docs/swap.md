# Swap

Base: `https://api.jup.ag/swap/v2`.

Implemented:

- `GET /order` via `client.Swap.GetOrder`.
- `GET /build` via `client.Swap.GetBuild`.
- `POST /execute` via `client.Swap.Execute`.

Use `/order` plus `/execute` for managed landing. Use `/build` only when the caller needs custom transaction assembly and self-managed RPC submission. `/execute` is not automatically retried because it submits a signed transaction.

`GetBuild` includes priority-fee safety helpers for custom execution:

- It forwards `priorityFeeLamports`, `router`, `payer`, referral, receiver, and slippage parameters to `/build`.
- It parses `computeBudgetInstructions` for `SetComputeUnitLimit` and `SetComputeUnitPrice`.
- `PriorityFeeMicrolamports` overrides the returned `SetComputeUnitPrice` instruction in `BuildResponse.Raw`.
- `MaxFeeLamports` rejects responses where `ceil(CU limit * CU price / 1_000_000)` exceeds the caller's limit.
- If a CU price exists without a CU limit and `MaxFeeLamports` is set, the client returns an error because it cannot estimate the upper bound safely.
- Instruction data supports base64 and base58 through `github.com/mr-tron/base58`; the SDK no longer maintains its own base58 codec.

Always treat `Amount` as raw integer token units. Keep quote freshness checks in caller strategy; never treat a quote as executable forever.
