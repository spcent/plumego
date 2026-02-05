// Package semver provides semantic versioning support without external dependencies.
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Metadata   string
	original   string
}

// Parse parses a semantic version string.
func Parse(v string) (*Version, error) {
	if v == "" {
		return nil, fmt.Errorf("version string is empty")
	}

	original := v

	// Remove leading 'v' if present
	if v[0] == 'v' || v[0] == 'V' {
		v = v[1:]
	}

	ver := &Version{original: original}

	// Split metadata (+)
	parts := strings.SplitN(v, "+", 2)
	if len(parts) == 2 {
		ver.Metadata = parts[1]
		v = parts[0]
	}

	// Split prerelease (-)
	parts = strings.SplitN(v, "-", 2)
	if len(parts) == 2 {
		ver.Prerelease = parts[1]
		v = parts[0]
	}

	// Parse major.minor.patch
	parts = strings.Split(v, ".")
	if len(parts) < 1 || len(parts) > 3 {
		return nil, fmt.Errorf("invalid version format: %s", original)
	}

	// Parse major
	major, err := strconv.Atoi(parts[0])
	if err != nil || major < 0 {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}
	ver.Major = major

	// Parse minor (default to 0 if not present)
	if len(parts) > 1 {
		minor, err := strconv.Atoi(parts[1])
		if err != nil || minor < 0 {
			return nil, fmt.Errorf("invalid minor version: %s", parts[1])
		}
		ver.Minor = minor
	}

	// Parse patch (default to 0 if not present)
	if len(parts) > 2 {
		patch, err := strconv.Atoi(parts[2])
		if err != nil || patch < 0 {
			return nil, fmt.Errorf("invalid patch version: %s", parts[2])
		}
		ver.Patch = patch
	}

	return ver, nil
}

// MustParse parses a version string and panics on error.
func MustParse(v string) *Version {
	ver, err := Parse(v)
	if err != nil {
		panic(err)
	}
	return ver
}

// String returns the string representation of the version.
func (v *Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		s += "-" + v.Prerelease
	}
	if v.Metadata != "" {
		s += "+" + v.Metadata
	}
	return s
}

// Original returns the original version string.
func (v *Version) Original() string {
	return v.original
}

// Compare compares two versions.
// Returns:
//   -1 if v < other
//    0 if v == other
//    1 if v > other
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		if v.Major > other.Major {
			return 1
		}
		return -1
	}

	if v.Minor != other.Minor {
		if v.Minor > other.Minor {
			return 1
		}
		return -1
	}

	if v.Patch != other.Patch {
		if v.Patch > other.Patch {
			return 1
		}
		return -1
	}

	// Compare prerelease
	// No prerelease > prerelease
	if v.Prerelease == "" && other.Prerelease != "" {
		return 1
	}
	if v.Prerelease != "" && other.Prerelease == "" {
		return -1
	}

	// Both have prerelease, compare lexicographically
	if v.Prerelease != other.Prerelease {
		if v.Prerelease > other.Prerelease {
			return 1
		}
		return -1
	}

	// Metadata is ignored in comparison per semver spec
	return 0
}

// Equal returns true if versions are equal.
func (v *Version) Equal(other *Version) bool {
	return v.Compare(other) == 0
}

// GreaterThan returns true if v > other.
func (v *Version) GreaterThan(other *Version) bool {
	return v.Compare(other) > 0
}

// LessThan returns true if v < other.
func (v *Version) LessThan(other *Version) bool {
	return v.Compare(other) < 0
}

// GreaterThanOrEqual returns true if v >= other.
func (v *Version) GreaterThanOrEqual(other *Version) bool {
	return v.Compare(other) >= 0
}

// LessThanOrEqual returns true if v <= other.
func (v *Version) LessThanOrEqual(other *Version) bool {
	return v.Compare(other) <= 0
}

// IsPrerelease returns true if the version is a prerelease.
func (v *Version) IsPrerelease() bool {
	return v.Prerelease != ""
}

// Collection is a type that can be sorted.
type Collection []*Version

func (c Collection) Len() int {
	return len(c)
}

func (c Collection) Less(i, j int) bool {
	return c[i].LessThan(c[j])
}

func (c Collection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Sort sorts a collection of versions.
func Sort(versions []*Version) {
	c := Collection(versions)
	// Simple bubble sort (good enough for small collections)
	for i := 0; i < c.Len(); i++ {
		for j := i + 1; j < c.Len(); j++ {
			if c.Less(j, i) {
				c.Swap(i, j)
			}
		}
	}
}

// NewVersion creates a new version.
func NewVersion(major, minor, patch int) *Version {
	return &Version{
		Major:    major,
		Minor:    minor,
		Patch:    patch,
		original: fmt.Sprintf("%d.%d.%d", major, minor, patch),
	}
}

// Constraint represents a version constraint.
type Constraint struct {
	operator string
	version  *Version
}

// ParseConstraint parses a version constraint.
// Supported operators: =, !=, >, <, >=, <=, ^, ~
func ParseConstraint(c string) (*Constraint, error) {
	c = strings.TrimSpace(c)
	if c == "" {
		return nil, fmt.Errorf("empty constraint")
	}

	var op string
	var vstr string

	// Try two-character operators first
	if len(c) >= 2 {
		twoChar := c[:2]
		if twoChar == ">=" || twoChar == "<=" || twoChar == "!=" {
			op = twoChar
			vstr = strings.TrimSpace(c[2:])
		}
	}

	// Try single-character operators
	if op == "" && len(c) >= 1 {
		oneChar := c[0]
		if oneChar == '=' || oneChar == '>' || oneChar == '<' || oneChar == '^' || oneChar == '~' {
			op = string(oneChar)
			vstr = strings.TrimSpace(c[1:])
		}
	}

	// Default to exact match
	if op == "" {
		op = "="
		vstr = c
	}

	ver, err := Parse(vstr)
	if err != nil {
		return nil, fmt.Errorf("invalid version in constraint: %w", err)
	}

	return &Constraint{
		operator: op,
		version:  ver,
	}, nil
}

// Check checks if a version satisfies the constraint.
func (c *Constraint) Check(v *Version) bool {
	switch c.operator {
	case "=", "==":
		return v.Equal(c.version)
	case "!=":
		return !v.Equal(c.version)
	case ">":
		return v.GreaterThan(c.version)
	case "<":
		return v.LessThan(c.version)
	case ">=":
		return v.GreaterThanOrEqual(c.version)
	case "<=":
		return v.LessThanOrEqual(c.version)
	case "^":
		// Caret: allows changes that do not modify left-most non-zero digit
		// ^1.2.3 := >=1.2.3 <2.0.0
		// ^0.2.3 := >=0.2.3 <0.3.0
		// ^0.0.3 := >=0.0.3 <0.0.4
		if !v.GreaterThanOrEqual(c.version) {
			return false
		}
		if c.version.Major > 0 {
			return v.Major == c.version.Major
		}
		if c.version.Minor > 0 {
			return v.Major == c.version.Major && v.Minor == c.version.Minor
		}
		return v.Major == c.version.Major && v.Minor == c.version.Minor && v.Patch == c.version.Patch
	case "~":
		// Tilde: allows patch-level changes
		// ~1.2.3 := >=1.2.3 <1.3.0
		// ~1.2 := >=1.2.0 <1.3.0
		if !v.GreaterThanOrEqual(c.version) {
			return false
		}
		return v.Major == c.version.Major && v.Minor == c.version.Minor
	default:
		return false
	}
}

// String returns the string representation of the constraint.
func (c *Constraint) String() string {
	return c.operator + c.version.String()
}
