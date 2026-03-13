# x/gateway/protocol

`x/gateway/protocol` defines protocol adapter contracts for gateway-specific integrations such as gRPC and GraphQL.

Use it with `x/gateway/protocolmw` when you want an HTTP request to be transformed and dispatched through a protocol adapter without adding non-stdlib dependencies to the stable core.
