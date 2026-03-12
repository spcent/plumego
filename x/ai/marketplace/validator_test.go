package marketplace

import (
	"testing"
	"time"
)

func TestValidator_ValidateAgent(t *testing.T) {
	validator := NewValidator()

	t.Run("Valid Agent", func(t *testing.T) {
		metadata := &AgentMetadata{
			ID:          "valid-agent",
			Name:        "Valid Agent",
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "This is a valid test agent with sufficient description length",
			Category:    CategoryDataAnalysis,
			Tags:        []string{"test"},
			Provider:    "claude",
			Model:       "claude-3-5-sonnet-20241022",
			Prompt: PromptTemplate{
				System: "You are a helpful assistant",
			},
			Config: AgentConfig{
				Temperature: 0.7,
				MaxTokens:   1000,
				TopP:        0.9,
			},
			License:   "MIT",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		result := validator.ValidateAgent(metadata)
		if !result.IsValid() {
			t.Errorf("expected valid agent, got errors: %v", result.Errors)
		}
	})

	t.Run("Missing ID", func(t *testing.T) {
		metadata := &AgentMetadata{
			Name:        "Test Agent",
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "Test description with sufficient length for validation",
			Category:    CategoryGeneral,
			Provider:    "claude",
			Model:       "test-model",
			Prompt:      PromptTemplate{System: "Test"},
			License:     "MIT",
		}

		result := validator.ValidateAgent(metadata)
		if result.IsValid() {
			t.Error("expected validation error for missing ID")
		}

		foundError := false
		for _, err := range result.Errors {
			if err.Field == "id" {
				foundError = true
				break
			}
		}
		if !foundError {
			t.Error("expected error for field 'id'")
		}
	})

	t.Run("Invalid ID Format", func(t *testing.T) {
		metadata := &AgentMetadata{
			ID:          "Invalid_Agent_ID",
			Name:        "Test Agent",
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "Test description with sufficient length for validation",
			Category:    CategoryGeneral,
			Provider:    "claude",
			Model:       "test-model",
			Prompt:      PromptTemplate{System: "Test"},
			License:     "MIT",
		}

		result := validator.ValidateAgent(metadata)
		if result.IsValid() {
			t.Error("expected validation error for invalid ID format")
		}
	})

	t.Run("Invalid Version", func(t *testing.T) {
		metadata := &AgentMetadata{
			ID:          "test-agent",
			Name:        "Test Agent",
			Version:     "invalid-version",
			Author:      "Test Author",
			Description: "Test description with sufficient length for validation",
			Category:    CategoryGeneral,
			Provider:    "claude",
			Model:       "test-model",
			Prompt:      PromptTemplate{System: "Test"},
			License:     "MIT",
		}

		result := validator.ValidateAgent(metadata)
		if result.IsValid() {
			t.Error("expected validation error for invalid version")
		}

		foundError := false
		for _, err := range result.Errors {
			if err.Field == "version" {
				foundError = true
				break
			}
		}
		if !foundError {
			t.Error("expected error for field 'version'")
		}
	})

	t.Run("Description Too Short", func(t *testing.T) {
		metadata := &AgentMetadata{
			ID:          "test-agent",
			Name:        "Test Agent",
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "Too short",
			Category:    CategoryGeneral,
			Provider:    "claude",
			Model:       "test-model",
			Prompt:      PromptTemplate{System: "Test"},
			License:     "MIT",
		}

		result := validator.ValidateAgent(metadata)
		if result.IsValid() {
			t.Error("expected validation error for short description")
		}
	})

	t.Run("Invalid Category", func(t *testing.T) {
		metadata := &AgentMetadata{
			ID:          "test-agent",
			Name:        "Test Agent",
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "Test description with sufficient length for validation",
			Category:    AgentCategory("invalid_category"),
			Provider:    "claude",
			Model:       "test-model",
			Prompt:      PromptTemplate{System: "Test"},
			License:     "MIT",
		}

		result := validator.ValidateAgent(metadata)
		if result.IsValid() {
			t.Error("expected validation error for invalid category")
		}
	})

	t.Run("Invalid Provider", func(t *testing.T) {
		metadata := &AgentMetadata{
			ID:          "test-agent",
			Name:        "Test Agent",
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "Test description with sufficient length for validation",
			Category:    CategoryGeneral,
			Provider:    "unknown-provider",
			Model:       "test-model",
			Prompt:      PromptTemplate{System: "Test"},
			License:     "MIT",
		}

		result := validator.ValidateAgent(metadata)
		if result.IsValid() {
			t.Error("expected validation error for invalid provider")
		}
	})

	t.Run("Invalid Temperature", func(t *testing.T) {
		metadata := &AgentMetadata{
			ID:          "test-agent",
			Name:        "Test Agent",
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "Test description with sufficient length for validation",
			Category:    CategoryGeneral,
			Provider:    "claude",
			Model:       "test-model",
			Prompt:      PromptTemplate{System: "Test"},
			Config: AgentConfig{
				Temperature: 3.0, // Invalid: > 2.0
			},
			License: "MIT",
		}

		result := validator.ValidateAgent(metadata)
		if result.IsValid() {
			t.Error("expected validation error for invalid temperature")
		}
	})

	t.Run("Invalid Dependencies", func(t *testing.T) {
		metadata := &AgentMetadata{
			ID:           "test-agent",
			Name:         "Test Agent",
			Version:      "1.0.0",
			Author:       "Test Author",
			Description:  "Test description with sufficient length for validation",
			Category:     CategoryGeneral,
			Provider:     "claude",
			Model:        "test-model",
			Prompt:       PromptTemplate{System: "Test"},
			Dependencies: []string{"invalid-dependency-format"},
			License:      "MIT",
		}

		result := validator.ValidateAgent(metadata)
		if result.IsValid() {
			t.Error("expected validation error for invalid dependency format")
		}
	})
}

func TestValidator_ValidateWorkflow(t *testing.T) {
	validator := NewValidator()

	t.Run("Valid Workflow", func(t *testing.T) {
		template := &WorkflowTemplate{
			ID:          "valid-workflow",
			Name:        "Valid Workflow",
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "This is a valid test workflow template with sufficient description",
			Category:    WorkflowCategoryDataPipeline,
			Tags:        []string{"test"},
			Steps: []StepTemplate{
				{
					Name:     "step1",
					Type:     "sequential",
					AgentRef: "agent-1@1.0.0",
				},
			},
			Variables: []VariableDefinition{
				{
					Name:     "input",
					Type:     "string",
					Required: true,
				},
			},
			Outputs: []OutputDefinition{
				{
					Name: "result",
					Type: "string",
				},
			},
			License: "MIT",
		}

		result := validator.ValidateWorkflow(template)
		if !result.IsValid() {
			t.Errorf("expected valid workflow, got errors: %v", result.Errors)
		}
	})

	t.Run("Missing ID", func(t *testing.T) {
		template := &WorkflowTemplate{
			Name:        "Test Workflow",
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "Test description with sufficient length for validation",
			Category:    WorkflowCategoryGeneral,
			Steps:       []StepTemplate{{Name: "step1", Type: "sequential"}},
			License:     "MIT",
		}

		result := validator.ValidateWorkflow(template)
		if result.IsValid() {
			t.Error("expected validation error for missing ID")
		}
	})

	t.Run("Invalid Version", func(t *testing.T) {
		template := &WorkflowTemplate{
			ID:          "test-workflow",
			Name:        "Test Workflow",
			Version:     "not-a-version",
			Author:      "Test Author",
			Description: "Test description with sufficient length for validation",
			Category:    WorkflowCategoryGeneral,
			Steps:       []StepTemplate{{Name: "step1", Type: "sequential"}},
			License:     "MIT",
		}

		result := validator.ValidateWorkflow(template)
		if result.IsValid() {
			t.Error("expected validation error for invalid version")
		}
	})

	t.Run("No Steps", func(t *testing.T) {
		template := &WorkflowTemplate{
			ID:          "test-workflow",
			Name:        "Test Workflow",
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "Test description with sufficient length for validation",
			Category:    WorkflowCategoryGeneral,
			Steps:       []StepTemplate{},
			License:     "MIT",
		}

		result := validator.ValidateWorkflow(template)
		if result.IsValid() {
			t.Error("expected validation error for missing steps")
		}
	})

	t.Run("Invalid Step Type", func(t *testing.T) {
		template := &WorkflowTemplate{
			ID:          "test-workflow",
			Name:        "Test Workflow",
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "Test description with sufficient length for validation",
			Category:    WorkflowCategoryGeneral,
			Steps: []StepTemplate{
				{
					Name: "step1",
					Type: "invalid-type",
				},
			},
			License: "MIT",
		}

		result := validator.ValidateWorkflow(template)
		if result.IsValid() {
			t.Error("expected validation error for invalid step type")
		}
	})

	t.Run("Missing Agent Reference", func(t *testing.T) {
		template := &WorkflowTemplate{
			ID:          "test-workflow",
			Name:        "Test Workflow",
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "Test description with sufficient length for validation",
			Category:    WorkflowCategoryGeneral,
			Steps: []StepTemplate{
				{
					Name: "step1",
					Type: "sequential",
					// Missing AgentRef
				},
			},
			License: "MIT",
		}

		result := validator.ValidateWorkflow(template)
		if result.IsValid() {
			t.Error("expected validation error for missing agent reference")
		}
	})

	t.Run("Invalid Variable", func(t *testing.T) {
		template := &WorkflowTemplate{
			ID:          "test-workflow",
			Name:        "Test Workflow",
			Version:     "1.0.0",
			Author:      "Test Author",
			Description: "Test description with sufficient length for validation",
			Category:    WorkflowCategoryGeneral,
			Steps:       []StepTemplate{{Name: "step1", Type: "sequential", AgentRef: "a@1.0.0"}},
			Variables: []VariableDefinition{
				{
					Name: "", // Empty name
					Type: "string",
				},
			},
			License: "MIT",
		}

		result := validator.ValidateWorkflow(template)
		if result.IsValid() {
			t.Error("expected validation error for invalid variable")
		}
	})
}

func TestValidator_ValidateRating(t *testing.T) {
	validator := NewValidator()

	t.Run("Valid Rating", func(t *testing.T) {
		rating := &Rating{
			TargetID:  "test-agent",
			UserID:    "user1",
			Score:     4.5,
			Comment:   "Great agent!",
			CreatedAt: time.Now(),
		}

		result := validator.ValidateRating(rating)
		if !result.IsValid() {
			t.Errorf("expected valid rating, got errors: %v", result.Errors)
		}
	})

	t.Run("Score Too Low", func(t *testing.T) {
		rating := &Rating{
			TargetID:  "test-agent",
			UserID:    "user1",
			Score:     0.5,
			CreatedAt: time.Now(),
		}

		result := validator.ValidateRating(rating)
		if result.IsValid() {
			t.Error("expected validation error for score too low")
		}
	})

	t.Run("Score Too High", func(t *testing.T) {
		rating := &Rating{
			TargetID:  "test-agent",
			UserID:    "user1",
			Score:     6.0,
			CreatedAt: time.Now(),
		}

		result := validator.ValidateRating(rating)
		if result.IsValid() {
			t.Error("expected validation error for score too high")
		}
	})

	t.Run("Missing Target ID", func(t *testing.T) {
		rating := &Rating{
			UserID:    "user1",
			Score:     4.0,
			CreatedAt: time.Now(),
		}

		result := validator.ValidateRating(rating)
		if result.IsValid() {
			t.Error("expected validation error for missing target ID")
		}
	})
}

func TestIsValidID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"my-agent", true},
		{"agent-123", true},
		{"data-analyzer", true},
		{"a", false},         // Too short
		{"ab", false},        // Too short
		{"abc", true},        // Minimum length
		{"My-Agent", false},  // Uppercase
		{"my_agent", false},  // Underscore
		{"my.agent", false},  // Dot
		{"123-agent", false}, // Starts with number
		{"-agent", false},    // Starts with hyphen
		{"agent-", true},     // Ends with hyphen (allowed)
		{"my--agent", true},  // Double hyphen (allowed)
	}

	for _, tt := range tests {
		result := isValidID(tt.id)
		if result != tt.valid {
			t.Errorf("isValidID(%q) = %v, want %v", tt.id, result, tt.valid)
		}
	}
}

func TestIsValidDependency(t *testing.T) {
	tests := []struct {
		dep   string
		valid bool
	}{
		{"agent-1@1.0.0", true},
		{"my-agent@2.1.3", true},
		{"agent", false},         // No version
		{"agent@", false},        // Empty version
		{"@1.0.0", false},        // No ID
		{"agent@invalid", false}, // Invalid version
		{"Agent@1.0.0", false},   // Invalid ID (uppercase)
		{"agent-1@1.0", true},    // Valid version (2 parts)
	}

	for _, tt := range tests {
		result := isValidDependency(tt.dep)
		if result != tt.valid {
			t.Errorf("isValidDependency(%q) = %v, want %v", tt.dep, result, tt.valid)
		}
	}
}
