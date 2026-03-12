# Function Calling (Tools)

> **Package**: `github.com/spcent/plumego/x/ai/tool` | **Feature**: Agent capabilities

## Overview

The `tool` package provides a framework for defining and executing AI agent tools (function calling). Tools allow LLMs to interact with external systems — querying databases, calling APIs, reading files, and more.

## Quick Start

```go
import "github.com/spcent/plumego/x/ai/tool"

// Define a tool
getWeather := tool.New("get_weather",
    "Get current weather for a location",
    tool.Schema{
        Type: "object",
        Properties: map[string]tool.Property{
            "location": {
                Type:        "string",
                Description: "City name or coordinates",
            },
            "units": {
                Type:        "string",
                Enum:        []string{"celsius", "fahrenheit"},
                Description: "Temperature units",
            },
        },
        Required: []string{"location"},
    },
    func(ctx context.Context, input map[string]any) (string, error) {
        location := input["location"].(string)
        units, _ := input["units"].(string)
        if units == "" {
            units = "celsius"
        }
        // Call weather API
        weather, err := weatherAPI.Get(location, units)
        if err != nil {
            return "", err
        }
        result, _ := json.Marshal(weather)
        return string(result), nil
    },
)
```

## Tool Registry

```go
// Register tools globally
registry := tool.NewRegistry()
registry.Register(getWeather)
registry.Register(searchDatabase)
registry.Register(readFile)
registry.Register(sendEmail)

// Use in LLM request
req := &provider.CompletionRequest{
    Model:    "claude-opus-4-6",
    Messages: messages,
    Tools:    registry.ProviderTools(), // Convert to provider format
}
```

## Tool Definition

### Simple Tool

```go
searchWeb := tool.New("search_web",
    "Search the web for current information",
    tool.Schema{
        Type: "object",
        Properties: map[string]tool.Property{
            "query": {
                Type:        "string",
                Description: "Search query",
            },
            "max_results": {
                Type:        "integer",
                Description: "Maximum number of results",
                Default:     5,
            },
        },
        Required: []string{"query"},
    },
    func(ctx context.Context, input map[string]any) (string, error) {
        query := input["query"].(string)
        maxResults := 5
        if n, ok := input["max_results"].(float64); ok {
            maxResults = int(n)
        }

        results, err := searchEngine.Search(ctx, query, maxResults)
        if err != nil {
            return "", fmt.Errorf("search failed: %w", err)
        }

        out, _ := json.Marshal(results)
        return string(out), nil
    },
)
```

### Database Query Tool

```go
queryDatabase := tool.New("query_database",
    "Execute a read-only SQL query against the application database",
    tool.Schema{
        Type: "object",
        Properties: map[string]tool.Property{
            "sql": {
                Type:        "string",
                Description: "SELECT SQL query",
            },
        },
        Required: []string{"sql"},
    },
    func(ctx context.Context, input map[string]any) (string, error) {
        query := input["sql"].(string)

        // Security: only allow SELECT statements
        if !strings.HasPrefix(strings.TrimSpace(strings.ToUpper(query)), "SELECT") {
            return "", errors.New("only SELECT queries are allowed")
        }

        rows, err := db.QueryContext(ctx, query)
        if err != nil {
            return "", fmt.Errorf("query error: %w", err)
        }
        defer rows.Close()

        // Serialize results
        columns, _ := rows.Columns()
        var results []map[string]any

        for rows.Next() {
            values := make([]any, len(columns))
            valuePtrs := make([]any, len(columns))
            for i := range values {
                valuePtrs[i] = &values[i]
            }
            rows.Scan(valuePtrs...)

            row := make(map[string]any)
            for i, col := range columns {
                row[col] = values[i]
            }
            results = append(results, row)
        }

        out, _ := json.Marshal(results)
        return string(out), nil
    },
)
```

## Agentic Loop

Execute tools in a loop until the LLM completes without requesting tools:

```go
func runAgentLoop(ctx context.Context, messages []provider.Message, tools *tool.Registry) (string, error) {
    const maxIterations = 10

    for i := 0; i < maxIterations; i++ {
        // Call LLM with tools
        resp, err := claude.Complete(ctx, &provider.CompletionRequest{
            Model:      "claude-opus-4-6",
            Messages:   messages,
            Tools:      tools.ProviderTools(),
            ToolChoice: provider.ToolChoiceAuto,
            MaxTokens:  4096,
        })
        if err != nil {
            return "", err
        }

        // Add assistant response to history
        messages = append(messages, provider.Message{
            Role:    "assistant",
            Content: resp.RawContent,
        })

        // Check if done (no tool calls)
        if resp.StopReason == "end_turn" {
            // Extract text from final response
            for _, block := range resp.Content {
                if block.Type == "text" {
                    return block.Text, nil
                }
            }
            return "", nil
        }

        // Execute tool calls
        var toolResults []provider.ContentPart
        for _, block := range resp.Content {
            if block.Type != "tool_use" {
                continue
            }

            result, err := tools.Execute(ctx, block.Name, block.Input)
            if err != nil {
                result = fmt.Sprintf("Error: %v", err)
            }

            toolResults = append(toolResults, provider.ContentPart{
                Type:      "tool_result",
                ToolUseID: block.ID,
                Content:   result,
            })
        }

        // Add tool results to history
        messages = append(messages, provider.Message{
            Role:  "user",
            Parts: toolResults,
        })
    }

    return "", errors.New("max iterations reached")
}
```

## Built-in Tools

```go
import "github.com/spcent/plumego/x/ai/tool/builtin"

registry := tool.NewRegistry()

// HTTP request tool
registry.Register(builtin.HTTPTool(builtin.HTTPConfig{
    AllowedHosts: []string{"api.example.com"},
    Timeout:      10 * time.Second,
}))

// Time tool
registry.Register(builtin.TimeTool())

// JSON manipulation tool
registry.Register(builtin.JSONTool())
```

## Tool Schema Reference

```go
tool.Schema{
    Type: "object",
    Properties: map[string]tool.Property{
        // String
        "name": {
            Type:        "string",
            Description: "User's full name",
            MinLength:   1,
            MaxLength:   100,
        },
        // Number
        "amount": {
            Type:        "number",
            Description: "Amount in USD",
            Minimum:     0.01,
            Maximum:     10000.00,
        },
        // Integer
        "page": {
            Type:        "integer",
            Description: "Page number",
            Minimum:     1,
            Default:     1,
        },
        // Boolean
        "include_deleted": {
            Type:        "boolean",
            Description: "Include deleted records",
            Default:     false,
        },
        // Enum
        "status": {
            Type:        "string",
            Enum:        []string{"active", "inactive", "pending"},
            Description: "Account status",
        },
        // Array
        "tags": {
            Type:        "array",
            Description: "Tag list",
            Items:       &tool.Property{Type: "string"},
        },
        // Nested object
        "address": {
            Type:        "object",
            Description: "Mailing address",
            Properties: map[string]tool.Property{
                "street": {Type: "string"},
                "city":   {Type: "string"},
                "zip":    {Type: "string"},
            },
        },
    },
    Required: []string{"name", "amount"},
}
```

## Tool Choice

```go
// Auto: LLM decides whether to use tools
ToolChoice: provider.ToolChoiceAuto

// Any: LLM must use at least one tool
ToolChoice: provider.ToolChoiceAny

// Specific tool: Force use of a specific tool
ToolChoice: provider.ToolChoiceSpecific("get_weather")

// None: Disable tool use for this request
ToolChoice: provider.ToolChoiceNone
```

## Error Handling in Tools

```go
func(ctx context.Context, input map[string]any) (string, error) {
    id, ok := input["id"].(string)
    if !ok {
        return "", errors.New("id must be a string")
    }

    record, err := db.Get(ctx, id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            // Return meaningful error for the LLM to reason about
            return "", fmt.Errorf("record %q not found", id)
        }
        return "", fmt.Errorf("database error: %w", err)
    }

    out, _ := json.Marshal(record)
    return string(out), nil
}
```

## Observability

```go
// Wrap tool execution with logging
registry.Use(tool.LoggingMiddleware(logger))
registry.Use(tool.MetricsMiddleware(metrics))
registry.Use(tool.TracingMiddleware(tracer))
```

## Best Practices

### 1. Single Responsibility
```go
// ✅ Good: One clear purpose
tool.New("get_user_email", "Get email for a user ID", ...)

// ❌ Bad: Multiple unrelated concerns
tool.New("get_user_and_send_email", "Get user and send them an email", ...)
```

### 2. Clear Descriptions
```go
// ✅ Good: Clear, specific description
Description: "Search orders by customer ID and return order details including status and total"

// ❌ Bad: Vague
Description: "Get orders"
```

### 3. Validate Inputs
```go
func(ctx context.Context, input map[string]any) (string, error) {
    // Always validate before using
    email, ok := input["email"].(string)
    if !ok || !strings.Contains(email, "@") {
        return "", errors.New("valid email address required")
    }
    // ...
}
```

## Related Documentation

- [Provider Interface](provider.md) — Tool integration with providers
- [Orchestration](orchestration.md) — Multi-step tool workflows
- [Session Management](session.md) — Tool call history in sessions
