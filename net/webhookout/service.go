package webhookout

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"crypto/rand"
	"encoding/hex"
)

type Metrics struct {
	Enqueued uint64
	Dropped  uint64
	SentOK   uint64
	SentFail uint64
	Retried  uint64
	Dead     uint64
}

type Service struct {
	cfg   Config
	store Store

	queue *Queue
	ds    *DelayScheduler

	http *http.Client

	wg      sync.WaitGroup
	cancel  context.CancelFunc
	running atomic.Bool

	metrics Metrics
}

func NewService(store Store, cfg Config) *Service {
	q := NewQueue(cfg.QueueSize, cfg.DropPolicy, cfg.BlockWait)
	s := &Service{
		cfg:   cfg,
		store: store,
		queue: q,
		http:  NewHTTPClient(cfg.DefaultTimeout),
	}
	s.ds = NewDelayScheduler(q)
	return s
}

func (s *Service) Start(ctx context.Context) {
	if !s.cfg.Enabled {
		return
	}
	if s.running.Swap(true) {
		return
	}
	cctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	// delay scheduler
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.ds.Run(cctx)
	}()

	// workers
	for i := 0; i < s.cfg.Workers; i++ {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.workerLoop(cctx)
		}()
	}
}

func (s *Service) Stop() {
	if !s.running.Swap(false) {
		return
	}
	if s.cancel != nil {
		s.cancel()
	}
	s.ds.Stop()

	// best-effort drain: wait up to DrainMax
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(s.cfg.DrainMax):
	}
}

func (s *Service) Metrics() Metrics { return s.metrics }

func (s *Service) CreateTarget(ctx context.Context, t Target) (Target, error) {
	if err := validateTarget(t); err != nil {
		return Target{}, err
	}
	t.ID = newID()
	if t.Headers == nil {
		t.Headers = map[string]string{}
	}
	if t.Enabled == false {
		// ok
	}
	// defaults
	if t.TimeoutMs == 0 {
		t.TimeoutMs = int(s.cfg.DefaultTimeout / time.Millisecond)
	}
	if t.MaxRetries == 0 {
		t.MaxRetries = s.cfg.DefaultMaxRetries
	}
	if t.BackoffBaseMs == 0 {
		t.BackoffBaseMs = int(s.cfg.BackoffBase / time.Millisecond)
	}
	if t.BackoffMaxMs == 0 {
		t.BackoffMaxMs = int(s.cfg.BackoffMax / time.Millisecond)
	}
	return s.store.CreateTarget(ctx, t)
}

func (s *Service) UpdateTarget(ctx context.Context, id string, patch TargetPatch) (Target, error) {
	// validate URL if set
	if patch.URL != nil {
		if err := validateURL(*patch.URL, s.cfg.AllowPrivateNetwork); err != nil {
			return Target{}, err
		}
	}
	if patch.Events != nil && len(*patch.Events) == 0 {
		return Target{}, errors.New("events cannot be empty")
	}
	if patch.Secret != nil && len(strings.TrimSpace(*patch.Secret)) < 8 {
		return Target{}, errors.New("secret too short")
	}
	return s.store.UpdateTarget(ctx, id, patch)
}

func (s *Service) GetTarget(ctx context.Context, id string) (Target, bool) {
	return s.store.GetTarget(ctx, id)
}

func (s *Service) ListTargets(ctx context.Context, f TargetFilter) ([]Target, error) {
	return s.store.ListTargets(ctx, f)
}

func (s *Service) GetDelivery(ctx context.Context, id string) (Delivery, bool) {
	return s.store.GetDelivery(ctx, id)
}

func (s *Service) ListDeliveries(ctx context.Context, f DeliveryFilter) ([]Delivery, error) {
	return s.store.ListDeliveries(ctx, f)
}
func (s *Service) ReplayDelivery(ctx context.Context, deliveryID string) (Delivery, error) {
	d, ok := s.store.GetDelivery(ctx, deliveryID)
	if !ok {
		return Delivery{}, ErrNotFound
	}
	if len(d.PayloadJSON) == 0 {
		return Delivery{}, errors.New("delivery has no payload_json")
	}

	newID := newID()

	// Optional: keep the original payload verbatim, but update meta.delivery_id for the new delivery.
	// Recommended: update delivery_id to keep the receiver's idempotency key correct.
	raw := d.PayloadJSON
	raw2, err := rewriteDeliveryIDInPayload(raw, newID)
	if err != nil {
		// fallback to raw as-is
		raw2 = raw
	}

	newD := Delivery{
		ID:          newID,
		TargetID:    d.TargetID,
		EventID:     d.EventID,
		EventType:   d.EventType,
		PayloadJSON: raw2,
		Attempt:     0,
		Status:      DeliveryPending,
	}
	newD, err = s.store.CreateDelivery(ctx, newD)
	if err != nil {
		return Delivery{}, err
	}
	if err := s.enqueue(ctx, Task{DeliveryID: newD.ID}); err != nil {
		return Delivery{}, err
	}
	return newD, nil
}

func rewriteDeliveryIDInPayload(raw []byte, newDeliveryID string) ([]byte, error) {
	// payload schema:
	// { id, type, occurred_at, data, meta: { delivery_id, ... } }
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	meta, ok := m["meta"].(map[string]any)
	if !ok || meta == nil {
		meta = map[string]any{}
	}
	meta["delivery_id"] = newDeliveryID
	m["meta"] = meta
	return json.Marshal(m)
}

func (s *Service) TriggerEvent(ctx context.Context, e Event) (int, error) {
	if strings.TrimSpace(e.Type) == "" {
		return 0, errors.New("event type required")
	}
	if e.Data == nil {
		e.Data = map[string]any{}
	}
	if e.Meta == nil {
		e.Meta = map[string]any{}
	}

	e.ID = newID()
	e.OccurredAt = time.Now().UTC()

	targets, err := s.store.ListTargets(ctx, TargetFilter{Enabled: boolPtr(true), Event: e.Type})
	if err != nil {
		return 0, err
	}

	count := 0
	for _, t := range targets {
		// Delivery ID first, so payload can include it
		deliveryID := newID()

		payload := map[string]any{
			"id":          e.ID,
			"type":        e.Type,
			"occurred_at": e.OccurredAt.Format(time.RFC3339Nano),
			"data":        e.Data,
			"meta": mergeMeta(e.Meta, map[string]any{
				"delivery_id": deliveryID,
				"target_id":   t.ID,
			}),
		}
		raw, _ := json.Marshal(payload)

		d := Delivery{
			ID:          deliveryID,
			TargetID:    t.ID,
			EventID:     e.ID,
			EventType:   e.Type,
			PayloadJSON: raw,
			Attempt:     0,
			Status:      DeliveryPending,
		}

		if _, err := s.store.CreateDelivery(ctx, d); err != nil {
			continue
		}
		if err := s.enqueue(ctx, Task{DeliveryID: d.ID}); err != nil {
			atomic.AddUint64(&s.metrics.Dropped, 1)
			continue
		}
		count++
	}
	return count, nil
}

func mergeMeta(a, b map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func (s *Service) enqueue(ctx context.Context, t Task) error {
	if !s.cfg.Enabled {
		return errors.New("webhook disabled")
	}
	err := s.queue.Enqueue(ctx, t)
	if err == nil {
		atomic.AddUint64(&s.metrics.Enqueued, 1)
	}
	return err
}

func (s *Service) workerLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case t, ok := <-s.queue.Chan():
			if !ok {
				return
			}
			s.handleTask(ctx, t)
		}
	}
}

func (s *Service) handleTask(ctx context.Context, t Task) {
	d, ok := s.store.GetDelivery(ctx, t.DeliveryID)
	if !ok {
		return
	}
	target, ok := s.store.GetTarget(ctx, d.TargetID)
	if !ok || !target.Enabled {
		return
	}

	// payload must exist for replayable delivery
	raw := d.PayloadJSON
	if len(raw) == 0 {
		// mark failed permanently (data corruption)
		attempt := d.Attempt + 1
		msg := "missing payload_json"
		st := DeliveryDead
		patch := DeliveryPatch{
			Attempt:   &attempt,
			Status:    &st,
			LastError: &msg,
		}
		_, _ = s.store.UpdateDelivery(ctx, d.ID, patch)
		atomic.AddUint64(&s.metrics.Dead, 1)
		return
	}

	attempt := d.Attempt + 1

	start := time.Now()
	status, snippet, err := s.sendOnce(ctx, target, d, raw, attempt)
	durMs := int(time.Since(start) / time.Millisecond)

	patch := DeliveryPatch{
		Attempt:        &attempt,
		LastHTTPStatus: &status,
		LastDurationMs: &durMs,
	}
	if snippet != "" {
		patch.LastRespSnippet = &snippet
	}
	if err != nil {
		msg := err.Error()
		patch.LastError = &msg
	}

	maxRetries := target.MaxRetries
	if maxRetries <= 0 {
		maxRetries = s.cfg.DefaultMaxRetries
	}
	retryOn429 := s.cfg.RetryOn429
	if target.RetryOn429 != nil {
		retryOn429 = *target.RetryOn429
	}

	if ShouldRetry(status, err, attempt, maxRetries, retryOn429) {
		atomic.AddUint64(&s.metrics.Retried, 1)
		nextAt := NextBackoff(
			time.Now().UTC(),
			attempt,
			ms(target.BackoffBaseMs, s.cfg.BackoffBase),
			ms(target.BackoffMaxMs, s.cfg.BackoffMax),
		)
		st := DeliveryRetry
		patch.Status = &st
		patch.NextAt = &nextAt
		_, _ = s.store.UpdateDelivery(ctx, d.ID, patch)
		s.ds.Schedule(nextAt, Task{DeliveryID: d.ID})
		return
	}

	if err == nil && status >= 200 && status <= 299 {
		atomic.AddUint64(&s.metrics.SentOK, 1)
		st := DeliverySuccess
		patch.Status = &st
		patch.NextAt = nil
		_, _ = s.store.UpdateDelivery(ctx, d.ID, patch)
		return
	}

	atomic.AddUint64(&s.metrics.SentFail, 1)
	st := DeliveryFailed
	if attempt >= maxRetries {
		st = DeliveryDead
		atomic.AddUint64(&s.metrics.Dead, 1)
	}
	patch.Status = &st
	patch.NextAt = nil
	_, _ = s.store.UpdateDelivery(ctx, d.ID, patch)
}

func (s *Service) sendOnce(ctx context.Context, target Target, d Delivery, raw []byte, attempt int) (int, string, error) {
	// SSRF check
	if err := validateURL(target.URL, s.cfg.AllowPrivateNetwork); err != nil {
		return 0, "", err
	}

	timeout := ms(target.TimeoutMs, s.cfg.DefaultTimeout)
	client := NewHTTPClient(timeout)

	req, err := http.NewRequestWithContext(ctx, "POST", target.URL, bytes.NewReader(raw))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")

	// custom headers
	for k, v := range target.Headers {
		if strings.TrimSpace(k) == "" {
			continue
		}
		req.Header.Set(k, v)
	}

	// signature headers
	ts := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	sig := SignV1(target.Secret, ts, raw)
	req.Header.Set("X-Webhook-Id", target.ID)
	req.Header.Set("X-Webhook-Event", d.EventType)
	req.Header.Set("X-Webhook-Delivery", d.ID)
	req.Header.Set("X-Webhook-Timestamp", ts)
	req.Header.Set("X-Webhook-Signature", "v1="+sig)
	req.Header.Set("X-Webhook-Attempt", strconv.Itoa(attempt))

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	// read small snippet only
	const maxRead = 4096
	b, _ := io.ReadAll(io.LimitReader(resp.Body, maxRead))
	snippet := strings.TrimSpace(string(b))
	if len(snippet) > 1024 {
		snippet = snippet[:1024]
	}
	return resp.StatusCode, snippet, nil
}

func validateTarget(t Target) error {
	if strings.TrimSpace(t.Name) == "" {
		return errors.New("name required")
	}
	if len(strings.TrimSpace(t.Secret)) < 8 {
		return errors.New("secret too short")
	}
	if len(t.Events) == 0 {
		return errors.New("events required")
	}
	return validateURL(t.URL, false) // strict by default; allow override at service-level update
}

func validateURL(raw string, allowPrivate bool) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return errors.New("invalid url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("url scheme must be http/https")
	}
	if u.Host == "" {
		return errors.New("url host required")
	}
	host := u.Hostname()
	if host == "" {
		return errors.New("url hostname required")
	}
	if allowPrivate {
		return nil
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		// if DNS fails, reject to be safe
		return errors.New("dns lookup failed")
	}
	for _, ip := range ips {
		if isPrivateIP(ip) {
			return errors.New("private network address not allowed")
		}
	}
	return nil
}

func isPrivateIP(ip net.IP) bool {
	// IPv4 private ranges + loopback + link-local
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	ip4 := ip.To4()
	if ip4 != nil {
		switch {
		case ip4[0] == 10:
			return true
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return true
		case ip4[0] == 192 && ip4[1] == 168:
			return true
		case ip4[0] == 169 && ip4[1] == 254:
			return true
		}
		return false
	}
	// IPv6 unique local: fc00::/7
	if len(ip) == 16 && (ip[0]&0xfe) == 0xfc {
		return true
	}
	return false
}

func ms(v int, def time.Duration) time.Duration {
	if v <= 0 {
		return def
	}
	return time.Duration(v) * time.Millisecond
}

func boolPtr(v bool) *bool { return &v }

func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
