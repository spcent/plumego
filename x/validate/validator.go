package validate

// Validator validates a decoded request value.
type Validator interface {
	Validate(v any) error
}
