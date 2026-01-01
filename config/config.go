package config

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Source represents a configuration source that can provide and watch for changes
type Source interface {
	// Load loads all configuration from the source
	Load(ctx context.Context) (map[string]any, error)
	// Watch watches for configuration changes
	Watch(ctx context.Context) (<-chan map[string]any, <-chan error)
	// Name returns the source name
	Name() string
}

// Config represents the main configuration manager
type Config struct {
	sources  []Source
	data     map[string]any
	mu       sync.RWMutex
	watchers map[string][]chan map[string]any
	ctx      context.Context
	cancel   context.CancelFunc
}

// New creates a new Config instance
func New() *Config {
	ctx, cancel := context.WithCancel(context.Background())
	return &Config{
		sources:  make([]Source, 0),
		data:     make(map[string]any),
		watchers: make(map[string][]chan map[string]any),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// AddSource adds a configuration source
func (c *Config) AddSource(source Source) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sources = append(c.sources, source)
}

// Load loads configuration from all sources
func (c *Config) Load(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear existing data
	c.data = make(map[string]any)

	// Load from all sources
	for _, source := range c.sources {
		sourceData, err := source.Load(ctx)
		if err != nil {
			return fmt.Errorf("failed to load from source %s: %w", source.Name(), err)
		}

		// Merge data (later sources override earlier ones)
		for key, value := range sourceData {
			c.data[key] = value
		}
	}

	return nil
}

// Get retrieves a configuration value as string
func (c *Config) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if value, exists := c.data[key]; exists {
		return c.toString(value)
	}
	return ""
}

// GetString retrieves a configuration value as string with default
func (c *Config) GetString(key, defaultValue string) string {
	value := c.Get(key)
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}

// GetInt retrieves a configuration value as int
func (c *Config) GetInt(key string, defaultValue int) int {
	value := c.Get(key)
	if value == "" {
		return defaultValue
	}

	// Trim whitespace before parsing
	value = strings.TrimSpace(value)
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// GetBool retrieves a configuration value as bool
func (c *Config) GetBool(key string, defaultValue bool) bool {
	value := c.Get(key)
	if value == "" {
		return defaultValue
	}

	// Trim whitespace before checking
	value = strings.TrimSpace(value)
	switch strings.ToLower(value) {
	case "1", "true", "yes", "y", "on", "t":
		return true
	case "0", "false", "no", "n", "off", "f":
		return false
	default:
		return defaultValue
	}
}

// GetFloat retrieves a configuration value as float64
func (c *Config) GetFloat(key string, defaultValue float64) float64 {
	value := c.Get(key)
	if value == "" {
		return defaultValue
	}

	// Trim whitespace before parsing
	value = strings.TrimSpace(value)
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return floatValue
}

// GetDurationMs retrieves a configuration value as time.Duration (milliseconds)
func (c *Config) GetDurationMs(key string, defaultValueMs int) time.Duration {
	return time.Duration(c.GetInt(key, defaultValueMs)) * time.Millisecond
}

// GetAll returns all configuration as a map
func (c *Config) GetAll() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create a copy to prevent external modification
	result := make(map[string]any, len(c.data))
	for key, value := range c.data {
		result[key] = value
	}
	return result
}

// Watch registers a callback for configuration changes
func (c *Config) Watch(key string, callback func(newValue map[string]any)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.watchers[key]; !exists {
		c.watchers[key] = make([]chan map[string]any, 0)
	}

	ch := make(chan map[string]any, 1)
	c.watchers[key] = append(c.watchers[key], ch)

	// Start a goroutine to listen for changes
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			case data := <-ch:
				callback(data)
			}
		}
	}()

	return nil
}

// WatchKey registers a callback for changes to a specific configuration key
func (c *Config) WatchKey(key string, callback func(oldValue, newValue any)) error {
	return c.Watch(key, func(newData map[string]any) {
		c.mu.RLock()
		oldValue := c.data[key]
		c.mu.RUnlock()

		newValue := newData[key]
		callback(oldValue, newValue)
	})
}

// StartWatchers starts watching all configured sources for changes
func (c *Config) StartWatchers(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, source := range c.sources {
		updates, errs := source.Watch(ctx)

		go func(src Source) {
			for {
				select {
				case <-ctx.Done():
					return
				case update := <-updates:
					c.handleSourceUpdate(src.Name(), update)
				case err := <-errs:
					// Log error but continue watching
					fmt.Printf("Watch error from source %s: %v\n", src.Name(), err)
				}
			}
		}(source)
	}

	return nil
}

// handleSourceUpdate handles configuration updates from a source
func (c *Config) handleSourceUpdate(sourceName string, newData map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Detect changes
	changes := make(map[string]struct{})
	for key, newValue := range newData {
		oldValue, exists := c.data[key]
		if !exists || !valuesEqual(oldValue, newValue) {
			changes[key] = struct{}{}
		}
	}

	// Update data
	for key, value := range newData {
		c.data[key] = value
	}

	// Notify watchers
	if len(changes) > 0 {
		c.notifyWatchers(changes, newData)
	}
}

// notifyWatchers notifies all registered watchers of configuration changes
func (c *Config) notifyWatchers(changes map[string]struct{}, newData map[string]any) {
	for key := range changes {
		for _, ch := range c.watchers[key] {
			select {
			case ch <- newData:
			default:
				// Skip if channel is full
			}
		}
	}

	// Notify global watchers (those watching for any change)
	for _, ch := range c.watchers["*"] {
		select {
		case ch <- newData:
		default:
			// Skip if channel is full
		}
	}
}

// Reload reloads configuration from all sources
func (c *Config) Reload(ctx context.Context) error {
	return c.Load(ctx)
}

// valuesEqual compares two values for equality
func valuesEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Convert to strings for comparison
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// Close stops all watchers and releases resources
func (c *Config) Close() error {
	c.cancel()
	return nil
}

// toString converts any value to string
func (c *Config) toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Unmarshal populates a struct with configuration values
func (c *Config) Unmarshal(dst any) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val := reflect.ValueOf(dst)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("dst must be a non-nil pointer")
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// Get configuration key from tag or field name
		key := field.Tag.Get("config")
		if key == "" {
			key = strings.ToUpper(field.Name)
		}

		if value, exists := c.data[key]; exists {
			if err := c.setField(fieldValue, value); err != nil {
				return fmt.Errorf("failed to set field %s: %w", field.Name, err)
			}
		}
	}

	return nil
}

// setField sets a field value based on its type
func (c *Config) setField(fieldValue reflect.Value, value any) error {
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(c.toString(value))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(c.toString(value), 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(c.toString(value), 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(c.toString(value), 64)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(floatValue)
	case reflect.Bool:
		switch strings.ToLower(c.toString(value)) {
		case "1", "true", "yes", "y", "on", "t":
			fieldValue.SetBool(true)
		case "0", "false", "no", "n", "off", "f":
			fieldValue.SetBool(false)
		default:
			return fmt.Errorf("invalid boolean value: %s", c.toString(value))
		}
	default:
		return fmt.Errorf("unsupported field type: %v", fieldValue.Kind())
	}

	return nil
}
