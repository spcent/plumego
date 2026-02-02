package frontend_test

import (
	"net/http"

	"github.com/spcent/plumego/frontend"
	"github.com/spcent/plumego/router"
)

// ExampleRegisterFromDir demonstrates basic frontend serving
func ExampleRegisterFromDir() {
	r := router.NewRouter()

	// Serve a Next.js/Vite build output directory
	err := frontend.RegisterFromDir(r, "./dist",
		frontend.WithPrefix("/"),
		frontend.WithCacheControl("public, max-age=31536000"),
		frontend.WithIndexCacheControl("no-cache"),
		frontend.WithFallback(true), // Enable SPA routing
	)
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":8080", r)
}

// ExampleWithPrecompressed demonstrates pre-compressed file serving
func ExampleWithPrecompressed() {
	r := router.NewRouter()

	// Enable serving pre-compressed .gz and .br files
	// When app.js.br exists and client supports Brotli, serve that instead
	err := frontend.RegisterFromDir(r, "./dist",
		frontend.WithPrecompressed(true), // Enable .gz/.br file serving
		frontend.WithCacheControl("public, max-age=31536000"),
	)
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":8080", r)
}

// ExampleWithNotFoundPage demonstrates custom 404 pages
func ExampleWithNotFoundPage() {
	r := router.NewRouter()

	// Serve custom 404 error page
	err := frontend.RegisterFromDir(r, "./dist",
		frontend.WithNotFoundPage("404.html"), // Custom 404 page
		frontend.WithFallback(false),          // Disable SPA fallback
	)
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":8080", r)
}

// ExampleWithMIMETypes demonstrates custom MIME type mapping
func ExampleWithMIMETypes() {
	r := router.NewRouter()

	// Set custom MIME types for specific file extensions
	err := frontend.RegisterFromDir(r, "./dist",
		frontend.WithMIMETypes(map[string]string{
			".wasm": "application/wasm",                // WebAssembly
			".json": "application/json; charset=utf-8", // JSON with charset
			".xml":  "application/xml; charset=utf-8",  // XML with charset
			".webp": "image/webp",                      // WebP images
		}),
	)
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":8080", r)
}

// ExampleRegisterFS demonstrates embedded filesystem serving
func ExampleRegisterFS() {
	r := router.NewRouter()

	// Use Go 1.16+ embed.FS for embedded assets
	//
	// In your main.go:
	//   import "embed"
	//   //go:embed dist/*
	//   var distFS embed.FS
	//
	// Then use:
	//   subFS, _ := fs.Sub(distFS, "dist")
	//   frontend.RegisterFS(r, http.FS(subFS), ...)

	// Example with http.Dir (for demonstration)
	err := frontend.RegisterFS(r, http.Dir("./dist"),
		frontend.WithPrefix("/app"),
		frontend.WithPrecompressed(true),
		frontend.WithCacheControl("public, max-age=31536000"),
	)
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":8080", r)
}

// ExampleWithPrefix demonstrates mounting at a non-root path
func ExampleWithPrefix() {
	r := router.NewRouter()

	// Mount frontend at /app instead of /
	err := frontend.RegisterFromDir(r, "./dist",
		frontend.WithPrefix("/app"), // Access via /app, /app/assets/*, etc.
	)
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":8080", r)
}

// ExampleWithHeaders demonstrates custom response headers
func ExampleWithHeaders() {
	r := router.NewRouter()

	// Add custom headers to all responses
	err := frontend.RegisterFromDir(r, "./dist",
		frontend.WithHeaders(map[string]string{
			"X-Frame-Options":        "DENY",
			"X-Content-Type-Options": "nosniff",
			"Referrer-Policy":        "strict-origin-when-cross-origin",
		}),
	)
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":8080", r)
}

// FullConfigurationExample demonstrates all options together
// This is not a runnable example (no Example prefix) since it starts a server
func FullConfigurationExample() {
	r := router.NewRouter()

	// Production-ready configuration with all features
	err := frontend.RegisterFromDir(r, "./dist",
		// Mounting
		frontend.WithPrefix("/"),
		frontend.WithIndex("index.html"),

		// Caching
		frontend.WithCacheControl("public, max-age=31536000, immutable"),
		frontend.WithIndexCacheControl("no-cache, must-revalidate"),

		// Performance
		frontend.WithPrecompressed(true), // Serve .gz/.br files

		// Routing
		frontend.WithFallback(true), // SPA mode

		// Custom pages
		frontend.WithNotFoundPage("404.html"),
		frontend.WithErrorPage("500.html"),

		// MIME types
		frontend.WithMIMETypes(map[string]string{
			".wasm": "application/wasm",
			".json": "application/json; charset=utf-8",
		}),

		// Security headers
		frontend.WithHeaders(map[string]string{
			"X-Frame-Options":        "DENY",
			"X-Content-Type-Options": "nosniff",
			"Referrer-Policy":        "strict-origin-when-cross-origin",
		}),
	)
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":8080", r)
}
