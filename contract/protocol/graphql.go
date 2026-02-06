package protocol

import (
	"context"
	"io"
)

// GraphQLAdapter is a specialized adapter contract for GraphQL gateway
//
// Users implement this interface using their preferred GraphQL library:
//   - github.com/graphql-go/graphql
//   - github.com/graph-gophers/graphql-go
//   - github.com/99designs/gqlgen
//
// This contract defines what plumego expects from a GraphQL adapter
// without importing any GraphQL library.
//
// Example implementation:
//
//	type MyGraphQLAdapter struct {
//		schema graphql.Schema  // User imports graphql library
//	}
//
//	func (a *MyGraphQLAdapter) Name() string {
//		return "graphql"
//	}
//
//	func (a *MyGraphQLAdapter) Handles(req *HTTPRequest) bool {
//		return strings.HasPrefix(req.URL, "/graphql")
//	}
//
//	func (a *MyGraphQLAdapter) Schema() string {
//		return mySchemaString
//	}
//
//	func (a *MyGraphQLAdapter) Validate(query string) error {
//		// User's validation using graphql library
//	}
type GraphQLAdapter interface {
	ProtocolAdapter

	// Schema returns the GraphQL schema definition (SDL format)
	// Example: "type Query { user(id: ID!): User }"
	Schema() string

	// Validate validates a GraphQL query against the schema
	// Returns error if query is invalid
	Validate(query string) error
}

// GraphQLRequest represents a GraphQL request
//
// Standard GraphQL request format:
//
//	{
//	  "query": "query GetUser($id: ID!) { user(id: $id) { name } }",
//	  "operationName": "GetUser",
//	  "variables": {"id": "123"}
//	}
type GraphQLRequest struct {
	// Query is the GraphQL query string
	Query string `json:"query"`

	// OperationName is the optional operation name
	OperationName string `json:"operationName,omitempty"`

	// Variables is the optional variables map
	Variables map[string]any `json:"variables,omitempty"`

	// Extensions is optional extensions map (for persisted queries, etc.)
	Extensions map[string]any `json:"extensions,omitempty"`

	// HeaderMap holds HTTP headers
	HeaderMap map[string][]string `json:"-"`

	// BodyReader holds the raw request body
	BodyReader io.Reader `json:"-"`
}

// Method returns the operation name or "query"
func (r *GraphQLRequest) Method() string {
	if r.OperationName != "" {
		return r.OperationName
	}
	return "query"
}

// Headers returns request headers
func (r *GraphQLRequest) Headers() map[string][]string {
	if r.HeaderMap == nil {
		return make(map[string][]string)
	}
	return r.HeaderMap
}

// Body returns the request body reader
func (r *GraphQLRequest) Body() io.Reader {
	return r.BodyReader
}

// Metadata returns GraphQL-specific metadata
func (r *GraphQLRequest) Metadata() map[string]any {
	m := make(map[string]any)
	m["query"] = r.Query
	m["operation_name"] = r.OperationName
	m["variables"] = r.Variables
	if r.Extensions != nil {
		m["extensions"] = r.Extensions
	}
	return m
}

// GraphQLResponse represents a GraphQL response
//
// Standard GraphQL response format:
//
//	{
//	  "data": { "user": { "name": "John" } },
//	  "errors": [
//	    {
//	      "message": "Field error",
//	      "locations": [{"line": 1, "column": 2}],
//	      "path": ["user", "email"]
//	    }
//	  ]
//	}
type GraphQLResponse struct {
	// Data holds the response data
	Data any `json:"data,omitempty"`

	// Errors holds any GraphQL errors
	Errors []GraphQLError `json:"errors,omitempty"`

	// Extensions holds optional extensions (timing, tracing, etc.)
	Extensions map[string]any `json:"extensions,omitempty"`

	// Status is the HTTP status code
	Status int `json:"-"`

	// HeaderMap holds response headers
	HeaderMap map[string][]string `json:"-"`

	// BodyReader holds the raw response body
	BodyReader io.Reader `json:"-"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	// Message is the error message
	Message string `json:"message"`

	// Locations indicates where in the query the error occurred
	Locations []GraphQLLocation `json:"locations,omitempty"`

	// Path indicates the path to the field that caused the error
	Path []any `json:"path,omitempty"`

	// Extensions holds additional error information
	Extensions map[string]any `json:"extensions,omitempty"`
}

// GraphQLLocation represents a location in a GraphQL document
type GraphQLLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// StatusCode returns HTTP status code
func (r *GraphQLResponse) StatusCode() int {
	if r.Status == 0 {
		// GraphQL always returns 200 even for errors
		// unless there's a network/protocol error
		return 200
	}
	return r.Status
}

// Headers returns response headers
func (r *GraphQLResponse) Headers() map[string][]string {
	if r.HeaderMap == nil {
		return make(map[string][]string)
	}
	return r.HeaderMap
}

// Body returns response body reader
func (r *GraphQLResponse) Body() io.Reader {
	return r.BodyReader
}

// Metadata returns GraphQL-specific metadata
func (r *GraphQLResponse) Metadata() map[string]any {
	m := make(map[string]any)
	if len(r.Errors) > 0 {
		m["has_errors"] = true
		m["error_count"] = len(r.Errors)
	}
	if r.Extensions != nil {
		m["extensions"] = r.Extensions
	}
	return m
}

// HasErrors returns true if the response contains GraphQL errors
func (r *GraphQLResponse) HasErrors() bool {
	return len(r.Errors) > 0
}

// GraphQLExecutor is an interface users can implement for custom GraphQL execution
//
// This provides a hook for users to customize how GraphQL queries are executed
type GraphQLExecutor interface {
	// Execute executes a GraphQL query
	Execute(ctx context.Context, req *GraphQLRequest) (*GraphQLResponse, error)
}

// GraphQLDirective represents a GraphQL directive
//
// Users can use this for implementing custom directive handling
type GraphQLDirective struct {
	Name      string
	Arguments map[string]any
}

// GraphQLIntrospectionQuery is the standard introspection query
// Users can use this to implement schema introspection
const GraphQLIntrospectionQuery = `
  query IntrospectionQuery {
    __schema {
      queryType { name }
      mutationType { name }
      subscriptionType { name }
      types {
        ...FullType
      }
      directives {
        name
        description
        locations
        args {
          ...InputValue
        }
      }
    }
  }

  fragment FullType on __Type {
    kind
    name
    description
    fields(includeDeprecated: true) {
      name
      description
      args {
        ...InputValue
      }
      type {
        ...TypeRef
      }
      isDeprecated
      deprecationReason
    }
    inputFields {
      ...InputValue
    }
    interfaces {
      ...TypeRef
    }
    enumValues(includeDeprecated: true) {
      name
      description
      isDeprecated
      deprecationReason
    }
    possibleTypes {
      ...TypeRef
    }
  }

  fragment InputValue on __InputValue {
    name
    description
    type { ...TypeRef }
    defaultValue
  }

  fragment TypeRef on __Type {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
                ofType {
                  kind
                  name
                }
              }
            }
          }
        }
      }
    }
  }
`
