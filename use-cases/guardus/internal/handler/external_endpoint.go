package handler

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"

	plumelog "github.com/spcent/plumego/log"

	"guardus/internal/config"
	"guardus/internal/domain/endpoint"
	"guardus/internal/metrics"
	"guardus/internal/storage"
	"guardus/internal/storage/common"
	"guardus/internal/watchdog"
)

// CreateExternalEndpointResult records a push from an external endpoint and
// fires alerting based on the result.
//
// The bearer token is checked in constant time to prevent timing oracles
// from leaking valid tokens.
func CreateExternalEndpointResult(cfg *config.Config, store storage.Store, wd *watchdog.Watchdog, m *metrics.Endpoints, logger plumelog.StructuredLogger) http.Handler {
	extraLabels := cfg.GetUniqueExtraMetricLabels()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		successParam := q.Get("success")
		if successParam != "true" && successParam != "false" {
			WriteErrorString(w, r, http.StatusBadRequest, "missing or invalid success query parameter")
			return
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			WriteErrorString(w, r, http.StatusUnauthorized, "invalid Authorization header")
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		if token == "" {
			WriteErrorString(w, r, http.StatusUnauthorized, "bearer token must not be empty")
			return
		}
		key := PathParam(r, "key")
		ee := cfg.GetExternalEndpointByKey(key)
		if ee == nil {
			logger.Error("CreateExternalEndpointResult: external endpoint not found", plumelog.Fields{"op": "handler.CreateExternalEndpointResult", "key": key})
			WriteErrorString(w, r, http.StatusNotFound, "not found")
			return
		}
		if subtle.ConstantTimeCompare([]byte(ee.Token), []byte(token)) != 1 {
			logger.Error("CreateExternalEndpointResult: invalid token", plumelog.Fields{"op": "handler.CreateExternalEndpointResult", "key": key})
			WriteErrorString(w, r, http.StatusUnauthorized, "invalid token")
			return
		}
		result := &endpoint.Result{
			Timestamp: time.Now(),
			Success:   successParam == "true",
			Errors:    []string{},
		}
		if d := q.Get("duration"); d != "" {
			parsed, err := time.ParseDuration(d)
			if err != nil {
				logger.Error("CreateExternalEndpointResult: invalid duration", plumelog.Fields{"op": "handler.CreateExternalEndpointResult", "duration": d, "err": err.Error()})
				WriteErrorString(w, r, http.StatusBadRequest, "invalid duration: "+err.Error())
				return
			}
			result.Duration = parsed
		}
		if errParam := q.Get("error"); !result.Success && errParam != "" {
			result.AddError(errParam)
		}
		converted := ee.ToEndpoint()
		if err := store.InsertEndpointResult(converted, result); err != nil {
			if errors.Is(err, common.ErrEndpointNotFound) {
				WriteErrorString(w, r, http.StatusNotFound, err.Error())
				return
			}
			logger.Error("CreateExternalEndpointResult: failed to insert result", plumelog.Fields{"op": "handler.CreateExternalEndpointResult", "key": key, "err": err.Error()})
			WriteErrorString(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		logger.Info("CreateExternalEndpointResult: inserted external result", plumelog.Fields{"op": "handler.CreateExternalEndpointResult", "key": key, "success": successParam})
		inWindow := false
		for _, mw := range ee.MaintenanceWindows {
			if mw.IsUnderMaintenance() {
				inWindow = true
				break
			}
		}
		if !cfg.Maintenance.IsUnderMaintenance() && !inWindow {
			wd.HandleAlertingExternal(ee, converted, result)
		}
		if cfg.Metrics && m != nil {
			m.PublishResult(converted, result, extraLabels)
		}
		w.WriteHeader(http.StatusOK)
	})
}
