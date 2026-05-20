package log

import "math/bits"

const maxMapCapacity = int(^uint(0) >> 1)

func safeMapCapacity(lengths ...int) int {
	total := 0
	for _, length := range lengths {
		next, ok := addMapCapacity(total, length)
		if !ok {
			return 0
		}
		total = next
	}
	return total
}

func addMapCapacity(base, extra int) (int, bool) {
	if base < 0 || extra < 0 {
		return 0, false
	}
	sum, carry := bits.Add(uint(base), uint(extra), 0)
	if carry != 0 || sum > uint(maxMapCapacity) {
		return 0, false
	}
	return int(sum), true
}

func fieldSetCapacity(fieldSets []Fields) int {
	total := 0
	for _, fields := range fieldSets {
		next, ok := addMapCapacity(total, len(fields))
		if !ok {
			return 0
		}
		total = next
	}
	return total
}

func fieldSetsEmpty(fieldSets []Fields) bool {
	for _, fields := range fieldSets {
		if len(fields) > 0 {
			return false
		}
	}
	return true
}

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

	merged := make(Fields, safeMapCapacity(len(base), len(override)))
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
	capacity := fieldSetCapacity(fieldSets)
	if capacity == 0 && fieldSetsEmpty(fieldSets) {
		return Fields{}
	}

	merged := make(Fields, capacity)
	for _, fields := range fieldSets {
		for k, v := range fields {
			merged[k] = v
		}
	}
	return merged
}
