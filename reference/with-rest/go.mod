module github.com/spcent/plumego/reference/with-rest

go 1.24.0

toolchain go1.24.4

require (
	github.com/spcent/plumego v0.0.0
	github.com/spcent/plumego/x/validate v0.0.0
	github.com/spcent/plumego/x/validate/playground v0.0.0
)

require (
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.27.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
)

replace github.com/spcent/plumego => ../..

replace github.com/spcent/plumego/x/validate => ../../x/validate

replace github.com/spcent/plumego/x/validate/playground => ../../x/validate/playground
