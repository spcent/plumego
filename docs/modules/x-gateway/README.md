# x/gateway

`x/gateway` owns API-gateway-specific transport helpers and protocol integration.

Use this extension layer for:

- reverse proxy helpers
- gateway routing and rewriting
- protocol adapter contracts
- protocol-aware HTTP middleware

Stable `contract` and `middleware` packages should not own these gateway-specific surfaces.
