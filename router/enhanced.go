package router

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// EnhancedRouter extends Router with advanced metadata and documentation capabilities
type EnhancedRouter struct {
	*Router
	mu              sync.RWMutex
	routeMetadata   map[string]EnhancedRouteMeta // method+path -> metadata
	handlerRegistry map[string]HandlerInfo       // handler name -> info
}

// EnhancedRouteMeta contains comprehensive metadata for a route
type EnhancedRouteMeta struct {
	Base        RouteMeta              `json:"base"`
	Description string                 `json:"description,omitempty"`
	OperationID string                 `json:"operation_id,omitempty"`
	Summary     string                 `json:"summary,omitempty"`
	Parameters  []ParameterInfo        `json:"parameters,omitempty"`
	Responses   []ResponseInfo         `json:"responses,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Middleware  []string               `json:"middleware,omitempty"`
	Examples    map[string]Example     `json:"examples,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ParameterInfo describes a route parameter
type ParameterInfo struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // path, query, header, body
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description,omitempty"`
	Example     interface{} `json:"example,omitempty"`
	Constraints interface{} `json:"constraints,omitempty"`
}

// ResponseInfo describes a response
type ResponseInfo struct {
	Code        int                    `json:"code"`
	Description string                 `json:"description,omitempty"`
	ContentType string                 `json:"content_type,omitempty"`
	Example     interface{}            `json:"example,omitempty"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
}

// Example demonstrates usage
type Example struct {
	Summary string      `json:"summary,omitempty"`
	Value   interface{} `json:"value"`
}

// HandlerInfo contains information about a handler
type HandlerInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Package     string   `json:"package,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// NewEnhancedRouter creates a new enhanced router
func NewEnhancedRouter(opts ...RouterOption) *EnhancedRouter {
	return &EnhancedRouter{
		Router:          NewRouter(opts...),
		routeMetadata:   make(map[string]EnhancedRouteMeta),
		handlerRegistry: make(map[string]HandlerInfo),
	}
}

// AddRouteWithMeta adds a route with comprehensive metadata
func (er *EnhancedRouter) AddRouteWithMeta(
	method, path string,
	handler Handler,
	description string,
	params []ParameterInfo,
	responses []ResponseInfo,
	tags []string,
) error {
	// Add the route normally
	if err := er.Router.AddRoute(method, path, handler); err != nil {
		return err
	}

	// Store enhanced metadata
	key := er.routeKey(method, path)
	er.mu.Lock()
	er.routeMetadata[key] = EnhancedRouteMeta{
		Base: RouteMeta{
			Name: er.generateRouteName(method, path),
			Tags: tags,
		},
		Description: description,
		OperationID: er.generateOperationID(method, path),
		Summary:     description,
		Parameters:  params,
		Responses:   responses,
		Tags:        tags,
		Middleware:  er.getCurrentMiddlewareNames(),
	}
	er.mu.Unlock()

	return nil
}

// AddRouteWithOptions adds a route with options and metadata
func (er *EnhancedRouter) AddRouteWithOptions(
	method, path string,
	handler Handler,
	opts ...RouteOption,
) error {
	// Add the route
	if err := er.Router.AddRoute(method, path, handler); err != nil {
		return err
	}

	// Apply options
	meta := RouteMeta{}
	for _, opt := range opts {
		if opt != nil {
			opt(&meta)
		}
	}

	// Store metadata
	key := er.routeKey(method, path)
	er.mu.Lock()
	existing, exists := er.routeMetadata[key]
	if !exists {
		existing = EnhancedRouteMeta{}
	}
	existing.Base = meta
	er.routeMetadata[key] = existing
	er.mu.Unlock()

	return nil
}

// SetRouteDescription sets the description for a route
func (er *EnhancedRouter) SetRouteDescription(method, path, description string) {
	key := er.routeKey(method, path)
	er.mu.Lock()
	defer er.mu.Unlock()

	meta := er.routeMetadata[key]
	meta.Description = description
	meta.Summary = description
	er.routeMetadata[key] = meta
}

// SetRouteParameters sets parameter information for a route
func (er *EnhancedRouter) SetRouteParameters(method, path string, params []ParameterInfo) {
	key := er.routeKey(method, path)
	er.mu.Lock()
	defer er.mu.Unlock()

	meta := er.routeMetadata[key]
	meta.Parameters = params
	er.routeMetadata[key] = meta
}

// SetRouteResponses sets response information for a route
func (er *EnhancedRouter) SetRouteResponses(method, path string, responses []ResponseInfo) {
	key := er.routeKey(method, path)
	er.mu.Lock()
	defer er.mu.Unlock()

	meta := er.routeMetadata[key]
	meta.Responses = responses
	er.routeMetadata[key] = meta
}

// AddRouteExample adds an example to a route
func (er *EnhancedRouter) AddRouteExample(method, path, name string, example Example) {
	key := er.routeKey(method, path)
	er.mu.Lock()
	defer er.mu.Unlock()

	meta := er.routeMetadata[key]
	if meta.Examples == nil {
		meta.Examples = make(map[string]Example)
	}
	meta.Examples[name] = example
	er.routeMetadata[key] = meta
}

// RegisterHandler registers a handler with metadata
func (er *EnhancedRouter) RegisterHandler(name, description, pkg string, tags []string) {
	er.mu.Lock()
	defer er.mu.Unlock()

	er.handlerRegistry[name] = HandlerInfo{
		Name:        name,
		Description: description,
		Package:     pkg,
		Tags:        tags,
	}
}

// GetRouteMetadata returns the enhanced metadata for a route
func (er *EnhancedRouter) GetRouteMetadata(method, path string) (EnhancedRouteMeta, bool) {
	key := er.routeKey(method, path)
	er.mu.RLock()
	defer er.mu.RUnlock()

	meta, exists := er.routeMetadata[key]
	return meta, exists
}

// GetRoutesWithMetadata returns all routes with their metadata
func (er *EnhancedRouter) GetRoutesWithMetadata() []EnhancedRouteInfo {
	er.mu.RLock()
	defer er.mu.RUnlock()

	var result []EnhancedRouteInfo

	// Get base routes
	baseRoutes := er.Router.Routes()

	for _, route := range baseRoutes {
		key := route.Method + " " + route.Path
		meta := er.routeMetadata[key]

		result = append(result, EnhancedRouteInfo{
			Method:       route.Method,
			Path:         route.Path,
			BaseMeta:     route.Meta,
			EnhancedMeta: meta,
		})
	}

	// Sort for consistent output
	sort.Slice(result, func(i, j int) bool {
		if result[i].Method == result[j].Method {
			return result[i].Path < result[j].Path
		}
		return result[i].Method < result[j].Method
	})

	return result
}

// GenerateDocumentation generates comprehensive API documentation
func (er *EnhancedRouter) GenerateDocumentation() string {
	var builder strings.Builder

	routes := er.GetRoutesWithMetadata()

	builder.WriteString("# API Documentation\n\n")
	builder.WriteString("Generated comprehensive API documentation with route metadata.\n\n")

	// Group by tags
	tags := make(map[string][]EnhancedRouteInfo)
	for _, route := range routes {
		if len(route.EnhancedMeta.Tags) > 0 {
			for _, tag := range route.EnhancedMeta.Tags {
				tags[tag] = append(tags[tag], route)
			}
		} else {
			tags["untagged"] = append(tags["untagged"], route)
		}
	}

	// Generate documentation by tag
	for tag, tagRoutes := range tags {
		builder.WriteString(fmt.Sprintf("## %s\n\n", tag))

		for _, route := range tagRoutes {
			builder.WriteString(fmt.Sprintf("### %s %s\n\n", route.Method, route.Path))

			if route.EnhancedMeta.Summary != "" {
				builder.WriteString(fmt.Sprintf("**Summary:** %s\n\n", route.EnhancedMeta.Summary))
			}

			if route.EnhancedMeta.Description != "" {
				builder.WriteString(fmt.Sprintf("**Description:** %s\n\n", route.EnhancedMeta.Description))
			}

			if route.BaseMeta.Name != "" {
				builder.WriteString(fmt.Sprintf("**Name:** %s\n\n", route.BaseMeta.Name))
			}

			// Parameters
			if len(route.EnhancedMeta.Parameters) > 0 {
				builder.WriteString("**Parameters:**\n\n")
				builder.WriteString("| Name | In | Type | Required | Description |\n")
				builder.WriteString("|------|----|------|----------|-------------|\n")
				for _, param := range route.EnhancedMeta.Parameters {
					required := "No"
					if param.Required {
						required = "Yes"
					}
					builder.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s | %s |\n",
						param.Name, param.In, param.Type, required, param.Description))
				}
				builder.WriteString("\n")
			}

			// Responses
			if len(route.EnhancedMeta.Responses) > 0 {
				builder.WriteString("**Responses:**\n\n")
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
				builder.WriteString("**Examples:**\n\n")
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
				builder.WriteString(fmt.Sprintf("**Middleware:** %s\n\n", strings.Join(route.EnhancedMeta.Middleware, ", ")))
			}

			builder.WriteString("---\n\n")
		}
	}

	return builder.String()
}

// GenerateSwaggerSpec generates OpenAPI/Swagger compatible specification
func (er *EnhancedRouter) GenerateSwaggerSpec() map[string]interface{} {
	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":   "Plumego API",
			"version": "1.0.0",
		},
		"paths": make(map[string]interface{}),
	}

	routes := er.GetRoutesWithMetadata()
	paths := spec["paths"].(map[string]interface{})

	for _, route := range routes {
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
				params = append(params, paramSpec)
			}
			operation["parameters"] = params
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
					responseSpec["content"] = map[string]interface{}{
						resp.ContentType: map[string]interface{}{
							"schema":  resp.Schema,
							"example": resp.Example,
						},
					}
				}
				responses[codeStr] = responseSpec
			}
			operation["responses"] = responses
		}

		pathItemMap[method] = operation
	}

	return spec
}

// GenerateMermaidDiagram generates a Mermaid diagram of the routing structure
func (er *EnhancedRouter) GenerateMermaidDiagram() string {
	var builder strings.Builder

	builder.WriteString("```mermaid\n")
	builder.WriteString("graph TD\n")
	builder.WriteString("    subgraph \"Routes\"\n")

	routes := er.GetRoutesWithMetadata()
	for _, route := range routes {
		nodeName := fmt.Sprintf("%s_%s", route.Method, strings.ReplaceAll(route.Path, "/", "_"))
		nodeLabel := fmt.Sprintf("%s %s", route.Method, route.Path)
		if len(route.EnhancedMeta.Tags) > 0 {
			nodeLabel += fmt.Sprintf(" [%s]", strings.Join(route.EnhancedMeta.Tags, ","))
		}
		builder.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", nodeName, nodeLabel))
	}

	builder.WriteString("    end\n")
	builder.WriteString("```\n")

	return builder.String()
}

// GetMiddlewareChain returns the current middleware chain with names
func (er *EnhancedRouter) GetMiddlewareChain() []string {
	return er.getCurrentMiddlewareNames()
}

// Helper methods

func (er *EnhancedRouter) generateRouteName(method, path string) string {
	// Convert /users/:id -> GetUserById
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	var nameParts []string

	nameParts = append(nameParts, strings.Title(strings.ToLower(method)))

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			nameParts = append(nameParts, "By"+strings.Title(part[1:]))
		} else if strings.HasPrefix(part, "*") {
			nameParts = append(nameParts, "All")
		} else {
			nameParts = append(nameParts, strings.Title(part))
		}
	}

	return strings.Join(nameParts, "")
}

func (er *EnhancedRouter) generateOperationID(method, path string) string {
	return er.generateRouteName(method, path)
}

func (er *EnhancedRouter) getCurrentMiddlewareNames() []string {
	// Get middleware names from the router's middleware manager
	// This would need to be enhanced to track middleware names
	return []string{} // Placeholder - would need middleware name tracking
}

func (er *EnhancedRouter) routeKey(method, path string) string {
	return method + " " + er.Router.fullPath(path)
}

// EnhancedRouteInfo combines base route info with enhanced metadata
type EnhancedRouteInfo struct {
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	BaseMeta     RouteMeta         `json:"base_meta,omitempty"`
	EnhancedMeta EnhancedRouteMeta `json:"enhanced_meta,omitempty"`
}

// RouteDocumentation represents complete documentation for a route
type RouteDocumentation struct {
	Route    EnhancedRouteInfo      `json:"route"`
	Swagger  map[string]interface{} `json:"swagger,omitempty"`
	Mermaid  string                 `json:"mermaid,omitempty"`
	Examples []Example              `json:"examples,omitempty"`
}

// GetRouteDocumentation gets complete documentation for a specific route
func (er *EnhancedRouter) GetRouteDocumentation(method, path string) (RouteDocumentation, error) {
	routes := er.GetRoutesWithMetadata()

	for _, route := range routes {
		if route.Method == method && route.Path == path {
			doc := RouteDocumentation{
				Route: route,
			}

			// Generate Swagger spec for this route
			spec := er.GenerateSwaggerSpec()
			if paths, ok := spec["paths"].(map[string]interface{}); ok {
				if pathItem, ok := paths[path].(map[string]interface{}); ok {
					doc.Swagger = pathItem
				}
			}

			// Generate Mermaid diagram
			doc.Mermaid = er.GenerateMermaidDiagram()

			// Add examples
			if len(route.EnhancedMeta.Examples) > 0 {
				for _, ex := range route.EnhancedMeta.Examples {
					doc.Examples = append(doc.Examples, ex)
				}
			}

			return doc, nil
		}
	}

	return RouteDocumentation{}, fmt.Errorf("route %s %s not found", method, path)
}
