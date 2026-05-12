# Price

Base: `https://lite-api.jup.ag/price/v3`.

Implemented:

- `client.Price.Get(ctx, price.GetRequest{IDs: []string{...}})`

The client validates that at least one mint is supplied and caps requests at `price.MaxIDsPerRequest`. Missing prices should be handled explicitly by callers.

