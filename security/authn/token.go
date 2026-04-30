package authn

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"
)

const staticTokenSubject = "static-token"
const bearerScheme = "Bearer"

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
	if a.token == "" || !constantTimeTokenEqual(extractToken(r), a.token) {
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
	if len(authz) <= len(bearerScheme) || !strings.EqualFold(authz[:len(bearerScheme)], bearerScheme) {
		return ""
	}
	if !isBearerDelimiter(authz[len(bearerScheme)]) {
		return ""
	}

	tok := strings.TrimSpace(authz[len(bearerScheme):])
	if tok == "" || strings.ContainsAny(tok, " \t\r\n") {
		return ""
	}

	return tok
}

func extractToken(r *http.Request) string {
	if r == nil {
		return ""
	}
	if token := ExtractBearerToken(r); token != "" {
		return token
	}
	return strings.TrimSpace(r.Header.Get(HeaderXToken))
}

func isBearerDelimiter(ch byte) bool {
	return ch == ' ' || ch == '\t'
}

func constantTimeTokenEqual(got, want string) bool {
	gotDigest := sha256.Sum256([]byte(got))
	wantDigest := sha256.Sum256([]byte(want))
	return subtle.ConstantTimeCompare(gotDigest[:], wantDigest[:]) == 1
}
