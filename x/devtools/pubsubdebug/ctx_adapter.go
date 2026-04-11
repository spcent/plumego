package pubsubdebug

import (
	"net/http"

	"github.com/spcent/plumego/contract"
)

func adaptCtx(handler func(*contract.Ctx)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := contract.RequestContextFromContext(r.Context())
		ctx := contract.NewCtx(w, r, rc.Params)
		handler(ctx)
	})
}
