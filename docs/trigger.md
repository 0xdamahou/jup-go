# Trigger

Base: `https://api.jup.ag/trigger/v2`.

Implemented helpers cover challenge/verify auth, price order creation, cancel initiation, cancel confirmation, open-order query, and history query.

Order mutation requires wallet auth and local signing. The client does not sign transactions and does not retry order mutation calls automatically.

