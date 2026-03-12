package marketplace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Manager provides a high-level interface for agent and workflow management.
// It combines registry, installer, and validator functionality.
type Manager struct {
	agentRegistry    AgentRegistry
	workflowRegistry WorkflowRegistry
	installer        *Installer
	validator        *Validator
}

// NewManager creates a new marketplace manager.
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".plumego")

	// Create registries
	agentRegistry, err := NewLocalAgentRegistry(filepath.Join(baseDir, "agents"))
	if err != nil {
		return nil, fmt.Errorf("failed to create agent registry: %w", err)
	}

	workflowRegistry, err := NewLocalWorkflowRegistry(filepath.Join(baseDir, "workflows"))
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow registry: %w", err)
	}

	// Create installer
	installer, err := NewInstaller(agentRegistry, workflowRegistry, filepath.Join(baseDir, "installed"))
	if err != nil {
		return nil, fmt.Errorf("failed to create installer: %w", err)
	}

	return &Manager{
		agentRegistry:    agentRegistry,
		workflowRegistry: workflowRegistry,
		installer:        installer,
		validator:        NewValidator(),
	}, nil
}

// NewManagerWithPath creates a manager with custom base directory.
func NewManagerWithPath(basePath string) (*Manager, error) {
	// Create registries
	agentRegistry, err := NewLocalAgentRegistry(filepath.Join(basePath, "agents"))
	if err != nil {
		return nil, fmt.Errorf("failed to create agent registry: %w", err)
	}

	workflowRegistry, err := NewLocalWorkflowRegistry(filepath.Join(basePath, "workflows"))
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow registry: %w", err)
	}

	// Create installer
	installer, err := NewInstaller(agentRegistry, workflowRegistry, filepath.Join(basePath, "installed"))
	if err != nil {
		return nil, fmt.Errorf("failed to create installer: %w", err)
	}

	return &Manager{
		agentRegistry:    agentRegistry,
		workflowRegistry: workflowRegistry,
		installer:        installer,
		validator:        NewValidator(),
	}, nil
}

// PublishAgent publishes an agent to the marketplace after validation.
func (m *Manager) PublishAgent(ctx context.Context, metadata *AgentMetadata) error {
	// Validate
	result := m.validator.ValidateAgent(metadata)
	if !result.IsValid() {
		return &result.Errors[0]
	}

	// Publish
	return m.agentRegistry.Publish(ctx, metadata)
}

// PublishWorkflow publishes a workflow to the marketplace after validation.
func (m *Manager) PublishWorkflow(ctx context.Context, template *WorkflowTemplate) error {
	// Validate
	result := m.validator.ValidateWorkflow(template)
	if !result.IsValid() {
		return &result.Errors[0]
	}

	// Publish
	return m.workflowRegistry.Publish(ctx, template)
}

// InstallAgent installs an agent and its dependencies.
func (m *Manager) InstallAgent(ctx context.Context, id string, version string) error {
	return m.installer.InstallAgent(ctx, id, version)
}

// UninstallAgent removes an installed agent.
func (m *Manager) UninstallAgent(ctx context.Context, id string, version string) error {
	return m.installer.UninstallAgent(ctx, id, version)
}

// InstallWorkflow installs a workflow template.
func (m *Manager) InstallWorkflow(ctx context.Context, id string, version string) error {
	return m.installer.InstallWorkflow(ctx, id, version)
}

// UninstallWorkflow removes an installed workflow.
func (m *Manager) UninstallWorkflow(ctx context.Context, id string, version string) error {
	return m.installer.UninstallWorkflow(ctx, id, version)
}

// SearchAgents searches for agents in the marketplace.
func (m *Manager) SearchAgents(ctx context.Context, query SearchQuery) ([]*AgentMetadata, error) {
	return m.agentRegistry.Search(ctx, query)
}

// SearchWorkflows searches for workflows in the marketplace.
func (m *Manager) SearchWorkflows(ctx context.Context, query SearchQuery) ([]*WorkflowTemplate, error) {
	return m.workflowRegistry.Search(ctx, query)
}

// GetAgent retrieves agent metadata.
func (m *Manager) GetAgent(ctx context.Context, id string, version string) (*AgentMetadata, error) {
	return m.agentRegistry.Get(ctx, id, version)
}

// GetWorkflow retrieves workflow template.
func (m *Manager) GetWorkflow(ctx context.Context, id string, version string) (*WorkflowTemplate, error) {
	return m.workflowRegistry.Get(ctx, id, version)
}

// ListInstalledAgents lists all installed agents.
func (m *Manager) ListInstalledAgents() ([]*AgentMetadata, error) {
	return m.installer.ListInstalledAgents()
}

// ListInstalledWorkflows lists all installed workflows.
func (m *Manager) ListInstalledWorkflows() ([]*WorkflowTemplate, error) {
	return m.installer.ListInstalledWorkflows()
}

// IsAgentInstalled checks if an agent is installed.
func (m *Manager) IsAgentInstalled(id string, version string) bool {
	return m.installer.IsAgentInstalled(id, version)
}

// IsWorkflowInstalled checks if a workflow is installed.
func (m *Manager) IsWorkflowInstalled(id string, version string) bool {
	return m.installer.IsWorkflowInstalled(id, version)
}

// ListAgentVersions lists all available versions of an agent.
func (m *Manager) ListAgentVersions(ctx context.Context, id string) ([]string, error) {
	return m.agentRegistry.ListVersions(ctx, id)
}

// ListWorkflowVersions lists all available versions of a workflow.
func (m *Manager) ListWorkflowVersions(ctx context.Context, id string) ([]string, error) {
	return m.workflowRegistry.ListVersions(ctx, id)
}

// RateAgent adds a rating for an agent.
func (m *Manager) RateAgent(ctx context.Context, rating *Rating) error {
	// Validate rating
	result := m.validator.ValidateRating(rating)
	if !result.IsValid() {
		return &result.Errors[0]
	}

	return m.agentRegistry.Rate(ctx, rating)
}

// RateWorkflow adds a rating for a workflow.
func (m *Manager) RateWorkflow(ctx context.Context, rating *Rating) error {
	// Validate rating
	result := m.validator.ValidateRating(rating)
	if !result.IsValid() {
		return &result.Errors[0]
	}

	return m.workflowRegistry.Rate(ctx, rating)
}

// GetAgentRatings retrieves ratings for an agent.
func (m *Manager) GetAgentRatings(ctx context.Context, agentID string) ([]*Rating, error) {
	return m.agentRegistry.GetRatings(ctx, agentID)
}

// GetWorkflowRatings retrieves ratings for a workflow.
func (m *Manager) GetWorkflowRatings(ctx context.Context, workflowID string) ([]*Rating, error) {
	return m.workflowRegistry.GetRatings(ctx, workflowID)
}

// DeleteAgent removes an agent from the marketplace.
func (m *Manager) DeleteAgent(ctx context.Context, id string, version string) error {
	return m.agentRegistry.Delete(ctx, id, version)
}

// DeleteWorkflow removes a workflow from the marketplace.
func (m *Manager) DeleteWorkflow(ctx context.Context, id string, version string) error {
	return m.workflowRegistry.Delete(ctx, id, version)
}

// UpdateAgent updates an installed agent to a new version.
func (m *Manager) UpdateAgent(ctx context.Context, id string, newVersion string) error {
	return m.installer.UpdateAgent(ctx, id, newVersion)
}

// UpdateWorkflow updates an installed workflow to a new version.
func (m *Manager) UpdateWorkflow(ctx context.Context, id string, newVersion string) error {
	return m.installer.UpdateWorkflow(ctx, id, newVersion)
}

// Installer returns the underlying installer for advanced operations.
func (m *Manager) Installer() *Installer {
	return m.installer
}

// AgentRegistry returns the agent registry for direct access.
func (m *Manager) AgentRegistry() AgentRegistry {
	return m.agentRegistry
}

// WorkflowRegistry returns the workflow registry for direct access.
func (m *Manager) WorkflowRegistry() WorkflowRegistry {
	return m.workflowRegistry
}

// Validator returns the validator for direct access.
func (m *Manager) Validator() *Validator {
	return m.validator
}
