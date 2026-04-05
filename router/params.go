package router

import (
	"net/http"

	"github.com/spcent/plumego/contract"
)

// ParamFromRequest returns the named route parameter from the request context,
// along with a boolean indicating whether the parameter was present.
// For handler code using Ctx, prefer Ctx.Param instead.
func ParamFromRequest(r *http.Request, name string) (string, bool) {
	rc := contract.RequestContextFromContext(r.Context())
	if rc.Params == nil {
		return "", false
	}
	v, ok := rc.Params[name]
	return v, ok
}
