package contract

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// TraceID represents a unique identifier for a trace.
type TraceID string

// SpanID represents a unique identifier for a span.
type SpanID string

// TraceFlags are flags that control tracing behavior.
type TraceFlags uint8

const (
	// TraceFlagsSampled indicates that the trace should be sampled.
	TraceFlagsSampled TraceFlags = 0x01
)

// Trace represents a distributed trace containing multiple spans.
type Trace struct {
	ID         TraceID        `json:"trace_id"`
	RootSpanID SpanID         `json:"root_span_id"`
	StartTime  time.Time      `json:"start_time"`
	EndTime    *time.Time     `json:"end_time,omitempty"`
	Spans      []*Span        `json:"spans"`
	Attributes map[string]any `json:"attributes"`
	Links      []TraceLink    `json:"links,omitempty"`
}

// Span represents a single operation in a distributed trace.
type Span struct {
	ID           SpanID         `json:"span_id"`
	ParentSpanID *SpanID        `json:"parent_span_id,omitempty"`
	TraceID      TraceID        `json:"trace_id"`
	Name         string         `json:"name"`
	Kind         SpanKind       `json:"kind"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      *time.Time     `json:"end_time,omitempty"`
	Duration     time.Duration  `json:"duration,omitempty"`
	Status       SpanStatus     `json:"status"`
	Attributes   map[string]any `json:"attributes"`
	Events       []SpanEvent    `json:"events,omitempty"`
	Links        []SpanLink     `json:"links,omitempty"`
}

// SpanKind represents the kind of span.
type SpanKind string

const (
	SpanKindInternal SpanKind = "internal"
	SpanKindServer   SpanKind = "server"
	SpanKindClient   SpanKind = "client"
	SpanKindProducer SpanKind = "producer"
	SpanKindConsumer SpanKind = "consumer"
)

// SpanStatus represents the status of a span.
type SpanStatus struct {
	Code        StatusCode `json:"code"`
	Description string     `json:"description,omitempty"`
}

// StatusCode represents the status of a span.
type StatusCode int32

const (
	StatusCodeUnset StatusCode = 0
	StatusCodeOk    StatusCode = 1
	StatusCodeError StatusCode = 2
)

// SpanEvent represents an event that occurred during a span.
type SpanEvent struct {
	Name       string         `json:"name"`
	Timestamp  time.Time      `json:"timestamp"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// SpanLink represents a link to another span.
type SpanLink struct {
	TraceID    TraceID        `json:"trace_id"`
	SpanID     SpanID         `json:"span_id"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// TraceLink represents a link between traces.
type TraceLink struct {
	TraceID    TraceID        `json:"trace_id"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// TraceContext represents the current tracing context.
type TraceContext struct {
	TraceID      TraceID           `json:"trace_id"`
	SpanID       SpanID            `json:"span_id"`
	ParentSpanID *SpanID           `json:"parent_span_id,omitempty"`
	Baggage      map[string]string `json:"baggage,omitempty"`
	Flags        TraceFlags        `json:"flags"`
	Sampled      bool              `json:"sampled"`
}

// Tracer manages trace creation and propagation.
type Tracer struct {
	generator   IDGenerator
	collector   TraceCollector
	sampler     Sampler
	config      TracerConfig
	mu          sync.RWMutex
	activeSpans map[SpanID]*Span
}

// IDGenerator generates unique IDs for traces and spans.
type IDGenerator interface {
	GenerateTraceID() TraceID
	GenerateSpanID() SpanID
}

// RandomIDGenerator generates random IDs using crypto/rand.
type RandomIDGenerator struct{}

// NewRandomIDGenerator creates a new random ID generator.
func NewRandomIDGenerator() *RandomIDGenerator {
	return &RandomIDGenerator{}
}

// GenerateTraceID generates a new trace ID.
func (g *RandomIDGenerator) GenerateTraceID() TraceID {
	var id [16]byte
	_, _ = rand.Read(id[:])
	return TraceID(hex.EncodeToString(id[:]))
}

// GenerateSpanID generates a new span ID.
func (g *RandomIDGenerator) GenerateSpanID() SpanID {
	var id [8]byte
	_, _ = rand.Read(id[:])
	return SpanID(hex.EncodeToString(id[:]))
}

// TraceCollector collects and stores traces.
type TraceCollector interface {
	Collect(trace *Trace)
	GetTrace(traceID TraceID) (*Trace, bool)
	GetTraces(filter TraceFilter) []*Trace
}

// SimpleTraceCollector is a simple in-memory trace collector.
type SimpleTraceCollector struct {
	mu     sync.RWMutex
	traces map[TraceID]*Trace
	maxAge time.Duration
}

// NewSimpleTraceCollector creates a new simple trace collector.
func NewSimpleTraceCollector() *SimpleTraceCollector {
	return &SimpleTraceCollector{
		traces: make(map[TraceID]*Trace),
	}
}

// SetMaxAge configures how long completed traces are retained.
func (c *SimpleTraceCollector) SetMaxAge(maxAge time.Duration) {
	c.mu.Lock()
	c.maxAge = maxAge
	c.pruneLocked(time.Now())
	c.mu.Unlock()
}

// Collect stores a trace.
func (c *SimpleTraceCollector) Collect(trace *Trace) {
	if trace == nil {
		return
	}
	c.mu.Lock()
	c.traces[trace.ID] = copyTrace(trace)
	c.pruneLocked(time.Now())
	c.mu.Unlock()
}

// GetTrace retrieves a trace by ID.
func (c *SimpleTraceCollector) GetTrace(traceID TraceID) (*Trace, bool) {
	c.mu.Lock()
	c.pruneLocked(time.Now())
	trace, exists := c.traces[traceID]
	c.mu.Unlock()
	if !exists {
		return nil, false
	}
	return copyTrace(trace), true
}

// GetTraces retrieves traces matching the filter.
func (c *SimpleTraceCollector) GetTraces(filter TraceFilter) []*Trace {
	c.mu.Lock()
	c.pruneLocked(time.Now())
	traces := make([]*Trace, 0, len(c.traces))
	for _, trace := range c.traces {
		traces = append(traces, copyTrace(trace))
	}
	c.mu.Unlock()

	if filter == nil {
		return traces
	}

	var result []*Trace
	for _, trace := range traces {
		if filter(trace) {
			result = append(result, trace)
		}
	}
	return result
}

func (c *SimpleTraceCollector) pruneLocked(now time.Time) {
	if c.maxAge <= 0 {
		return
	}
	cutoff := now.Add(-c.maxAge)
	for id, trace := range c.traces {
		if trace == nil {
			delete(c.traces, id)
			continue
		}
		timestamp := trace.StartTime
		if trace.EndTime != nil {
			timestamp = *trace.EndTime
		}
		if timestamp.Before(cutoff) {
			delete(c.traces, id)
		}
	}
}

// TraceFilter is a function that filters traces.
type TraceFilter func(trace *Trace) bool

// Sampler determines whether a trace should be sampled.
type Sampler interface {
	ShouldSample(traceID TraceID) bool
}

// ProbabilitySampler samples traces based on probability.
type ProbabilitySampler struct {
	probability float64
}

// NewProbabilitySampler creates a new probability sampler.
func NewProbabilitySampler(probability float64) *ProbabilitySampler {
	return &ProbabilitySampler{probability: probability}
}

// ShouldSample returns true if the trace should be sampled.
func (s *ProbabilitySampler) ShouldSample(traceID TraceID) bool {
	// Use the entire trace ID to calculate a hash
	if len(traceID) == 0 {
		return false
	}

	// Calculate a simple hash of the trace ID
	var hash uint64
	for _, b := range traceID {
		hash = hash*31 + uint64(b)
	}

	// Get a value between 0-99
	value := hash % 100
	threshold := s.probability * 100.0
	result := float64(value) < threshold
	return result
}

// TracerConfig holds configuration for the tracer.
type TracerConfig struct {
	ServiceName      string
	ServiceVersion   string
	Environment      string
	SamplingRate     float64
	MaxSpansPerTrace int
	MaxTraceAge      time.Duration
}

// DefaultTracerConfig returns default tracer configuration.
func DefaultTracerConfig() TracerConfig {
	return TracerConfig{
		ServiceName:      "plumego",
		ServiceVersion:   "1.0.0",
		Environment:      "production",
		SamplingRate:     0.1, // 10% sampling rate
		MaxSpansPerTrace: 1000,
		MaxTraceAge:      24 * time.Hour,
	}
}

// NewTracer creates a new tracer with the given configuration.
func NewTracer(config TracerConfig) *Tracer {
	tracer := &Tracer{
		generator:   NewRandomIDGenerator(),
		collector:   NewSimpleTraceCollector(),
		sampler:     NewProbabilitySampler(config.SamplingRate),
		config:      config,
		activeSpans: make(map[SpanID]*Span),
	}
	if collector, ok := tracer.collector.(*SimpleTraceCollector); ok {
		collector.SetMaxAge(config.MaxTraceAge)
	}
	return tracer
}

// StartTrace starts a new trace.
func (t *Tracer) StartTrace(ctx context.Context, name string, options ...TraceOption) (context.Context, *Span) {
	traceID := t.generator.GenerateTraceID()
	spanID := t.generator.GenerateSpanID()

	trace := &Trace{
		ID:         traceID,
		RootSpanID: spanID,
		StartTime:  time.Now(),
		Spans:      make([]*Span, 0),
		Attributes: make(map[string]any),
	}

	span := &Span{
		ID:        spanID,
		TraceID:   traceID,
		Name:      name,
		Kind:      SpanKindServer,
		StartTime: time.Now(),
		Status: SpanStatus{
			Code: StatusCodeUnset,
		},
		Attributes: make(map[string]any),
	}

	// Apply options
	for _, option := range options {
		option(trace, span)
	}

	// Determine sampling
	spanContext := TraceContext{
		TraceID: traceID,
		SpanID:  spanID,
		Flags:   0,
		Sampled: t.sampler.ShouldSample(traceID),
		Baggage: make(map[string]string),
	}

	if spanContext.Sampled {
		spanContext.Flags |= TraceFlagsSampled
		if t.canAddSpan(trace) {
			// Add the root span to the trace
			trace.Spans = append(trace.Spans, span)
			t.collector.Collect(trace)
		}
	}

	ctx = ContextWithTraceContext(ctx, spanContext)

	t.mu.Lock()
	t.activeSpans[spanID] = span
	t.mu.Unlock()

	return ctx, span
}

// StartChildSpan starts a child span for the given parent span.
func (t *Tracer) StartChildSpan(ctx context.Context, parentSpan *Span, name string, options ...TraceOption) (context.Context, *Span) {
	if parentSpan == nil {
		return t.StartTrace(ctx, name, options...)
	}

	spanID := t.generator.GenerateSpanID()

	span := &Span{
		ID:           spanID,
		ParentSpanID: &parentSpan.ID,
		TraceID:      parentSpan.TraceID,
		Name:         name,
		Kind:         SpanKindInternal,
		StartTime:    time.Now(),
		Status: SpanStatus{
			Code: StatusCodeUnset,
		},
		Attributes: make(map[string]any),
	}

	// Apply options
	for _, option := range options {
		// Options might need to access parent span
		option(&Trace{}, span)
	}

	sampled := false
	flags := TraceFlags(0)
	if parentCtx := TraceContextFromContext(ctx); parentCtx != nil && parentCtx.TraceID == parentSpan.TraceID {
		sampled = parentCtx.Sampled
		flags = parentCtx.Flags
	} else {
		sampled = t.sampler.ShouldSample(parentSpan.TraceID)
	}
	if sampled {
		flags |= TraceFlagsSampled
	} else {
		flags &^= TraceFlagsSampled
	}

	spanContext := TraceContext{
		TraceID:      parentSpan.TraceID,
		SpanID:       spanID,
		ParentSpanID: &parentSpan.ID,
		Flags:        flags,
		Sampled:      sampled,
		Baggage:      make(map[string]string),
	}

	// Add the child span to the existing trace
	if sampled {
		if trace, exists := t.collector.GetTrace(parentSpan.TraceID); exists {
			if t.canAddSpan(trace) {
				trace.Spans = append(trace.Spans, span)
			}
			t.collector.Collect(trace)
		}
	}

	ctx = ContextWithTraceContext(ctx, spanContext)

	t.mu.Lock()
	t.activeSpans[spanID] = span
	t.mu.Unlock()

	return ctx, span
}

// EndSpan ends the given span.
func (t *Tracer) EndSpan(span *Span, options ...SpanOption) {
	if span == nil {
		return
	}
	now := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()

	span.EndTime = &now
	span.Duration = span.EndTime.Sub(span.StartTime)

	// Apply end options
	for _, option := range options {
		option(span)
	}

	delete(t.activeSpans, span.ID)

	// Update the trace if it's sampled
	if trace, exists := t.collector.GetTrace(span.TraceID); exists {
		trace = mergeSpanIntoTrace(trace, span, t.config.MaxSpansPerTrace)

		// Check if this was the root span and if so, mark the trace as completed
		if span.ID == trace.RootSpanID {
			endTime := now
			trace.EndTime = &endTime
		}

		t.collector.Collect(trace)
	}
}

// GetActiveSpan returns the active span from the context.
func (t *Tracer) GetActiveSpan(ctx context.Context) *Span {
	spanContext := TraceContextFromContext(ctx)
	if spanContext == nil {
		return nil
	}

	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.activeSpans[spanContext.SpanID]
}

// RecordSpanEvent records an event for the given span.
func (t *Tracer) RecordSpanEvent(span *Span, name string, options ...EventOption) {
	if span == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	event := SpanEvent{
		Name:       name,
		Timestamp:  time.Now(),
		Attributes: make(map[string]any),
	}

	for _, option := range options {
		option(&event)
	}

	span.Events = append(span.Events, event)
	t.updateCollectedSpan(span)
}

// RecordError records an error for the given span.
func (t *Tracer) RecordError(span *Span, err error, options ...ErrorOption) {
	if span == nil || err == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	span.Status.Code = StatusCodeError
	span.Status.Description = err.Error()

	// Add error details as attributes
	errorAttrs := map[string]any{
		"error.type":    fmt.Sprintf("%T", err),
		"error.message": err.Error(),
		"error.stack":   fmt.Sprintf("%+v", err),
	}

	for _, option := range options {
		option(errorAttrs)
	}

	for k, v := range errorAttrs {
		span.Attributes[k] = v
	}
	t.updateCollectedSpan(span)
}

// TraceOption is a function that configures a trace.
type TraceOption func(trace *Trace, span *Span)

// WithTraceAttributes sets attributes for the trace.
func WithTraceAttributes(attrs map[string]any) TraceOption {
	return func(trace *Trace, span *Span) {
		for k, v := range attrs {
			trace.Attributes[k] = v
		}
	}
}

// WithSpanKind sets the span kind.
func WithSpanKind(kind SpanKind) TraceOption {
	return func(trace *Trace, span *Span) {
		span.Kind = kind
	}
}

// SpanOption is a function that configures a span.
type SpanOption func(span *Span)

// WithSpanAttributes sets attributes for the span.
func WithSpanAttributes(attrs map[string]any) SpanOption {
	return func(span *Span) {
		for k, v := range attrs {
			span.Attributes[k] = v
		}
	}
}

// WithSpanStatus sets the span status.
func WithSpanStatus(status SpanStatus) SpanOption {
	return func(span *Span) {
		span.Status = status
	}
}

// EventOption is a function that configures an event.
type EventOption func(event *SpanEvent)

// WithEventAttributes sets attributes for the event.
func WithEventAttributes(attrs map[string]any) EventOption {
	return func(event *SpanEvent) {
		for k, v := range attrs {
			event.Attributes[k] = v
		}
	}
}

// ErrorOption is a function that configures error recording.
type ErrorOption func(attrs map[string]any)

// WithErrorAttributes sets additional attributes for the error.
func WithErrorAttributes(attrs map[string]any) ErrorOption {
	return func(errorAttrs map[string]any) {
		for k, v := range attrs {
			errorAttrs[k] = v
		}
	}
}

// Context management functions.

type traceContextKey struct{}

var traceContextKeyVar traceContextKey

// ContextWithTraceContext adds trace context to the context.
func ContextWithTraceContext(ctx context.Context, traceContext TraceContext) context.Context {
	return context.WithValue(ctx, traceContextKeyVar, &traceContext)
}

// TraceContextFromContext retrieves trace context from the context.
func TraceContextFromContext(ctx context.Context) *TraceContext {
	if v := ctx.Value(traceContextKeyVar); v != nil {
		if tc, ok := v.(*TraceContext); ok {
			return tc
		}
	}
	return nil
}

// Legacy TraceIDKey for backward compatibility
type TraceIDKey struct{}

// TraceIDFromContext extracts the trace id injected by the Logging middleware.
func TraceIDFromContext(ctx context.Context) string {
	if tc := TraceContextFromContext(ctx); tc != nil {
		return string(tc.TraceID)
	}
	if v, ok := ctx.Value(TraceIDKey{}).(string); ok {
		return v
	}
	return ""
}

// Utility functions

// ParseTraceID parses a trace ID from string.
func ParseTraceID(id string) (TraceID, error) {
	if len(id) != 32 {
		return "", fmt.Errorf("invalid trace ID length: expected 32, got %d", len(id))
	}

	// Validate hex string
	if _, err := hex.DecodeString(id); err != nil {
		return "", fmt.Errorf("invalid trace ID format: %v", err)
	}

	return TraceID(id), nil
}

// ParseSpanID parses a span ID from string.
func ParseSpanID(id string) (SpanID, error) {
	if len(id) != 16 {
		return "", fmt.Errorf("invalid span ID length: expected 16, got %d", len(id))
	}

	// Validate hex string
	if _, err := hex.DecodeString(id); err != nil {
		return "", fmt.Errorf("invalid span ID format: %v", err)
	}

	return SpanID(id), nil
}

// IsValidTraceID checks if a trace ID is valid.
func IsValidTraceID(traceID TraceID) bool {
	_, err := ParseTraceID(string(traceID))
	return err == nil
}

// IsValidSpanID checks if a span ID is valid.
func IsValidSpanID(spanID SpanID) bool {
	_, err := ParseSpanID(string(spanID))
	return err == nil
}

// TraceStatistics provides statistics about traces.
type TraceStatistics struct {
	TotalTraces     int64                `json:"total_traces"`
	ActiveTraces    int64                `json:"active_traces"`
	CompletedTraces int64                `json:"completed_traces"`
	ByStatus        map[StatusCode]int64 `json:"by_status"`
	AverageDuration time.Duration        `json:"average_duration"`
	TotalSpans      int64                `json:"total_spans"`
	LastTraceTime   time.Time            `json:"last_trace_time"`
}

// GetTraceStatistics returns statistics about collected traces.
func (t *Tracer) GetTraceStatistics() TraceStatistics {
	stats := TraceStatistics{
		ByStatus: make(map[StatusCode]int64),
	}

	traces := t.collector.GetTraces(func(trace *Trace) bool {
		return true // Include all traces
	})

	var totalDuration time.Duration
	var completedCount int64

	for _, trace := range traces {
		stats.TotalTraces++
		stats.TotalSpans += int64(len(trace.Spans))

		if trace.EndTime != nil {
			completedCount++
			totalDuration += trace.EndTime.Sub(trace.StartTime)
		} else {
			stats.ActiveTraces++
		}

		// Count by status
		for _, span := range trace.Spans {
			stats.ByStatus[span.Status.Code]++
		}

		if trace.StartTime.After(stats.LastTraceTime) {
			stats.LastTraceTime = trace.StartTime
		}
	}

	stats.CompletedTraces = completedCount
	if completedCount > 0 {
		stats.AverageDuration = totalDuration / time.Duration(completedCount)
	}

	return stats
}

func (t *Tracer) updateCollectedSpan(span *Span) {
	if span == nil {
		return
	}
	trace, exists := t.collector.GetTrace(span.TraceID)
	if !exists {
		return
	}
	trace = mergeSpanIntoTrace(trace, span, t.config.MaxSpansPerTrace)
	t.collector.Collect(trace)
}

func (t *Tracer) canAddSpan(trace *Trace) bool {
	if trace == nil {
		return false
	}
	if t.config.MaxSpansPerTrace <= 0 {
		return true
	}
	return len(trace.Spans) < t.config.MaxSpansPerTrace
}

func mergeSpanIntoTrace(trace *Trace, span *Span, maxSpans int) *Trace {
	if trace == nil || span == nil {
		return trace
	}
	for i, existingSpan := range trace.Spans {
		if existingSpan.ID == span.ID {
			trace.Spans[i] = copySpan(span)
			return trace
		}
	}
	if maxSpans > 0 && len(trace.Spans) >= maxSpans {
		return trace
	}
	trace.Spans = append(trace.Spans, copySpan(span))
	return trace
}

func copyTrace(src *Trace) *Trace {
	if src == nil {
		return nil
	}

	var endTime *time.Time
	if src.EndTime != nil {
		val := *src.EndTime
		endTime = &val
	}

	trace := &Trace{
		ID:         src.ID,
		RootSpanID: src.RootSpanID,
		StartTime:  src.StartTime,
		EndTime:    endTime,
		Attributes: copyAttributes(src.Attributes),
		Links:      copyTraceLinks(src.Links),
	}

	if len(src.Spans) > 0 {
		trace.Spans = make([]*Span, len(src.Spans))
		for i, span := range src.Spans {
			trace.Spans[i] = copySpan(span)
		}
	}

	return trace
}

func copySpan(src *Span) *Span {
	if src == nil {
		return nil
	}

	var parent *SpanID
	if src.ParentSpanID != nil {
		val := *src.ParentSpanID
		parent = &val
	}

	span := &Span{
		ID:           src.ID,
		ParentSpanID: parent,
		TraceID:      src.TraceID,
		Name:         src.Name,
		Kind:         src.Kind,
		StartTime:    src.StartTime,
		EndTime:      copyTimePtr(src.EndTime),
		Duration:     src.Duration,
		Status:       src.Status,
		Attributes:   copyAttributes(src.Attributes),
		Links:        copySpanLinks(src.Links),
	}

	if len(src.Events) > 0 {
		span.Events = make([]SpanEvent, len(src.Events))
		for i, event := range src.Events {
			span.Events[i] = copySpanEvent(event)
		}
	}

	return span
}

func copyTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	val := *value
	return &val
}

func copySpanEvent(src SpanEvent) SpanEvent {
	return SpanEvent{
		Name:       src.Name,
		Timestamp:  src.Timestamp,
		Attributes: copyAttributes(src.Attributes),
	}
}

func copySpanLinks(links []SpanLink) []SpanLink {
	if len(links) == 0 {
		return nil
	}
	out := make([]SpanLink, len(links))
	for i, link := range links {
		out[i] = SpanLink{
			TraceID:    link.TraceID,
			SpanID:     link.SpanID,
			Attributes: copyAttributes(link.Attributes),
		}
	}
	return out
}

func copyTraceLinks(links []TraceLink) []TraceLink {
	if len(links) == 0 {
		return nil
	}
	out := make([]TraceLink, len(links))
	for i, link := range links {
		out[i] = TraceLink{
			TraceID:    link.TraceID,
			Attributes: copyAttributes(link.Attributes),
		}
	}
	return out
}

func copyAttributes(attrs map[string]any) map[string]any {
	if attrs == nil {
		return nil
	}
	out := make(map[string]any, len(attrs))
	for k, v := range attrs {
		out[k] = v
	}
	return out
}
