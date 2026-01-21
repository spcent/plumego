package main

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spcent/plumego/config"
	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/frontend"
	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/metrics"
	webhookout "github.com/spcent/plumego/net/webhookout"
	"github.com/spcent/plumego/pubsub"
)

//go:embed ui/*
var staticFS embed.FS

// AppContext holds application-wide dependencies and configuration
type AppContext struct {
	App        *core.App
	Bus        *pubsub.InProcPubSub
	WebhookSvc *webhookout.Service
	Prom       *metrics.PrometheusCollector
	Tracer     *metrics.OpenTelemetryTracer
	DocSite    *DocSite
}

// Config holds all application configuration
type Config struct {
	Addr           string
	WSecret        string
	GitHubSecret   string
	StripeSecret   string
	WebhookToken   string
	EnableDocs     bool
	EnableMetrics  bool
	EnableWebhooks bool
}

// LoadConfig loads configuration from environment
func LoadConfig() Config {
	return Config{
		Addr:           config.GetString("ADDR", ":8080"),
		WSecret:        config.GetString("WS_SECRET", "dev-secret"),
		GitHubSecret:   config.GetString("GITHUB_WEBHOOK_SECRET", "dev-github-secret"),
		StripeSecret:   config.GetString("STRIPE_WEBHOOK_SECRET", "whsec_dev"),
		WebhookToken:   config.GetString("WEBHOOK_TRIGGER_TOKEN", "dev-trigger"),
		EnableDocs:     config.GetBool("ENABLE_DOCS", true),
		EnableMetrics:  config.GetBool("ENABLE_METRICS", true),
		EnableWebhooks: config.GetBool("ENABLE_WEBHOOKS", true),
	}
}

// NewAppContext creates and initializes the application context
func NewAppContext(cfg Config) (*AppContext, error) {
	// Initialize in-process pub/sub bus
	bus := pubsub.New()

	// Initialize webhook service
	webhookStore := webhookout.NewMemStore()
	webhookCfg := webhookout.ConfigFromEnv()
	webhookCfg.Enabled = cfg.EnableWebhooks
	webhookSvc := webhookout.NewService(webhookStore, webhookCfg)

	// Initialize metrics
	prom := metrics.NewPrometheusCollector("plumego_reference")
	tracer := metrics.NewOpenTelemetryTracer("plumego-reference")

	// Create core application
	app := core.New(
		core.WithAddr(cfg.Addr),
		core.WithDebug(),
		core.WithPubSub(bus),
		core.WithMetricsCollector(prom),
		core.WithTracer(tracer),
		core.WithWebhookIn(core.WebhookInConfig{
			Enabled:           cfg.EnableWebhooks,
			Pub:               bus,
			GitHubSecret:      cfg.GitHubSecret,
			StripeSecret:      cfg.StripeSecret,
			MaxBodyBytes:      1 << 20,
			StripeTolerance:   5 * time.Minute,
			TopicPrefixGitHub: "in.github.",
			TopicPrefixStripe: "in.stripe.",
		}),
		core.WithWebhookOut(core.WebhookOutConfig{
			Enabled:          cfg.EnableWebhooks,
			Service:          webhookSvc,
			TriggerToken:     cfg.WebhookToken,
			BasePath:         "/webhooks",
			IncludeStats:     true,
			DefaultPageLimit: 50,
		}),
	)

	// Enable standard middleware
	app.EnableRecovery()
	app.EnableLogging()
	app.EnableCORS()

	// Load documentation site
	var docSite *DocSite
	if cfg.EnableDocs {
		if ds, err := NewDocSite(); err == nil {
			docSite = ds
		} else {
			log.Printf("docs disabled: %v", err)
		}
	}

	return &AppContext{
		App:        app,
		Bus:        bus,
		WebhookSvc: webhookSvc,
		Prom:       prom,
		Tracer:     tracer,
		DocSite:    docSite,
	}, nil
}

// RegisterRoutes registers all application routes
func (ctx *AppContext) RegisterRoutes() error {
	// Register documentation routes
	if ctx.DocSite != nil {
		ctx.App.Get("/docs", ctx.DocSite.Handler())
		ctx.App.Get("/docs/*path", ctx.DocSite.Handler())
	}

	// Register static frontend
	if err := frontend.RegisterFS(ctx.App.Router(), http.FS(staticFS), frontend.WithPrefix("/")); err != nil {
		return fmt.Errorf("failed to register frontend: %w", err)
	}

	// Register static files (CSS, JS) under /static path
	staticSubFS, err := fs.Sub(staticFS, "ui/static")
	if err != nil {
		return fmt.Errorf("failed to get static subdirectory: %w", err)
	}
	if err := frontend.RegisterFS(ctx.App.Router(), http.FS(staticSubFS), frontend.WithPrefix("/static")); err != nil {
		return fmt.Errorf("failed to register static files: %w", err)
	}

	// Register health endpoints
	ctx.App.GetHandler("/health/ready", health.ReadinessHandler())
	ctx.App.GetHandler("/health/build", health.BuildInfoHandler())

	// Configure WebSocket
	wsCfg := core.DefaultWebSocketConfig()
	wsCfg.Secret = []byte(config.GetString("WS_SECRET", "dev-secret"))
	if _, err := ctx.App.ConfigureWebSocketWithOptions(wsCfg); err != nil {
		return fmt.Errorf("failed to configure websocket: %w", err)
	}

	// Register API endpoints
	ctx.registerAPIEndpoints()

	// Register test endpoints
	ctx.registerTestEndpoints()

	// Register metrics endpoint
	if ctx.Prom != nil {
		ctx.App.GetHandler("/metrics", ctx.Prom.Handler())
	}

	return nil
}

// registerAPIEndpoints registers core API endpoints
func (ctx *AppContext) registerAPIEndpoints() {
	// Hello endpoint - demonstrates basic API response
	ctx.App.Get("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"message":   "hello from plumego reference",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "1.0.0",
			"features": []string{
				"WebSocket",
				"Documentation",
				"Webhook",
				"Metrics",
				"Health Check",
				"Middleware",
				"Pub/Sub",
			},
			"endpoints": map[string]string{
				"docs":      "/docs",
				"webhooks":  "/webhooks",
				"metrics":   "/metrics",
				"health":    "/health/ready",
				"websocket": "/ws",
				"api":       "/api",
			},
		}
		ctx.writeJSON(w, response)
	})

	// Status endpoint - returns comprehensive system status
	ctx.App.Get("/api/status", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now().Add(-time.Hour) // Simulate 1 hour uptime

		response := map[string]any{
			"status":  "healthy",
			"service": "plumego-reference",
			"version": "1.0.0",
			"system": map[string]any{
				"uptime":     time.Since(startTime).String(),
				"timestamp":  time.Now().Format(time.RFC3339),
				"go_version": "1.21+",
			},
			"components": map[string]any{
				"websocket":   "enabled",
				"webhook_in":  "enabled",
				"webhook_out": "enabled",
				"metrics":     "enabled",
				"docs":        "enabled",
				"pubsub":      "enabled",
			},
			"endpoints": map[string]string{
				"root":      "/",
				"docs":      "/docs",
				"api":       "/api",
				"metrics":   "/metrics",
				"health":    "/health/ready",
				"websocket": "/ws",
			},
		}
		ctx.writeJSON(w, response)
	})

	// API test endpoint with multiple response formats
	ctx.App.Get("/api/test", func(w http.ResponseWriter, r *http.Request) {
		format := r.URL.Query().Get("format")
		delay := r.URL.Query().Get("delay")

		// Optional delay for testing
		if delay != "" {
			if d, err := time.ParseDuration(delay); err == nil {
				const maxDelay = 2 * time.Second
				if d > maxDelay {
					d = maxDelay
				}
				select {
				case <-time.After(d):
				case <-r.Context().Done():
					return
				}
			}
		}

		switch format {
		case "xml":
			w.Header().Set("Content-Type", "application/xml")
			w.Write([]byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<response>
  <timestamp>%s</timestamp>
  <format>xml</format>
  <status>success</status>
</response>`, time.Now().Format(time.RFC3339))))
		case "csv":
			w.Header().Set("Content-Type", "text/csv")
			w.Write([]byte("timestamp,format,status\n"))
			w.Write([]byte(fmt.Sprintf("%s,csv,success\n", time.Now().Format(time.RFC3339))))
		case "plain":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(fmt.Sprintf("Plain text response at %s", time.Now().Format(time.RFC3339))))
		default:
			response := map[string]any{
				"format":       "json",
				"timestamp":    time.Now().Format(time.RFC3339),
				"status":       "success",
				"query_params": r.URL.Query().Encode(),
			}
			ctx.writeJSON(w, response)
		}
	})
}

// registerTestEndpoints registers test and demo endpoints
func (ctx *AppContext) registerTestEndpoints() {
	// Pub/Sub test endpoint
	ctx.App.Get("/test/pubsub", func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Query().Get("topic")
		if topic == "" {
			topic = "test.default"
		}

		message := pubsub.Message{
			Topic: topic,
			Type:  "test",
			Data:  fmt.Sprintf("Test message published at %s", time.Now().Format(time.RFC3339)),
			Time:  time.Now(),
		}

		if err := ctx.Bus.Publish(topic, message); err != nil {
			ctx.writeError(w, http.StatusInternalServerError, "publish failed", err)
			return
		}

		ctx.writeJSON(w, map[string]any{
			"status":    "success",
			"topic":     topic,
			"message":   fmt.Sprint(message.Data),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Webhook test endpoint
	ctx.App.Post("/test/webhook", func(w http.ResponseWriter, r *http.Request) {
		const maxWebhookBody = int64(1 << 20)
		body, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookBody+1))
		if err != nil {
			ctx.writeError(w, http.StatusBadRequest, "read body failed", err)
			return
		}
		if int64(len(body)) > maxWebhookBody {
			ctx.writeError(w, http.StatusRequestEntityTooLarge, "body too large", nil)
			return
		}

		ctx.writeJSON(w, map[string]any{
			"status":         "webhook_received",
			"content_length": len(body),
			"timestamp":      time.Now().Format(time.RFC3339),
			"headers": map[string]string{
				"user_agent":   r.UserAgent(),
				"content_type": r.Header.Get("Content-Type"),
			},
		})
	})
}

// writeJSON writes a JSON response
func (ctx *AppContext) writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode JSON response: %v", err)
	}
}

// writeError writes an error response
func (ctx *AppContext) writeError(w http.ResponseWriter, status int, message string, err error) {
	response := map[string]any{
		"status":  "error",
		"message": message,
	}
	if err != nil {
		response["error"] = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		log.Printf("failed to encode error response: %v", encodeErr)
	}
}

// Start starts the application
func (ctx *AppContext) Start() error {
	// Start webhook service
	if ctx.WebhookSvc != nil {
		ctx.WebhookSvc.Start(context.Background())
		defer ctx.WebhookSvc.Stop()
	}

	// Boot the application
	if err := ctx.App.Boot(); err != nil {
		return fmt.Errorf("server stopped: %w", err)
	}

	return nil
}

// DocSite represents a documentation site
type DocSite struct {
	fs          fs.FS
	nav         map[string][]DocPage
	defaultLang string
}

// DocPage represents a documentation page
type DocPage struct {
	Lang  string
	Slug  string
	Title string
}

// NewDocSite creates a new documentation site
func NewDocSite() (*DocSite, error) {
	path := locateDocsPath()
	if path == "" {
		return nil, fmt.Errorf("docs directory not found")
	}

	f := os.DirFS(path)
	nav, err := buildDocNav(f)
	if err != nil {
		return nil, fmt.Errorf("build docs nav: %w", err)
	}

	defaultLang := "zh"
	if _, ok := nav[defaultLang]; !ok {
		for lang := range nav {
			defaultLang = lang
			break
		}
	}

	return &DocSite{
		fs:          f,
		nav:         nav,
		defaultLang: defaultLang,
	}, nil
}

// locateDocsPath finds the documentation directory
func locateDocsPath() string {
	candidates := []string{}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "examples", "docs"),
			filepath.Join(cwd, "docs"),
			filepath.Join(filepath.Dir(cwd), "docs"),
		)
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "docs"),
			filepath.Join(exeDir, "examples", "docs"),
		)
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}

// buildDocNav builds the documentation navigation
func buildDocNav(f fs.FS) (map[string][]DocPage, error) {
	nav := make(map[string][]DocPage)
	langs, err := fs.ReadDir(f, ".")
	if err != nil {
		return nil, err
	}

	for _, langEntry := range langs {
		if !langEntry.IsDir() {
			continue
		}
		lang := langEntry.Name()
		files, err := fs.ReadDir(f, lang)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			if file.IsDir() || filepath.Ext(file.Name()) != ".md" {
				continue
			}
			slug := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			title := extractTitle(f, path.Join(lang, file.Name()))
			nav[lang] = append(nav[lang], DocPage{Lang: lang, Slug: slug, Title: title})
		}
		sort.Slice(nav[lang], func(i, j int) bool { return nav[lang][i].Title < nav[lang][j].Title })
	}
	return nav, nil
}

// extractTitle extracts the title from a markdown file
func extractTitle(f fs.FS, filePath string) string {
	file, err := f.Open(filePath)
	if err != nil {
		return filePath
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			return strings.TrimLeft(line, "# ")
		}
	}
	if err := scanner.Err(); err != nil {
		return strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	}
	return strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
}

// Handler returns the HTTP handler for the documentation site
func (d *DocSite) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := d.defaultLang
		slug := ""
		trimmed := strings.TrimPrefix(r.URL.Path, "/docs")
		trimmed = strings.Trim(trimmed, "/")
		if trimmed != "" {
			parts := strings.Split(trimmed, "/")
			if len(parts) > 0 && parts[0] != "" {
				lang = parts[0]
			}
			if len(parts) > 1 {
				slug = parts[1]
			}
		}
		if slug == "" {
			d.renderIndex(w, lang)
			return
		}
		d.renderDoc(w, r, lang, slug)
	}
}

// renderIndex renders the documentation index
func (d *DocSite) renderIndex(w http.ResponseWriter, lang string) {
	content := markdownToHTML("# Documentation\nSelect a language and module on the left to start reading, or directly access /docs/{lang}/{module}.\n\nUse the navigation on the left to browse modules by language.")
	d.renderPage(w, docTemplateData{
		Title:        "Plumego Docs",
		Lang:         lang,
		CurrentSlug:  "",
		Content:      content,
		Navigation:   d.nav,
		LanguageList: d.languages(),
	})
}

// renderDoc renders a specific documentation page
func (d *DocSite) renderDoc(w http.ResponseWriter, r *http.Request, lang, slug string) {
	filePath := path.Join(lang, slug+".md")
	data, err := fs.ReadFile(d.fs, filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	content := markdownToHTML(string(data))
	pageTitle := slug
	if pages, ok := d.nav[lang]; ok {
		for _, p := range pages {
			if p.Slug == slug {
				pageTitle = p.Title
				break
			}
		}
	}
	d.renderPage(w, docTemplateData{
		Title:        pageTitle,
		Lang:         lang,
		CurrentSlug:  slug,
		Content:      content,
		Navigation:   d.nav,
		LanguageList: d.languages(),
	})
}

// languages returns the list of available languages
func (d *DocSite) languages() []string {
	langs := make([]string, 0, len(d.nav))
	for lang := range d.nav {
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	return langs
}

// renderPage renders the documentation page template
func (d *DocSite) renderPage(w http.ResponseWriter, data docTemplateData) {
	buf := &bytes.Buffer{}
	if err := docTemplate.Execute(buf, data); err != nil {
		http.Error(w, fmt.Sprintf("render docs: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

// docTemplateData represents the data for the documentation template
type docTemplateData struct {
	Title        string
	Lang         string
	CurrentSlug  string
	Content      template.HTML
	Navigation   map[string][]DocPage
	LanguageList []string
}

// docTemplate is the HTML template for documentation pages
var docTemplate = template.Must(template.New("docs").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>{{.Title}} | Plumego Docs</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { 
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif; 
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); 
      min-height: 100vh; 
      color: #161616; 
    }
    .container { 
      display: flex; 
      height: 100vh; 
      background: rgba(255, 255, 255, 0.95); 
      backdrop-filter: blur(10px); 
      margin: 2rem; 
      border-radius: 24px; 
      overflow: hidden; 
      box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25); 
    }
    nav { 
      width: 280px; 
      background: #ffffff; 
      border-right: 1px solid #e5e7eb; 
      padding: 24px 16px; 
      overflow-y: auto; 
    }
    nav h1 { 
      font-size: 20px; 
      margin: 0 0 20px 12px; 
      color: #111827; 
      font-weight: 700; 
    }
    nav h2 { 
      font-size: 12px; 
      margin: 20px 0 8px 12px; 
      text-transform: uppercase; 
      letter-spacing: 0.08em; 
      color: #6b7280; 
      font-weight: 600; 
    }
    nav ul { 
      list-style: none; 
      padding: 0 0 0 12px; 
      margin: 0; 
    }
    nav li { 
      margin: 4px 0; 
    }
    nav a { 
      color: #0f172a; 
      text-decoration: none; 
      font-size: 14px; 
      padding: 8px 12px; 
      display: block; 
      border-radius: 8px; 
      transition: all 0.2s ease; 
      font-weight: 500; 
    }
    nav a:hover { 
      background: #f1f5f9; 
      color: #4f46e5; 
    }
    nav a.active { 
      background: linear-gradient(135deg, #6366f1, #8b5cf6); 
      color: #ffffff; 
      box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3); 
    }
    main { 
      flex: 1; 
      padding: 40px; 
      overflow-y: auto; 
      background: #fafafa; 
    }
    main h1 { 
      font-size: 32px; 
      color: #111827; 
      margin-bottom: 24px; 
      font-weight: 800; 
      background: linear-gradient(135deg, #111827, #374151); 
      -webkit-background-clip: text; 
      -webkit-text-fill-color: transparent; 
    }
    main h2 { 
      font-size: 24px; 
      color: #111827; 
      margin: 32px 0 16px; 
      font-weight: 700; 
      border-bottom: 2px solid #e5e7eb; 
      padding-bottom: 8px; 
    }
    main h3 { 
      font-size: 18px; 
      color: #374151; 
      margin: 24px 0 12px; 
      font-weight: 600; 
    }
    main p { 
      line-height: 1.7; 
      color: #4b5563; 
      margin-bottom: 16px; 
    }
    main ul { 
      padding-left: 24px; 
      margin-bottom: 16px; 
    }
    main li { 
      margin: 8px 0; 
      color: #4b5563; 
      line-height: 1.6; 
    }
    main pre { 
      background: #1e293b; 
      color: #e2e8f0; 
      padding: 16px; 
      border-radius: 12px; 
      overflow-x: auto; 
      margin: 16px 0; 
      font-family: 'JetBrains Mono', 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace; 
      font-size: 14px; 
      line-height: 1.5; 
      border: 1px solid #334155; 
      box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1); 
    }
    main code { 
      font-family: 'JetBrains Mono', 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace; 
      background: #f3f4f6; 
      padding: 2px 6px; 
      border-radius: 4px; 
      font-size: 14px; 
      color: #dc2626; 
    }
    main pre code { 
      background: none; 
      padding: 0; 
      color: inherit; 
    }
    main a { 
      color: #4f46e5; 
      text-decoration: none; 
      font-weight: 600; 
      transition: all 0.2s ease; 
    }
    main a:hover { 
      color: #7c3aed; 
      text-decoration: underline; 
    }
    main blockquote { 
      border-left: 4px solid #6366f1; 
      padding-left: 16px; 
      margin: 16px 0; 
      color: #6b7280; 
      font-style: italic; 
      background: #f9fafb; 
      padding: 12px 16px; 
      border-radius: 0 8px 8px 0; 
    }
    .header-bar { 
      background: linear-gradient(135deg, #6366f1, #8b5cf6); 
      color: white; 
      padding: 16px 24px; 
      margin: -40px -40px 24px -40px; 
      display: flex; 
      align-items: center; 
      justify-content: space-between; 
    }
    .header-bar h2 { 
      margin: 0; 
      font-size: 18px; 
      font-weight: 700; 
      border: none; 
      color: white; 
    }
    .back-link { 
      color: white; 
      text-decoration: none; 
      font-size: 14px; 
      font-weight: 600; 
      padding: 8px 12px; 
      background: rgba(255, 255, 255, 0.2); 
      border-radius: 6px; 
      transition: all 0.2s ease; 
    }
    .back-link:hover { 
      background: rgba(255, 255, 255, 0.3); 
      text-decoration: none; 
    }
    @media (max-width: 768px) { 
      .container { 
        margin: 0; 
        border-radius: 0; 
        flex-direction: column; 
      }
      nav { 
        width: 100%; 
        height: auto; 
        max-height: 200px; 
        border-right: none; 
        border-bottom: 1px solid #e5e7eb; 
      }
      main { 
        padding: 24px; 
      }
      .header-bar { 
        margin: -24px -24px 16px -24px; 
        padding: 12px 16px; 
      }
    }
  </style>
</head>
<body>
  <div class="container">
    <nav>
      <h1>📚 Plumego Docs</h1>
      {{range .LanguageList}}
        <h2>{{.}}</h2>
        <ul>
          {{range $page := index $.Navigation .}}
            <li><a class="{{if and (eq $.Lang $page.Lang) (eq $.CurrentSlug $page.Slug)}}active{{end}}" href="/docs/{{$page.Lang}}/{{$page.Slug}}">{{$page.Title}}</a></li>
          {{end}}
        </ul>
      {{end}}
    </nav>
    <main>
      {{if .CurrentSlug}}
        <div class="header-bar">
          <h2>{{.Title}}</h2>
          <a href="/docs/{{.Lang}}" class="back-link">← Back to Index</a>
        </div>
      {{end}}
      {{.Content}}
    </main>
  </div>
</body>
</html>`))

// markdownToHTML converts markdown to HTML (simplified)
func markdownToHTML(md string) template.HTML {
	scanner := bufio.NewScanner(strings.NewReader(md))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var b strings.Builder
	inList := false
	inCode := false
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "```"):
			if inCode {
				b.WriteString("</code></pre>")
				inCode = false
			} else {
				b.WriteString("<pre><code>")
				inCode = true
			}
			continue
		case inCode:
			b.WriteString(html.EscapeString(line))
			b.WriteString("\n")
			continue
		case strings.HasPrefix(line, "- "):
			if !inList {
				b.WriteString("<ul>")
				inList = true
			}
			b.WriteString("<li>" + html.EscapeString(strings.TrimSpace(line[2:])) + "</li>")
			continue
		default:
			if inList {
				b.WriteString("</ul>")
				inList = false
			}
		}

		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "### "):
			b.WriteString("<h3>" + html.EscapeString(strings.TrimSpace(trimmed[4:])) + "</h3>")
		case strings.HasPrefix(trimmed, "## "):
			b.WriteString("<h2>" + html.EscapeString(strings.TrimSpace(trimmed[3:])) + "</h2>")
		case strings.HasPrefix(trimmed, "# "):
			b.WriteString("<h1>" + html.EscapeString(strings.TrimSpace(trimmed[2:])) + "</h1>")
		case trimmed == "":
			b.WriteString("<p></p>")
		default:
			b.WriteString("<p>" + html.EscapeString(line) + "</p>")
		}
	}
	if inList {
		b.WriteString("</ul>")
	}
	if inCode {
		b.WriteString("</code></pre>")
	}
	return template.HTML(b.String())
}

func main() {
	// Load configuration
	cfg := LoadConfig()

	// Initialize application context
	appCtx, err := NewAppContext(cfg)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	// Register routes
	if err := appCtx.RegisterRoutes(); err != nil {
		log.Fatalf("failed to register routes: %v", err)
	}

	// Start the application
	log.Printf("Starting Plumego Reference on %s", cfg.Addr)
	if err := appCtx.Start(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
