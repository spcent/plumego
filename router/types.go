package router

// MatchResult represents the result of route matching
type MatchResult struct {
	Handler          Handler
	ParamValues      []string
	ParamKeys        []string
	RouteMiddlewares []interface{} // Using interface{} to avoid import cycles
}
