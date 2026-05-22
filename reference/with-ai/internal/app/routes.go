package app

import (
	"net/http"

	"with-ai/internal/handler"
)

// RegisterRoutes wires all HTTP routes for the with-ai demo.
func (a *App) RegisterRoutes() error {
	ai := handler.NewAIHandler()

	if err := a.Core.Post("/api/chat", http.HandlerFunc(ai.Chat)); err != nil {
		return err
	}
	return a.Core.Get("/api/ai/status", http.HandlerFunc(ai.Status))
}
