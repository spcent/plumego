package handler

import plumelog "github.com/spcent/plumego/log"

// logWriteErr logs err at Warn level when both logger and err are non-nil.
// Pass the error returned by contract.WriteResponse or contract.WriteError.
// Response write errors (e.g., client disconnects after headers are sent)
// are not recoverable; logging them surfaces unexpected disconnects for operators.
func logWriteErr(logger plumelog.StructuredLogger, err error) {
	if err == nil || logger == nil {
		return
	}
	logger.Warn("write response failed", plumelog.Fields{"error": err.Error()})
}
