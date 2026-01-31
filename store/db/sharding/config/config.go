package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ShardingConfig represents the complete sharding configuration
type ShardingConfig struct {
	// Shards is the list of database shards
	Shards []ShardConfig `json:"shards"`

	// ShardingRules defines how tables are sharded
	ShardingRules []ShardingRuleConfig `json:"sharding_rules"`

	// CrossShardPolicy determines how cross-shard queries are handled
	CrossShardPolicy string `json:"cross_shard_policy"`

	// DefaultShardIndex is the default shard when routing fails (-1 to disable)
	DefaultShardIndex int `json:"default_shard_index"`

	// EnableMetrics enables metrics collection
	EnableMetrics bool `json:"enable_metrics"`

	// EnableTracing enables distributed tracing
	EnableTracing bool `json:"enable_tracing"`

	// LogLevel sets the logging level (debug, info, warn, error)
	LogLevel string `json:"log_level"`
}

// ShardConfig represents a single database shard configuration
type ShardConfig struct {
	// Name is the shard identifier
	Name string `json:"name"`

	// Primary is the primary database configuration
	Primary DatabaseConfig `json:"primary"`

	// Replicas are the read replica configurations
	Replicas []DatabaseConfig `json:"replicas"`

	// ReplicaWeights for load balancing (optional)
	ReplicaWeights []int `json:"replica_weights"`

	// FallbackToPrimary when all replicas are down
	FallbackToPrimary bool `json:"fallback_to_primary"`

	// HealthCheck configuration
	HealthCheck HealthCheckConfig `json:"health_check"`
}

// DatabaseConfig represents a database connection configuration
type DatabaseConfig struct {
	// Driver is the database driver (mysql, postgres, sqlite3)
	Driver string `json:"driver"`

	// Host is the database host
	Host string `json:"host"`

	// Port is the database port
	Port int `json:"port"`

	// Database is the database name
	Database string `json:"database"`

	// Username for authentication
	Username string `json:"username"`

	// Password for authentication
	Password string `json:"password"`

	// DSN is the full data source name (overrides other fields if set)
	DSN string `json:"dsn"`

	// MaxOpenConns is the maximum number of open connections
	MaxOpenConns int `json:"max_open_conns"`

	// MaxIdleConns is the maximum number of idle connections
	MaxIdleConns int `json:"max_idle_conns"`

	// ConnMaxLifetime is the maximum connection lifetime (e.g., "30m")
	ConnMaxLifetime string `json:"conn_max_lifetime"`

	// ConnMaxIdleTime is the maximum connection idle time (e.g., "5m")
	ConnMaxIdleTime string `json:"conn_max_idle_time"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	// Enabled enables health checking
	Enabled bool `json:"enabled"`

	// Interval is the health check interval (e.g., "30s")
	Interval string `json:"interval"`

	// Timeout is the health check timeout (e.g., "5s")
	Timeout string `json:"timeout"`

	// FailureThreshold is the number of failures before marking unhealthy
	FailureThreshold int `json:"failure_threshold"`

	// RecoveryThreshold is the number of successes before marking healthy
	RecoveryThreshold int `json:"recovery_threshold"`
}

// ShardingRuleConfig represents a sharding rule configuration
type ShardingRuleConfig struct {
	// TableName is the logical table name
	TableName string `json:"table_name"`

	// ShardKeyColumn is the column name used for sharding
	ShardKeyColumn string `json:"shard_key_column"`

	// Strategy is the sharding strategy (hash, mod, range, list)
	Strategy string `json:"strategy"`

	// StrategyConfig is strategy-specific configuration
	StrategyConfig StrategyConfig `json:"strategy_config"`

	// ActualTableNames maps shard index to physical table name
	ActualTableNames map[int]string `json:"actual_table_names"`

	// DefaultShard is the default shard for this table (-1 to disable)
	DefaultShard int `json:"default_shard"`
}

// StrategyConfig represents strategy-specific configuration
type StrategyConfig struct {
	// Ranges for range-based sharding
	Ranges []RangeConfig `json:"ranges"`

	// Mapping for list-based sharding
	Mapping map[string]int `json:"mapping"`

	// DefaultShard for list strategy when value not in mapping
	DefaultShard int `json:"default_shard"`
}

// RangeConfig represents a range definition for range-based sharding
type RangeConfig struct {
	// Start is the range start (inclusive)
	Start interface{} `json:"start"`

	// End is the range end (exclusive)
	End interface{} `json:"end"`

	// Shard is the target shard index
	Shard int `json:"shard"`
}

// DefaultConfig returns a default sharding configuration
func DefaultConfig() ShardingConfig {
	return ShardingConfig{
		Shards:            []ShardConfig{},
		ShardingRules:     []ShardingRuleConfig{},
		CrossShardPolicy:  "deny",
		DefaultShardIndex: -1,
		EnableMetrics:     true,
		EnableTracing:     false,
		LogLevel:          "info",
	}
}

// Validate validates the configuration
func (c *ShardingConfig) Validate() error {
	if len(c.Shards) == 0 {
		return fmt.Errorf("at least one shard is required")
	}

	if len(c.ShardingRules) == 0 {
		return fmt.Errorf("at least one sharding rule is required")
	}

	// Validate cross-shard policy
	validPolicies := map[string]bool{
		"deny":  true,
		"first": true,
		"all":   true,
	}
	if !validPolicies[c.CrossShardPolicy] {
		return fmt.Errorf("invalid cross-shard policy: %s (must be deny, first, or all)", c.CrossShardPolicy)
	}

	// Validate default shard index
	if c.DefaultShardIndex >= len(c.Shards) {
		return fmt.Errorf("default shard index %d exceeds shard count %d", c.DefaultShardIndex, len(c.Shards))
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.LogLevel)
	}

	// Validate each shard
	for i, shard := range c.Shards {
		if err := shard.Validate(); err != nil {
			return fmt.Errorf("shard %d (%s): %w", i, shard.Name, err)
		}
	}

	// Validate each sharding rule
	for i, rule := range c.ShardingRules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("sharding rule %d (%s): %w", i, rule.TableName, err)
		}
	}

	return nil
}

// Validate validates a shard configuration
func (s *ShardConfig) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("shard name is required")
	}

	if err := s.Primary.Validate(); err != nil {
		return fmt.Errorf("primary database: %w", err)
	}

	for i, replica := range s.Replicas {
		if err := replica.Validate(); err != nil {
			return fmt.Errorf("replica %d: %w", i, err)
		}
	}

	if len(s.ReplicaWeights) > 0 && len(s.ReplicaWeights) != len(s.Replicas) {
		return fmt.Errorf("replica weights count (%d) must match replicas count (%d)",
			len(s.ReplicaWeights), len(s.Replicas))
	}

	return nil
}

// Validate validates a database configuration
func (d *DatabaseConfig) Validate() error {
	// If DSN is provided, skip other validations
	if d.DSN != "" {
		return nil
	}

	if d.Driver == "" {
		return fmt.Errorf("driver is required")
	}

	if d.Host == "" {
		return fmt.Errorf("host is required")
	}

	if d.Database == "" {
		return fmt.Errorf("database is required")
	}

	return nil
}

// BuildDSN builds a DSN from the configuration
func (d *DatabaseConfig) BuildDSN() string {
	if d.DSN != "" {
		return d.DSN
	}

	switch d.Driver {
	case "mysql":
		// mysql: user:password@tcp(host:port)/database
		password := d.Password
		if password != "" {
			password = ":" + password
		}
		port := d.Port
		if port == 0 {
			port = 3306
		}
		return fmt.Sprintf("%s%s@tcp(%s:%d)/%s", d.Username, password, d.Host, port, d.Database)

	case "postgres":
		// postgres: host=localhost port=5432 user=user password=pass dbname=db sslmode=disable
		port := d.Port
		if port == 0 {
			port = 5432
		}
		dsn := fmt.Sprintf("host=%s port=%d dbname=%s", d.Host, port, d.Database)
		if d.Username != "" {
			dsn += " user=" + d.Username
		}
		if d.Password != "" {
			dsn += " password=" + d.Password
		}
		dsn += " sslmode=disable"
		return dsn

	case "sqlite3":
		// sqlite3: /path/to/database.db
		return d.Database

	default:
		return d.DSN
	}
}

// ParseDuration parses a duration string (e.g., "30m", "5s")
func (d *DatabaseConfig) ParseConnMaxLifetime() (time.Duration, error) {
	if d.ConnMaxLifetime == "" {
		return 0, nil
	}
	return time.ParseDuration(d.ConnMaxLifetime)
}

// ParseConnMaxIdleTime parses the connection max idle time
func (d *DatabaseConfig) ParseConnMaxIdleTime() (time.Duration, error) {
	if d.ConnMaxIdleTime == "" {
		return 0, nil
	}
	return time.ParseDuration(d.ConnMaxIdleTime)
}

// Validate validates a sharding rule configuration
func (r *ShardingRuleConfig) Validate() error {
	if r.TableName == "" {
		return fmt.Errorf("table name is required")
	}

	if r.ShardKeyColumn == "" {
		return fmt.Errorf("shard key column is required")
	}

	validStrategies := map[string]bool{
		"hash":  true,
		"mod":   true,
		"range": true,
		"list":  true,
	}
	if !validStrategies[r.Strategy] {
		return fmt.Errorf("invalid strategy: %s (must be hash, mod, range, or list)", r.Strategy)
	}

	// Validate strategy-specific configuration
	if r.Strategy == "range" && len(r.StrategyConfig.Ranges) == 0 {
		return fmt.Errorf("range strategy requires range definitions")
	}

	return nil
}

// LoadFromJSON loads configuration from JSON bytes
func LoadFromJSON(data []byte) (*ShardingConfig, error) {
	var config ShardingConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// LoadFromJSONFile loads configuration from a JSON file
func LoadFromJSONFile(path string) (*ShardingConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return LoadFromJSON(data)
}

// LoadFromFile loads configuration from a JSON file (alias for LoadFromJSONFile)
func LoadFromFile(path string) (*ShardingConfig, error) {
	return LoadFromJSONFile(path)
}

// MergeWithEnv merges configuration with environment variables
// Environment variables override config file values
func (c *ShardingConfig) MergeWithEnv() error {
	// Override top-level config
	if val := os.Getenv("DB_SHARD_CROSS_SHARD_POLICY"); val != "" {
		c.CrossShardPolicy = val
	}

	if val := os.Getenv("DB_SHARD_DEFAULT_INDEX"); val != "" {
		if idx, err := strconv.Atoi(val); err == nil {
			c.DefaultShardIndex = idx
		}
	}

	if val := os.Getenv("DB_SHARD_ENABLE_METRICS"); val != "" {
		c.EnableMetrics = parseBool(val)
	}

	if val := os.Getenv("DB_SHARD_ENABLE_TRACING"); val != "" {
		c.EnableTracing = parseBool(val)
	}

	if val := os.Getenv("DB_SHARD_LOG_LEVEL"); val != "" {
		c.LogLevel = val
	}

	return nil
}

// parseBool parses a boolean string
func parseBool(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "1" || s == "yes" || s == "on"
}
