package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

const (
	DefaultJSONBodyLimitBytes = 16 << 20
	DefaultExportRows         = 10_000
	MaxExportRows             = 100_000
	MaxESExportRows           = 10_000
	DefaultQueryRows          = 1_000
	MaxPageSize               = 500
	MaxMongoQueryLimit        = 500
	MaxESSearchSize           = 500
	MaxBulkImportDocuments    = 10_000
	DefaultRedisExportKeys    = 1_000
	MaxRedisExportKeys        = 10_000
	MaxRedisContainerEntries  = 10_000
)

func decodeJSONLimited(w http.ResponseWriter, r *http.Request, logger plumelog.StructuredLogger, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, DefaultJSONBodyLimitBytes)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		if isBodyTooLarge(err) {
			logWriteErr(logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypePayloadTooLarge).
				Message("request body is too large").
				Detail("max_bytes", DefaultJSONBodyLimitBytes).
				Build()))
			return false
		}
		logWriteErr(logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("invalid request body").Build()))
		return false
	}
	return true
}

func isBodyTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr) || strings.Contains(err.Error(), "http: request body too large")
}

func parseExportLimit(raw string) int {
	return parseBoundedLimit(raw, DefaultExportRows, MaxExportRows)
}

func parseBoundedLimit(raw string, fallback, max int) int {
	limit := fallback
	if raw == "" {
		return limit
	}
	n, err := strconvAtoi(raw)
	if err != nil || n <= 0 {
		return limit
	}
	if n > max {
		return max
	}
	return n
}

func strconvAtoi(raw string) (int, error) {
	n := 0
	if raw == "" {
		return 0, errors.New("empty")
	}
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return 0, errors.New("invalid integer")
		}
		n = n*10 + int(ch-'0')
	}
	return n, nil
}
