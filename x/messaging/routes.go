package messaging

import (
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
)

// RegisterRoutes binds the canonical messaging HTTP routes with explicit wiring.
func RegisterRoutes(r *router.Router, svc *Service, prefix string) error {
	if r == nil || svc == nil {
		return nil
	}
	if prefix == "" {
		prefix = "/api/v1/messages"
	}
	prefix = strings.TrimRight(prefix, "/")

	if err := r.AddRoute(http.MethodPost, prefix+"/send", contract.AdaptCtxHandler(svc.HandleSend)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodPost, prefix+"/batch", contract.AdaptCtxHandler(svc.HandleBatchSend)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodGet, prefix+"/stats", contract.AdaptCtxHandler(svc.HandleStats)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodGet, prefix+"/receipts", contract.AdaptCtxHandler(svc.HandleListReceipts)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodGet, prefix+"/:id/receipt", contract.AdaptCtxHandler(svc.HandleGetReceipt)); err != nil {
		return err
	}
	return r.AddRoute(http.MethodGet, prefix+"/channels", contract.AdaptCtxHandler(svc.HandleChannelHealth))
}
