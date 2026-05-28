package handler

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"

	"dbadmin/internal/domain/connection"
)

func logWriteErr(logger plumelog.StructuredLogger, err error) {
	if err == nil {
		return
	}
	logger.Warn("write response failed", plumelog.Fields{"error": err.Error()})
}

// guardReadonly writes a 403 READONLY_VIOLATION response and returns true when
// conn.Readonly is set. The caller must return immediately when true is returned.
func guardReadonly(conn *connection.Connection, w http.ResponseWriter, r *http.Request, log plumelog.StructuredLogger) bool {
	if !conn.Readonly {
		return false
	}
	logWriteErr(log, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeForbidden).
		Message("this connection is read-only").
		Detail("code", "READONLY_VIOLATION").
		Build()))
	return true
}
