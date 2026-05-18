package openapi

import (
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
)

const openAPIVersion = "3.1.0"

var (
	String  = Schema{Type: "string"}
	Integer = Schema{Type: "integer"}
	Boolean = Schema{Type: "boolean"}
	Array   = Schema{Type: "array"}
)

// Op carries optional operation hints for one route.
type Op struct {
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	Params      []Param             `json:"parameters,omitempty"`
	Body        *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses,omitempty"`
}

// Param describes an OpenAPI operation parameter.
type Param struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Schema      Schema `json:"schema,omitempty"`
}

// Schema describes a minimal OpenAPI schema object.
type Schema struct {
	Type       string            `json:"type,omitempty"`
	Format     string            `json:"format,omitempty"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
	Ref        string            `json:"$ref,omitempty"`
}

// RequestBody describes an OpenAPI requestBody object.
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// Response describes an OpenAPI response object.
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// MediaType describes an OpenAPI media type object.
type MediaType struct {
	Schema Schema `json:"schema,omitempty"`
}

// Document is the root OpenAPI document.
type Document struct {
	OpenAPI string              `json:"openapi"`
	Info    Info                `json:"info"`
	Paths   map[string]PathItem `json:"paths"`
}

// Info describes OpenAPI document metadata.
type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

// PathItem holds operations for one route path.
type PathItem struct {
	Get     *Operation `json:"get,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Head    *Operation `json:"head,omitempty"`
	Options *Operation `json:"options,omitempty"`
}

// Operation describes one OpenAPI operation.
type Operation struct {
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	Parameters  []Param             `json:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
}

// Generator builds OpenAPI documents from router route metadata.
type Generator struct {
	info Info
}

// New creates a generator with default document metadata.
func New() *Generator {
	return &Generator{
		info: Info{
			Title:   "Plumego API",
			Version: "0.0.0",
		},
	}
}

// Generate converts route metadata and optional operation hints into a document.
func (g *Generator) Generate(routes []router.RouteInfo, hints map[string]Op) Document {
	info := Info{Title: "Plumego API", Version: "0.0.0"}
	if g != nil {
		info = g.info
	}
	doc := Document{
		OpenAPI: openAPIVersion,
		Info:    info,
		Paths:   make(map[string]PathItem),
	}

	for _, route := range routes {
		path := routePathTemplate(route.Path)
		item := doc.Paths[path]
		op := operationFromHint(lookupHint(route, hints))
		setOperation(&item, route.Method, op)
		doc.Paths[path] = item
	}
	return doc
}

// PathParam creates a required path parameter.
func PathParam(name string, schema Schema) Param {
	return Param{Name: name, In: "path", Required: true, Schema: schema}
}

// QueryParam creates an optional query parameter.
func QueryParam(name string, schema Schema) Param {
	return Param{Name: name, In: "query", Schema: schema}
}

// HeaderParam creates an optional header parameter.
func HeaderParam(name string, schema Schema) Param {
	return Param{Name: name, In: "header", Schema: schema}
}

// JSONContent creates an application/json media type map for request and response hints.
func JSONContent(schema Schema) map[string]MediaType {
	return map[string]MediaType{
		contract.ContentTypeJSON: {Schema: schema},
	}
}

func lookupHint(route router.RouteInfo, hints map[string]Op) Op {
	if len(hints) == 0 {
		return Op{}
	}
	if route.Meta.Name != "" {
		if hint, ok := hints[route.Meta.Name]; ok {
			return hint
		}
	}
	if hint, ok := hints[route.Method+" "+route.Path]; ok {
		return hint
	}
	return Op{}
}

func operationFromHint(hint Op) *Operation {
	responses := hint.Responses
	if len(responses) == 0 {
		responses = defaultResponses()
	}
	return &Operation{
		Summary:     hint.Summary,
		Description: hint.Description,
		Tags:        append([]string(nil), hint.Tags...),
		Parameters:  append([]Param(nil), hint.Params...),
		RequestBody: hint.Body,
		Responses:   cloneResponses(responses),
	}
}

func defaultResponses() map[string]Response {
	return map[string]Response{
		"200": {Description: "OK"},
	}
}

func cloneResponses(in map[string]Response) map[string]Response {
	out := make(map[string]Response, len(in))
	for code, response := range in {
		out[code] = response
	}
	return out
}

func routePathTemplate(path string) string {
	if path == "" || path == "/" {
		return "/"
	}
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if len(part) < 2 {
			continue
		}
		switch part[0] {
		case ':', '*':
			parts[i] = "{" + part[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

func setOperation(item *PathItem, method string, op *Operation) {
	switch method {
	case http.MethodGet:
		item.Get = op
	case http.MethodPost:
		item.Post = op
	case http.MethodPut:
		item.Put = op
	case http.MethodPatch:
		item.Patch = op
	case http.MethodDelete:
		item.Delete = op
	case http.MethodHead:
		item.Head = op
	case http.MethodOptions:
		item.Options = op
	}
}
