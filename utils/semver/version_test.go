package semver

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Version
		wantErr bool
	}{
		{
			name:  "simple version",
			input: "1.2.3",
			want:  &Version{Major: 1, Minor: 2, Patch: 3, original: "1.2.3"},
		},
		{
			name:  "with v prefix",
			input: "v1.2.3",
			want:  &Version{Major: 1, Minor: 2, Patch: 3, original: "v1.2.3"},
		},
		{
			name:  "with prerelease",
			input: "1.2.3-alpha",
			want:  &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha", original: "1.2.3-alpha"},
		},
		{
			name:  "with metadata",
			input: "1.2.3+build123",
			want:  &Version{Major: 1, Minor: 2, Patch: 3, Metadata: "build123", original: "1.2.3+build123"},
		},
		{
			name:  "with prerelease and metadata",
			input: "1.2.3-beta.1+sha.abc123",
			want:  &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "beta.1", Metadata: "sha.abc123", original: "1.2.3-beta.1+sha.abc123"},
		},
		{
			name:  "major only",
			input: "1",
			want:  &Version{Major: 1, Minor: 0, Patch: 0, original: "1"},
		},
		{
			name:  "major.minor",
			input: "1.2",
			want:  &Version{Major: 1, Minor: 2, Patch: 0, original: "1.2"},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid major",
			input:   "abc.2.3",
			wantErr: true,
		},
		{
			name:    "invalid minor",
			input:   "1.abc.3",
			wantErr: true,
		},
		{
			name:    "invalid patch",
			input:   "1.2.abc",
			wantErr: true,
		},
		{
			name:    "negative version",
			input:   "-1.2.3",
			wantErr: true,
		},
		{
			name:    "too many parts",
			input:   "1.2.3.4",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got.Major != tt.want.Major || got.Minor != tt.want.Minor || got.Patch != tt.want.Patch {
					t.Errorf("Parse() = %v, want %v", got, tt.want)
				}
				if got.Prerelease != tt.want.Prerelease {
					t.Errorf("Parse() prerelease = %v, want %v", got.Prerelease, tt.want.Prerelease)
				}
				if got.Metadata != tt.want.Metadata {
					t.Errorf("Parse() metadata = %v, want %v", got.Metadata, tt.want.Metadata)
				}
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name string
		v    *Version
		want string
	}{
		{
			name: "simple version",
			v:    &Version{Major: 1, Minor: 2, Patch: 3},
			want: "1.2.3",
		},
		{
			name: "with prerelease",
			v:    &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha"},
			want: "1.2.3-alpha",
		},
		{
			name: "with metadata",
			v:    &Version{Major: 1, Minor: 2, Patch: 3, Metadata: "build123"},
			want: "1.2.3+build123",
		},
		{
			name: "with both",
			v:    &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "beta", Metadata: "build123"},
			want: "1.2.3-beta+build123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("Version.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersion_Compare(t *testing.T) {
	tests := []struct {
		name  string
		v1    string
		v2    string
		want  int
	}{
		{name: "equal", v1: "1.2.3", v2: "1.2.3", want: 0},
		{name: "major greater", v1: "2.0.0", v2: "1.9.9", want: 1},
		{name: "major less", v1: "1.0.0", v2: "2.0.0", want: -1},
		{name: "minor greater", v1: "1.3.0", v2: "1.2.9", want: 1},
		{name: "minor less", v1: "1.2.0", v2: "1.3.0", want: -1},
		{name: "patch greater", v1: "1.2.4", v2: "1.2.3", want: 1},
		{name: "patch less", v1: "1.2.3", v2: "1.2.4", want: -1},
		{name: "release > prerelease", v1: "1.2.3", v2: "1.2.3-alpha", want: 1},
		{name: "prerelease < release", v1: "1.2.3-alpha", v2: "1.2.3", want: -1},
		{name: "prerelease alpha < beta", v1: "1.2.3-alpha", v2: "1.2.3-beta", want: -1},
		{name: "prerelease beta > alpha", v1: "1.2.3-beta", v2: "1.2.3-alpha", want: 1},
		{name: "metadata ignored", v1: "1.2.3+build1", v2: "1.2.3+build2", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, _ := Parse(tt.v1)
			v2, _ := Parse(tt.v2)
			got := v1.Compare(v2)
			if got != tt.want {
				t.Errorf("Compare(%s, %s) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestVersion_Equal(t *testing.T) {
	v1, _ := Parse("1.2.3")
	v2, _ := Parse("1.2.3")
	v3, _ := Parse("1.2.4")

	if !v1.Equal(v2) {
		t.Error("Equal versions not equal")
	}
	if v1.Equal(v3) {
		t.Error("Different versions are equal")
	}
}

func TestVersion_GreaterThan(t *testing.T) {
	v1, _ := Parse("2.0.0")
	v2, _ := Parse("1.9.9")

	if !v1.GreaterThan(v2) {
		t.Error("v1 should be greater than v2")
	}
	if v2.GreaterThan(v1) {
		t.Error("v2 should not be greater than v1")
	}
}

func TestVersion_LessThan(t *testing.T) {
	v1, _ := Parse("1.9.9")
	v2, _ := Parse("2.0.0")

	if !v1.LessThan(v2) {
		t.Error("v1 should be less than v2")
	}
	if v2.LessThan(v1) {
		t.Error("v2 should not be less than v1")
	}
}

func TestVersion_IsPrerelease(t *testing.T) {
	v1, _ := Parse("1.2.3")
	v2, _ := Parse("1.2.3-alpha")

	if v1.IsPrerelease() {
		t.Error("v1 should not be prerelease")
	}
	if !v2.IsPrerelease() {
		t.Error("v2 should be prerelease")
	}
}

func TestSort(t *testing.T) {
	versions := []*Version{
		MustParse("2.0.0"),
		MustParse("1.0.0"),
		MustParse("1.5.0"),
		MustParse("1.2.3"),
		MustParse("1.2.3-alpha"),
		MustParse("1.2.3-beta"),
	}

	Sort(versions)

	expected := []string{
		"1.0.0",
		"1.2.3-alpha",
		"1.2.3-beta",
		"1.2.3",
		"1.5.0",
		"2.0.0",
	}

	for i, v := range versions {
		if v.String() != expected[i] {
			t.Errorf("Sort()[%d] = %s, want %s", i, v.String(), expected[i])
		}
	}
}

func TestNewVersion(t *testing.T) {
	v := NewVersion(1, 2, 3)
	if v.Major != 1 || v.Minor != 2 || v.Patch != 3 {
		t.Errorf("NewVersion(1, 2, 3) = %v", v)
	}
	if v.String() != "1.2.3" {
		t.Errorf("NewVersion(1, 2, 3).String() = %s, want 1.2.3", v.String())
	}
}

func TestParseConstraint(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOp  string
		wantVer string
		wantErr bool
	}{
		{name: "exact", input: "=1.2.3", wantOp: "=", wantVer: "1.2.3"},
		{name: "exact implicit", input: "1.2.3", wantOp: "=", wantVer: "1.2.3"},
		{name: "not equal", input: "!=1.2.3", wantOp: "!=", wantVer: "1.2.3"},
		{name: "greater", input: ">1.2.3", wantOp: ">", wantVer: "1.2.3"},
		{name: "less", input: "<1.2.3", wantOp: "<", wantVer: "1.2.3"},
		{name: "greater or equal", input: ">=1.2.3", wantOp: ">=", wantVer: "1.2.3"},
		{name: "less or equal", input: "<=1.2.3", wantOp: "<=", wantVer: "1.2.3"},
		{name: "caret", input: "^1.2.3", wantOp: "^", wantVer: "1.2.3"},
		{name: "tilde", input: "~1.2.3", wantOp: "~", wantVer: "1.2.3"},
		{name: "empty", input: "", wantErr: true},
		{name: "invalid version", input: ">=abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConstraint(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConstraint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got.operator != tt.wantOp {
					t.Errorf("ParseConstraint() operator = %v, want %v", got.operator, tt.wantOp)
				}
				if got.version.String() != tt.wantVer {
					t.Errorf("ParseConstraint() version = %v, want %v", got.version.String(), tt.wantVer)
				}
			}
		})
	}
}

func TestConstraint_Check(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		version    string
		want       bool
	}{
		// Exact match
		{name: "exact match", constraint: "=1.2.3", version: "1.2.3", want: true},
		{name: "exact no match", constraint: "=1.2.3", version: "1.2.4", want: false},

		// Not equal
		{name: "not equal true", constraint: "!=1.2.3", version: "1.2.4", want: true},
		{name: "not equal false", constraint: "!=1.2.3", version: "1.2.3", want: false},

		// Greater than
		{name: "greater true", constraint: ">1.2.3", version: "1.2.4", want: true},
		{name: "greater false", constraint: ">1.2.3", version: "1.2.3", want: false},

		// Less than
		{name: "less true", constraint: "<1.2.3", version: "1.2.2", want: true},
		{name: "less false", constraint: "<1.2.3", version: "1.2.3", want: false},

		// Greater or equal
		{name: "gte equal", constraint: ">=1.2.3", version: "1.2.3", want: true},
		{name: "gte greater", constraint: ">=1.2.3", version: "1.2.4", want: true},
		{name: "gte less", constraint: ">=1.2.3", version: "1.2.2", want: false},

		// Less or equal
		{name: "lte equal", constraint: "<=1.2.3", version: "1.2.3", want: true},
		{name: "lte less", constraint: "<=1.2.3", version: "1.2.2", want: true},
		{name: "lte greater", constraint: "<=1.2.3", version: "1.2.4", want: false},

		// Caret (major)
		{name: "caret major same", constraint: "^1.2.3", version: "1.9.9", want: true},
		{name: "caret major different", constraint: "^1.2.3", version: "2.0.0", want: false},
		{name: "caret major less", constraint: "^1.2.3", version: "1.2.2", want: false},

		// Caret (0.minor)
		{name: "caret 0.minor same", constraint: "^0.2.3", version: "0.2.5", want: true},
		{name: "caret 0.minor different", constraint: "^0.2.3", version: "0.3.0", want: false},

		// Caret (0.0.patch)
		{name: "caret 0.0.patch exact", constraint: "^0.0.3", version: "0.0.3", want: true},
		{name: "caret 0.0.patch different", constraint: "^0.0.3", version: "0.0.4", want: false},

		// Tilde
		{name: "tilde patch", constraint: "~1.2.3", version: "1.2.5", want: true},
		{name: "tilde minor", constraint: "~1.2.3", version: "1.3.0", want: false},
		{name: "tilde major", constraint: "~1.2.3", version: "2.0.0", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := ParseConstraint(tt.constraint)
			if err != nil {
				t.Fatalf("ParseConstraint() error = %v", err)
			}
			v, err := Parse(tt.version)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			got := c.Check(v)
			if got != tt.want {
				t.Errorf("Check(%s, %s) = %v, want %v", tt.constraint, tt.version, got, tt.want)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	// Should not panic
	v := MustParse("1.2.3")
	if v.String() != "1.2.3" {
		t.Errorf("MustParse(\"1.2.3\") = %s", v.String())
	}
}

func TestMustParse_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse should panic on invalid input")
		}
	}()
	MustParse("invalid")
}
