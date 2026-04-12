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

type forcePrimaryContextKey struct{}
type preferReplicaContextKey struct{}
type inTransactionContextKey struct{}

// WithForcePrimary returns a context that forces all queries to use the primary database.
func WithForcePrimary(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, forcePrimaryContextKey{}, true)
}

// WithPreferReplica returns a context that prefers routing to replicas.
func WithPreferReplica(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, preferReplicaContextKey{}, true)
}

// ForcePrimaryFromContext checks if the context forces primary usage.
func ForcePrimaryFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	v, ok := ctx.Value(forcePrimaryContextKey{}).(bool)
	return ok && v
}

// PreferReplicaFromContext checks if the context prefers replicas.
func PreferReplicaFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	v, ok := ctx.Value(preferReplicaContextKey{}).(bool)
	return ok && v
}

// WithInTransaction marks the context as being in a transaction.
func WithInTransaction(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, inTransactionContextKey{}, true)
}

// InTransactionFromContext checks if the context is in a transaction.
func InTransactionFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	v, ok := ctx.Value(inTransactionContextKey{}).(bool)
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
	if ForcePrimaryFromContext(ctx) {
		return true
	}

	if PreferReplicaFromContext(ctx) {
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
	if InTransactionFromContext(ctx) {
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
