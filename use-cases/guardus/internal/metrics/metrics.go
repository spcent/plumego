// Package metrics exposes endpoint-probe Prometheus metrics for guardus.
//
// Unlike upstream gatus, the registerer is owned by Endpoints (not a package
// global), so Reload can swap registries cleanly and tests can use isolated
// registries without affecting prometheus.DefaultRegisterer.
package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"guardus/internal/domain/endpoint"
)

const namespace = "guardus"

// Endpoints owns the per-endpoint Prometheus collectors.
//
// The same Endpoints instance is shared across the watchdog and the /metrics
// handler. Reload should construct a fresh Endpoints (with a fresh registry)
// rather than mutating an existing one — concurrent updates to label sets are
// undefined.
type Endpoints struct {
	registry prometheus.Registerer
	gatherer prometheus.Gatherer

	resultTotal                        *prometheus.CounterVec
	resultDurationSeconds              *prometheus.GaugeVec
	resultConnectedTotal               *prometheus.CounterVec
	resultCodeTotal                    *prometheus.CounterVec
	resultCertificateExpirationSeconds *prometheus.GaugeVec
	resultDomainExpirationSeconds      *prometheus.GaugeVec
	resultEndpointSuccess              *prometheus.GaugeVec
}

// NewEndpoints registers endpoint metrics on a fresh isolated registry.
//
// extraLabels are appended after the default label set on every metric;
// PublishResult fills missing labels with empty strings.
func NewEndpoints(extraLabels []string) *Endpoints {
	reg := prometheus.NewRegistry()
	return newEndpointsWithRegistry(reg, reg, extraLabels)
}

func newEndpointsWithRegistry(reg prometheus.Registerer, gat prometheus.Gatherer, extraLabels []string) *Endpoints {
	e := &Endpoints{registry: reg, gatherer: gat}
	e.resultTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "results_total",
		Help:      "Number of results per endpoint",
	}, append([]string{"key", "group", "name", "type", "success"}, extraLabels...))
	e.resultDurationSeconds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "results_duration_seconds",
		Help:      "Duration of the request in seconds",
	}, append([]string{"key", "group", "name", "type"}, extraLabels...))
	e.resultConnectedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "results_connected_total",
		Help:      "Total number of results in which a connection was successfully established",
	}, append([]string{"key", "group", "name", "type"}, extraLabels...))
	e.resultCodeTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "results_code_total",
		Help:      "Total number of results by code",
	}, append([]string{"key", "group", "name", "type", "code"}, extraLabels...))
	e.resultCertificateExpirationSeconds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "results_certificate_expiration_seconds",
		Help:      "Number of seconds until the certificate expires",
	}, append([]string{"key", "group", "name", "type"}, extraLabels...))
	e.resultDomainExpirationSeconds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "results_domain_expiration_seconds",
		Help:      "Number of seconds until the domain expires",
	}, append([]string{"key", "group", "name", "type"}, extraLabels...))
	e.resultEndpointSuccess = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "results_endpoint_success",
		Help:      "Whether the endpoint was a success (1) or not (0)",
	}, append([]string{"key", "group", "name", "type"}, extraLabels...))
	reg.MustRegister(
		e.resultTotal,
		e.resultDurationSeconds,
		e.resultConnectedTotal,
		e.resultCodeTotal,
		e.resultCertificateExpirationSeconds,
		e.resultDomainExpirationSeconds,
		e.resultEndpointSuccess,
	)
	return e
}

// Gatherer returns the Prometheus gatherer wired to the /metrics handler.
func (e *Endpoints) Gatherer() prometheus.Gatherer { return e.gatherer }

// PublishResult records a probe outcome. extraLabels supplies the label
// ordering: any label missing from ep.ExtraLabels becomes the empty string.
func (e *Endpoints) PublishResult(ep *endpoint.Endpoint, result *endpoint.Result, extraLabels []string) {
	if e == nil {
		return
	}
	values := make([]string, 0, len(extraLabels))
	for _, label := range extraLabels {
		values = append(values, ep.ExtraLabels[label])
	}
	endpointType := string(ep.Type())
	base := []string{ep.Key(), ep.Group, ep.Name, endpointType}

	e.resultTotal.WithLabelValues(append(append([]string{}, base[0], base[1], base[2], base[3], strconv.FormatBool(result.Success)), values...)...).Inc()
	e.resultDurationSeconds.WithLabelValues(append(append([]string{}, base...), values...)...).Set(result.Duration.Seconds())
	if result.Connected {
		e.resultConnectedTotal.WithLabelValues(append(append([]string{}, base...), values...)...).Inc()
	}
	if result.DNSRCode != "" {
		e.resultCodeTotal.WithLabelValues(append(append([]string{}, base[0], base[1], base[2], base[3], result.DNSRCode), values...)...).Inc()
	}
	if result.HTTPStatus != 0 {
		e.resultCodeTotal.WithLabelValues(append(append([]string{}, base[0], base[1], base[2], base[3], strconv.Itoa(result.HTTPStatus)), values...)...).Inc()
	}
	if result.CertificateExpiration != 0 {
		e.resultCertificateExpirationSeconds.WithLabelValues(append(append([]string{}, base...), values...)...).Set(result.CertificateExpiration.Seconds())
	}
	if result.DomainExpiration != 0 {
		e.resultDomainExpirationSeconds.WithLabelValues(append(append([]string{}, base...), values...)...).Set(result.DomainExpiration.Seconds())
	}
	if result.Success {
		e.resultEndpointSuccess.WithLabelValues(append(append([]string{}, base...), values...)...).Set(1)
	} else {
		e.resultEndpointSuccess.WithLabelValues(append(append([]string{}, base...), values...)...).Set(0)
	}
}
