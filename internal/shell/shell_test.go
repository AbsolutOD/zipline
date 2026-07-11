package shell_test

import (
	"strings"
	"testing"

	"github.com/AbsolutOD/zipline/internal/shell"
)

func TestScriptFull(t *testing.T) {
	got, err := shell.Script(shell.Data{
		Shell:   "zsh",
		Cmd:     "zl",
		Aliases: []string{"dive", "runlike"},
	})
	if err != nil {
		t.Fatalf("Script: %v", err)
	}
	for _, want := range []string{
		`dive() { \command zipline run dive -- "$@"; }`,
		`runlike() { \command zipline run runlike -- "$@"; }`,
		"zl() {",
		`eval "$(\command zipline init zsh --funcs-only)"`,
		`unset -f "$2" 2>/dev/null`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("script missing %q:\n%s", want, got)
		}
	}
}

func TestScriptFuncsOnly(t *testing.T) {
	got, err := shell.Script(shell.Data{
		Shell:     "bash",
		Cmd:       "zl",
		Aliases:   []string{"dive"},
		FuncsOnly: true,
	})
	if err != nil {
		t.Fatalf("Script: %v", err)
	}
	if !strings.Contains(got, `dive() {`) {
		t.Errorf("funcs-only script missing alias function:\n%s", got)
	}
	if strings.Contains(got, "zl() {") {
		t.Errorf("funcs-only script must not define the wrapper:\n%s", got)
	}
}

func TestScriptRejectsBadInput(t *testing.T) {
	if _, err := shell.Script(shell.Data{Shell: "fish", Cmd: "zl"}); err == nil {
		t.Error("fish accepted, want error")
	}
	if _, err := shell.Script(shell.Data{Shell: "zsh", Cmd: "bad name"}); err == nil {
		t.Error("bad wrapper name accepted, want error")
	}
	if _, err := shell.Script(shell.Data{Shell: "zsh", Cmd: "zl", Aliases: []string{"ok", "not ok"}}); err == nil {
		t.Error("bad alias name accepted, want error")
	}
}
