package webhook

func mapGitHubVerifyError(err error) error {
	switch ErrorCodeOf(err) {
	case CodeMissingSignature:
		return ErrGitHubMissingHeader
	case CodeInvalidSignature:
		return ErrGitHubSignature
	case CodeInvalidEncoding:
		return ErrInvalidHexEncoding
	default:
		return err
	}
}

func mapStripeVerifyError(err error) error {
	if err == nil {
		return ErrStripeSignature
	}

	switch ErrorCodeOf(err) {
	case CodeMissingSignature:
		return ErrStripeMissingHeader
	case CodeInvalidSignature:
		return ErrStripeSignature
	case CodeInvalidEncoding:
		return ErrInvalidHexEncoding
	case CodeTimestampExpired:
		return ErrStripeExpired
	case CodeInvalidTimestamp:
		return ErrStripeInvalidTimestamp
	default:
		return err
	}
}
