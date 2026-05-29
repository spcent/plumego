package config

// SecurityConfig is the security configuration for guardus.
//
// v1 supports HTTP Basic only — OIDC is intentionally out of scope.
type SecurityConfig struct {
	Basic *BasicConfig `json:"basic,omitempty"`
}

// BasicConfig is the configuration for Basic authentication.
//
// Either PasswordBcryptHashBase64Encoded (compatible with upstream gatus) or
// PasswordPBKDF2Hash (plumego/security/password format) must be set.
type BasicConfig struct {
	Username                        string `json:"username"`
	PasswordBcryptHashBase64Encoded string `json:"password-bcrypt-base64,omitempty"`
	PasswordPBKDF2Hash              string `json:"password-pbkdf2,omitempty"`
}

// IsValid returns true when the basic config has a username and at least one hashed password set.
func (b *BasicConfig) IsValid() bool {
	if b == nil {
		return false
	}
	return len(b.Username) > 0 && (len(b.PasswordBcryptHashBase64Encoded) > 0 || len(b.PasswordPBKDF2Hash) > 0)
}

// ValidateAndSetDefaults reports whether the security configuration is valid.
func (c *SecurityConfig) ValidateAndSetDefaults() bool {
	return c.Basic == nil || c.Basic.IsValid()
}
