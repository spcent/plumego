package resolve_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/spcent/plumego/contract"
	tenantcore "github.com/spcent/plumego/x/tenant/core"
	"github.com/spcent/plumego/x/tenant/resolve"
)

func ExampleMiddleware_principalFirst() {
	var source string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, tenantcore.TenantIDFromContext(r.Context()))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-ID", "header-tenant")
	req = contract.RequestWithPrincipal(req, &contract.Principal{TenantID: "principal-tenant"})

	rec := httptest.NewRecorder()
	mw := resolve.Middleware(resolve.Options{
		Hooks: tenantcore.Hooks{
			OnResolve: func(_ context.Context, info tenantcore.ResolveInfo) {
				source = info.Source
			},
		},
	})
	mw(handler).ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Print(rec.Body.String())
	fmt.Println(source)

	// Output:
	// 200
	// principal-tenant
	// principal
}

func ExampleMiddleware_customExtractor() {
	var source string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, tenantcore.TenantIDFromContext(r.Context()))
	})

	req := httptest.NewRequest(http.MethodGet, "/?tenant=query-tenant", nil)
	rec := httptest.NewRecorder()

	mw := resolve.Middleware(resolve.Options{
		DisablePrincipal: true,
		Extractor:        tenantcore.FromQuery("tenant"),
		Hooks: tenantcore.Hooks{
			OnResolve: func(_ context.Context, info tenantcore.ResolveInfo) {
				source = info.Source
			},
		},
	})
	mw(handler).ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Print(rec.Body.String())
	fmt.Println(source)

	// Output:
	// 200
	// query-tenant
	// extractor
}
