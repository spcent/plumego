package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	plumelog "github.com/spcent/plumego/log"

	"guardus/internal/config"
	"guardus/internal/storage"
	"guardus/internal/storage/common"
	"guardus/internal/storage/common/paging"
)

const statusesCacheTTL = 10 * time.Second

// statusCache is a tiny TTL cache for the /endpoints/statuses payload.
//
// Upstream gatus uses gocache; guardus uses sync.Map + monotonic deadline
// to keep the dependency footprint flat. Items expire on lookup; eviction is
// passive.
type statusCache struct {
	mu      sync.RWMutex
	entries map[string]statusCacheEntry
}

type statusCacheEntry struct {
	body    []byte
	expires time.Time
}

func (c *statusCache) get(key string) ([]byte, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(entry.expires) {
		return nil, false
	}
	return entry.body, true
}

func (c *statusCache) set(key string, body []byte, ttl time.Duration) {
	c.mu.Lock()
	if c.entries == nil {
		c.entries = make(map[string]statusCacheEntry)
	}
	c.entries[key] = statusCacheEntry{body: body, expires: time.Now().Add(ttl)}
	c.mu.Unlock()
}

// EndpointStatuses returns paginated endpoint statuses.
//
// The response is cached for statusesCacheTTL because GetAllEndpointStatuses
// is the hottest read path (the SPA polls it every 10s).
func EndpointStatuses(cfg *config.Config, store storage.Store, logger plumelog.StructuredLogger) http.Handler {
	cache := &statusCache{}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, pageSize := ExtractPageAndPageSize(r, cfg.Storage.MaximumNumberOfResults)
		key := fmt.Sprintf("endpoint-status-%d-%d", page, pageSize)
		if data, ok := cache.get(key); ok {
			WriteRawJSON(w, http.StatusOK, data)
			return
		}
		statuses, err := store.GetAllEndpointStatuses(paging.NewEndpointStatusParams().WithResults(page, pageSize))
		if err != nil {
			logger.Error("EndpointStatuses: failed to retrieve endpoint statuses", plumelog.Fields{"op": "handler.EndpointStatuses", "err": err.Error()})
			WriteErrorString(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		body, err := json.Marshal(statuses)
		if err != nil {
			logger.Error("EndpointStatuses: marshal failure", plumelog.Fields{"op": "handler.EndpointStatuses", "err": err.Error()})
			WriteErrorString(w, r, http.StatusInternalServerError, "unable to marshal object to JSON")
			return
		}
		cache.set(key, body, statusesCacheTTL)
		WriteRawJSON(w, http.StatusOK, body)
	})
}

// EndpointStatus returns a single endpoint status by URL-decoded key.
func EndpointStatus(cfg *config.Config, store storage.Store, logger plumelog.StructuredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, pageSize := ExtractPageAndPageSize(r, cfg.Storage.MaximumNumberOfResults)
		key, err := PathKey(r)
		if err != nil {
			logger.Error("EndpointStatus: failed to decode key", plumelog.Fields{"op": "handler.EndpointStatus", "err": err.Error()})
			WriteErrorString(w, r, http.StatusBadRequest, "invalid key encoding")
			return
		}
		params := paging.NewEndpointStatusParams().
			WithResults(page, pageSize).
			WithEvents(1, cfg.Storage.MaximumNumberOfEvents)
		status, err := store.GetEndpointStatusByKey(key, params)
		if err != nil {
			if errors.Is(err, common.ErrEndpointNotFound) {
				WriteErrorString(w, r, http.StatusNotFound, err.Error())
				return
			}
			logger.Error("EndpointStatus: failed to retrieve endpoint status", plumelog.Fields{"op": "handler.EndpointStatus", "key": key, "err": err.Error()})
			WriteErrorString(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		if status == nil {
			WriteErrorString(w, r, http.StatusNotFound, "not found")
			return
		}
		body, err := json.Marshal(status)
		if err != nil {
			logger.Error("EndpointStatus: marshal failure", plumelog.Fields{"op": "handler.EndpointStatus", "key": key, "err": err.Error()})
			WriteErrorString(w, r, http.StatusInternalServerError, "unable to marshal object to JSON")
			return
		}
		WriteRawJSON(w, http.StatusOK, body)
	})
}
