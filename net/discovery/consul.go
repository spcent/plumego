package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Consul implements Discovery using HashiCorp Consul
//
// Example:
//
//	sd, err := discovery.NewConsul("localhost:8500", discovery.ConsulConfig{
//		Datacenter: "dc1",
//		Token:      "secret-token",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer sd.Close()
//
//	backends, err := sd.Resolve(ctx, "user-service")
type Consul struct {
	address  string
	config   ConsulConfig
	client   *http.Client
	watchers map[string]*consulWatcher
	mu       sync.RWMutex
}

// ConsulConfig holds Consul-specific configuration
type ConsulConfig struct {
	// Datacenter to query (default: "dc1")
	Datacenter string

	// Token for ACL authentication
	Token string

	// Namespace for Consul Enterprise (default: "default")
	Namespace string

	// WaitTime for long polling (default: 30s)
	WaitTime time.Duration

	// Timeout for HTTP requests (default: 60s)
	Timeout time.Duration

	// OnlyHealthy filters to only return healthy instances (default: true)
	OnlyHealthy bool

	// Tag filters services by tag
	Tag string

	// Scheme is the URL scheme (default: "http")
	Scheme string
}

// NewConsul creates a new Consul discovery client
func NewConsul(address string, config ConsulConfig) (*Consul, error) {
	// Apply defaults
	if config.Datacenter == "" {
		config.Datacenter = "dc1"
	}
	if config.WaitTime == 0 {
		config.WaitTime = 30 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.Scheme == "" {
		config.Scheme = "http"
	}
	config.OnlyHealthy = true // Default to only healthy

	return &Consul{
		address:  address,
		config:   config,
		client:   &http.Client{Timeout: config.Timeout},
		watchers: make(map[string]*consulWatcher),
	}, nil
}

// Resolve returns the list of backend URLs for a service
func (c *Consul) Resolve(ctx context.Context, serviceName string) ([]string, error) {
	// Build query URL
	queryURL := fmt.Sprintf("%s://%s/v1/health/service/%s",
		c.config.Scheme, c.address, serviceName)

	params := url.Values{}
	params.Set("dc", c.config.Datacenter)
	if c.config.OnlyHealthy {
		params.Set("passing", "true")
	}
	if c.config.Tag != "" {
		params.Set("tag", c.config.Tag)
	}

	queryURL += "?" + params.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	// Add ACL token if configured
	if c.config.Token != "" {
		req.Header.Set("X-Consul-Token", c.config.Token)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrServiceNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("consul returned status %d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var entries []consulServiceEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, ErrNoInstances
	}

	// Build backend URLs
	backends := make([]string, 0, len(entries))
	for _, entry := range entries {
		backend := fmt.Sprintf("http://%s:%d",
			entry.Service.Address, entry.Service.Port)
		backends = append(backends, backend)
	}

	return backends, nil
}

// Watch returns a channel for backend updates
func (c *Consul) Watch(ctx context.Context, serviceName string) (<-chan []string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if watcher already exists
	if watcher, exists := c.watchers[serviceName]; exists {
		// Subscribe to existing watcher
		return watcher.subscribe(ctx), nil
	}

	// Create new watcher
	watcher := newConsulWatcher(c, serviceName)
	c.watchers[serviceName] = watcher

	// Start watching
	go watcher.watch()

	// Subscribe to watcher
	return watcher.subscribe(ctx), nil
}

// Register registers a service instance
func (c *Consul) Register(ctx context.Context, instance Instance) error {
	// Build registration payload
	reg := consulRegistration{
		ID:      instance.ID,
		Name:    instance.Name,
		Address: instance.Address,
		Port:    instance.Port,
		Tags:    instance.Tags,
		Meta:    instance.Metadata,
	}

	// Note: Health checks can be added via Consul's agent API separately
	// This simple registration doesn't include health checks
	// Users can configure health checks in Consul directly

	// Marshal to JSON
	body, err := json.Marshal(reg)
	if err != nil {
		return err
	}

	// Build URL
	registerURL := fmt.Sprintf("%s://%s/v1/agent/service/register",
		c.config.Scheme, c.address)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "PUT", registerURL,
		io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.Token != "" {
		req.Header.Set("X-Consul-Token", c.config.Token)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("consul returned status %d", resp.StatusCode)
	}

	return nil
}

// Deregister removes a service instance
func (c *Consul) Deregister(ctx context.Context, serviceID string) error {
	deregisterURL := fmt.Sprintf("%s://%s/v1/agent/service/deregister/%s",
		c.config.Scheme, c.address, serviceID)

	req, err := http.NewRequestWithContext(ctx, "PUT", deregisterURL, nil)
	if err != nil {
		return err
	}

	if c.config.Token != "" {
		req.Header.Set("X-Consul-Token", c.config.Token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("consul returned status %d", resp.StatusCode)
	}

	return nil
}

// Health updates the health status of an instance
func (c *Consul) Health(ctx context.Context, serviceID string, healthy bool) error {
	// Consul uses TTL-based health checks
	// This would require registering with a TTL check first
	return ErrNotSupported
}

// Close closes the Consul client
func (c *Consul) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Stop all watchers
	for _, watcher := range c.watchers {
		watcher.stop()
	}

	c.watchers = make(map[string]*consulWatcher)

	return nil
}

// consulServiceEntry represents a Consul service entry
type consulServiceEntry struct {
	Service consulService `json:"Service"`
}

// consulService represents a Consul service
type consulService struct {
	Address string `json:"Address"`
	Port    int    `json:"Port"`
}

// consulRegistration represents a Consul service registration
type consulRegistration struct {
	ID      string            `json:"ID"`
	Name    string            `json:"Name"`
	Address string            `json:"Address"`
	Port    int               `json:"Port"`
	Tags    []string          `json:"Tags,omitempty"`
	Meta    map[string]string `json:"Meta,omitempty"`
	Check   *consulCheck      `json:"Check,omitempty"`
}

// consulCheck represents a Consul health check
type consulCheck struct {
	HTTP     string `json:"HTTP"`
	Interval string `json:"Interval"`
	Timeout  string `json:"Timeout"`
}

// consulWatcher watches a service for changes
type consulWatcher struct {
	consul      *Consul
	serviceName string
	subscribers []chan []string
	stopChan    chan struct{}
	mu          sync.RWMutex
}

// newConsulWatcher creates a new watcher
func newConsulWatcher(consul *Consul, serviceName string) *consulWatcher {
	return &consulWatcher{
		consul:      consul,
		serviceName: serviceName,
		subscribers: make([]chan []string, 0),
		stopChan:    make(chan struct{}),
	}
}

// subscribe subscribes to watcher updates
func (w *consulWatcher) subscribe(ctx context.Context) <-chan []string {
	ch := make(chan []string, 1)

	w.mu.Lock()
	w.subscribers = append(w.subscribers, ch)
	w.mu.Unlock()

	// Remove subscriber when context is done
	go func() {
		<-ctx.Done()
		w.mu.Lock()
		defer w.mu.Unlock()

		for i, sub := range w.subscribers {
			if sub == ch {
				w.subscribers = append(w.subscribers[:i], w.subscribers[i+1:]...)
				close(ch)
				break
			}
		}
	}()

	return ch
}

// watch watches for service changes
func (w *consulWatcher) watch() {
	ticker := time.NewTicker(w.consul.config.WaitTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Query for updates
			ctx, cancel := context.WithTimeout(context.Background(), w.consul.config.Timeout)
			backends, err := w.consul.Resolve(ctx, w.serviceName)
			cancel()

			if err == nil && len(backends) > 0 {
				w.notify(backends)
			}

		case <-w.stopChan:
			return
		}
	}
}

// notify sends backends to all subscribers
func (w *consulWatcher) notify(backends []string) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for _, ch := range w.subscribers {
		select {
		case ch <- backends:
		default:
			// Skip if channel is full
		}
	}
}

// stop stops the watcher
func (w *consulWatcher) stop() {
	close(w.stopChan)
}
