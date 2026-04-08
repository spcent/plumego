package authn

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

const staticTokenSubject = "static-token"

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

func extractToken(r *http.Request) string {
	if r == nil {
		return ""
	}

	authz := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(authz) > len("Bearer ") && strings.EqualFold(authz[:len("Bearer")], "Bearer") {
		if tok := strings.TrimSpace(authz[len("Bearer"):]); tok != "" {
			return tok
		}
	}

	return strings.TrimSpace(r.Header.Get("X-Token"))
}
