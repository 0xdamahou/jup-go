# Security

Do not log private keys, seed phrases, signed transactions, bearer tokens, API keys, or full authorization headers.

The SDK does not accept private keys and does not sign transactions. Decode, sign, and custody keys in caller-controlled code. Use raw token units for amounts and validate wallet/mint public keys before calling mutation endpoints.

