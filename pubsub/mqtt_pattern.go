package pubsub

import (
	"strings"
)

const (
	// MQTTSeparator is the topic level separator
	MQTTSeparator = "/"

	// MQTTSingleLevelWildcard matches exactly one level
	MQTTSingleLevelWildcard = "+"

	// MQTTMultiLevelWildcard matches zero or more levels (must be last)
	MQTTMultiLevelWildcard = "#"
)

// MQTTMatcher handles MQTT-style topic pattern matching.
type MQTTMatcher struct{}

// Match checks if a topic matches an MQTT-style pattern.
//
// Pattern rules:
//   - "/" is the level separator
//   - "+" matches exactly one level (e.g., "a/+/c" matches "a/b/c" but not "a/b/d/c")
//   - "#" matches zero or more levels and must be the last character (e.g., "a/#" matches "a", "a/b", "a/b/c")
//
// Examples:
//
//	MQTTMatch("user/+/profile", "user/123/profile") // true
//	MQTTMatch("user/+/profile", "user/123/456/profile") // false
//	MQTTMatch("user/#", "user") // true
//	MQTTMatch("user/#", "user/123/profile") // true
//	MQTTMatch("+/+/+", "a/b/c") // true
//	MQTTMatch("#", "any/topic/here") // true
func (m *MQTTMatcher) Match(pattern, topic string) bool {
	return MQTTMatch(pattern, topic)
}

// MQTTMatch checks if a topic matches an MQTT-style pattern.
func MQTTMatch(pattern, topic string) bool {
	// Empty pattern only matches empty topic
	if pattern == "" {
		return topic == ""
	}

	// "#" alone matches everything
	if pattern == MQTTMultiLevelWildcard {
		return true
	}

	patternParts := strings.Split(pattern, MQTTSeparator)
	topicParts := strings.Split(topic, MQTTSeparator)

	return matchParts(patternParts, topicParts)
}

// matchParts recursively matches pattern parts against topic parts.
func matchParts(patternParts, topicParts []string) bool {
	pi, ti := 0, 0

	for pi < len(patternParts) {
		part := patternParts[pi]

		switch part {
		case MQTTMultiLevelWildcard:
			// "#" must be the last part
			if pi != len(patternParts)-1 {
				return false
			}
			// Matches all remaining (including zero) levels
			return true

		case MQTTSingleLevelWildcard:
			// "+" matches exactly one level
			if ti >= len(topicParts) {
				return false
			}
			// Match one level
			pi++
			ti++

		default:
			// Literal match
			if ti >= len(topicParts) || part != topicParts[ti] {
				return false
			}
			pi++
			ti++
		}
	}

	// All pattern parts consumed, check if topic parts also consumed
	return ti == len(topicParts)
}

// ValidateMQTTPattern validates an MQTT-style pattern.
// Returns an error if the pattern is invalid.
func ValidateMQTTPattern(pattern string) error {
	if pattern == "" {
		return ErrInvalidPattern
	}

	parts := strings.Split(pattern, MQTTSeparator)

	for i, part := range parts {
		// "#" must be the last part
		if part == MQTTMultiLevelWildcard && i != len(parts)-1 {
			return NewError(ErrCodeInvalidPattern, "validate", "multi-level wildcard '#' must be the last character")
		}

		// Check for invalid wildcard combinations
		if part != MQTTSingleLevelWildcard && part != MQTTMultiLevelWildcard {
			// Wildcards cannot be mixed with other characters in a level
			if strings.Contains(part, MQTTSingleLevelWildcard) || strings.Contains(part, MQTTMultiLevelWildcard) {
				return NewError(ErrCodeInvalidPattern, "validate", "wildcards cannot be mixed with other characters")
			}
		}
	}

	return nil
}

// ContainsMQTTWildcard checks if a pattern contains MQTT wildcards.
func ContainsMQTTWildcard(pattern string) bool {
	return strings.Contains(pattern, MQTTSingleLevelWildcard) ||
		strings.Contains(pattern, MQTTMultiLevelWildcard)
}

// mqttPatternSubscribers stores subscribers using MQTT patterns.
type mqttPatternSubscribers struct {
	patterns map[string]map[uint64]*subscriber // pattern -> id -> subscriber
}

// newMQTTPatternSubscribers creates a new MQTT pattern subscriber store.
func newMQTTPatternSubscribers() *mqttPatternSubscribers {
	return &mqttPatternSubscribers{
		patterns: make(map[string]map[uint64]*subscriber),
	}
}

// Add adds a subscriber for a pattern.
func (mps *mqttPatternSubscribers) Add(pattern string, id uint64, sub *subscriber) {
	if mps.patterns[pattern] == nil {
		mps.patterns[pattern] = make(map[uint64]*subscriber)
	}
	mps.patterns[pattern][id] = sub
}

// Remove removes a subscriber from a pattern.
func (mps *mqttPatternSubscribers) Remove(pattern string, id uint64) bool {
	if subs, ok := mps.patterns[pattern]; ok {
		delete(subs, id)
		if len(subs) == 0 {
			delete(mps.patterns, pattern)
		}
		return true
	}
	return false
}

// Match returns all subscribers whose patterns match the given topic.
func (mps *mqttPatternSubscribers) Match(topic string) []*subscriber {
	var result []*subscriber

	for pattern, subs := range mps.patterns {
		if MQTTMatch(pattern, topic) {
			for _, sub := range subs {
				result = append(result, sub)
			}
		}
	}

	return result
}

// List returns all patterns.
func (mps *mqttPatternSubscribers) List() []string {
	patterns := make([]string, 0, len(mps.patterns))
	for pattern := range mps.patterns {
		patterns = append(patterns, pattern)
	}
	return patterns
}

// Count returns the number of subscribers for a pattern.
func (mps *mqttPatternSubscribers) Count(pattern string) int {
	return len(mps.patterns[pattern])
}

// All returns all subscribers.
func (mps *mqttPatternSubscribers) All() []*subscriber {
	var result []*subscriber
	for _, subs := range mps.patterns {
		for _, sub := range subs {
			result = append(result, sub)
		}
	}
	return result
}

// Clear removes all subscribers.
func (mps *mqttPatternSubscribers) Clear() {
	mps.patterns = make(map[string]map[uint64]*subscriber)
}
