package app

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"

	"cloud-vault/internal/web"
)

// RegisterRoutes registers all application routes.
func (a *App) RegisterRoutes() error {
	// Health endpoints
	root := newRouteReg(a.Core)
	root.get("/healthz", http.HandlerFunc(a.healthLive))
	root.get("/readyz", http.HandlerFunc(a.healthLive))
	if root.err != nil {
		return root.err
	}

	// API v1
	v1 := newGroupRouteReg(a.Core.Group("/api/v1"))

	// Health
	v1.get("/health", http.HandlerFunc(a.healthLive))

	// Documents — CRUD + versions
	v1.get("/documents", http.HandlerFunc(a.Docs.List))
	v1.post("/documents", http.HandlerFunc(a.Docs.Create))
	v1.get("/documents/:id", http.HandlerFunc(a.Docs.GetByID))
	v1.put("/documents/:id", http.HandlerFunc(a.Docs.Update))
	v1.delete("/documents/:id", http.HandlerFunc(a.Docs.Delete))
	v1.get("/documents/:id/versions", http.HandlerFunc(a.Docs.GetVersions))
	v1.get("/documents/:id/versions/:version", http.HandlerFunc(a.Docs.GetVersion))

	// Documents — V0.2 management (batch before :id to avoid ambiguity)
	v1.put("/documents/batch-status", http.HandlerFunc(a.Docs.BatchUpdateStatus))
	v1.put("/documents/:id/favorite", http.HandlerFunc(a.Docs.UpdateFavorite))
	v1.put("/documents/:id/status", http.HandlerFunc(a.Docs.UpdateStatus))
	v1.put("/documents/:id/review-status", http.HandlerFunc(a.Docs.UpdateReviewStatus))

	// Documents — tags
	v1.get("/documents/:id/tags", http.HandlerFunc(a.Tags.GetDocumentTags))
	v1.put("/documents/:id/tags", http.HandlerFunc(a.Tags.SetDocumentTags))
	v1.delete("/documents/:id/tags/:tag_id", http.HandlerFunc(a.Tags.RemoveDocumentTag))

	// Search — static sub-paths before /:id to avoid ambiguity.
	v1.get("/search/index-status", http.HandlerFunc(a.Search.GetIndexStatus))
	v1.post("/search/reindex", http.HandlerFunc(a.Search.Reindex))
	v1.get("/search/history", http.HandlerFunc(a.Search.GetHistory))
	v1.delete("/search/history", http.HandlerFunc(a.Search.ClearHistory))
	v1.get("/search", http.HandlerFunc(a.Search.Search))

	// Tags
	v1.get("/tags", http.HandlerFunc(a.Tags.ListTags))
	v1.post("/tags", http.HandlerFunc(a.Tags.CreateTag))
	v1.put("/tags/:id", http.HandlerFunc(a.Tags.UpdateTag))
	v1.delete("/tags/:id", http.HandlerFunc(a.Tags.DeleteTag))

	// Import jobs
	v1.get("/imports", http.HandlerFunc(a.Importer.ListJobs))
	v1.post("/imports", http.HandlerFunc(a.Importer.CreateJob))
	v1.get("/imports/:id", http.HandlerFunc(a.Importer.GetJob))
	v1.post("/imports/:id/start", http.HandlerFunc(a.Importer.StartJob))
	v1.post("/imports/:id/pause", http.HandlerFunc(a.Importer.PauseJob))
	v1.post("/imports/:id/cancel", http.HandlerFunc(a.Importer.CancelJob))
	v1.post("/imports/:id/retry", http.HandlerFunc(a.Importer.RetryJob))
	v1.get("/imports/:id/items", http.HandlerFunc(a.Importer.ListItems))

	if v1.err != nil {
		return v1.err
	}

	// Static SPA — must be last; catches all paths not matched above.
	// Register both the root and the wildcard so "/" is also served.
	spaHandler := web.NewHandler()
	if err := a.Core.Get("/", spaHandler); err != nil {
		return err
	}
	if err := a.Core.Get("/*filepath", spaHandler); err != nil {
		return err
	}

	return nil
}

func (a *App) healthLive(w http.ResponseWriter, r *http.Request) {
	logWriteErr(a.Core.Logger(), contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"status":    "ok",
		"service":   "cloud-vault",
		"version":   a.Cfg.App.Version,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, nil))
}

func logWriteErr(logger plumelog.StructuredLogger, err error) {
	if err != nil && logger != nil {
		logger.Error("write response", plumelog.Fields{"error": err.Error()})
	}
}
