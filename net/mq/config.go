package mq

import (
	"fmt"
	"strings"
	"time"
)

// Config holds broker configuration.
type Config struct {
	// EnableHealthCheck enables health check endpoint
	EnableHealthCheck bool

	// MaxTopics limits the number of topics (0 = no limit)
	MaxTopics int

	// MaxSubscribers limits the number of subscribers per topic (0 = no limit)
	MaxSubscribers int

	// DefaultBufferSize is the default buffer size for new subscriptions
	DefaultBufferSize int

	// EnableMetrics enables metrics collection
	EnableMetrics bool

	// HealthCheckInterval is the interval for health checks (default: 30s)
	HealthCheckInterval time.Duration

	// MessageTTL is the default time-to-live for messages (0 = no TTL)
	MessageTTL time.Duration

	// EnablePriorityQueue enables priority queue support
	EnablePriorityQueue bool

	// EnableAckSupport enables message acknowledgment support
	EnableAckSupport bool

	// DefaultAckTimeout is the default timeout for acknowledgment (default: 30s)
	DefaultAckTimeout time.Duration

	// MaxMemoryUsage limits memory usage in bytes (0 = no limit)
	MaxMemoryUsage uint64

	// EnableTriePattern enables Trie-based pattern matching
	EnableTriePattern bool

	// EnableCluster enables distributed cluster mode
	EnableCluster bool

	// ClusterNodeID is the unique identifier for this node in the cluster
	ClusterNodeID string

	// ClusterNodes is the list of peer nodes in the cluster (format: "node-id@host:port")
	ClusterNodes []string

	// ClusterReplicationFactor is the number of replicas for each message (default: 1)
	ClusterReplicationFactor int

	// ClusterSyncInterval is the interval for cluster state synchronization (default: 5s)
	ClusterSyncInterval time.Duration

	// EnablePersistence enables persistent storage backend
	EnablePersistence bool

	// PersistencePath is the directory path for persistent storage
	PersistencePath string

	// EnableDeadLetterQueue enables dead letter queue support
	EnableDeadLetterQueue bool

	// DeadLetterTopic is the topic for dead letter messages
	DeadLetterTopic string

	// EnableTransactions enables transaction support
	EnableTransactions bool

	// TransactionTimeout is the timeout for transactions (default: 30s)
	TransactionTimeout time.Duration

	// EnableMQTT enables MQTT protocol support
	EnableMQTT bool

	// MQTTPort is the port for MQTT protocol (default: 1883)
	MQTTPort int

	// EnableAMQP enables AMQP protocol support
	EnableAMQP bool

	// AMQPPort is the port for AMQP protocol (default: 5672)
	AMQPPort int
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		EnableHealthCheck:        true,
		MaxTopics:                0, // No limit
		MaxSubscribers:           0, // No limit
		DefaultBufferSize:        DefaultBufferSize,
		EnableMetrics:            true,
		HealthCheckInterval:      DefaultHealthCheckInterval,
		MessageTTL:               0, // No TTL by default
		EnablePriorityQueue:      true,
		EnableAckSupport:         false, // Disabled by default for backward compatibility
		DefaultAckTimeout:        DefaultAckTimeoutDuration,
		MaxMemoryUsage:           0,     // No limit by default
		EnableTriePattern:        false, // Disabled by default for backward compatibility
		EnableCluster:            false, // Disabled by default
		ClusterReplicationFactor: 1,     // No replication by default
		ClusterSyncInterval:      DefaultClusterSyncInterval,
		EnablePersistence:        false, // Disabled by default
		EnableDeadLetterQueue:    false, // Disabled by default
		EnableTransactions:       false, // Disabled by default
		TransactionTimeout:       DefaultTransactionTimeoutDuration,
		EnableMQTT:               false, // Disabled by default
		MQTTPort:                 DefaultMQTTPort,
		EnableAMQP:               false, // Disabled by default
		AMQPPort:                 DefaultAMQPPort,
	}
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	// Basic validation
	if c.DefaultBufferSize <= 0 {
		return fmt.Errorf("%w: DefaultBufferSize must be positive", ErrInvalidConfig)
	}
	if c.HealthCheckInterval < 0 {
		return fmt.Errorf("%w: HealthCheckInterval cannot be negative", ErrInvalidConfig)
	}
	if c.MaxTopics < 0 {
		return fmt.Errorf("%w: MaxTopics cannot be negative", ErrInvalidConfig)
	}
	if c.MaxSubscribers < 0 {
		return fmt.Errorf("%w: MaxSubscribers cannot be negative", ErrInvalidConfig)
	}

	// Cluster configuration validation
	if c.EnableCluster {
		if strings.TrimSpace(c.ClusterNodeID) == "" {
			return fmt.Errorf("%w: ClusterNodeID is required when cluster mode is enabled", ErrInvalidConfig)
		}
		if c.ClusterReplicationFactor < 1 {
			return fmt.Errorf("%w: ClusterReplicationFactor must be at least 1", ErrInvalidConfig)
		}
		if c.ClusterSyncInterval < 0 {
			return fmt.Errorf("%w: ClusterSyncInterval cannot be negative", ErrInvalidConfig)
		}
	}

	// Persistence configuration validation
	if c.EnablePersistence {
		if strings.TrimSpace(c.PersistencePath) == "" {
			return fmt.Errorf("%w: PersistencePath is required when persistence is enabled", ErrInvalidConfig)
		}
	}

	// Dead letter queue configuration validation
	if c.EnableDeadLetterQueue {
		if strings.TrimSpace(c.DeadLetterTopic) == "" {
			return fmt.Errorf("%w: DeadLetterTopic is required when dead letter queue is enabled", ErrInvalidConfig)
		}
	}

	// Acknowledgment configuration validation
	if c.EnableAckSupport {
		if c.DefaultAckTimeout < 0 {
			return fmt.Errorf("%w: DefaultAckTimeout cannot be negative", ErrInvalidConfig)
		}
	}

	// Transaction configuration validation
	if c.EnableTransactions {
		if c.TransactionTimeout < 0 {
			return fmt.Errorf("%w: TransactionTimeout cannot be negative", ErrInvalidConfig)
		}
	}

	// Protocol configuration validation
	if c.EnableMQTT {
		if c.MQTTPort <= 0 || c.MQTTPort > 65535 {
			return fmt.Errorf("%w: MQTTPort must be between 1 and 65535", ErrInvalidConfig)
		}
	}
	if c.EnableAMQP {
		if c.AMQPPort <= 0 || c.AMQPPort > 65535 {
			return fmt.Errorf("%w: AMQPPort must be between 1 and 65535", ErrInvalidConfig)
		}
	}

	return nil
}
