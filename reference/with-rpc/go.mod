module github.com/spcent/plumego/reference/with-rpc

go 1.24.0

toolchain go1.24.4

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0
	github.com/spcent/plumego v0.0.0
	github.com/spcent/plumego/x/rpc v0.0.0
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
)

require (
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260209200024-4cfbd4190f57 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260209200024-4cfbd4190f57 // indirect
)

replace github.com/spcent/plumego => ../..

replace github.com/spcent/plumego/x/rpc => ../../x/rpc
