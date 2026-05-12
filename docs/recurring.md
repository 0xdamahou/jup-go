# Recurring

Base: `https://api.jup.ag/recurring/v1`.

Implemented:

- `CreateOrder`
- `Execute`
- `CancelOrder`
- `Orders`

Recurring orders have schedule, amount, wallet ownership, liquidity, ATA, and Token-2022 risks. The client validates raw amounts and minimum order count, but caller-side business rules should also enforce notional and ownership constraints.

