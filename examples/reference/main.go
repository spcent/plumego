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
		core.WithRecovery(),
		core.WithLogging(),
		core.WithCORS(),
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
	template    *template.Template
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

	// Default to English
	defaultLang := "en"
	if _, ok := nav[defaultLang]; !ok {
		for lang := range nav {
			defaultLang = lang
			break
		}
	}

	// Load template from embedded filesystem
	tmplPath := filepath.Join("ui", "templates", "docs.html")
	tmplContent, err := fs.ReadFile(staticFS, tmplPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	tmpl, err := template.New("docs").Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &DocSite{
		fs:          f,
		nav:         nav,
		defaultLang: defaultLang,
		template:    tmpl,
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
	if err := d.template.Execute(buf, data); err != nil {
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
