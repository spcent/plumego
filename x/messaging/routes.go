package messaging

import (
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
)

// RegisterRoutes binds the canonical messaging HTTP routes with explicit wiring.
func RegisterRoutes(r *router.Router, svc *Service, prefix string) {
	if r == nil || svc == nil {
		return
	}
	if prefix == "" {
		prefix = "/api/v1/messages"
	}
	prefix = strings.TrimRight(prefix, "/")

	r.Post(prefix+"/send", contract.AdaptCtxHandler(svc.HandleSend))
	r.Post(prefix+"/batch", contract.AdaptCtxHandler(svc.HandleBatchSend))
	r.Get(prefix+"/stats", contract.AdaptCtxHandler(svc.HandleStats))
	r.Get(prefix+"/receipts", contract.AdaptCtxHandler(svc.HandleListReceipts))
	r.Get(prefix+"/:id/receipt", contract.AdaptCtxHandler(svc.HandleGetReceipt))
	r.Get(prefix+"/channels", contract.AdaptCtxHandler(svc.HandleChannelHealth))
}
