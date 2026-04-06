package observability

import (
	"fmt"
	"net/http"
)

type Exporter interface {
	Handler() http.Handler
}

type PrometheusExporter struct {
	collector *PrometheusCollector
}

func NewPrometheusExporter(collector *PrometheusCollector) *PrometheusExporter {
	if collector == nil {
		panic("observability prometheus exporter requires a collector")
	}
	return &PrometheusExporter{collector: collector}
}

func (e *PrometheusExporter) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests, durations, uptime := e.collector.snapshot()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")

		fmt.Fprintf(w, "# HELP %s_http_requests_total Total number of HTTP requests processed.\n", e.collector.namespace)
		fmt.Fprintf(w, "# TYPE %s_http_requests_total counter\n", e.collector.namespace)

		for _, key := range sortedKeys(requests) {
			fmt.Fprintf(w, "%s_http_requests_total{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
				e.collector.namespace,
				escapeLabelValue(key.method),
				escapeLabelValue(key.path),
				escapeLabelValue(key.status),
				requests[key],
			)
		}

		fmt.Fprintln(w)
		fmt.Fprintf(w, "# HELP %s_http_request_duration_seconds_sum Sum of HTTP request latencies in seconds.\n", e.collector.namespace)
		fmt.Fprintf(w, "# TYPE %s_http_request_duration_seconds_summary summary\n", e.collector.namespace)

		for _, key := range sortedKeys(durations) {
			stats := durations[key]
			method := escapeLabelValue(key.method)
			path := escapeLabelValue(key.path)
			status := escapeLabelValue(key.status)
			fmt.Fprintf(w, "%s_http_request_duration_seconds_sum{method=\"%s\",path=\"%s\",status=\"%s\"} %.9f\n",
				e.collector.namespace, method, path, status, stats.sum)
			fmt.Fprintf(w, "%s_http_request_duration_seconds_count{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
				e.collector.namespace, method, path, status, stats.count)
			fmt.Fprintf(w, "%s_http_request_duration_seconds_min{method=\"%s\",path=\"%s\",status=\"%s\"} %.9f\n",
				e.collector.namespace, method, path, status, stats.min)
			fmt.Fprintf(w, "%s_http_request_duration_seconds_max{method=\"%s\",path=\"%s\",status=\"%s\"} %.9f\n",
				e.collector.namespace, method, path, status, stats.max)
		}

		fmt.Fprintln(w)
		fmt.Fprintf(w, "# HELP %s_uptime_seconds Total uptime in seconds.\n", e.collector.namespace)
		fmt.Fprintf(w, "# TYPE %s_uptime_seconds gauge\n", e.collector.namespace)
		fmt.Fprintf(w, "%s_uptime_seconds %.3f\n", e.collector.namespace, uptime.Seconds())

		fmt.Fprintln(w)
		fmt.Fprintf(w, "# HELP %s_http_requests_total_all Total requests across all labels.\n", e.collector.namespace)
		fmt.Fprintf(w, "# TYPE %s_http_requests_total_all counter\n", e.collector.namespace)
		var totalRequests uint64
		for _, count := range requests {
			totalRequests += count
		}
		fmt.Fprintf(w, "%s_http_requests_total_all %d\n", e.collector.namespace, totalRequests)
	})
}
