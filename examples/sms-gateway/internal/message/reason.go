package message

type ReasonCode string

const (
	ReasonValidationError   ReasonCode = "validation_error"
	ReasonAuthFailed        ReasonCode = "auth_failed"
	ReasonQuotaExceeded     ReasonCode = "quota_exceeded"
	ReasonRateLimited       ReasonCode = "rate_limited"
	ReasonTemplateInvalid   ReasonCode = "template_invalid"
	ReasonSignatureInvalid  ReasonCode = "signature_invalid"
	ReasonBlacklisted       ReasonCode = "recipient_blacklisted"
	ReasonProviderTransient ReasonCode = "provider_transient"
	ReasonProviderPermanent ReasonCode = "provider_permanent"
	ReasonDLQ               ReasonCode = "dead_letter"
	ReasonNetworkError      ReasonCode = "network_error"
	ReasonTimeout           ReasonCode = "timeout"
	ReasonUnknown           ReasonCode = "unknown"
)

type Reason struct {
	Code      ReasonCode
	Detail    string
	Retryable bool
}

type ProviderErrorMapper interface {
	Normalize(provider string, code string, detail string) Reason
}
