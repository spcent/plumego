package marketplace

import "errors"

// Common errors.
var (
	// ErrAgentNotFound is returned when an agent is not found in the registry.
	ErrAgentNotFound = errors.New("agent not found")

	// ErrWorkflowNotFound is returned when a workflow is not found in the registry.
	ErrWorkflowNotFound = errors.New("workflow not found")

	// ErrDuplicateAgent is returned when trying to publish an agent that already exists.
	ErrDuplicateAgent = errors.New("agent with this ID and version already exists")

	// ErrDuplicateWorkflow is returned when trying to publish a workflow that already exists.
	ErrDuplicateWorkflow = errors.New("workflow with this ID and version already exists")

	// ErrInvalidVersion is returned when a version string is not valid semantic versioning.
	ErrInvalidVersion = errors.New("invalid semantic version")

	// ErrDependencyNotFound is returned when a dependency cannot be resolved.
	ErrDependencyNotFound = errors.New("dependency not found")
)
