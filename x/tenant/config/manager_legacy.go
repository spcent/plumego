package config

import tenant "github.com/spcent/plumego/x/tenant/core"

func applyLegacyMinuteQuota(cfg *tenant.Config, requestsPerMinute, tokensPerMinute int) {
	if len(cfg.Quota.Limits) > 0 || (requestsPerMinute <= 0 && tokensPerMinute <= 0) {
		return
	}

	cfg.Quota.Limits = []tenant.QuotaLimit{{
		Window:   tenant.QuotaWindowMinute,
		Requests: int64(requestsPerMinute),
		Tokens:   int64(tokensPerMinute),
	}}
}
