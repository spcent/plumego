package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func (c *Collector) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(c.PrometheusText()))
	})
}

func (c *Collector) PrometheusText() string {
	snap := c.snapshot()
	var b strings.Builder
	writeSimpleMetrics(&b, snap.gauges, MetricKindGauge)
	writeSimpleMetrics(&b, snap.counters, MetricKindCounter)
	writeHistogramMetrics(&b, snap)
	return b.String()
}

func writeSimpleMetrics(b *strings.Builder, values map[seriesKey]float64, kind MetricKind) {
	for _, name := range sortedMetricNames(values) {
		spec := catalogByName()[name]
		writeHelpAndType(b, name, kind, spec.Description)
		for _, key := range sortedSeries(values, name) {
			fmt.Fprintf(b, "%s%s %.9f\n", key.name, labelSuffix(key.labels), values[key])
		}
		b.WriteByte('\n')
	}
}

func writeHistogramMetrics(b *strings.Builder, snap collectorSnapshot) {
	for _, name := range sortedHistogramNames(snap.histograms) {
		spec := catalogByName()[name]
		writeHelpAndType(b, name, MetricKindHistogram, spec.Description)
		for _, key := range sortedHistogramSeries(snap.histograms, name) {
			state := snap.histograms[key]
			for i, bucket := range snap.buckets {
				labels := appendLabel(key.labels, "le", formatBucket(bucket))
				fmt.Fprintf(b, "%s_bucket%s %d\n", key.name, labelSuffix(labels), state.buckets[i])
			}
			fmt.Fprintf(b, "%s_bucket%s %d\n", key.name, labelSuffix(appendLabel(key.labels, "le", "+Inf")), state.count)
			fmt.Fprintf(b, "%s_sum%s %.9f\n", key.name, labelSuffix(key.labels), state.sum)
			fmt.Fprintf(b, "%s_count%s %d\n", key.name, labelSuffix(key.labels), state.count)
		}
		b.WriteByte('\n')
	}
}

func writeHelpAndType(b *strings.Builder, name string, kind MetricKind, help string) {
	if help == "" {
		help = name
	}
	fmt.Fprintf(b, "# HELP %s %s\n", name, escapeLabelValue(help))
	fmt.Fprintf(b, "# TYPE %s %s\n", name, kind)
}

func sortedMetricNames(values map[seriesKey]float64) []string {
	set := map[string]struct{}{}
	for key := range values {
		set[key.name] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func sortedHistogramNames(values map[seriesKey]histogramSnapshot) []string {
	set := map[string]struct{}{}
	for key := range values {
		set[key.name] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func sortedSeries(values map[seriesKey]float64, name string) []seriesKey {
	out := make([]seriesKey, 0)
	for key := range values {
		if key.name == name {
			out = append(out, key)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].labels < out[j].labels })
	return out
}

func sortedHistogramSeries(values map[seriesKey]histogramSnapshot, name string) []seriesKey {
	out := make([]seriesKey, 0)
	for key := range values {
		if key.name == name {
			out = append(out, key)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].labels < out[j].labels })
	return out
}

func labelSuffix(labels string) string {
	if labels == "" {
		return ""
	}
	parts := strings.Split(labels, ",")
	for i, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		parts[i] = fmt.Sprintf(`%s="%s"`, key, escapeLabelValue(value))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func appendLabel(labels string, key string, value string) string {
	if labels == "" {
		return key + "=" + value
	}
	return labels + "," + key + "=" + value
}

func formatBucket(bucket float64) string {
	return strconv.FormatFloat(bucket, 'f', -1, 64)
}

func escapeLabelValue(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return value
}

func catalogByName() map[string]MetricSpec {
	out := map[string]MetricSpec{}
	for _, spec := range Catalog() {
		out[spec.Name] = spec
	}
	return out
}
