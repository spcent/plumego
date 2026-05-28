// Package authn provides the session-cookie authenticator for dbadmin.
package authn

import (
	"net/http"

	"github.com/spcent/plumego/security/authn"

	"dbadmin/internal/domain/session"
)

const cookieName = "dbadmin_session"

// SessionAuthenticator implements authn.Authenticator using HttpOnly session cookies.
type SessionAuthenticator struct {
	Sessions *session.Store
}

// Authenticate reads the session cookie and resolves the principal.
func (a *SessionAuthenticator) Authenticate(r *http.Request) (*authn.Principal, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, authn.ErrUnauthenticated
	}
	token := cookie.Value
	if token == "" {
		return nil, authn.ErrUnauthenticated
	}
	sess, err := a.Sessions.Get(token)
	if err != nil {
		return nil, authn.ErrUnauthenticated
	}
	return &authn.Principal{Subject: sess.User}, nil
}

// CookieName returns the cookie name used by the session authenticator.
func CookieName() string { return cookieName }
