package protocol

import (
	"context"
	"io"
)

// GraphQLAdapter is a specialized adapter contract for GraphQL gateway use cases.
type GraphQLAdapter interface {
	ProtocolAdapter
	Schema() string
	Validate(query string) error
}

// GraphQLRequest represents a GraphQL request.
type GraphQLRequest struct {
	Query         string              `json:"query"`
	OperationName string              `json:"operationName,omitempty"`
	Variables     map[string]any      `json:"variables,omitempty"`
	Extensions    map[string]any      `json:"extensions,omitempty"`
	HeaderMap     map[string][]string `json:"-"`
	BodyReader    io.Reader           `json:"-"`
}

func (r *GraphQLRequest) Method() string {
	if r.OperationName != "" {
		return r.OperationName
	}
	return "query"
}

func (r *GraphQLRequest) Headers() map[string][]string {
	if r.HeaderMap == nil {
		return make(map[string][]string)
	}
	return r.HeaderMap
}

func (r *GraphQLRequest) Body() io.Reader { return r.BodyReader }

func (r *GraphQLRequest) Metadata() map[string]any {
	m := map[string]any{
		"query":          r.Query,
		"operation_name": r.OperationName,
		"variables":      r.Variables,
	}
	if r.Extensions != nil {
		m["extensions"] = r.Extensions
	}
	return m
}

// GraphQLResponse represents a GraphQL response.
type GraphQLResponse struct {
	Data       any                 `json:"data,omitempty"`
	Errors     []GraphQLError      `json:"errors,omitempty"`
	Extensions map[string]any      `json:"extensions,omitempty"`
	Status     int                 `json:"-"`
	HeaderMap  map[string][]string `json:"-"`
	BodyReader io.Reader           `json:"-"`
}

// GraphQLError represents a GraphQL error.
type GraphQLError struct {
	Message    string            `json:"message"`
	Locations  []GraphQLLocation `json:"locations,omitempty"`
	Path       []any             `json:"path,omitempty"`
	Extensions map[string]any    `json:"extensions,omitempty"`
}

// GraphQLLocation represents a location in a GraphQL document.
type GraphQLLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func (r *GraphQLResponse) StatusCode() int {
	if r.Status == 0 {
		return 200
	}
	return r.Status
}

func (r *GraphQLResponse) Headers() map[string][]string {
	if r.HeaderMap == nil {
		return make(map[string][]string)
	}
	return r.HeaderMap
}

func (r *GraphQLResponse) Body() io.Reader { return r.BodyReader }

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

func (r *GraphQLResponse) HasErrors() bool { return len(r.Errors) > 0 }

// GraphQLExecutor is an interface users can implement for custom GraphQL execution.
type GraphQLExecutor interface {
	Execute(ctx context.Context, req *GraphQLRequest) (*GraphQLResponse, error)
}

// GraphQLDirective represents a GraphQL directive.
type GraphQLDirective struct {
	Name      string
	Arguments map[string]any
}

// GraphQLIntrospectionQuery is the standard introspection query.
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
  }
`
