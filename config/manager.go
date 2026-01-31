package config

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/spcent/plumego/log"
)

var errConfigClosed = errors.New("config manager is closed")

// WatcherCallback represents a configuration change callback.
type WatcherCallback func(oldValue, newValue any)

// Manager manages application configuration with dependency injection.
type Manager struct {
	sources  []Source
	data     map[string]any
	mu       sync.RWMutex
	watchers map[string][]WatcherCallback
	ctx      context.Context
	cancel   context.CancelFunc
	watchWg  sync.WaitGroup
	logger   log.StructuredLogger
}

// NewManager creates a new Manager instance.
func NewManager(logger log.StructuredLogger) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	if logger == nil {
		logger = log.NewGLogger()
	}
	return &Manager{
		sources:  make([]Source, 0),
		data:     make(map[string]any),
		watchers: make(map[string][]WatcherCallback),
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
	}
}

// AddSource adds a configuration source.
func (m *Manager) AddSource(source Source) error {
	if source == nil {
		return errors.New("source cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ctx.Err() != nil {
		return errConfigClosed
	}

	m.sources = append(m.sources, source)
	return nil
}

// Load loads configuration from all sources.
func (m *Manager) Load(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	sources, err := m.snapshotSources()
	if err != nil {
		return err
	}

	data := make(map[string]any)
	for _, source := range sources {
		sourceData, err := source.Load(ctx)
		if err != nil {
			m.logger.Error("Failed to load configuration", log.Fields{
				"source": source.Name(),
				"error":  err,
			})
			return fmt.Errorf("failed to load from source %s: %w", source.Name(), err)
		}

		normalized := normalizeData(sourceData, m.logger)
		for key, value := range normalized {
			data[key] = value
		}

		m.logger.Info("Loaded configuration", log.Fields{
			"source": source.Name(),
			"keys":   len(normalized),
		})
	}

	m.mu.Lock()
	if m.ctx.Err() != nil {
		m.mu.Unlock()
		return errConfigClosed
	}
	m.data = data
	m.mu.Unlock()

	return nil
}

// LoadBestEffort loads configuration from all sources, skipping failures.
// It returns an error only if all sources fail.
func (m *Manager) LoadBestEffort(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	sources, err := m.snapshotSources()
	if err != nil {
		return err
	}

	data := make(map[string]any)
	loaded := 0
	var lastErr error

	for _, source := range sources {
		sourceData, err := source.Load(ctx)
		if err != nil {
			lastErr = err
			m.logger.Error("Failed to load configuration", log.Fields{
				"source": source.Name(),
				"error":  err,
			})
			continue
		}

		loaded++
		normalized := normalizeData(sourceData, m.logger)
		for key, value := range normalized {
			data[key] = value
		}

		m.logger.Info("Loaded configuration", log.Fields{
			"source": source.Name(),
			"keys":   len(normalized),
		})
	}

	if loaded == 0 && lastErr != nil {
		return fmt.Errorf("failed to load any configuration sources: %w", lastErr)
	}

	m.mu.Lock()
	if m.ctx.Err() != nil {
		m.mu.Unlock()
		return errConfigClosed
	}
	m.data = data
	m.mu.Unlock()

	return nil
}

func (m *Manager) snapshotSources() ([]Source, error) {
	m.mu.RLock()
	if m.ctx.Err() != nil {
		m.mu.RUnlock()
		return nil, errConfigClosed
	}
	sources := append([]Source(nil), m.sources...)
	m.mu.RUnlock()
	return sources, nil
}

func (m *Manager) getValue(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return lookupValue(m.data, key)
}

// Get retrieves a configuration value as string.
func (m *Manager) Get(key string) string {
	value, exists := m.getValue(key)
	if exists {
		return toString(value)
	}
	return ""
}

// GetString retrieves a configuration value as string with default.
func (m *Manager) GetString(key, defaultValue string) string {
	value := m.Get(key)
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}

// GetInt retrieves a configuration value as int.
func (m *Manager) GetInt(key string, defaultValue int) int {
	value, exists := m.getValue(key)
	if !exists {
		return defaultValue
	}
	return toInt(value, defaultValue)
}

// GetBool retrieves a configuration value as bool.
func (m *Manager) GetBool(key string, defaultValue bool) bool {
	value, exists := m.getValue(key)
	if !exists {
		return defaultValue
	}
	return toBool(value, defaultValue)
}

// GetFloat retrieves a configuration value as float64.
func (m *Manager) GetFloat(key string, defaultValue float64) float64 {
	value, exists := m.getValue(key)
	if !exists {
		return defaultValue
	}
	return toFloat64(value, defaultValue)
}

// GetDurationMs retrieves a configuration value as time.Duration (milliseconds).
func (m *Manager) GetDurationMs(key string, defaultValueMs int) time.Duration {
	return time.Duration(m.GetInt(key, defaultValueMs)) * time.Millisecond
}

// GetAll returns all configuration as a map.
func (m *Manager) GetAll() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]any, len(m.data))
	for key, value := range m.data {
		result[key] = value
	}
	return result
}

// Type-safe configuration access methods with validation.

// String returns a type-safe string configuration with optional validation.
func (m *Manager) String(key string, defaultValue string, validators ...Validator) (string, error) {
	value := m.GetString(key, defaultValue)

	for _, validator := range validators {
		if err := validator.Validate(value, key); err != nil {
			return defaultValue, err
		}
	}

	return value, nil
}

// Int returns a type-safe int configuration with optional validation.
func (m *Manager) Int(key string, defaultValue int, validators ...Validator) (int, error) {
	value := m.GetInt(key, defaultValue)

	hasRangeValidator := false
	for _, validator := range validators {
		if _, ok := validator.(*Range); ok {
			hasRangeValidator = true
			break
		}
	}

	if !hasRangeValidator {
		validators = append(validators, &Range{Min: -2147483648, Max: 2147483647})
	}

	for _, validator := range validators {
		if err := validator.Validate(value, key); err != nil {
			return defaultValue, err
		}
	}

	return value, nil
}

// Bool returns a type-safe bool configuration with optional validation.
func (m *Manager) Bool(key string, defaultValue bool, validators ...Validator) (bool, error) {
	value := m.GetBool(key, defaultValue)

	for _, validator := range validators {
		if err := validator.Validate(value, key); err != nil {
			return defaultValue, err
		}
	}

	return value, nil
}

// Float returns a type-safe float64 configuration with optional validation.
func (m *Manager) Float(key string, defaultValue float64, validators ...Validator) (float64, error) {
	value := m.GetFloat(key, defaultValue)

	for _, validator := range validators {
		if err := validator.Validate(value, key); err != nil {
			return defaultValue, err
		}
	}

	return value, nil
}

// DurationMs returns a type-safe time.Duration configuration with optional validation.
func (m *Manager) DurationMs(key string, defaultValueMs int, validators ...Validator) (time.Duration, error) {
	value := m.GetDurationMs(key, defaultValueMs)

	intValue := int(value.Milliseconds())
	for _, validator := range validators {
		if err := validator.Validate(intValue, key); err != nil {
			return time.Duration(defaultValueMs) * time.Millisecond, err
		}
	}

	return value, nil
}

// Watch registers a callback for configuration changes to a specific key.
func (m *Manager) Watch(key string, callback WatcherCallback) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ctx.Err() != nil {
		return errConfigClosed
	}

	if key != "*" {
		key = normalizeKey(key)
	}

	if _, exists := m.watchers[key]; !exists {
		m.watchers[key] = make([]WatcherCallback, 0)
	}

	m.watchers[key] = append(m.watchers[key], callback)
	return nil
}

// WatchGlobal registers a callback for any configuration change.
func (m *Manager) WatchGlobal(callback WatcherCallback) error {
	return m.Watch("*", callback)
}

// StartWatchers starts watching all configured sources for changes.
func (m *Manager) StartWatchers(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	sources, err := m.snapshotSources()
	if err != nil {
		return err
	}

	for _, source := range sources {
		results := source.Watch(ctx)

		m.watchWg.Add(1)
		go func(src Source, results <-chan WatchResult) {
			defer m.watchWg.Done()
			for {
				select {
				case <-m.ctx.Done():
					return
				case <-ctx.Done():
					return
				case result, ok := <-results:
					if !ok {
						return
					}
					if result.Err != nil {
						m.logger.Error("Watch error", log.Fields{
							"source": src.Name(),
							"error":  result.Err,
						})
					} else if result.Data != nil {
						m.handleSourceUpdate(src.Name(), result.Data)
					}
				}
			}
		}(source, results)
	}

	return nil
}

// handleSourceUpdate handles configuration updates from a source.
func (m *Manager) handleSourceUpdate(sourceName string, newData map[string]any) {
	if len(newData) == 0 {
		return
	}

	newData = normalizeData(newData, m.logger)
	if len(newData) == 0 {
		return
	}

	m.mu.Lock()

	changes := make(map[string]struct{})
	oldValues := make(map[string]any)
	for key, newValue := range newData {
		oldValue, exists := m.data[key]
		if !exists || !valuesEqual(oldValue, newValue) {
			changes[key] = struct{}{}
			oldValues[key] = oldValue
		}
	}

	for key, value := range newData {
		m.data[key] = value
	}

	if len(changes) > 0 {
		keyWatchers := make(map[string][]WatcherCallback, len(changes))
		for key := range changes {
			if watchers, exists := m.watchers[key]; exists && len(watchers) > 0 {
				copied := make([]WatcherCallback, len(watchers))
				copy(copied, watchers)
				keyWatchers[key] = copied
			}
		}

		globalWatchers := append([]WatcherCallback(nil), m.watchers["*"]...)

		m.mu.Unlock()

		m.notifyWatchers(changes, oldValues, newData, keyWatchers, globalWatchers)
		m.logger.Info("Configuration updated", log.Fields{
			"source":  sourceName,
			"changes": len(changes),
		})
		return
	}

	m.mu.Unlock()
}

// notifyWatchers notifies all registered watchers of configuration changes.
func (m *Manager) notifyWatchers(changes map[string]struct{}, oldValues map[string]any, newData map[string]any, keyWatchers map[string][]WatcherCallback, globalWatchers []WatcherCallback) {
	for key := range changes {
		watchers := keyWatchers[key]
		if len(watchers) == 0 {
			continue
		}
		oldValue := oldValues[key]
		newValue := newData[key]

		for _, callback := range watchers {
			m.callWatcher(callback, oldValue, newValue)
		}
	}

	if len(globalWatchers) == 0 {
		return
	}

	dataSnapshot := make(map[string]any, len(newData))
	for key, value := range newData {
		dataSnapshot[key] = value
	}

	for _, callback := range globalWatchers {
		m.callWatcher(callback, nil, dataSnapshot)
	}
}

func (m *Manager) callWatcher(callback WatcherCallback, old, new any) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("Watcher callback panicked", log.Fields{
				"error": r,
			})
		}
	}()
	callback(old, new)
}

// Reload reloads configuration from all sources.
func (m *Manager) Reload(ctx context.Context) error {
	return m.Load(ctx)
}

// ReloadWithValidation reloads configuration and validates before committing.
func (m *Manager) ReloadWithValidation(ctx context.Context, validate func(map[string]any) error) error {
	if validate == nil {
		return m.Reload(ctx)
	}

	m.mu.RLock()
	snapshot := make(map[string]any, len(m.data))
	for key, value := range m.data {
		snapshot[key] = value
	}
	m.mu.RUnlock()

	if err := m.Load(ctx); err != nil {
		return err
	}

	m.mu.RLock()
	data := make(map[string]any, len(m.data))
	for key, value := range m.data {
		data[key] = value
	}
	m.mu.RUnlock()

	if err := validate(data); err != nil {
		m.mu.Lock()
		m.data = snapshot
		m.mu.Unlock()
		return err
	}

	return nil
}

// Close stops all watchers and releases resources.
func (m *Manager) Close() error {
	m.mu.Lock()
	if m.ctx.Err() != nil {
		m.mu.Unlock()
		return nil
	}
	m.cancel()
	m.mu.Unlock()

	m.watchWg.Wait()
	m.logger.Info("Config manager closed", nil)
	return nil
}

// Unmarshal populates a struct with configuration values.
func (m *Manager) Unmarshal(dst any) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val := reflect.ValueOf(dst)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("dst must be a non-nil pointer")
	}

	return m.unmarshalValue(val.Elem())
}

func (m *Manager) unmarshalValue(val reflect.Value) error {
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("dst must point to a struct")
	}

	typ := val.Type()
	timeType := reflect.TypeOf(time.Time{})

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		if fieldValue.Kind() == reflect.Struct && fieldValue.Type() != timeType {
			if err := m.unmarshalValue(fieldValue); err != nil {
				return err
			}
			continue
		}

		if fieldValue.Kind() == reflect.Ptr && fieldValue.Type().Elem().Kind() == reflect.Struct && fieldValue.Type().Elem() != timeType {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
			if err := m.unmarshalValue(fieldValue.Elem()); err != nil {
				return err
			}
			continue
		}

		key := field.Tag.Get("config")
		if key == "" {
			key = field.Name
		}

		if value, exists := lookupValue(m.data, key); exists {
			if err := m.setField(fieldValue, value); err != nil {
				return fmt.Errorf("failed to set field %s: %w", field.Name, err)
			}
		}
	}

	return nil
}

// setField sets a field value based on its type.
func (m *Manager) setField(fieldValue reflect.Value, value any) error {
	strValue := toString(value)

	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(strValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(strValue, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(strValue, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(floatValue)
	case reflect.Bool:
		switch strings.ToLower(strValue) {
		case "1", "true", "yes", "y", "on", "t":
			fieldValue.SetBool(true)
		case "0", "false", "no", "n", "off", "f":
			fieldValue.SetBool(false)
		default:
			return fmt.Errorf("invalid boolean value: %s", strValue)
		}
	default:
		return fmt.Errorf("unsupported field type: %v", fieldValue.Kind())
	}

	return nil
}
