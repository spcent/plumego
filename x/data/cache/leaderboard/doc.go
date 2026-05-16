// Package leaderboard provides Plumego-local in-memory ranked-data cache
// behavior on top of the stable store/cache primitives.
//
// The package exposes sorted-set style operations, but it is not a Redis
// sorted-set compatibility layer. Missing-key behavior, lifecycle errors, and
// range performance are Plumego-local contracts documented by the package tests
// and x/data/cache module primer. LeaderboardConfig.DefaultTTL uses the package
// default when zero; use NoExpirationTTL when sorted sets should not expire.
package leaderboard
