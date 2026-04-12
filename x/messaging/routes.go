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

	if err := r.AddRoute(http.MethodPost, prefix+"/send", adaptCtx(svc.HandleSend)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodPost, prefix+"/batch", adaptCtx(svc.HandleBatchSend)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodGet, prefix+"/stats", adaptCtx(svc.HandleStats)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodGet, prefix+"/receipts", adaptCtx(svc.HandleListReceipts)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodGet, prefix+"/:id/receipt", adaptCtx(svc.HandleGetReceipt)); err != nil {
		return err
	}
	return r.AddRoute(http.MethodGet, prefix+"/channels", adaptCtx(svc.HandleChannelHealth))
}

func adaptCtx(handler func(*contract.Ctx)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := contract.RequestContextFromContext(r.Context())
		ctx := contract.NewCtx(w, r, rc.Params)
		handler(ctx)
	})
}
