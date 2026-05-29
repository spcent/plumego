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

	// V0.4: Organize — static sub-paths before :id to avoid ambiguity.
	v1.post("/organize/detect-duplicates", http.HandlerFunc(a.Organize.DetectDuplicates))
	v1.get("/organize/duplicates", http.HandlerFunc(a.Organize.ListDuplicates))
	v1.post("/organize/duplicates/resolve", http.HandlerFunc(a.Organize.ResolveDuplicates))
	v1.post("/organize/detect-similarity", http.HandlerFunc(a.Organize.DetectSimilarity))
	v1.post("/organize/suggest-tags", http.HandlerFunc(a.Organize.SuggestTags))
	v1.post("/organize/build-topics", http.HandlerFunc(a.Organize.BuildTopics))
	v1.post("/organize/score-quality", http.HandlerFunc(a.Organize.ScoreQuality))
	v1.post("/organize/detect-prompt-candidates", http.HandlerFunc(a.Organize.DetectPromptCandidates))
	v1.post("/organize/run-all", http.HandlerFunc(a.Organize.RunAll))
	v1.get("/organize/jobs", http.HandlerFunc(a.Organize.ListJobs))
	v1.get("/organize/jobs/:id", http.HandlerFunc(a.Organize.GetJob))
	v1.post("/organize/similarity/:id/ignore", http.HandlerFunc(a.Organize.IgnoreSimilarity))
	v1.post("/organize/similarity/:id/confirm", http.HandlerFunc(a.Organize.ConfirmSimilarity))

	// V0.4: Topics
	v1.get("/topics", http.HandlerFunc(a.Organize.ListTopics))
	v1.get("/topics/:id", http.HandlerFunc(a.Organize.GetTopic))

	// V0.4: Document similar + tag suggestions (under /documents/:id)
	v1.get("/documents/:id/similar", http.HandlerFunc(a.Organize.GetSimilarDocuments))
	v1.get("/documents/:id/tag-suggestions", http.HandlerFunc(a.Organize.GetTagSuggestions))

	// V0.4: Tag suggestions actions — batch before :id to avoid ambiguity.
	v1.post("/tag-suggestions/batch/accept", http.HandlerFunc(a.Organize.BatchAcceptTagSuggestions))
	v1.post("/tag-suggestions/:id/accept", http.HandlerFunc(a.Organize.AcceptTagSuggestion))
	v1.post("/tag-suggestions/:id/reject", http.HandlerFunc(a.Organize.RejectTagSuggestion))

	// V0.4: Collections — from-search before :id to avoid ambiguity.
	v1.post("/collections/from-search", http.HandlerFunc(a.Collection.CreateFromSearch))
	v1.get("/collections", http.HandlerFunc(a.Collection.List))
	v1.post("/collections", http.HandlerFunc(a.Collection.Create))
	v1.get("/collections/:id", http.HandlerFunc(a.Collection.GetByID))
	v1.put("/collections/:id", http.HandlerFunc(a.Collection.Update))
	v1.delete("/collections/:id", http.HandlerFunc(a.Collection.Delete))
	v1.post("/collections/:id/documents", http.HandlerFunc(a.Collection.AddDocument))
	v1.put("/collections/:id/documents/reorder", http.HandlerFunc(a.Collection.Reorder))
	v1.delete("/collections/:id/documents/:document_id", http.HandlerFunc(a.Collection.RemoveDocument))

	// V0.4: Review queue
	v1.get("/review/queue", http.HandlerFunc(a.Organize.GetReviewQueue))

	// V0.6: System observability.
	v1.get("/system/health", http.HandlerFunc(a.System.Health))
	v1.get("/system/stats", http.HandlerFunc(a.System.Stats))
	v1.post("/system/doctor", http.HandlerFunc(a.System.Doctor))

	// V0.5: AI — static sub-paths before :id to avoid ambiguity.
	v1.post("/ai/tasks/summary", http.HandlerFunc(a.AI.EnqueueSummary))
	v1.post("/ai/tasks/qa", http.HandlerFunc(a.AI.EnqueueQA))
	v1.post("/ai/tasks/prompt-extract", http.HandlerFunc(a.AI.EnqueuePromptExtract))
	v1.get("/ai/tasks", http.HandlerFunc(a.AI.ListTasks))
	v1.get("/ai/tasks/:id", http.HandlerFunc(a.AI.GetTask))
	v1.post("/ai/tasks/:id/cancel", http.HandlerFunc(a.AI.CancelTask))
	v1.get("/ai/prompts", http.HandlerFunc(a.AI.ListPrompts))
	v1.get("/ai/prompts/:id", http.HandlerFunc(a.AI.GetPrompt))
	v1.delete("/ai/prompts/:id", http.HandlerFunc(a.AI.DeletePrompt))
	v1.get("/ai/documents/:id/summary", http.HandlerFunc(a.AI.GetDocumentSummary))

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
