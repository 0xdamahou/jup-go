# Trigger

Base: `https://api.jup.ag/trigger/v2`.

Implemented helpers cover challenge/verify auth, price order creation, cancel initiation, cancel confirmation, open-order query, and history query.

`CreateOCOPriceOrder` and `CreateOTOCOPriceOrder` provide typed helpers for OCO and OTOCO flows. They validate that linked legs share wallet/token pair and that OCO legs use different trigger conditions before posting to `/trigger/v2/orders/price`.

Order mutation requires wallet auth and local signing. The client does not sign transactions and does not retry order mutation calls automatically.
