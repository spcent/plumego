package handler

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var openAPISpecYAML []byte

//go:embed openapi.json
var openAPISpecJSON []byte

// swaggerUIPage renders a minimal Swagger UI page backed by the public
// CDN-hosted swagger-ui-dist bundle. It only points at the spec endpoint
// served by this binary; no bundler or new frontend dependency is required.
const swaggerUIPage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>dbadmin API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>body { margin: 0; }</style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: "/openapi.json",
        dom_id: "#swagger-ui",
        presets: [SwaggerUIBundle.presets.apis],
      });
    };
  </script>
</body>
</html>
`

// DocsHandler serves the static OpenAPI spec and an interactive docs page.
// All three routes are public (no session required) since they only
// describe the API shape, not data.
type DocsHandler struct{}

// SpecJSON serves the OpenAPI 3.0 document as JSON.
func (DocsHandler) SpecJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(openAPISpecJSON)
}

// SpecYAML serves the OpenAPI 3.0 document as YAML.
func (DocsHandler) SpecYAML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	_, _ = w.Write(openAPISpecYAML)
}

// UI serves a minimal Swagger UI page pointed at SpecJSON.
func (DocsHandler) UI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(swaggerUIPage))
}
