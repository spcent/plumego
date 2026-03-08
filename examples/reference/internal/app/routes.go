package app

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/examples/reference/internal/docsite"
	"github.com/spcent/plumego/examples/reference/internal/handler"
	"github.com/spcent/plumego/frontend"
	"github.com/spcent/plumego/health"
)

// RegisterRoutes wires all HTTP routes for the reference application.
func (a *App) RegisterRoutes() error {
	if err := a.registerDocs(); err != nil {
		return err
	}
	if err := a.registerPages(); err != nil {
		return err
	}
	if err := a.registerStatic(); err != nil {
		return err
	}
	a.registerHealth()
	a.registerWebSocket()
	a.registerAPI()
	a.registerTest()
	a.registerMetrics()
	return nil
}

func (a *App) registerDocs() error {
	if !a.Cfg.EnableDocs {
		return nil
	}
	footerLinks := docsFooterLinks()
	site, err := docsite.New(a.StaticFS, footerLinks)
	if err != nil {
		log.Printf("docs disabled: %v", err)
		return nil // docs are optional
	}
	a.Core.Get("/docs", site.Handler())
	a.Core.Get("/docs/*path", site.Handler())
	return nil
}

func (a *App) registerPages() error {
	tmpl, err := template.ParseFS(a.StaticFS,
		"templates/partials.html",
		"templates/index.html",
		"templates/webhooks.html",
		"templates/health.html",
	)
	if err != nil {
		return fmt.Errorf("parse page templates: %w", err)
	}
	pages := handler.NewPageHandler(tmpl)
	a.Core.Get("/", pages.Home)
	a.Core.Get("/webhooks", pages.Webhooks)
	a.Core.Get("/health/detailed", pages.HealthDetail)
	return nil
}

func (a *App) registerStatic() error {
	uiFS, err := fs.Sub(a.StaticFS, ".")
	if err != nil {
		return fmt.Errorf("get ui directory: %w", err)
	}
	if err := frontend.RegisterFS(a.Core.Router(), http.FS(uiFS), frontend.WithPrefix("/")); err != nil {
		return fmt.Errorf("register frontend: %w", err)
	}

	staticSubFS, err := fs.Sub(uiFS, "static")
	if err != nil {
		return fmt.Errorf("get static subdirectory: %w", err)
	}
	if err := frontend.RegisterFS(a.Core.Router(), http.FS(staticSubFS), frontend.WithPrefix("/static")); err != nil {
		return fmt.Errorf("register static files: %w", err)
	}
	return nil
}

func (a *App) registerHealth() {
	a.Core.Get("/health/ready", health.ReadinessHandler().ServeHTTP)
	a.Core.Get("/health/build", health.BuildInfoHandler().ServeHTTP)
}

func (a *App) registerWebSocket() {
	wsCfg := core.DefaultWebSocketConfig()
	wsCfg.Secret = []byte(a.Cfg.WebSocketSecret)
	if _, err := a.Core.ConfigureWebSocketWithOptions(wsCfg); err != nil {
		log.Printf("websocket disabled: %v", err)
	}
}

func (a *App) registerAPI() {
	api := handler.APIHandler{}
	a.Core.Get("/api/hello", api.Hello)
	a.Core.Get("/api/status", api.Status)
	a.Core.Get("/api/test", api.Test)
}

func (a *App) registerTest() {
	test := handler.NewTestHandler(a.Bus)
	a.Core.Get("/test/pubsub", test.PubSub)
	a.Core.Post("/test/webhook", test.Webhook)
}

func (a *App) registerMetrics() {
	if a.Prom != nil {
		a.Core.Get("/metrics", a.Prom.Handler().ServeHTTP)
	}
}

// docsFooterLinks returns the footer links rendered on documentation pages.
func docsFooterLinks() []docsite.FooterLink {
	return []docsite.FooterLink{
		{Href: "/docs", Label: "Docs"},
		{Href: "/health/ready", Label: "Health Check"},
		{Href: "/health/build", Label: "Build Info"},
		{Href: "/api/status", Label: "System Status"},
	}
}
