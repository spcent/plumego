package commands

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestCLI_CommandHelpReturnsUsage(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"--format", "text", "new", "--help"}, "")
	if err != nil {
		t.Fatalf("command help failed: %v\noutput: %s", err, stdout)
	}
	if !strings.Contains(stdout, "plumego [global-flags] new") {
		t.Fatalf("expected command usage, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Create new project from template") {
		t.Fatalf("expected command summary, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--template <name>") || !strings.Contains(stdout, "--module <path>") {
		t.Fatalf("expected new command flags, got: %s", stdout)
	}
}

func TestCLI_CommandHelpReflectsCurrentContracts(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want []string
	}{
		{name: "config", cmd: "config", want: []string{"--dir <path>", "--resolve", "--show-secrets"}},
		{name: "generate", cmd: "generate", want: []string{"--dir <path>", "middleware RateLimit", "handler User", "model Invoice"}},
		{name: "serve", cmd: "serve", want: []string{"[directory]", "-a, --addr <addr>"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, err := runCLI(t, []string{"--format", "text", tt.cmd, "--help"}, "")
			if err != nil {
				t.Fatalf("%s help failed: %v\noutput: %s", tt.cmd, err, stdout)
			}
			for _, want := range tt.want {
				if !strings.Contains(stdout, want) {
					t.Fatalf("expected %q in %s help, got: %s", want, tt.cmd, stdout)
				}
			}
		})
	}
}

func TestCommandHelpListsDeclaredFlags(t *testing.T) {
	tests := []struct {
		name       string
		cmd        Command
		sourceFile string
		ignore     map[string]bool
	}{
		{name: "new", cmd: &NewCmd{}, sourceFile: "new.go"},
		{name: "generate", cmd: &GenerateCmd{}, sourceFile: "generate.go"},
		{name: "add", cmd: &AddCmd{}, sourceFile: "add.go"},
		{name: "dev", cmd: NewDevCmd(), sourceFile: "dev.go"},
		{name: "routes", cmd: &RoutesCmd{}, sourceFile: "routes.go"},
		{name: "check", cmd: &CheckCmd{}, sourceFile: "check.go"},
		{name: "config", cmd: &ConfigCmd{}, sourceFile: "config.go", ignore: map[string]bool{"redact": true}},
		{name: "migrate", cmd: &MigrateCmd{}, sourceFile: "migrate.go"},
		{name: "test", cmd: &TestCmd{}, sourceFile: "test.go"},
		{name: "build", cmd: &BuildCmd{}, sourceFile: "build.go"},
		{name: "inspect", cmd: &InspectCmd{}, sourceFile: "inspect.go"},
		{name: "serve", cmd: &ServeCmd{}, sourceFile: "serve.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := declaredFlagNames(t, tt.sourceFile)
			help := commandHelp(tt.cmd)
			for _, flagName := range flags {
				if tt.ignore[flagName] {
					continue
				}
				want := "--" + flagName
				if len(flagName) == 1 {
					want = "-" + flagName
				}
				if !strings.Contains(help, want) {
					t.Fatalf("help for %s missing declared flag %q:\n%s", tt.name, want, help)
				}
			}
		})
	}
}

func declaredFlagNames(t *testing.T, sourceFile string) []string {
	t.Helper()

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filepath.Join(".", sourceFile), nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", sourceFile, err)
	}

	seen := make(map[string]struct{})
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		argIndex, ok := flagNameArgIndex(sel.Sel.Name)
		if !ok || len(call.Args) <= argIndex {
			return true
		}
		lit, ok := call.Args[argIndex].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		name := strings.Trim(lit.Value, `"`)
		if name != "" {
			seen[name] = struct{}{}
		}
		return true
	})

	flags := make([]string, 0, len(seen))
	for name := range seen {
		flags = append(flags, name)
	}
	sort.Strings(flags)
	return flags
}

func flagNameArgIndex(method string) (int, bool) {
	switch method {
	case "String", "Bool", "Int":
		return 0, true
	case "StringVar", "BoolVar", "IntVar":
		return 1, true
	default:
		return 0, false
	}
}

func TestCLI_HelpListsStableCommandSurface(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"--help"}, "")
	if err != nil {
		t.Fatalf("top-level help failed: %v\noutput: %s", err, stdout)
	}

	for _, command := range []string{
		"new", "generate", "add", "dev", "routes", "check", "config",
		"migrate", "test", "build", "inspect", "serve", "version",
	} {
		if !strings.Contains(stdout, command) {
			t.Fatalf("expected command %q in help, got: %s", command, stdout)
		}
	}
}

func TestCLI_MigrateHelpIncludesSubcommandsAndFlags(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"--format", "text", "migrate", "--help"}, "")
	if err != nil {
		t.Fatalf("migrate help failed: %v\noutput: %s", err, stdout)
	}
	for _, want := range []string{"Subcommands:", "create <name>", "status", "--driver <name>", "--db-url <url>"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in migrate help, got: %s", want, stdout)
		}
	}
}

func TestCLI_OutputFormatSmokeForVersionAndHelp(t *testing.T) {
	jsonOut, _, err := runCLI(t, []string{"--format", "json", "version"}, "")
	if err != nil {
		t.Fatalf("json version failed: %v", err)
	}
	if !strings.Contains(jsonOut, `"status": "success"`) {
		t.Fatalf("expected json success envelope, got: %s", jsonOut)
	}

	yamlOut, _, err := runCLI(t, []string{"--format", "yaml", "version"}, "")
	if err != nil {
		t.Fatalf("yaml version failed: %v", err)
	}
	if !strings.Contains(yamlOut, "status: success") || !strings.Contains(yamlOut, "message: Plumego CLI") {
		t.Fatalf("expected yaml success envelope, got: %s", yamlOut)
	}

	textOut, _, err := runCLI(t, []string{"--format", "text", "help", "version"}, "")
	if err != nil {
		t.Fatalf("text help failed: %v", err)
	}
	if !strings.Contains(textOut, "Global Flags:") {
		t.Fatalf("expected text command help with Global Flags section, got: %s", textOut)
	}

	jsonHelp, _, err := runCLI(t, []string{"--format", "json", "help", "check"}, "")
	if err != nil {
		t.Fatalf("json help failed: %v", err)
	}
	var helpPayload struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Kind    string `json:"kind"`
			Command string `json:"command"`
			Help    string `json:"help"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(jsonHelp), &helpPayload); err != nil {
		t.Fatalf("failed to parse json help: %v\noutput: %s", err, jsonHelp)
	}
	if helpPayload.Status != "success" || helpPayload.Message != "Command help" || helpPayload.Data.Command != "check" {
		t.Fatalf("unexpected help payload: %#v", helpPayload)
	}
	for _, want := range []string{"--dir <path>", "--updates"} {
		if !strings.Contains(helpPayload.Data.Help, want) {
			t.Fatalf("expected check %s in help, got: %s", want, helpPayload.Data.Help)
		}
	}

	yamlHelp, _, err := runCLI(t, []string{"--format", "yaml", "help", "version"}, "")
	if err != nil {
		t.Fatalf("yaml help failed: %v", err)
	}
	if !strings.Contains(yamlHelp, "status: success") ||
		!strings.Contains(yamlHelp, "message: Command help") ||
		!strings.Contains(yamlHelp, "command: version") {
		t.Fatalf("expected yaml help envelope, got: %s", yamlHelp)
	}
}

func TestCLI_ServeHelpReturnsUsage(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"--format", "text", "serve", "--help"}, "")
	if err != nil {
		t.Fatalf("serve help failed: %v\noutput: %s", err, stdout)
	}
	if !strings.Contains(stdout, "plumego [global-flags] serve") {
		t.Fatalf("expected serve usage, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Start static file server") {
		t.Fatalf("expected serve summary, got: %s", stdout)
	}
}

func TestCLI_ServeHelpUsesMachineEnvelope(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"--format", "json", "serve", "--addr", ":0", "--help"}, "")
	if err != nil {
		t.Fatalf("serve help failed: %v\noutput: %s", err, stdout)
	}

	var payload struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Command string `json:"command"`
			Help    string `json:"help"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse serve help output: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "success" || payload.Message != "Command help" || payload.Data.Command != "serve" {
		t.Fatalf("unexpected serve help payload: %#v", payload)
	}
}
