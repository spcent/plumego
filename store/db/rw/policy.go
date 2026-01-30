package rw

import (
	"context"
	"strings"
)

// RoutingPolicy defines how queries are routed between primary and replicas
type RoutingPolicy interface {
	// ShouldUsePrimary determines if a query should use the primary database
	ShouldUsePrimary(ctx context.Context, query string) bool

	// Name returns the policy name
	Name() string
}

// ctxKey is the type for context keys
type ctxKey int

const (
	// keyUsePrimary forces routing to primary
	keyUsePrimary ctxKey = iota

	// keyPreferReplica prefers routing to replica even for writes
	keyPreferReplica

	// keyInTransaction marks context as being in a transaction
	keyInTransaction
)

// ForcePrimary returns a context that forces all queries to use the primary database
func ForcePrimary(ctx context.Context) context.Context {
	return context.WithValue(ctx, keyUsePrimary, true)
}

// PreferReplica returns a context that prefers routing to replicas
func PreferReplica(ctx context.Context) context.Context {
	return context.WithValue(ctx, keyPreferReplica, true)
}

// IsForcePrimary checks if the context forces primary usage
func IsForcePrimary(ctx context.Context) bool {
	v, ok := ctx.Value(keyUsePrimary).(bool)
	return ok && v
}

// IsPreferReplica checks if the context prefers replicas
func IsPreferReplica(ctx context.Context) bool {
	v, ok := ctx.Value(keyPreferReplica).(bool)
	return ok && v
}

// MarkInTransaction marks the context as being in a transaction
func MarkInTransaction(ctx context.Context) context.Context {
	return context.WithValue(ctx, keyInTransaction, true)
}

// IsInTransaction checks if the context is in a transaction
func IsInTransaction(ctx context.Context) bool {
	v, ok := ctx.Value(keyInTransaction).(bool)
	return ok && v
}

// SQLTypePolicy routes queries based on SQL statement type
type SQLTypePolicy struct{}

// NewSQLTypePolicy creates a new SQL type-based routing policy
func NewSQLTypePolicy() *SQLTypePolicy {
	return &SQLTypePolicy{}
}

// ShouldUsePrimary determines routing based on SQL statement type
func (p *SQLTypePolicy) ShouldUsePrimary(ctx context.Context, query string) bool {
	// Check context hints first
	if IsForcePrimary(ctx) {
		return true
	}

	if IsPreferReplica(ctx) {
		return false
	}

	// Parse SQL and determine if it's a write operation
	return isWriteOperation(query)
}

// Name returns the policy name
func (p *SQLTypePolicy) Name() string {
	return "sql_type"
}

// isWriteOperation checks if a SQL query is a write operation
func isWriteOperation(query string) bool {
	// Normalize query
	query = strings.TrimSpace(query)
	if query == "" {
		return false
	}

	// Convert to uppercase for comparison
	upper := strings.ToUpper(query)

	// Write operations
	writeKeywords := []string{
		"INSERT",
		"UPDATE",
		"DELETE",
		"CREATE",
		"ALTER",
		"DROP",
		"TRUNCATE",
		"REPLACE",
	}

	for _, kw := range writeKeywords {
		if strings.HasPrefix(upper, kw) {
			return true
		}
	}

	// SELECT ... FOR UPDATE must go to primary
	if strings.HasPrefix(upper, "SELECT") && strings.Contains(upper, "FOR UPDATE") {
		return true
	}

	// All other queries (SELECT, SHOW, DESCRIBE, etc.) go to replicas
	return false
}

// TransactionAwarePolicy wraps another policy and ensures transactions use primary
type TransactionAwarePolicy struct {
	base RoutingPolicy
}

// NewTransactionAwarePolicy creates a transaction-aware policy wrapper
func NewTransactionAwarePolicy(base RoutingPolicy) *TransactionAwarePolicy {
	if base == nil {
		base = NewSQLTypePolicy()
	}
	return &TransactionAwarePolicy{base: base}
}

// ShouldUsePrimary routes to primary if in transaction
func (p *TransactionAwarePolicy) ShouldUsePrimary(ctx context.Context, query string) bool {
	// All queries in a transaction must use primary
	if IsInTransaction(ctx) {
		return true
	}

	return p.base.ShouldUsePrimary(ctx, query)
}

// Name returns the policy name
func (p *TransactionAwarePolicy) Name() string {
	return "transaction_aware(" + p.base.Name() + ")"
}

// CompositePolicy combines multiple policies with AND logic
type CompositePolicy struct {
	policies []RoutingPolicy
}

// NewCompositePolicy creates a composite policy
func NewCompositePolicy(policies ...RoutingPolicy) *CompositePolicy {
	return &CompositePolicy{policies: policies}
}

// ShouldUsePrimary returns true if any policy requires primary
func (p *CompositePolicy) ShouldUsePrimary(ctx context.Context, query string) bool {
	for _, policy := range p.policies {
		if policy.ShouldUsePrimary(ctx, query) {
			return true
		}
	}
	return false
}

// Name returns the policy name
func (p *CompositePolicy) Name() string {
	names := make([]string, len(p.policies))
	for i, policy := range p.policies {
		names[i] = policy.Name()
	}
	return "composite(" + strings.Join(names, ",") + ")"
}

// AlwaysPrimaryPolicy always routes to primary (useful for testing or maintenance)
type AlwaysPrimaryPolicy struct{}

// NewAlwaysPrimaryPolicy creates a policy that always uses primary
func NewAlwaysPrimaryPolicy() *AlwaysPrimaryPolicy {
	return &AlwaysPrimaryPolicy{}
}

// ShouldUsePrimary always returns true
func (p *AlwaysPrimaryPolicy) ShouldUsePrimary(ctx context.Context, query string) bool {
	return true
}

// Name returns the policy name
func (p *AlwaysPrimaryPolicy) Name() string {
	return "always_primary"
}
