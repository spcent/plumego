package authn

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

const staticTokenSubject = "static-token"

// HeaderAuthorization is the standard HTTP Authorization header name.
const HeaderAuthorization = "Authorization"

// HeaderXToken is the X-Token header name used as an alternative bearer credential.
const HeaderXToken = "X-Token"

type staticTokenAuthenticator struct {
	token string
}

// StaticToken authenticates a fixed bearer or X-Token credential.
func StaticToken(token string) Authenticator {
	return staticTokenAuthenticator{token: strings.TrimSpace(token)}
}

func (a staticTokenAuthenticator) Authenticate(r *http.Request) (*Principal, error) {
	if a.token == "" || subtle.ConstantTimeCompare([]byte(extractToken(r)), []byte(a.token)) != 1 {
		return nil, ErrUnauthenticated
	}
	return &Principal{Subject: staticTokenSubject}, nil
}

// ExtractBearerToken returns the bearer token from the Authorization header.
// Query-string tokens are intentionally ignored for security.
func ExtractBearerToken(r *http.Request) string {
	if r == nil {
		return ""
	}

	authz := strings.TrimSpace(r.Header.Get(HeaderAuthorization))
	if len(authz) > len("Bearer ") && strings.EqualFold(authz[:len("Bearer")], "Bearer") {
		if tok := strings.TrimSpace(authz[len("Bearer"):]); tok != "" {
			return tok
		}
	}

	return ""
}

func extractToken(r *http.Request) string {
	if token := ExtractBearerToken(r); token != "" {
		return token
	}
	return strings.TrimSpace(r.Header.Get(HeaderXToken))
}
