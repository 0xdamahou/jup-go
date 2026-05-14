# Recurring

Base: `https://api.jup.ag/recurring/v1`.

Implemented:

- `CreateOrder`
- `Execute`
- `CancelOrder`
- `Orders`
- `History`
- `GetOrders`

`CreateOrder` sends the documented time-based shape:

```json
{
  "user": "...",
  "inputMint": "...",
  "outputMint": "...",
  "params": {
    "time": {
      "inAmount": "1000000",
      "numberOfOrders": 2,
      "interval": 86400
    }
  }
}
```

`History` uses the documented `GET /getRecurringOrders` endpoint with `orderStatus=history`; there is no guessed history endpoint.

Recurring orders have schedule, amount, wallet ownership, liquidity, ATA, and Token-2022 risks. The client validates raw amounts, positive interval, future `startAt`, minimum order count, and optional price range, but caller-side business rules should also enforce notional and ownership constraints.
