package handler

import (
	"html/template"
	"log"
	"net/http"
)

// PageLink is a navigation link rendered in page footers.
type PageLink struct {
	Href  string
	Label string
}

// PageHeader contains the banner content for a rendered HTML page.
type PageHeader struct {
	IconClass string
	Title     string
	Subtitle  string
}

// PageData is the template payload for the main HTML pages.
type PageData struct {
	Header      PageHeader
	FooterLinks []PageLink
}

// DefaultFooterLinks returns the standard footer navigation links shared
// across all UI pages.
func DefaultFooterLinks() []PageLink {
	return []PageLink{
		{Href: "/docs", Label: "Docs"},
		{Href: "/health/ready", Label: "Health Check"},
		{Href: "/health/build", Label: "Build Info"},
		{Href: "/api/status", Label: "System Status"},
	}
}

// PageHandler serves the main HTML UI pages using a shared template set.
type PageHandler struct {
	tmpl     *template.Template
	home     PageData
	webhooks PageData
	health   PageData
}

// NewPageHandler creates a PageHandler with pre-built page data.
func NewPageHandler(tmpl *template.Template) *PageHandler {
	links := DefaultFooterLinks()
	return &PageHandler{
		tmpl: tmpl,
		home: PageData{
			Header: PageHeader{
				IconClass: "fas fa-feather",
				Title:     "Plumego Reference",
				Subtitle:  "Modern Go Web Framework Demo Platform - Showcasing Elegant Architecture Design",
			},
			FooterLinks: links,
		},
		webhooks: PageData{
			Header: PageHeader{
				IconClass: "fas fa-plug",
				Title:     "Webhook Management",
				Subtitle:  "Test and verify webhook integration, supporting GitHub and Stripe with automatic retry and status monitoring.",
			},
			FooterLinks: links,
		},
		health: PageData{
			Header: PageHeader{
				IconClass: "fas fa-heartbeat",
				Title:     "Health Overview",
				Subtitle:  "Status, readiness, and build diagnostics for the reference service.",
			},
			FooterLinks: links,
		},
	}
}

// Home renders the landing page.
func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "index.html", h.home); err != nil {
		log.Printf("failed to render homepage: %v", err)
	}
}

// Webhooks renders the webhook management page.
func (h *PageHandler) Webhooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "webhooks.html", h.webhooks); err != nil {
		log.Printf("failed to render webhooks page: %v", err)
	}
}

// HealthDetail renders the health overview page.
func (h *PageHandler) HealthDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "health.html", h.health); err != nil {
		log.Printf("failed to render health page: %v", err)
	}
}
