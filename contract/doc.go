// Package contract provides Plumego's canonical HTTP response, error, and
// request metadata contracts.
//
// The package owns transport primitives only: structured API errors, success
// response envelopes, request ID context accessors, route metadata carriers,
// and lightweight trace metadata. It deliberately avoids route matching,
// application bootstrap, authentication policy, persistence, and business
// response shapes.
//
// # Response envelopes
//
// WriteResponse encodes a success envelope:
//
//	{"data": <payload>, "meta": <meta>, "request_id": "<id>"}
//
// WriteError encodes an error envelope:
//
//	{"error": {"code":"RESOURCE_NOT_FOUND", "message":"...", "category":"client_error"}, "request_id":"<id>"}
//
// # Canonical handler pattern
//
//	func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
//	    id := contract.RequestParamFromContext(r.Context(), "id")
//
//	    item, err := h.repo.Find(r.Context(), id)
//	    if apiErr, ok := contract.AsAPIError(err); ok {
//	        _ = contract.WriteError(w, r, apiErr)
//	        return
//	    }
//	    if err != nil {
//	        _ = contract.WriteError(w, r, contract.NewErrorBuilder().
//	            Type(contract.TypeInternal).Wrap(err).Build())
//	        return
//	    }
//	    _ = contract.WriteResponse(w, r, http.StatusOK, item, nil)
//	}
package contract
