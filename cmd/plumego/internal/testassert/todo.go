package testassert

import (
	"strings"
	"testing"
)

// NoBareTODO fails when generated content contains a placeholder TODO comment.
func NoBareTODO(t testing.TB, label string, content string) {
	t.Helper()

	if strings.Contains(content, "// TODO") {
		t.Errorf("%s contains '// TODO':\n%s", label, content)
	}
}
