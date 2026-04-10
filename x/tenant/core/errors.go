package tenant

import "errors"

var (
	ErrTenantNotFound         = errors.New("tenant not found")
	ErrInvalidTenantID        = errors.New("invalid tenant ID")
	ErrPolicyDenied           = errors.New("policy denied")
	ErrQuotaExceeded          = errors.New("quota exceeded")
	ErrRateLimitExceeded      = errors.New("rate limit exceeded")
	ErrRoutePolicyNotFound    = errors.New("route policy not found")
	ErrUnsupportedQuotaWindow = errors.New("unsupported quota window")
)
