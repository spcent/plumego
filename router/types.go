package router

import (
	"github.com/spcent/plumego/middleware"
)

// MatchResult represents the result of route matching
type MatchResult struct {
	Handler          Handler
	ParamValues      []string
	ParamKeys        []string
	RouteMiddlewares []middleware.Middleware // Now using direct type instead of any
}
