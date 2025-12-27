package main

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"fmt"
	"html"
	"html/template"
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

func main() {
	// Optional in-process pub/sub powers inbound webhook fan-out and WebSocket broadcasts.
	bus := pubsub.New()

	// Outbound webhook management runs alongside the HTTP server.
	webhookStore := webhookout.NewMemStore()
	webhookCfg := webhookout.ConfigFromEnv()
	webhookCfg.Enabled = true
	webhookSvc := webhookout.NewService(webhookStore, webhookCfg)

	// Prometheus + OpenTelemetry hooks plugged into logging middleware.
	prom := metrics.NewPrometheusCollector("plumego_example")
	tracer := metrics.NewOpenTelemetryTracer("plumego-example")

	app := core.New(
		core.WithAddr(":8080"),
		core.WithDebug(),
		core.WithPubSub(bus),
		core.WithMetricsCollector(prom),
		core.WithTracer(tracer),
		core.WithWebhookIn(core.WebhookInConfig{
			Enabled:           true,
			Pub:               bus,
			GitHubSecret:      config.GetString("GITHUB_WEBHOOK_SECRET", "dev-github-secret"),
			StripeSecret:      config.GetString("STRIPE_WEBHOOK_SECRET", "whsec_dev"),
			MaxBodyBytes:      1 << 20,
			StripeTolerance:   5 * time.Minute,
			TopicPrefixGitHub: "in.github.",
			TopicPrefixStripe: "in.stripe.",
		}),
		core.WithWebhookOut(core.WebhookOutConfig{
			Enabled:          true,
			Service:          webhookSvc,
			TriggerToken:     config.GetString("WEBHOOK_TRIGGER_TOKEN", "dev-trigger"),
			BasePath:         "/webhooks",
			IncludeStats:     true,
			DefaultPageLimit: 50,
		}),
	)

	app.EnableRecovery()
	app.EnableLogging()
	app.EnableCORS()

	docSite, docErr := loadDocSite()
	if docErr != nil {
		log.Printf("docs disabled: %v", docErr)
	} else {
		app.Get("/docs", docSite.handler())
		app.Get("/docs/*path", docSite.handler())
	}

	// Static frontend served from the embedded UI folder.
	_ = frontend.RegisterFS(app.Router(), http.FS(staticFS), frontend.WithPrefix("/"))

	// Minimal health endpoints for orchestration hooks.
	app.GetHandler("/health/ready", health.ReadinessHandler())
	app.GetHandler("/health/build", health.BuildInfoHandler())

	// WebSocket hub with broadcast endpoint and simple echoing demo.
	wsCfg := core.DefaultWebSocketConfig()
	wsCfg.Secret = []byte(config.GetString("WS_SECRET", "dev-secret"))
	_, err := app.ConfigureWebSocketWithOptions(wsCfg)
	if err != nil {
		log.Fatalf("configure websocket: %v", err)
	}
	webhookSvc.Start(context.Background())
	defer webhookSvc.Stop()

	// Example API route demonstrating middleware and tracing hooks.
	app.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Served-At", time.Now().Format(time.RFC3339))
		w.Write([]byte("hello from plumego"))
	})

	// Expose metrics for scraping.
	app.GetHandler("/metrics", prom.Handler())

	if err := app.Boot(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

type docPage struct {
	Lang  string
	Slug  string
	Title string
}

type docSite struct {
	fs          fs.FS
	nav         map[string][]docPage
	defaultLang string
}

func loadDocSite() (*docSite, error) {
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
	return &docSite{fs: f, nav: nav, defaultLang: defaultLang}, nil
}

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

func buildDocNav(f fs.FS) (map[string][]docPage, error) {
	nav := make(map[string][]docPage)
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
			nav[lang] = append(nav[lang], docPage{Lang: lang, Slug: slug, Title: title})
		}
		sort.Slice(nav[lang], func(i, j int) bool { return nav[lang][i].Title < nav[lang][j].Title })
	}
	return nav, nil
}

func extractTitle(f fs.FS, filePath string) string {
	file, err := f.Open(filePath)
	if err != nil {
		return filePath
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			return strings.TrimLeft(line, "# ")
		}
	}
	return strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
}

func (d *docSite) handler() http.HandlerFunc {
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

func (d *docSite) renderIndex(w http.ResponseWriter, lang string) {
	content := markdownToHTML("# Documentation\n选择左侧导航以浏览各模块，或直接访问 /docs/{lang}/{module}。\n\nSelect a language and module on the left to start reading.")
	d.renderPage(w, docTemplateData{
		Title:        "Plumego Docs",
		Lang:         lang,
		CurrentSlug:  "",
		Content:      content,
		Navigation:   d.nav,
		LanguageList: d.languages(),
	})
}

func (d *docSite) renderDoc(w http.ResponseWriter, r *http.Request, lang, slug string) {
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

func (d *docSite) languages() []string {
	langs := make([]string, 0, len(d.nav))
	for lang := range d.nav {
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	return langs
}

type docTemplateData struct {
	Title        string
	Lang         string
	CurrentSlug  string
	Content      template.HTML
	Navigation   map[string][]docPage
	LanguageList []string
}

var docTemplate = template.Must(template.New("docs").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>{{.Title}} | Plumego Docs</title>
  <style>
    body { font-family: "Helvetica Neue", Arial, sans-serif; margin: 0; display: flex; background: #f7f7f8; color: #161616; }
    nav { width: 280px; background: #ffffff; border-right: 1px solid #e5e7eb; padding: 24px 16px; box-sizing: border-box; height: 100vh; overflow-y: auto; }
    nav h1 { font-size: 18px; margin: 0 0 12px 12px; }
    nav h2 { font-size: 14px; margin: 16px 0 8px 12px; text-transform: uppercase; letter-spacing: 0.08em; color: #6b7280; }
    nav ul { list-style: none; padding: 0 0 0 12px; margin: 0; }
    nav li { margin: 6px 0; }
    nav a { color: #0f172a; text-decoration: none; font-size: 14px; padding: 6px 8px; display: inline-block; border-radius: 6px; }
    nav a.active { background: #0ea5e9; color: #ffffff; }
    main { flex: 1; padding: 32px; overflow-y: auto; }
    main h1, main h2, main h3 { color: #111827; }
    main p { line-height: 1.6; }
    main pre { background: #0f172a; color: #e2e8f0; padding: 12px; border-radius: 8px; overflow-x: auto; }
    main code { font-family: "JetBrains Mono", "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace; }
    main ul { padding-left: 20px; }
  </style>
</head>
<body>
  <nav>
    <h1>Plumego Docs</h1>
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
    {{.Content}}
  </main>
</body>
</html>`))

func (d *docSite) renderPage(w http.ResponseWriter, data docTemplateData) {
	buf := &bytes.Buffer{}
	if err := docTemplate.Execute(buf, data); err != nil {
		http.Error(w, fmt.Sprintf("render docs: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func markdownToHTML(md string) template.HTML {
	scanner := bufio.NewScanner(strings.NewReader(md))
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
