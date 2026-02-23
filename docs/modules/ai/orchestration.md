# AI Workflow Orchestration

> **Package**: `github.com/spcent/plumego/ai/orchestration` | **Feature**: Multi-step AI workflows

## Overview

The `orchestration` package enables multi-step AI workflows — coordinating sequences of LLM calls, tool executions, and data transformations into reliable pipelines.

## Quick Start

```go
import "github.com/spcent/plumego/ai/orchestration"

// Define workflow steps
workflow := orchestration.NewWorkflow("research-and-summarize")

workflow.Step("search", func(ctx context.Context, input any) (any, error) {
    query := input.(string)
    return searchEngine.Search(ctx, query)
})

workflow.Step("analyze", func(ctx context.Context, input any) (any, error) {
    results := input.([]SearchResult)
    text := formatResults(results)

    resp, err := claude.Complete(ctx, &provider.CompletionRequest{
        Model:   "claude-opus-4-6",
        Messages: []provider.Message{
            {Role: "user", Content: "Analyze these search results: " + text},
        },
    })
    return resp.Content[0].Text, err
})

workflow.Step("format", func(ctx context.Context, input any) (any, error) {
    analysis := input.(string)
    return formatMarkdown(analysis), nil
})

// Execute
result, err := workflow.Run(ctx, "Go generics best practices")
```

## Workflow Patterns

### Sequential Pipeline

```go
pipeline := orchestration.Pipeline(
    orchestration.Step("extract", extractDataFromRequest),
    orchestration.Step("validate", validateExtractedData),
    orchestration.Step("enrich", enrichWithLLM),
    orchestration.Step("store", saveToDatabase),
)

result, err := pipeline.Execute(ctx, input)
```

### Parallel Execution

```go
// Execute multiple steps concurrently
parallel := orchestration.Parallel(
    orchestration.Step("get_user", fetchUserProfile),
    orchestration.Step("get_orders", fetchUserOrders),
    orchestration.Step("get_prefs", fetchUserPreferences),
)

// All results available after parallel execution
results, err := parallel.Execute(ctx, userID)
userProfile := results["get_user"]
orders := results["get_orders"]
```

### Map-Reduce

```go
// Process each item through LLM, then aggregate
workflow := orchestration.MapReduce(
    // Map: process each document
    func(ctx context.Context, doc Document) (string, error) {
        resp, err := claude.Complete(ctx, &provider.CompletionRequest{
            Model:    "claude-haiku-4-5-20251001", // Fast model for map phase
            Messages: []provider.Message{
                {Role: "user", Content: "Summarize: " + doc.Content},
            },
            MaxTokens: 200,
        })
        if err != nil {
            return "", err
        }
        return resp.Content[0].Text, nil
    },

    // Reduce: combine summaries
    func(ctx context.Context, summaries []string) (string, error) {
        combined := strings.Join(summaries, "\n\n")
        resp, err := claude.Complete(ctx, &provider.CompletionRequest{
            Model:    "claude-opus-4-6",
            Messages: []provider.Message{
                {Role: "user", Content: "Synthesize these summaries: " + combined},
            },
            MaxTokens: 1024,
        })
        if err != nil {
            return "", err
        }
        return resp.Content[0].Text, nil
    },
)

// Process 100 documents concurrently
result, err := workflow.Execute(ctx, documents,
    orchestration.WithConcurrency(10),
)
```

### Conditional Branching

```go
workflow := orchestration.NewWorkflow("smart-router")

workflow.Step("classify", func(ctx context.Context, input any) (any, error) {
    query := input.(string)

    resp, err := claude.Complete(ctx, &provider.CompletionRequest{
        Model:    "claude-haiku-4-5-20251001",
        Messages: []provider.Message{
            {Role: "user", Content: fmt.Sprintf(
                "Classify this query as 'technical', 'billing', or 'general': %s\nReply with just the category.",
                query,
            )},
        },
        MaxTokens: 10,
    })
    return strings.TrimSpace(resp.Content[0].Text), err
})

workflow.Branch("classify",
    orchestration.When("technical", handleTechnicalQuery),
    orchestration.When("billing", handleBillingQuery),
    orchestration.Default(handleGeneralQuery),
)

result, err := workflow.Run(ctx, userQuery)
```

## Orchestrator

The `Orchestrator` manages multiple workflows with shared resources:

```go
orch := orchestration.NewOrchestrator(
    orchestration.WithProvider(claude),
    orchestration.WithTools(registry),
    orchestration.WithMaxConcurrency(50),
    orchestration.WithTimeout(120*time.Second),
)

// Register workflows
orch.Register("research", researchWorkflow)
orch.Register("summarize", summarizeWorkflow)
orch.Register("analyze", analyzeWorkflow)

// Execute named workflow
result, err := orch.Execute(ctx, "research", input)
```

## Retry and Error Handling

```go
workflow.Step("llm-call",
    func(ctx context.Context, input any) (any, error) {
        return claude.Complete(ctx, req)
    },
    orchestration.WithRetry(3, time.Second), // Retry 3x with 1s delay
    orchestration.WithTimeout(30*time.Second),
    orchestration.OnError(func(err error) any {
        // Return fallback value on permanent failure
        return "Unable to process request"
    }),
)
```

## State Management

```go
// Workflows can share state across steps
workflow := orchestration.NewWorkflow("stateful")
workflow.WithStateStore(stateStore)

workflow.Step("step1", func(ctx context.Context, input any) (any, error) {
    // Set state
    orchestration.SetState(ctx, "intermediate_result", computedValue)
    return result, nil
})

workflow.Step("step2", func(ctx context.Context, input any) (any, error) {
    // Get state from previous step
    intermediate := orchestration.GetState[string](ctx, "intermediate_result")
    return processWithContext(input, intermediate), nil
})
```

## Streaming Orchestration

```go
// Stream results as each step completes
streamWorkflow := orchestration.NewStreamingWorkflow("progressive-answer")

streamWorkflow.StreamStep("outline", func(ctx context.Context, input any, events chan<- string) error {
    resp, err := claude.Complete(ctx, outlineRequest(input.(string)))
    if err != nil {
        return err
    }
    events <- fmt.Sprintf("## Outline\n%s\n\n", resp.Content[0].Text)
    return nil
})

streamWorkflow.StreamStep("details", func(ctx context.Context, input any, events chan<- string) error {
    // Stream each section as it's generated
    reader, _ := claude.CompleteStream(ctx, detailsRequest(input.(string)))
    for {
        delta, err := reader.Next()
        if err == io.EOF {
            break
        }
        events <- delta.Text
    }
    return nil
})

// Stream to SSE
sseStream, _ := sse.NewStream(r.Context(), w)
events := make(chan string)

go streamWorkflow.Run(r.Context(), query, events)
for event := range events {
    sseStream.Send(sse.Event{Data: event})
}
```

## Observability

```go
orch := orchestration.NewOrchestrator(
    orchestration.WithTracing(tracer),
    orchestration.WithMetrics(metricsRegistry),
    orchestration.WithLogger(logger),
)

// Emits metrics:
// workflow_executions_total{workflow="research", status="success"}
// workflow_duration_seconds{workflow="research"}
// workflow_step_duration_seconds{workflow="research", step="search"}
```

## Related Documentation

- [Provider Interface](provider.md) — LLM integration
- [Tool Framework](tool.md) — Tool execution in workflows
- [Session Management](session.md) — State across workflow runs
- [Circuit Breaker](circuit-breaker.md) — Fault tolerance in workflows
