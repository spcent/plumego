package config

const redacted = "[REDACTED]"

// Masked returns a copy of cfg with sensitive fields replaced by "[REDACTED]".
// Safe to use in logs and error messages. Only non-empty sensitive values are masked;
// empty values remain empty so absence is observable.
func (c Config) Masked() Config {
	out := c
	if out.Qiniu.SecretKey != "" {
		out.Qiniu.SecretKey = redacted
	}
	if out.AI.APIKey != "" {
		out.AI.APIKey = redacted
	}
	if out.Auth.BootstrapAdminPassword != "" {
		out.Auth.BootstrapAdminPassword = redacted
	}
	return out
}

// HasSecrets reports whether any sensitive field is non-empty.
func (c Config) HasSecrets() bool {
	return c.Qiniu.SecretKey != "" || c.AI.APIKey != "" || c.Auth.BootstrapAdminPassword != ""
}
