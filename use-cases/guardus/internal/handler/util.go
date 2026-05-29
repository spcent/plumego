// Package handler implements guardus's HTTP handlers on net/http +
// plumego's contract package.
//
// SVG/text endpoints (badges, raw, charts) deliberately bypass
// contract.WriteResponse — see AGENTS.md for the documented exception.
package handler

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/spcent/plumego/router"
)

const (
	// DefaultPage is the default page when none is specified.
	DefaultPage = 1
	// DefaultPageSize is the default page size when none is specified.
	DefaultPageSize = 50
)

// ExtractPageAndPageSize parses page/pageSize query params with the upstream
// gatus clamp semantics: page=1 + pageSize > maximum is reduced to maximum.
func ExtractPageAndPageSize(r *http.Request, maximumNumberOfResults int) (page, pageSize int) {
	q := r.URL.Query()
	if pageParam := q.Get("page"); pageParam == "" {
		page = DefaultPage
	} else {
		v, err := strconv.Atoi(pageParam)
		if err != nil || v < 1 {
			page = DefaultPage
		} else {
			page = v
		}
	}
	if pageSizeParam := q.Get("pageSize"); pageSizeParam == "" {
		pageSize = DefaultPageSize
	} else {
		v, err := strconv.Atoi(pageSizeParam)
		if err != nil {
			pageSize = DefaultPageSize
		} else {
			pageSize = v
		}
	}
	if page == 1 && pageSize > maximumNumberOfResults {
		pageSize = maximumNumberOfResults
	} else if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	return
}

// PathKey returns a URL-decoded ":key" path param. Returns an error when
// the encoded form is malformed.
func PathKey(r *http.Request) (string, error) {
	raw := router.Param(r, "key")
	return url.QueryUnescape(raw)
}

// PathParam returns the named path parameter unchanged.
func PathParam(r *http.Request, name string) string {
	return router.Param(r, name)
}

// WriteText writes a plain-text response with cache-control disabled.
func WriteText(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Expires", "0")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// WriteSVG writes an SVG image response with cache-control disabled.
func WriteSVG(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Expires", "0")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// WriteRawJSON writes a pre-marshaled JSON payload, bypassing contract.WriteResponse.
// Used by handlers that mirror the upstream gatus wire format byte-for-byte.
func WriteRawJSON(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// WriteErrorString writes a plain-text error response.
//
// guardus mirrors the upstream gatus wire format, which sends plain-text
// error bodies rather than the contract JSON envelope.
func WriteErrorString(w http.ResponseWriter, _ *http.Request, status int, message string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(message))
}
