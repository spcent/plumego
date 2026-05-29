package security

import (
	"net/http"

	"github.com/spcent/plumego/security/authn"

	"guardus/internal/config"
)

// BasicAuthenticator implements authn.Authenticator backed by guardus's
// SecurityConfig.Basic.
//
// Username comparison is constant-time even when the stored hash is bcrypt,
// so timing channels can't reveal whether the user exists.
type BasicAuthenticator struct {
	cfg *config.BasicConfig
}

// NewBasicAuthenticator returns a BasicAuthenticator for the given config.
// The config must have already been validated by SecurityConfig.ValidateAndSetDefaults.
func NewBasicAuthenticator(cfg *config.BasicConfig) *BasicAuthenticator {
	return &BasicAuthenticator{cfg: cfg}
}

// Authenticate parses the Authorization header and returns the principal
// when the credentials match the configured Basic credentials.
//
// Failures collapse to authn.ErrUnauthenticated to avoid leaking whether the
// user existed.
func (a *BasicAuthenticator) Authenticate(r *http.Request) (*authn.Principal, error) {
	if a == nil || a.cfg == nil {
		return nil, authn.ErrUnauthenticated
	}
	user, pass, ok := r.BasicAuth()
	if !ok {
		return nil, authn.ErrUnauthenticated
	}
	usernameOK := EqualConstantTime(user, a.cfg.Username)
	stored := pickStoredHash(a.cfg)
	if stored == "" {
		return nil, authn.ErrUnauthenticated
	}
	if err := VerifyPassword(stored, pass); err != nil || !usernameOK {
		return nil, authn.ErrUnauthenticated
	}
	return &authn.Principal{
		Subject: a.cfg.Username,
		Roles:   []string{"basic"},
	}, nil
}

func pickStoredHash(cfg *config.BasicConfig) string {
	if cfg.PasswordPBKDF2Hash != "" {
		return cfg.PasswordPBKDF2Hash
	}
	return cfg.PasswordBcryptHashBase64Encoded
}

var _ authn.Authenticator = (*BasicAuthenticator)(nil)
