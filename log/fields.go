package log

// cloneFields returns a shallow copy of fields.
func cloneFields(fields Fields) Fields {
	if len(fields) == 0 {
		return Fields{}
	}
	cloned := make(Fields, len(fields))
	for k, v := range fields {
		cloned[k] = v
	}
	return cloned
}

// mergeFields returns a merged copy where override fields take precedence.
func mergeFields(base, override Fields) Fields {
	if len(base) == 0 && len(override) == 0 {
		return Fields{}
	}
	merged := make(Fields, len(base)+len(override))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range override {
		merged[k] = v
	}
	return merged
}

// mergeFieldSets returns a merged copy where later field sets take precedence.
func mergeFieldSets(fieldSets ...Fields) Fields {
	total := 0
	for _, fields := range fieldSets {
		total += len(fields)
	}
	if total == 0 {
		return Fields{}
	}

	merged := make(Fields, total)
	for _, fields := range fieldSets {
		for k, v := range fields {
			merged[k] = v
		}
	}
	return merged
}
