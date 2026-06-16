package handler

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"

	"dbadmin/internal/domain/connection"
)

// connectionExportFormatVersion identifies the export document shape so future
// changes can be detected and migrated on import.
const connectionExportFormatVersion = 1

// connectionExportDocument is the portable JSON document produced by Export
// and accepted by Import. Connections are always redacted: secret fields
// (Password, ESPassword, ESAPIKey, and any embedded MongoURI credentials) are
// stripped before being written, matching the same Redact() used for API
// responses. There is no "include secrets" mode — a migrated connection must
// have its credentials re-entered on the destination instance.
type connectionExportDocument struct {
	FormatVersion int                      `json:"format_version"`
	ExportedAt    time.Time                `json:"exported_at"`
	Connections   []*connection.Connection `json:"connections"`
}

// connectionImportResult reports the outcome of a single entry in an import
// batch. Entries are processed independently (partial-success semantics),
// matching the existing SQL ImportHandler.Import behavior, which executes
// each statement independently and reports a per-statement result list
// instead of failing the whole batch on the first error.
type connectionImportResult struct {
	Index   int    `json:"index"`
	Name    string `json:"name"`
	ID      string `json:"id,omitempty"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type connectionImportSummary struct {
	Imported int                      `json:"imported"`
	Failed   int                      `json:"failed"`
	Results  []connectionImportResult `json:"results"`
}

// Export returns all saved connections as a downloadable JSON document with
// secret fields redacted (passwords, Elasticsearch API keys, and embedded
// MongoDB URI credentials). This mirrors the redaction already applied to
// List/Get API responses via Connection.Redact — exporting plaintext or even
// encrypted-at-rest secrets to a portable file would defeat the at-rest
// encryption model, so redaction is unconditional and not configurable.
func (h ConnectionHandler) Export(w http.ResponseWriter, r *http.Request) {
	conns, err := h.Connections.List()
	if err != nil {
		h.Logger.Error("export connections", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to list connections").Build()))
		return
	}
	for _, c := range conns {
		c.Redact()
	}
	doc := connectionExportDocument{
		FormatVersion: connectionExportFormatVersion,
		ExportedAt:    time.Now().UTC(),
		Connections:   conns,
	}
	w.Header().Set("Content-Disposition", `attachment; filename="dbadmin-connections.json"`)
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, doc, map[string]any{"count": len(conns)}))
}

// Import accepts a connectionExportDocument and creates a new connection for
// each entry, applying the same validateConnection rules used by Create.
// Entries are processed independently: a validation or store failure on one
// entry is recorded in the result list and does not abort the remaining
// entries (partial-success semantics), consistent with ImportHandler.Import's
// per-statement result reporting for SQL imports.
//
// Imported connections will have empty secret fields (since exports are
// always redacted) and must have credentials re-entered manually afterward.
// validateConnection does not require a password at creation time, so this
// does not require any change to existing validation behavior.
func (h ConnectionHandler) Import(w http.ResponseWriter, r *http.Request) {
	var doc connectionExportDocument
	if !decodeJSONLimited(w, r, h.Logger, &doc) {
		return
	}
	if len(doc.Connections) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("connections list is empty").Build()))
		return
	}

	summary := connectionImportSummary{Results: make([]connectionImportResult, 0, len(doc.Connections))}
	for i, c := range doc.Connections {
		result := connectionImportResult{Index: i}
		if c == nil {
			result.Error = "connection entry is null"
			summary.Failed++
			summary.Results = append(summary.Results, result)
			continue
		}
		result.Name = c.Name

		// Import always creates new connections; never trust an imported ID,
		// and never honor any secret value that might still be present in a
		// hand-edited import file — exported documents are always redacted,
		// so any password/secret field on import is presumed stale or
		// untrusted and is dropped rather than persisted.
		imported := *c
		imported.ID = ""
		imported.Password = ""
		imported.ESPassword = ""
		imported.ESAPIKey = ""
		imported.MongoURI = connection.SanitizeMongoURI(imported.MongoURI)
		imported.SavePassword = false

		if err := validateConnection(&imported); err != nil {
			result.Error = err.Error()
			summary.Failed++
			summary.Results = append(summary.Results, result)
			continue
		}
		if err := h.Connections.Create(&imported); err != nil {
			if err == connection.ErrEncryptionRequired {
				result.Error = "DBADMIN_ENCRYPTION_KEY is required to save connection credentials"
			} else {
				h.Logger.Error("import connection", plumelog.Fields{"error": err.Error(), "name": c.Name})
				result.Error = "failed to save connection"
			}
			summary.Failed++
			summary.Results = append(summary.Results, result)
			continue
		}
		result.Success = true
		result.ID = imported.ID
		summary.Imported++
		summary.Results = append(summary.Results, result)
	}

	if summary.Imported == 0 {
		// Every entry failed validation/save — surface this as a request
		// error (not just a 200 with an all-failed result list), but still
		// include the per-entry results so the caller can see why.
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).
			Message("no connections were imported").
			Detail("results", summary.Results).
			Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, summary, map[string]any{
		"imported": summary.Imported,
		"failed":   summary.Failed,
	}))
}
