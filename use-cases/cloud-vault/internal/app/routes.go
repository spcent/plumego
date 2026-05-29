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

	// Health (public)
	v1.get("/health", http.HandlerFunc(a.healthLive))

	// V0.7: Auth endpoints
	// Setup, Status, and Login are always public
	v1.post("/auth/setup", http.HandlerFunc(a.Auth.Setup))
	v1.get("/auth/status", http.HandlerFunc(a.Auth.GetStatus))
	v1.post("/auth/login", http.HandlerFunc(a.Auth.Login))
	// Other auth endpoints require authentication when auth is enabled
	v1.post("/auth/logout", a.requireAuth(http.HandlerFunc(a.Auth.Logout)))
	v1.get("/auth/me", a.requireAuth(http.HandlerFunc(a.Auth.GetCurrentUser)))
	v1.put("/auth/me", a.requireAuth(http.HandlerFunc(a.Auth.UpdateProfile)))
	v1.post("/auth/change-password", a.requireAuth(http.HandlerFunc(a.Auth.ChangePassword)))
	v1.get("/auth/sessions", a.requireAuth(http.HandlerFunc(a.Auth.ListSessions)))
	v1.post("/auth/sessions/revoke-all", a.requireAuth(http.HandlerFunc(a.Auth.RevokeAllSessions)))

	// Documents — CRUD + versions (protected)
	v1.get("/documents", a.requireAuth(http.HandlerFunc(a.Docs.List)))
	v1.post("/documents", a.requireAuth(http.HandlerFunc(a.Docs.Create)))
	v1.get("/documents/:id", a.requireAuth(http.HandlerFunc(a.Docs.GetByID)))
	v1.put("/documents/:id", a.requireAuth(http.HandlerFunc(a.Docs.Update)))
	v1.delete("/documents/:id", a.requireAuth(http.HandlerFunc(a.Docs.Delete)))
	v1.get("/documents/:id/versions", a.requireAuth(http.HandlerFunc(a.Docs.GetVersions)))
	v1.get("/documents/:id/versions/:version", a.requireAuth(http.HandlerFunc(a.Docs.GetVersion)))

	// Documents — V0.2 management (batch before :id to avoid ambiguity, protected)
	v1.put("/documents/batch-status", a.requireAuth(http.HandlerFunc(a.Docs.BatchUpdateStatus)))
	v1.put("/documents/:id/favorite", a.requireAuth(http.HandlerFunc(a.Docs.UpdateFavorite)))
	v1.put("/documents/:id/status", a.requireAuth(http.HandlerFunc(a.Docs.UpdateStatus)))
	v1.put("/documents/:id/review-status", a.requireAuth(http.HandlerFunc(a.Docs.UpdateReviewStatus)))

	// Documents — tags (protected)
	v1.get("/documents/:id/tags", a.requireAuth(http.HandlerFunc(a.Tags.GetDocumentTags)))
	v1.put("/documents/:id/tags", a.requireAuth(http.HandlerFunc(a.Tags.SetDocumentTags)))
	v1.delete("/documents/:id/tags/:tag_id", a.requireAuth(http.HandlerFunc(a.Tags.RemoveDocumentTag)))

	// Search — static sub-paths before /:id to avoid ambiguity (protected).
	v1.get("/search/index-status", a.requireAuth(http.HandlerFunc(a.Search.GetIndexStatus)))
	v1.post("/search/reindex", a.requireAuth(http.HandlerFunc(a.Search.Reindex)))
	v1.get("/search/history", a.requireAuth(http.HandlerFunc(a.Search.GetHistory)))
	v1.delete("/search/history", a.requireAuth(http.HandlerFunc(a.Search.ClearHistory)))
	v1.get("/search", a.requireAuth(http.HandlerFunc(a.Search.Search)))

	// Tags (protected)
	v1.get("/tags", a.requireAuth(http.HandlerFunc(a.Tags.ListTags)))
	v1.post("/tags", a.requireAuth(http.HandlerFunc(a.Tags.CreateTag)))
	v1.put("/tags/:id", a.requireAuth(http.HandlerFunc(a.Tags.UpdateTag)))
	v1.delete("/tags/:id", a.requireAuth(http.HandlerFunc(a.Tags.DeleteTag)))

	// Import jobs (protected)
	v1.get("/imports", a.requireAuth(http.HandlerFunc(a.Importer.ListJobs)))
	v1.post("/imports", a.requireAuth(http.HandlerFunc(a.Importer.CreateJob)))
	v1.get("/imports/:id", a.requireAuth(http.HandlerFunc(a.Importer.GetJob)))
	v1.post("/imports/:id/start", a.requireAuth(http.HandlerFunc(a.Importer.StartJob)))
	v1.post("/imports/:id/pause", a.requireAuth(http.HandlerFunc(a.Importer.PauseJob)))
	v1.post("/imports/:id/cancel", a.requireAuth(http.HandlerFunc(a.Importer.CancelJob)))
	v1.post("/imports/:id/retry", a.requireAuth(http.HandlerFunc(a.Importer.RetryJob)))
	v1.get("/imports/:id/items", a.requireAuth(http.HandlerFunc(a.Importer.ListItems)))

	// V0.4: Organize — static sub-paths before :id to avoid ambiguity (protected).
	v1.post("/organize/detect-duplicates", a.requireAuth(http.HandlerFunc(a.Organize.DetectDuplicates)))
	v1.get("/organize/duplicates", a.requireAuth(http.HandlerFunc(a.Organize.ListDuplicates)))
	v1.post("/organize/duplicates/resolve", a.requireAuth(http.HandlerFunc(a.Organize.ResolveDuplicates)))
	v1.post("/organize/detect-similarity", a.requireAuth(http.HandlerFunc(a.Organize.DetectSimilarity)))
	v1.post("/organize/suggest-tags", a.requireAuth(http.HandlerFunc(a.Organize.SuggestTags)))
	v1.post("/organize/build-topics", a.requireAuth(http.HandlerFunc(a.Organize.BuildTopics)))
	v1.post("/organize/score-quality", a.requireAuth(http.HandlerFunc(a.Organize.ScoreQuality)))
	v1.post("/organize/detect-prompt-candidates", a.requireAuth(http.HandlerFunc(a.Organize.DetectPromptCandidates)))
	v1.post("/organize/run-all", a.requireAuth(http.HandlerFunc(a.Organize.RunAll)))
	v1.get("/organize/jobs", a.requireAuth(http.HandlerFunc(a.Organize.ListJobs)))
	v1.get("/organize/jobs/:id", a.requireAuth(http.HandlerFunc(a.Organize.GetJob)))
	v1.post("/organize/similarity/:id/ignore", a.requireAuth(http.HandlerFunc(a.Organize.IgnoreSimilarity)))
	v1.post("/organize/similarity/:id/confirm", a.requireAuth(http.HandlerFunc(a.Organize.ConfirmSimilarity)))

	// V0.4: Topics (protected)
	v1.get("/topics", a.requireAuth(http.HandlerFunc(a.Organize.ListTopics)))
	v1.get("/topics/:id", a.requireAuth(http.HandlerFunc(a.Organize.GetTopic)))

	// V0.4: Document similar + tag suggestions (under /documents/:id, protected)
	v1.get("/documents/:id/similar", a.requireAuth(http.HandlerFunc(a.Organize.GetSimilarDocuments)))
	v1.get("/documents/:id/tag-suggestions", a.requireAuth(http.HandlerFunc(a.Organize.GetTagSuggestions)))

	// V0.4: Tag suggestions actions — batch before :id to avoid ambiguity (protected).
	v1.post("/tag-suggestions/batch/accept", a.requireAuth(http.HandlerFunc(a.Organize.BatchAcceptTagSuggestions)))
	v1.post("/tag-suggestions/:id/accept", a.requireAuth(http.HandlerFunc(a.Organize.AcceptTagSuggestion)))
	v1.post("/tag-suggestions/:id/reject", a.requireAuth(http.HandlerFunc(a.Organize.RejectTagSuggestion)))

	// V0.4: Collections — from-search before :id to avoid ambiguity (protected).
	v1.post("/collections/from-search", a.requireAuth(http.HandlerFunc(a.Collection.CreateFromSearch)))
	v1.get("/collections", a.requireAuth(http.HandlerFunc(a.Collection.List)))
	v1.post("/collections", a.requireAuth(http.HandlerFunc(a.Collection.Create)))
	v1.get("/collections/:id", a.requireAuth(http.HandlerFunc(a.Collection.GetByID)))
	v1.put("/collections/:id", a.requireAuth(http.HandlerFunc(a.Collection.Update)))
	v1.delete("/collections/:id", a.requireAuth(http.HandlerFunc(a.Collection.Delete)))
	v1.post("/collections/:id/documents", a.requireAuth(http.HandlerFunc(a.Collection.AddDocument)))
	v1.put("/collections/:id/documents/reorder", a.requireAuth(http.HandlerFunc(a.Collection.Reorder)))
	v1.delete("/collections/:id/documents/:document_id", a.requireAuth(http.HandlerFunc(a.Collection.RemoveDocument)))

	// V0.4: Review queue (protected)
	v1.get("/review/queue", a.requireAuth(http.HandlerFunc(a.Organize.GetReviewQueue)))

	// V0.6: System observability (protected).
	v1.get("/system/health", a.requireAuth(http.HandlerFunc(a.System.Health)))
	v1.get("/system/stats", a.requireAuth(http.HandlerFunc(a.System.Stats)))
	v1.post("/system/doctor", a.requireAuth(http.HandlerFunc(a.System.Doctor)))

	// V0.8: Backup and restore endpoints (protected).
	v1.post("/system/backup", a.requireAuth(http.HandlerFunc(a.Backup.CreateBackup)))
	v1.get("/system/backups", a.requireAuth(http.HandlerFunc(a.Backup.ListBackups)))
	v1.get("/system/backups/:name/download", a.requireAuth(http.HandlerFunc(a.Backup.DownloadBackup)))
	v1.delete("/system/backups/:name", a.requireAuth(http.HandlerFunc(a.Backup.DeleteBackup)))
	v1.post("/system/restore", a.requireAuth(http.HandlerFunc(a.Backup.RestoreBackup)))

	// V0.5: AI — static sub-paths before :id to avoid ambiguity (protected).
	v1.post("/ai/tasks/summary", a.requireAuth(http.HandlerFunc(a.AI.EnqueueSummary)))
	v1.post("/ai/tasks/qa", a.requireAuth(http.HandlerFunc(a.AI.EnqueueQA)))
	v1.post("/ai/tasks/prompt-extract", a.requireAuth(http.HandlerFunc(a.AI.EnqueuePromptExtract)))
	v1.get("/ai/tasks", a.requireAuth(http.HandlerFunc(a.AI.ListTasks)))
	v1.get("/ai/tasks/:id", a.requireAuth(http.HandlerFunc(a.AI.GetTask)))
	v1.post("/ai/tasks/:id/cancel", a.requireAuth(http.HandlerFunc(a.AI.CancelTask)))
	v1.get("/ai/prompts", a.requireAuth(http.HandlerFunc(a.AI.ListPrompts)))
	v1.get("/ai/prompts/:id", a.requireAuth(http.HandlerFunc(a.AI.GetPrompt)))
	v1.delete("/ai/prompts/:id", a.requireAuth(http.HandlerFunc(a.AI.DeletePrompt)))
	v1.get("/ai/documents/:id/summary", a.requireAuth(http.HandlerFunc(a.AI.GetDocumentSummary)))

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

// requireAuth wraps a handler with auth middleware if auth is enabled.
// If auth is disabled, returns the handler unchanged.
func (a *App) requireAuth(h http.Handler) http.Handler {
	if a.authMiddleware != nil {
		return a.authMiddleware(h)
	}
	return h
}
