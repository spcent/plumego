// Package proxy provides a thin HTTP adapter for the gateway reverse proxy module.
package proxy

import gateway "github.com/spcent/plumego/net/gateway"

type (
	PathRewriteFunc            = gateway.PathRewriteFunc
	RegexRewriter              = gateway.RegexRewriter
	RequestModifier            = gateway.RequestModifier
	ResponseModifier           = gateway.ResponseModifier
	ProxyError                 = gateway.ProxyError
	CircuitBreaker             = gateway.CircuitBreaker
	Backend                    = gateway.Backend
	BackendStats               = gateway.BackendStats
	BackendPool                = gateway.BackendPool
	TransportConfig            = gateway.TransportConfig
	TransportPool              = gateway.TransportPool
	Proxy                      = gateway.Proxy
	ProxyStats                 = gateway.ProxyStats
	HealthChecker              = gateway.HealthChecker
	Config                     = gateway.Config
	ErrorHandlerFunc           = gateway.ErrorHandlerFunc
	HealthCheckConfig          = gateway.HealthCheckConfig
	ServiceDiscovery           = gateway.ServiceDiscovery
	BufferPool                 = gateway.BufferPool
	CircuitBreakerConfig       = gateway.CircuitBreakerConfig
	LoadBalancer               = gateway.LoadBalancer
	RoundRobinBalancer         = gateway.RoundRobinBalancer
	RandomBalancer             = gateway.RandomBalancer
	WeightedRoundRobinBalancer = gateway.WeightedRoundRobinBalancer
	IPHashBalancer             = gateway.IPHashBalancer
	LeastConnectionsBalancer   = gateway.LeastConnectionsBalancer
)

var (
	StripPrefix                   = gateway.StripPrefix
	AddPrefix                     = gateway.AddPrefix
	ReplacePrefix                 = gateway.ReplacePrefix
	RewriteMap                    = gateway.RewriteMap
	RewriteRegex                  = gateway.RewriteRegex
	Chain                         = gateway.Chain
	AddForwardedHeaders           = gateway.AddForwardedHeaders
	RemoveHopByHopHeaders         = gateway.RemoveHopByHopHeaders
	AddHeader                     = gateway.AddHeader
	SetHeader                     = gateway.SetHeader
	DelHeader                     = gateway.DelHeader
	AddResponseHeader             = gateway.AddResponseHeader
	SetResponseHeader             = gateway.SetResponseHeader
	DelResponseHeader             = gateway.DelResponseHeader
	ChainRequestModifiers         = gateway.ChainRequestModifiers
	ChainResponseModifiers        = gateway.ChainResponseModifiers
	NewProxyError                 = gateway.NewProxyError
	NewBackend                    = gateway.NewBackend
	NewBackendPool                = gateway.NewBackendPool
	DefaultTransportConfig        = gateway.DefaultTransportConfig
	NewTransportPool              = gateway.NewTransportPool
	New                           = gateway.New
	NewHealthChecker              = gateway.NewHealthChecker
	NewRoundRobinBalancer         = gateway.NewRoundRobinBalancer
	NewRandomBalancer             = gateway.NewRandomBalancer
	NewWeightedRoundRobinBalancer = gateway.NewWeightedRoundRobinBalancer
	NewIPHashBalancer             = gateway.NewIPHashBalancer
	NewLeastConnectionsBalancer   = gateway.NewLeastConnectionsBalancer
)
