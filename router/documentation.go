package router

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// RouteDocumentationGenerator generates comprehensive documentation for routes
type RouteDocumentationGenerator struct {
	mu              sync.RWMutex
	routes          []EnhancedRouteInfo
	middlewareNames map[string]string // middleware function -> name
}

// NewRouteDocumentationGenerator creates a new documentation generator
func NewRouteDocumentationGenerator() *RouteDocumentationGenerator {
	return &RouteDocumentationGenerator{
		routes:          make([]EnhancedRouteInfo, 0),
		middlewareNames: make(map[string]string),
	}
}

// RegisterRoute registers a route for documentation
func (rdg *RouteDocumentationGenerator) RegisterRoute(method, path string, meta EnhancedRouteMeta) {
	rdg.mu.Lock()
	defer rdg.mu.Unlock()

	// Check if route already exists
	for i, route := range rdg.routes {
		if route.Method == method && route.Path == path {
			rdg.routes[i] = EnhancedRouteInfo{
				Method:       method,
				Path:         path,
				BaseMeta:     meta.Base,
				EnhancedMeta: meta,
			}
			return
		}
	}

	// Add new route
	rdg.routes = append(rdg.routes, EnhancedRouteInfo{
		Method:       method,
		Path:         path,
		BaseMeta:     meta.Base,
		EnhancedMeta: meta,
	})
}

// RegisterMiddlewareName maps middleware functions to readable names
func (rdg *RouteDocumentationGenerator) RegisterMiddlewareName(middlewareFunc, name string) {
	rdg.mu.Lock()
	defer rdg.mu.Unlock()
	rdg.middlewareNames[middlewareFunc] = name
}

// GenerateMarkdown generates comprehensive markdown documentation
func (rdg *RouteDocumentationGenerator) GenerateMarkdown() string {
	var builder strings.Builder

	rdg.mu.RLock()
	defer rdg.mu.RUnlock()

	// Sort routes for consistent output
	sortedRoutes := make([]EnhancedRouteInfo, len(rdg.routes))
	copy(sortedRoutes, rdg.routes)
	sort.Slice(sortedRoutes, func(i, j int) bool {
		if sortedRoutes[i].Method == sortedRoutes[j].Method {
			return sortedRoutes[i].Path < sortedRoutes[j].Path
		}
		return sortedRoutes[i].Method < sortedRoutes[j].Method
	})

	builder.WriteString("# API Documentation\n\n")
	builder.WriteString("Comprehensive API documentation generated from route metadata.\n\n")

	// Table of Contents
	builder.WriteString("## Table of Contents\n\n")
	for _, route := range sortedRoutes {
		anchor := strings.ToLower(route.Method + "-" + strings.ReplaceAll(strings.Trim(route.Path, "/"), "/", "-"))
		title := fmt.Sprintf("%s %s", route.Method, route.Path)
		if route.EnhancedMeta.Summary != "" {
			title = route.EnhancedMeta.Summary
		}
		builder.WriteString(fmt.Sprintf("- [%s](#%s)\n", title, anchor))
	}
	builder.WriteString("\n")

	// Detailed Documentation
	for _, route := range sortedRoutes {
		builder.WriteString(fmt.Sprintf("## %s %s\n\n", route.Method, route.Path))

		// Summary and Description
		if route.EnhancedMeta.Summary != "" {
			builder.WriteString(fmt.Sprintf("**Summary:** %s\n\n", route.EnhancedMeta.Summary))
		}
		if route.EnhancedMeta.Description != "" {
			builder.WriteString(fmt.Sprintf("**Description:** %s\n\n", route.EnhancedMeta.Description))
		}
		if route.BaseMeta.Name != "" {
			builder.WriteString(fmt.Sprintf("**Route Name:** %s\n\n", route.BaseMeta.Name))
		}

		// Tags
		if len(route.EnhancedMeta.Tags) > 0 {
			builder.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(route.EnhancedMeta.Tags, ", ")))
		}

		// Operation ID
		if route.EnhancedMeta.OperationID != "" {
			builder.WriteString(fmt.Sprintf("**Operation ID:** `%s`\n\n", route.EnhancedMeta.OperationID))
		}

		// Parameters
		if len(route.EnhancedMeta.Parameters) > 0 {
			builder.WriteString("### Parameters\n\n")
			builder.WriteString("| Name | In | Type | Required | Description | Example |\n")
			builder.WriteString("|------|----|------|----------|-------------|---------|\n")
			for _, param := range route.EnhancedMeta.Parameters {
				example := ""
				if param.Example != nil {
					example = fmt.Sprintf("%v", param.Example)
				}
				required := "No"
				if param.Required {
					required = "Yes"
				}
				builder.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s | %s | %s |\n",
					param.Name, param.In, param.Type, required, param.Description, example))
			}
			builder.WriteString("\n")
		}

		// Responses
		if len(route.EnhancedMeta.Responses) > 0 {
			builder.WriteString("### Responses\n\n")
			builder.WriteString("| Code | Description | Content Type |\n")
			builder.WriteString("|------|-------------|--------------|\n")
			for _, resp := range route.EnhancedMeta.Responses {
				builder.WriteString(fmt.Sprintf("| %d | %s | %s |\n",
					resp.Code, resp.Description, resp.ContentType))
			}
			builder.WriteString("\n")
		}

		// Examples
		if len(route.EnhancedMeta.Examples) > 0 {
			builder.WriteString("### Examples\n\n")
			for name, example := range route.EnhancedMeta.Examples {
				builder.WriteString(fmt.Sprintf("#### %s\n", name))
				if example.Summary != "" {
					builder.WriteString(fmt.Sprintf("*%s*\n\n", example.Summary))
				}
				exampleJSON, _ := json.MarshalIndent(example.Value, "", "  ")
				builder.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", string(exampleJSON)))
			}
		}

		// Middleware
		if len(route.EnhancedMeta.Middleware) > 0 {
			builder.WriteString("**Middleware Chain:**\n\n")
			for i, mw := range route.EnhancedMeta.Middleware {
				builder.WriteString(fmt.Sprintf("%d. `%s`\n", i+1, mw))
			}
			builder.WriteString("\n")
		}

		// Metadata
		if len(route.EnhancedMeta.Metadata) > 0 {
			builder.WriteString("**Additional Metadata:**\n\n")
			metaJSON, _ := json.MarshalIndent(route.EnhancedMeta.Metadata, "", "  ")
			builder.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", string(metaJSON)))
		}

		builder.WriteString("---\n\n")
	}

	return builder.String()
}

// GenerateSwagger generates OpenAPI 3.0 specification
func (rdg *RouteDocumentationGenerator) GenerateSwagger() map[string]interface{} {
	rdg.mu.RLock()
	defer rdg.mu.RUnlock()

	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":   "Plumego API",
			"version": "1.0.0",
		},
		"paths": make(map[string]interface{}),
	}

	paths := spec["paths"].(map[string]interface{})

	for _, route := range rdg.routes {
		pathItem, exists := paths[route.Path]
		if !exists {
			pathItem = make(map[string]interface{})
			paths[route.Path] = pathItem
		}

		pathItemMap := pathItem.(map[string]interface{})
		method := strings.ToLower(route.Method)

		operation := map[string]interface{}{
			"summary":     route.EnhancedMeta.Summary,
			"description": route.EnhancedMeta.Description,
			"operationId": route.EnhancedMeta.OperationID,
			"tags":        route.EnhancedMeta.Tags,
		}

		// Parameters
		if len(route.EnhancedMeta.Parameters) > 0 {
			params := make([]map[string]interface{}, 0)
			for _, param := range route.EnhancedMeta.Parameters {
				paramSpec := map[string]interface{}{
					"name":        param.Name,
					"in":          param.In,
					"required":    param.Required,
					"description": param.Description,
					"schema": map[string]interface{}{
						"type": param.Type,
					},
				}
				if param.Example != nil {
					paramSpec["example"] = param.Example
				}
				if param.Constraints != nil {
					paramSpec["schema"].(map[string]interface{})["constraints"] = param.Constraints
				}
				params = append(params, paramSpec)
			}
			operation["parameters"] = params
		}

		// Request Body (for POST, PUT, PATCH)
		if route.Method == "POST" || route.Method == "PUT" || route.Method == "PATCH" {
			// Look for body parameter
			for _, param := range route.EnhancedMeta.Parameters {
				if param.In == "body" {
					requestBody := map[string]interface{}{
						"description": param.Description,
						"required":    param.Required,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": param.Type,
								},
								"example": param.Example,
							},
						},
					}
					operation["requestBody"] = requestBody
					break
				}
			}
		}

		// Responses
		if len(route.EnhancedMeta.Responses) > 0 {
			responses := make(map[string]interface{})
			for _, resp := range route.EnhancedMeta.Responses {
				codeStr := fmt.Sprintf("%d", resp.Code)
				responseSpec := map[string]interface{}{
					"description": resp.Description,
				}
				if resp.ContentType != "" {
					content := make(map[string]interface{})
					content[resp.ContentType] = map[string]interface{}{
						"schema": resp.Schema,
					}
					if resp.Example != nil {
						content[resp.ContentType].(map[string]interface{})["example"] = resp.Example
					}
					responseSpec["content"] = content
				}
				responses[codeStr] = responseSpec
			}
			operation["responses"] = responses
		}

		pathItemMap[method] = operation
	}

	return spec
}

// GenerateMermaid generates Mermaid diagram code
func (rdg *RouteDocumentationGenerator) GenerateMermaid() string {
	var builder strings.Builder

	rdg.mu.RLock()
	defer rdg.mu.RUnlock()

	builder.WriteString("```mermaid\n")
	builder.WriteString("graph TD\n")
	builder.WriteString("    subgraph \"API Routes\"\n")

	for _, route := range rdg.routes {
		nodeName := fmt.Sprintf("%s_%s", route.Method, strings.ReplaceAll(strings.Trim(route.Path, "/"), "/", "_"))
		nodeLabel := fmt.Sprintf("%s %s", route.Method, route.Path)
		if len(route.EnhancedMeta.Tags) > 0 {
			nodeLabel += fmt.Sprintf("\\n[%s]", strings.Join(route.EnhancedMeta.Tags, ","))
		}
		if route.EnhancedMeta.Summary != "" {
			nodeLabel += fmt.Sprintf("\\n%s", route.EnhancedMeta.Summary)
		}
		builder.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", nodeName, nodeLabel))
	}

	builder.WriteString("    end\n")
	builder.WriteString("```\n")

	return builder.String()
}

// GenerateJSON generates complete JSON documentation
func (rdg *RouteDocumentationGenerator) GenerateJSON() (string, error) {
	rdg.mu.RLock()
	defer rdg.mu.RUnlock()

	doc := map[string]interface{}{
		"routes":    rdg.routes,
		"swagger":   rdg.GenerateSwagger(),
		"mermaid":   rdg.GenerateMermaid(),
		"generated": "plumego-router-docs",
	}

	jsonBytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// GetRouteCount returns the number of documented routes
func (rdg *RouteDocumentationGenerator) GetRouteCount() int {
	rdg.mu.RLock()
	defer rdg.mu.RUnlock()
	return len(rdg.routes)
}

// Clear clears all documented routes
func (rdg *RouteDocumentationGenerator) Clear() {
	rdg.mu.Lock()
	defer rdg.mu.Unlock()
	rdg.routes = make([]EnhancedRouteInfo, 0)
}

// EnhancedRouterIntegration provides integration with EnhancedRouter
type EnhancedRouterIntegration struct {
	*EnhancedRouter
	docGenerator *RouteDocumentationGenerator
}

// NewEnhancedRouterIntegration creates a new integrated router with documentation
func NewEnhancedRouterIntegration(opts ...RouterOption) *EnhancedRouterIntegration {
	router := NewEnhancedRouter(opts...)
	docGen := NewRouteDocumentationGenerator()

	return &EnhancedRouterIntegration{
		EnhancedRouter: router,
		docGenerator:   docGen,
	}
}

// AddRouteWithDocumentation adds a route and automatically registers it for documentation
func (eri *EnhancedRouterIntegration) AddRouteWithDocumentation(
	method, path string,
	handler Handler,
	description string,
	params []ParameterInfo,
	responses []ResponseInfo,
	tags []string,
) error {
	// Add to router
	err := eri.EnhancedRouter.AddRouteWithMeta(method, path, handler, description, params, responses, tags)
	if err != nil {
		return err
	}

	// Register for documentation
	meta, _ := eri.GetRouteMetadata(method, path)
	eri.docGenerator.RegisterRoute(method, path, meta)

	return nil
}

// GenerateAllDocumentation generates all types of documentation
func (eri *EnhancedRouterIntegration) GenerateAllDocumentation() (map[string]interface{}, error) {
	jsonDoc, err := eri.docGenerator.GenerateJSON()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"markdown": eri.docGenerator.GenerateMarkdown(),
		"swagger":  eri.docGenerator.GenerateSwagger(),
		"mermaid":  eri.docGenerator.GenerateMermaid(),
		"json":     jsonDoc,
		"count":    eri.docGenerator.GetRouteCount(),
	}, nil
}
