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

// Source represents a configuration source
type Source interface {
	Load(ctx context.Context) (map[string]any, error)
	Watch(ctx context.Context) (<-chan map[string]any, <-chan error)
	Name() string
}

// ConfigManager manages application configuration with dependency injection
type ConfigManager struct {
	sources  []Source
	data     map[string]any
	mu       sync.RWMutex
	watchers map[string][]WatcherCallback
	ctx      context.Context
	cancel   context.CancelFunc
	watchWg  sync.WaitGroup
	logger   log.StructuredLogger
}

// Type-safe configuration access methods

// String returns a type-safe string configuration
func (cm *ConfigManager) String(key string, defaultValue string, validators ...Validator) (string, error) {
	value := cm.GetString(key, defaultValue)

	for _, validator := range validators {
		if err := validator.Validate(value, key); err != nil {
			return defaultValue, err
		}
	}

	return value, nil
}

// Int returns a type-safe int configuration
func (cm *ConfigManager) Int(key string, defaultValue int, validators ...Validator) (int, error) {
	value := cm.GetInt(key, defaultValue)

	// Add range validator for numeric values if none specified
	hasRangeValidator := false
	for _, validator := range validators {
		if _, ok := validator.(*Range); ok {
			hasRangeValidator = true
			break
		}
	}

	if !hasRangeValidator {
		// Add default range validator to prevent overflow
		validators = append(validators, &Range{Min: -2147483648, Max: 2147483647})
	}

	for _, validator := range validators {
		if err := validator.Validate(value, key); err != nil {
			return defaultValue, err
		}
	}

	return value, nil
}

// Bool returns a type-safe bool configuration
func (cm *ConfigManager) Bool(key string, defaultValue bool, validators ...Validator) (bool, error) {
	value := cm.GetBool(key, defaultValue)

	for _, validator := range validators {
		if err := validator.Validate(value, key); err != nil {
			return defaultValue, err
		}
	}

	return value, nil
}

// Float returns a type-safe float64 configuration
func (cm *ConfigManager) Float(key string, defaultValue float64, validators ...Validator) (float64, error) {
	value := cm.GetFloat(key, defaultValue)

	for _, validator := range validators {
		if err := validator.Validate(value, key); err != nil {
			return defaultValue, err
		}
	}

	return value, nil
}

// DurationMs returns a type-safe time.Duration configuration
func (cm *ConfigManager) DurationMs(key string, defaultValueMs int, validators ...Validator) (time.Duration, error) {
	value := cm.GetDurationMs(key, defaultValueMs)

	// Convert to int for validation
	intValue := int(value.Milliseconds())

	for _, validator := range validators {
		if err := validator.Validate(intValue, key); err != nil {
			return time.Duration(defaultValueMs) * time.Millisecond, err
		}
	}

	return value, nil
}

// WatcherCallback represents a configuration change callback
type WatcherCallback func(oldValue, newValue any)

// NewConfigManager creates a new ConfigManager instance
func NewConfigManager(logger log.StructuredLogger) *ConfigManager {
	ctx, cancel := context.WithCancel(context.Background())
	if logger == nil {
		logger = log.NewGLogger()
	}
	return &ConfigManager{
		sources:  make([]Source, 0),
		data:     make(map[string]any),
		watchers: make(map[string][]WatcherCallback),
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
	}
}

// AddSource adds a configuration source
func (cm *ConfigManager) AddSource(source Source) error {
	if source == nil {
		return errors.New("source cannot be nil")
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.ctx.Err() != nil {
		return errors.New("config manager is closed")
	}

	cm.sources = append(cm.sources, source)
	return nil
}

// Load loads configuration from all sources
func (cm *ConfigManager) Load(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.ctx.Err() != nil {
		return errors.New("config manager is closed")
	}

	// Clear existing data
	cm.data = make(map[string]any)

	// Load from all sources
	for _, source := range cm.sources {
		sourceData, err := source.Load(ctx)
		if err != nil {
			cm.logger.Error("Failed to load configuration", log.Fields{
				"source": source.Name(),
				"error":  err,
			})
			return fmt.Errorf("failed to load from source %s: %w", source.Name(), err)
		}

		// Merge data (later sources override earlier ones)
		for key, value := range sourceData {
			cm.data[key] = value
		}

		cm.logger.Info("Loaded configuration", log.Fields{
			"source": source.Name(),
			"keys":   len(sourceData),
		})
	}

	return nil
}

// LoadBestEffort loads configuration from all sources, skipping failures.
// It returns an error only if all sources fail.
func (cm *ConfigManager) LoadBestEffort(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.ctx.Err() != nil {
		return errors.New("config manager is closed")
	}

	cm.data = make(map[string]any)
	loaded := 0
	var lastErr error

	for _, source := range cm.sources {
		sourceData, err := source.Load(ctx)
		if err != nil {
			lastErr = err
			cm.logger.Error("Failed to load configuration", log.Fields{
				"source": source.Name(),
				"error":  err,
			})
			continue
		}

		loaded++
		for key, value := range sourceData {
			cm.data[key] = value
		}

		cm.logger.Info("Loaded configuration", log.Fields{
			"source": source.Name(),
			"keys":   len(sourceData),
		})
	}

	if loaded == 0 && lastErr != nil {
		return fmt.Errorf("failed to load any configuration sources: %w", lastErr)
	}

	return nil
}

// Get retrieves a configuration value as string
func (cm *ConfigManager) Get(key string) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if value, exists := cm.data[key]; exists {
		return cm.toString(value)
	}
	return ""
}

// GetString retrieves a configuration value as string with default
func (cm *ConfigManager) GetString(key, defaultValue string) string {
	value := cm.Get(key)
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}

// GetInt retrieves a configuration value as int
func (cm *ConfigManager) GetInt(key string, defaultValue int) int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if value, exists := cm.data[key]; exists {
		// Direct type assertion first for better performance
		switch v := value.(type) {
		case int:
			return v
		case int8:
			return int(v)
		case int16:
			return int(v)
		case int32:
			return int(v)
		case int64:
			return int(v)
		case uint:
			return int(v)
		case uint8:
			return int(v)
		case uint16:
			return int(v)
		case uint32:
			return int(v)
		case uint64:
			return int(v)
		case float32:
			return int(v)
		case float64:
			return int(v)
		case bool:
			if v {
				return 1
			}
			return 0
		case string:
			if v := strings.TrimSpace(v); v != "" {
				if intValue, err := strconv.Atoi(v); err == nil {
					return intValue
				}
			}
		}
	}
	return defaultValue
}

// GetBool retrieves a configuration value as bool
func (cm *ConfigManager) GetBool(key string, defaultValue bool) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Helper function to convert value to bool
	convertToBool := func(value any) (bool, bool) {
		// Direct type assertion first for better performance
		switch v := value.(type) {
		case bool:
			return v, true
		case int, int8, int16, int32, int64:
			return v != 0, true
		case uint, uint8, uint16, uint32, uint64:
			return v != 0, true
		case float32, float64:
			return v != 0, true
		case string:
			if v := strings.TrimSpace(v); v != "" {
				return parseBool(v, defaultValue), true
			}
		}
		return defaultValue, false
	}

	// Try exact key first
	if value, exists := cm.data[key]; exists {
		if boolValue, ok := convertToBool(value); ok {
			return boolValue
		}
	}

	// Try snake_case version
	snakeKey := toSnakeCase(key)
	if value, exists := cm.data[snakeKey]; exists {
		if boolValue, ok := convertToBool(value); ok {
			return boolValue
		}
	}

	// Try uppercase version
	upperKey := strings.ToUpper(key)
	if value, exists := cm.data[upperKey]; exists {
		if boolValue, ok := convertToBool(value); ok {
			return boolValue
		}
	}

	return defaultValue
}

// GetFloat retrieves a configuration value as float64
func (cm *ConfigManager) GetFloat(key string, defaultValue float64) float64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if value, exists := cm.data[key]; exists {
		// Direct type assertion first for better performance
		switch v := value.(type) {
		case float32:
			return float64(v)
		case float64:
			return v
		case int:
			return float64(v)
		case int8:
			return float64(v)
		case int16:
			return float64(v)
		case int32:
			return float64(v)
		case int64:
			return float64(v)
		case uint:
			return float64(v)
		case uint8:
			return float64(v)
		case uint16:
			return float64(v)
		case uint32:
			return float64(v)
		case uint64:
			return float64(v)
		case bool:
			if v {
				return 1.0
			}
			return 0.0
		case string:
			if v := strings.TrimSpace(v); v != "" {
				if floatValue, err := strconv.ParseFloat(v, 64); err == nil {
					return floatValue
				}
			}
		}
	}
	return defaultValue
}

// GetDurationMs retrieves a configuration value as time.Duration (milliseconds)
func (cm *ConfigManager) GetDurationMs(key string, defaultValueMs int) time.Duration {
	return time.Duration(cm.GetInt(key, defaultValueMs)) * time.Millisecond
}

// GetAll returns all configuration as a map
func (cm *ConfigManager) GetAll() map[string]any {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Create a copy to prevent external modification
	result := make(map[string]any, len(cm.data))
	for key, value := range cm.data {
		result[key] = value
	}
	return result
}

// Watch registers a callback for configuration changes to a specific key
func (cm *ConfigManager) Watch(key string, callback WatcherCallback) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.ctx.Err() != nil {
		return errors.New("config manager is closed")
	}

	if _, exists := cm.watchers[key]; !exists {
		cm.watchers[key] = make([]WatcherCallback, 0)
	}

	cm.watchers[key] = append(cm.watchers[key], callback)
	return nil
}

// WatchGlobal registers a callback for any configuration change
func (cm *ConfigManager) WatchGlobal(callback WatcherCallback) error {
	return cm.Watch("*", callback)
}

// StartWatchers starts watching all configured sources for changes
func (cm *ConfigManager) StartWatchers(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.ctx.Err() != nil {
		return errors.New("config manager is closed")
	}

	for _, source := range cm.sources {
		updates, errs := source.Watch(ctx)

		cm.watchWg.Add(1)
		go func(src Source, updates <-chan map[string]any, errs <-chan error) {
			defer cm.watchWg.Done()
			for updates != nil || errs != nil {
				select {
				case <-cm.ctx.Done():
					return
				case update, ok := <-updates:
					if !ok {
						updates = nil
						continue
					}
					if update != nil {
						cm.handleSourceUpdate(src.Name(), update)
					}
				case err, ok := <-errs:
					if !ok {
						errs = nil
						continue
					}
					if err != nil {
						cm.logger.Error("Watch error", log.Fields{
							"source": src.Name(),
							"error":  err,
						})
					}
				}
			}
		}(source, updates, errs)
	}

	return nil
}

// handleSourceUpdate handles configuration updates from a source
func (cm *ConfigManager) handleSourceUpdate(sourceName string, newData map[string]any) {
	cm.mu.Lock()

	if len(newData) == 0 {
		cm.mu.Unlock()
		return
	}

	// Detect changes
	changes := make(map[string]struct{})
	oldValues := make(map[string]any)
	for key, newValue := range newData {
		oldValue, exists := cm.data[key]
		if !exists || !valuesEqual(oldValue, newValue) {
			changes[key] = struct{}{}
			oldValues[key] = oldValue
		}
	}

	// Update data
	for key, value := range newData {
		cm.data[key] = value
	}

	// Notify watchers
	if len(changes) > 0 {
		keyWatchers := make(map[string][]WatcherCallback, len(changes))
		for key := range changes {
			if watchers, exists := cm.watchers[key]; exists && len(watchers) > 0 {
				copied := make([]WatcherCallback, len(watchers))
				copy(copied, watchers)
				keyWatchers[key] = copied
			}
		}

		globalWatchers := append([]WatcherCallback(nil), cm.watchers["*"]...)

		cm.mu.Unlock()

		cm.notifyWatchers(changes, oldValues, newData, keyWatchers, globalWatchers)
		cm.logger.Info("Configuration updated", log.Fields{
			"source":  sourceName,
			"changes": len(changes),
		})
		return
	}

	cm.mu.Unlock()
}

// notifyWatchers notifies all registered watchers of configuration changes
func (cm *ConfigManager) notifyWatchers(changes map[string]struct{}, oldValues map[string]any, newData map[string]any, keyWatchers map[string][]WatcherCallback, globalWatchers []WatcherCallback) {
	for key := range changes {
		watchers := keyWatchers[key]
		if len(watchers) == 0 {
			continue
		}
		oldValue := oldValues[key]
		newValue := newData[key]

		for _, callback := range watchers {
			cm.callWatcher(callback, oldValue, newValue)
		}
	}

	// Notify global watchers
	if len(globalWatchers) == 0 {
		return
	}

	dataSnapshot := make(map[string]any, len(newData))
	for key, value := range newData {
		dataSnapshot[key] = value
	}

	for _, callback := range globalWatchers {
		cm.callWatcher(callback, nil, dataSnapshot)
	}
}

func (cm *ConfigManager) callWatcher(callback WatcherCallback, old, new any) {
	defer func() {
		if r := recover(); r != nil {
			cm.logger.Error("Watcher callback panicked", log.Fields{
				"error": r,
			})
		}
	}()
	callback(old, new)
}

// Reload reloads configuration from all sources
func (cm *ConfigManager) Reload(ctx context.Context) error {
	return cm.Load(ctx)
}

// ReloadWithValidation reloads configuration and validates before committing.
func (cm *ConfigManager) ReloadWithValidation(ctx context.Context, validate func(map[string]any) error) error {
	if validate == nil {
		return cm.Reload(ctx)
	}

	snapshot := cm.GetAll()
	if err := cm.Load(ctx); err != nil {
		return err
	}

	data := cm.GetAll()
	if err := validate(data); err != nil {
		cm.mu.Lock()
		cm.data = snapshot
		cm.mu.Unlock()
		return err
	}

	return nil
}

// Close stops all watchers and releases resources
func (cm *ConfigManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.ctx.Err() != nil {
		return nil
	}

	cm.cancel()
	cm.watchWg.Wait() // Wait for all watcher goroutines to finish
	cm.logger.Info("Config manager closed", nil)
	return nil
}

// Unmarshal populates a struct with configuration values
func (cm *ConfigManager) Unmarshal(dst any) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	val := reflect.ValueOf(dst)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("dst must be a non-nil pointer")
	}

	return cm.unmarshalValue(val.Elem())
}

func (cm *ConfigManager) unmarshalValue(val reflect.Value) error {
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
			if err := cm.unmarshalValue(fieldValue); err != nil {
				return err
			}
			continue
		}

		if fieldValue.Kind() == reflect.Ptr && fieldValue.Type().Elem().Kind() == reflect.Struct && fieldValue.Type().Elem() != timeType {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
			if err := cm.unmarshalValue(fieldValue.Elem()); err != nil {
				return err
			}
			continue
		}

		// Get configuration key from tag or field name
		key := field.Tag.Get("config")
		if key == "" {
			key = field.Name
		}

		if value, exists := cm.lookupConfigValue(key); exists {
			if err := cm.setField(fieldValue, value); err != nil {
				return fmt.Errorf("failed to set field %s: %w", field.Name, err)
			}
		}
	}

	return nil
}

func (cm *ConfigManager) lookupConfigValue(key string) (any, bool) {
	if value, exists := cm.data[key]; exists {
		return value, true
	}

	snake := toSnakeCase(key)
	if value, exists := cm.data[snake]; exists {
		return value, true
	}

	upper := strings.ToUpper(key)
	if value, exists := cm.data[upper]; exists {
		return value, true
	}

	upperSnake := strings.ToUpper(snake)
	if value, exists := cm.data[upperSnake]; exists {
		return value, true
	}

	lower := strings.ToLower(key)
	if value, exists := cm.data[lower]; exists {
		return value, true
	}

	return nil, false
}

// setField sets a field value based on its type
func (cm *ConfigManager) setField(fieldValue reflect.Value, value any) error {
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(cm.toString(value))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(cm.toString(value), 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(cm.toString(value), 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(cm.toString(value), 64)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(floatValue)
	case reflect.Bool:
		switch strings.ToLower(cm.toString(value)) {
		case "1", "true", "yes", "y", "on", "t":
			fieldValue.SetBool(true)
		case "0", "false", "no", "n", "off", "f":
			fieldValue.SetBool(false)
		default:
			return fmt.Errorf("invalid boolean value: %s", cm.toString(value))
		}
	default:
		return fmt.Errorf("unsupported field type: %v", fieldValue.Kind())
	}

	return nil
}

// toString converts any value to string
func (cm *ConfigManager) toString(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		// Fallback to fmt.Sprintf for complex types
		return fmt.Sprintf("%v", v)
	}
}

// valuesEqual compares two values for equality
func valuesEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Type-aware comparison
	switch av := a.(type) {
	case string:
		if bv, ok := b.(string); ok {
			return av == bv
		}
	case int:
		if bv, ok := b.(int); ok {
			return av == bv
		}
	case float64:
		if bv, ok := b.(float64); ok {
			return av == bv
		}
	case bool:
		if bv, ok := b.(bool); ok {
			return av == bv
		}
	}

	// Fallback to string comparison for complex types
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// parseBool parses a string to bool
func parseBool(value string, defaultValue bool) bool {
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

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(s string) string {
	if s == "" {
		return s
	}

	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Check if previous character is not uppercase
			if i > 0 && s[i-1] >= 'a' && s[i-1] <= 'z' {
				result = append(result, '_')
			}
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}
