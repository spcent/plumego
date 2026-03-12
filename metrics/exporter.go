package metrics

import (
	"fmt"
	"net/http"
)

// Exporter exposes collected metrics through an HTTP handler.
type Exporter interface {
	Handler() http.Handler
}

// PrometheusExporter renders PrometheusCollector state using the Prometheus
// text exposition format.
type PrometheusExporter struct {
	collector *PrometheusCollector
}

// NewPrometheusExporter constructs an explicit HTTP exporter for a collector.
func NewPrometheusExporter(collector *PrometheusCollector) *PrometheusExporter {
	if collector == nil {
		panic("metrics prometheus exporter requires a collector")
	}
	return &PrometheusExporter{collector: collector}
}

// Handler returns an HTTP handler that emits the current metrics snapshot.
func (e *PrometheusExporter) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests, durations, uptime := e.collector.snapshot()
		smsSnapshot := e.collector.snapshotSMSGateway()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")

		fmt.Fprintf(w, "# HELP %s_http_requests_total Total number of HTTP requests processed.\n", e.collector.namespace)
		fmt.Fprintf(w, "# TYPE %s_http_requests_total counter\n", e.collector.namespace)

		reqKeys := sortedKeys(requests)
		for _, k := range reqKeys {
			fmt.Fprintf(w, "%s_http_requests_total{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
				e.collector.namespace,
				escapeLabelValue(k.method),
				escapeLabelValue(k.path),
				escapeLabelValue(k.status),
				requests[k],
			)
		}

		fmt.Fprintln(w)
		fmt.Fprintf(w, "# HELP %s_http_request_duration_seconds_sum Sum of HTTP request latencies in seconds.\n", e.collector.namespace)
		fmt.Fprintf(w, "# TYPE %s_http_request_duration_seconds_summary summary\n", e.collector.namespace)

		durKeys := sortedKeys(durations)
		for _, k := range durKeys {
			stats := durations[k]
			em := escapeLabelValue(k.method)
			ep := escapeLabelValue(k.path)
			es := escapeLabelValue(k.status)
			fmt.Fprintf(w, "%s_http_request_duration_seconds_sum{method=\"%s\",path=\"%s\",status=\"%s\"} %.9f\n",
				e.collector.namespace, em, ep, es, stats.sum)
			fmt.Fprintf(w, "%s_http_request_duration_seconds_count{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
				e.collector.namespace, em, ep, es, stats.count)
			fmt.Fprintf(w, "%s_http_request_duration_seconds_min{method=\"%s\",path=\"%s\",status=\"%s\"} %.9f\n",
				e.collector.namespace, em, ep, es, stats.min)
			fmt.Fprintf(w, "%s_http_request_duration_seconds_max{method=\"%s\",path=\"%s\",status=\"%s\"} %.9f\n",
				e.collector.namespace, em, ep, es, stats.max)
		}

		fmt.Fprintln(w)
		fmt.Fprintf(w, "# HELP %s_uptime_seconds Total uptime in seconds.\n", e.collector.namespace)
		fmt.Fprintf(w, "# TYPE %s_uptime_seconds gauge\n", e.collector.namespace)
		fmt.Fprintf(w, "%s_uptime_seconds %.3f\n", e.collector.namespace, uptime.Seconds())

		fmt.Fprintln(w)
		fmt.Fprintf(w, "# HELP %s_http_requests_total_all Total requests across all labels.\n", e.collector.namespace)
		fmt.Fprintf(w, "# TYPE %s_http_requests_total_all counter\n", e.collector.namespace)
		totalRequests := uint64(0)
		for _, count := range requests {
			totalRequests += count
		}
		fmt.Fprintf(w, "%s_http_requests_total_all %d\n", e.collector.namespace, totalRequests)

		if smsSnapshot != nil {
			e.collector.writeSMSGatewayMetrics(w, smsSnapshot)
		}
	})
}
