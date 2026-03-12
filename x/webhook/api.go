package webhook

import (
	"net/http"
	"time"

	webhookin "github.com/spcent/plumego/net/webhookin"
	webhookout "github.com/spcent/plumego/net/webhookout"
)

type (
	Service        = webhookout.Service
	Store          = webhookout.Store
	MemStore       = webhookout.MemStore
	Config         = webhookout.Config
	DropPolicy     = webhookout.DropPolicy
	Target         = webhookout.Target
	TargetPatch    = webhookout.TargetPatch
	TargetFilter   = webhookout.TargetFilter
	Event          = webhookout.Event
	Delivery       = webhookout.Delivery
	DeliveryStatus = webhookout.DeliveryStatus
	DeliveryFilter = webhookout.DeliveryFilter
	DeliveryPatch  = webhookout.DeliveryPatch

	Deduper             = webhookin.Deduper
	IPAllowlist         = webhookin.IPAllowlist
	NonceStore          = webhookin.NonceStore
	MemoryNonceStore    = webhookin.MemoryNonceStore
	SignatureEncoding   = webhookin.SignatureEncoding
	HashAlgorithm       = webhookin.HashAlgorithm
	HMACReplayConfig    = webhookin.HMACReplayConfig
	HMACConfig          = webhookin.HMACConfig
	SignedPayloadFormat = webhookin.SignedPayloadFormat
	VerifyResult        = webhookin.VerifyResult
	StripeVerifyOptions = webhookin.StripeVerifyOptions
	GitHubVerifyOptions = webhookin.GitHubVerifyOptions
	ErrorCode           = webhookin.ErrorCode
	VerifyError         = webhookin.VerifyError
)

const (
	DropNewest     = webhookout.DropNewest
	BlockWithLimit = webhookout.BlockWithLimit
	FailFast       = webhookout.FailFast

	DeliveryPending = webhookout.DeliveryPending
	DeliveryRetry   = webhookout.DeliveryRetry
	DeliverySuccess = webhookout.DeliverySuccess
	DeliveryFailed  = webhookout.DeliveryFailed
	DeliveryDead    = webhookout.DeliveryDead

	EncodingHex    = webhookin.EncodingHex
	EncodingBase64 = webhookin.EncodingBase64

	HashSHA256 = webhookin.HashSHA256
	HashSHA512 = webhookin.HashSHA512

	PayloadRaw                = webhookin.PayloadRaw
	PayloadTimestampBody      = webhookin.PayloadTimestampBody
	PayloadTimestampNonceBody = webhookin.PayloadTimestampNonceBody
	PayloadNonceBody          = webhookin.PayloadNonceBody

	CodeMissingSignature = webhookin.CodeMissingSignature
	CodeInvalidSignature = webhookin.CodeInvalidSignature
	CodeInvalidEncoding  = webhookin.CodeInvalidEncoding
	CodeMissingTimestamp = webhookin.CodeMissingTimestamp
	CodeInvalidTimestamp = webhookin.CodeInvalidTimestamp
	CodeTimestampExpired = webhookin.CodeTimestampExpired
	CodeMissingNonce     = webhookin.CodeMissingNonce
	CodeReplayDetected   = webhookin.CodeReplayDetected
	CodeIPDenied         = webhookin.CodeIPDenied
	CodeBodyRead         = webhookin.CodeBodyRead
	CodeConfigError      = webhookin.CodeConfigError
	CodeInternalError    = webhookin.CodeInternalError
)

var (
	ErrNotFound   = webhookout.ErrNotFound
	ErrQueueFull  = webhookout.ErrQueueFull
	ErrInvalidHex = webhookout.ErrInvalidHex
)

func NewMemStore() *MemStore {
	return webhookout.NewMemStore()
}

func NewService(store Store, cfg Config) *Service {
	return webhookout.NewService(store, cfg)
}

func ConfigFromEnv() Config {
	return webhookout.ConfigFromEnv()
}

func NewHTTPClient(timeout time.Duration) *http.Client {
	return webhookout.NewHTTPClient(timeout)
}

func NewDeduper(ttl time.Duration) *Deduper {
	return webhookin.NewDeduper(ttl)
}

func NewIPAllowlist(entries []string) (*IPAllowlist, error) {
	return webhookin.NewIPAllowlist(entries)
}

func NewMemoryNonceStore(ttl time.Duration) *MemoryNonceStore {
	return webhookin.NewMemoryNonceStore(ttl)
}

func VerifyGitHub(r *http.Request, secret string, maxBody int64) ([]byte, error) {
	return webhookin.VerifyGitHub(r, secret, maxBody)
}

func VerifyGitHubWithOptions(r *http.Request, secret string, opts GitHubVerifyOptions) ([]byte, error) {
	return webhookin.VerifyGitHubWithOptions(r, secret, opts)
}

func VerifyStripe(r *http.Request, endpointSecret string, opt StripeVerifyOptions) ([]byte, error) {
	return webhookin.VerifyStripe(r, endpointSecret, opt)
}

func VerifyHMAC(r *http.Request, cfg HMACConfig) (VerifyResult, error) {
	return webhookin.VerifyHMAC(r, cfg)
}

func ErrorCodeOf(err error) ErrorCode {
	return webhookin.ErrorCodeOf(err)
}

func HTTPStatus(err error) int {
	return webhookin.HTTPStatus(err)
}
