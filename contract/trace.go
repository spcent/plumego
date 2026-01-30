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

	// TraceIDLength is the expected length of a trace ID in hexadecimal format (32 hex chars = 16 bytes).
	TraceIDLength = 32

	// SpanIDLength is the expected length of a span ID in hexadecimal format (16 hex chars = 8 bytes).
	SpanIDLength = 16
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

// Clone returns a deep copy of the Trace.
// NOTE: Update this method when Trace struct fields change.
func (t *Trace) Clone() *Trace {
	if t == nil {
		return nil
	}

	trace := &Trace{
		ID:         t.ID,
		RootSpanID: t.RootSpanID,
		StartTime:  t.StartTime,
		EndTime:    cloneTimePtr(t.EndTime),
		Attributes: cloneAttributes(t.Attributes),
	}

	if len(t.Spans) > 0 {
		trace.Spans = make([]*Span, len(t.Spans))
		for i, span := range t.Spans {
			trace.Spans[i] = span.Clone()
		}
	}

	if len(t.Links) > 0 {
		trace.Links = make([]TraceLink, len(t.Links))
		for i, link := range t.Links {
			trace.Links[i] = link.Clone()
		}
	}

	return trace
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

// Clone returns a deep copy of the Span.
// NOTE: Update this method when Span struct fields change.
func (s *Span) Clone() *Span {
	if s == nil {
		return nil
	}

	span := &Span{
		ID:           s.ID,
		ParentSpanID: cloneSpanIDPtr(s.ParentSpanID),
		TraceID:      s.TraceID,
		Name:         s.Name,
		Kind:         s.Kind,
		StartTime:    s.StartTime,
		EndTime:      cloneTimePtr(s.EndTime),
		Duration:     s.Duration,
		Status:       s.Status,
		Attributes:   cloneAttributes(s.Attributes),
	}

	if len(s.Events) > 0 {
		span.Events = make([]SpanEvent, len(s.Events))
		for i, event := range s.Events {
			span.Events[i] = event.Clone()
		}
	}

	if len(s.Links) > 0 {
		span.Links = make([]SpanLink, len(s.Links))
		for i, link := range s.Links {
			span.Links[i] = link.Clone()
		}
	}

	return span
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

// Clone returns a deep copy of the SpanEvent.
// NOTE: Update this method when SpanEvent struct fields change.
func (e SpanEvent) Clone() SpanEvent {
	return SpanEvent{
		Name:       e.Name,
		Timestamp:  e.Timestamp,
		Attributes: cloneAttributes(e.Attributes),
	}
}

// SpanLink represents a link to another span.
type SpanLink struct {
	TraceID    TraceID        `json:"trace_id"`
	SpanID     SpanID         `json:"span_id"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// Clone returns a deep copy of the SpanLink.
// NOTE: Update this method when SpanLink struct fields change.
func (l SpanLink) Clone() SpanLink {
	return SpanLink{
		TraceID:    l.TraceID,
		SpanID:     l.SpanID,
		Attributes: cloneAttributes(l.Attributes),
	}
}

// TraceLink represents a link between traces.
type TraceLink struct {
	TraceID    TraceID        `json:"trace_id"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// Clone returns a deep copy of the TraceLink.
// NOTE: Update this method when TraceLink struct fields change.
func (l TraceLink) Clone() TraceLink {
	return TraceLink{
		TraceID:    l.TraceID,
		Attributes: cloneAttributes(l.Attributes),
	}
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
	mu            sync.RWMutex
	traces        map[TraceID]*Trace
	maxAge        time.Duration
	lastPruneTime time.Time
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
	c.lastPruneTime = time.Time{} // Reset to force next prune
	c.pruneLocked(time.Now())
	c.mu.Unlock()
}

// Collect stores a trace.
func (c *SimpleTraceCollector) Collect(trace *Trace) {
	if trace == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	// Don't store already-expired traces
	if c.maxAge > 0 {
		now := time.Now()
		cutoff := now.Add(-c.maxAge)
		timestamp := trace.StartTime
		if trace.EndTime != nil {
			timestamp = *trace.EndTime
		}
		if timestamp.Before(cutoff) {
			return // Skip storing expired trace
		}
	}

	c.traces[trace.ID] = trace.Clone()
	c.pruneLocked(time.Now())
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
	return trace.Clone(), true
}

// GetTraces retrieves traces matching the filter.
func (c *SimpleTraceCollector) GetTraces(filter TraceFilter) []*Trace {
	c.mu.Lock()
	c.pruneLocked(time.Now())
	traces := make([]*Trace, 0, len(c.traces))
	for _, trace := range c.traces {
		traces = append(traces, trace.Clone())
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

	// Only prune if enough time has passed since last prune
	// Avoid pruning on every operation
	pruneInterval := c.maxAge / 10
	if pruneInterval < 10*time.Second {
		pruneInterval = 10 * time.Second
	}
	if pruneInterval > time.Minute {
		pruneInterval = time.Minute
	}

	if now.Sub(c.lastPruneTime) < pruneInterval {
		return
	}
	c.lastPruneTime = now

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
	if len(id) != TraceIDLength {
		return "", fmt.Errorf("invalid trace ID length: expected %d, got %d", TraceIDLength, len(id))
	}

	// Validate hex string
	if _, err := hex.DecodeString(id); err != nil {
		return "", fmt.Errorf("invalid trace ID format: %v", err)
	}

	return TraceID(id), nil
}

// ParseSpanID parses a span ID from string.
func ParseSpanID(id string) (SpanID, error) {
	if len(id) != SpanIDLength {
		return "", fmt.Errorf("invalid span ID length: expected %d, got %d", SpanIDLength, len(id))
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

	// Check if we need to update or add
	spanIndex := -1
	for i, existingSpan := range trace.Spans {
		if existingSpan.ID == span.ID {
			spanIndex = i
			break
		}
	}

	// If span exists, create shallow copy and update the specific span
	if spanIndex >= 0 {
		newTrace := &Trace{
			ID:         trace.ID,
			RootSpanID: trace.RootSpanID,
			StartTime:  trace.StartTime,
			EndTime:    trace.EndTime,
			Attributes: trace.Attributes,
			Links:      trace.Links,
			Spans:      make([]*Span, len(trace.Spans)),
		}
		copy(newTrace.Spans, trace.Spans)
		newTrace.Spans[spanIndex] = span.Clone()
		return newTrace
	}

	// Add new span if not found
	if maxSpans > 0 && len(trace.Spans) >= maxSpans {
		return trace
	}

	// Create new trace with additional span
	newTrace := &Trace{
		ID:         trace.ID,
		RootSpanID: trace.RootSpanID,
		StartTime:  trace.StartTime,
		EndTime:    trace.EndTime,
		Attributes: trace.Attributes,
		Links:      trace.Links,
		Spans:      make([]*Span, len(trace.Spans)+1),
	}
	copy(newTrace.Spans, trace.Spans)
	newTrace.Spans[len(trace.Spans)] = span.Clone()
	return newTrace
}

// cloneTimePtr returns a deep copy of a time pointer.
func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	val := *value
	return &val
}

// cloneSpanIDPtr returns a deep copy of a SpanID pointer.
func cloneSpanIDPtr(value *SpanID) *SpanID {
	if value == nil {
		return nil
	}
	val := *value
	return &val
}

// cloneAttributes returns a shallow copy of the attributes map.
func cloneAttributes(attrs map[string]any) map[string]any {
	if attrs == nil {
		return nil
	}
	out := make(map[string]any, len(attrs))
	for k, v := range attrs {
		out[k] = v
	}
	return out
}
