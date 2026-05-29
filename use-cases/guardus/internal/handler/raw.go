package handler

import (
	"fmt"
	"net/http"
	"time"

	"guardus/internal/storage"
)

// UptimeRaw returns the raw uptime ratio (0..1) for the duration as text/plain.
func UptimeRaw(store storage.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duration := PathParam(r, "duration")
		from, ok := durationFromParam(duration)
		if !ok {
			WriteErrorString(w, r, http.StatusBadRequest, "Durations supported: 30d, 7d, 24h, 1h")
			return
		}
		key, err := PathKey(r)
		if err != nil {
			WriteErrorString(w, r, http.StatusBadRequest, "invalid key encoding")
			return
		}
		uptime, err := store.GetUptimeByKey(key, from, time.Now())
		if err != nil {
			writeStorageError(w, r, err)
			return
		}
		WriteText(w, http.StatusOK, []byte(fmt.Sprintf("%f", uptime)))
	})
}

// ResponseTimeRaw returns the raw average response time (ms) for the duration.
func ResponseTimeRaw(store storage.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duration := PathParam(r, "duration")
		from, ok := durationFromParam(duration)
		if !ok {
			WriteErrorString(w, r, http.StatusBadRequest, "Durations supported: 30d, 7d, 24h, 1h")
			return
		}
		key, err := PathKey(r)
		if err != nil {
			WriteErrorString(w, r, http.StatusBadRequest, "invalid key encoding")
			return
		}
		rt, err := store.GetAverageResponseTimeByKey(key, from, time.Now())
		if err != nil {
			writeStorageError(w, r, err)
			return
		}
		WriteText(w, http.StatusOK, []byte(fmt.Sprintf("%d", rt)))
	})
}
