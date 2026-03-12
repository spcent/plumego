package marketplace

import (
	"context"
)

// AgentRegistry defines the interface for agent registry operations.
type AgentRegistry interface {
	// Publish publishes an agent to the registry.
	Publish(ctx context.Context, metadata *AgentMetadata) error

	// Get retrieves agent metadata by ID and version.
	Get(ctx context.Context, id string, version string) (*AgentMetadata, error)

	// Search searches for agents matching the query.
	Search(ctx context.Context, query SearchQuery) ([]*AgentMetadata, error)

	// ListVersions lists all versions of an agent.
	ListVersions(ctx context.Context, id string) ([]string, error)

	// Delete removes an agent from the registry.
	Delete(ctx context.Context, id string, version string) error

	// Rate adds a rating for an agent.
	Rate(ctx context.Context, rating *Rating) error

	// GetRatings retrieves ratings for an agent.
	GetRatings(ctx context.Context, agentID string) ([]*Rating, error)

	// ResolveDependencies resolves all dependencies for an agent.
	ResolveDependencies(ctx context.Context, id string, version string) ([]*AgentMetadata, error)

	// UpdateDownloadCount increments download counter.
	UpdateDownloadCount(ctx context.Context, id string) error
}

// WorkflowRegistry defines the interface for workflow registry operations.
type WorkflowRegistry interface {
	// Publish publishes a workflow template to the registry.
	Publish(ctx context.Context, template *WorkflowTemplate) error

	// Get retrieves workflow template by ID and version.
	Get(ctx context.Context, id string, version string) (*WorkflowTemplate, error)

	// Search searches for workflows matching the query.
	Search(ctx context.Context, query SearchQuery) ([]*WorkflowTemplate, error)

	// ListVersions lists all versions of a workflow.
	ListVersions(ctx context.Context, id string) ([]string, error)

	// Delete removes a workflow from the registry.
	Delete(ctx context.Context, id string, version string) error

	// Rate adds a rating for a workflow.
	Rate(ctx context.Context, rating *Rating) error

	// GetRatings retrieves ratings for a workflow.
	GetRatings(ctx context.Context, workflowID string) ([]*Rating, error)
}
