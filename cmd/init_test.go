package cmd

import (
	"strings"
	"testing"

	"github.com/AbsolutOD/zipline/internal/store"
)

func TestInitZsh(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	st, err := openStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Add(&store.Alias{Name: "runlike", Command: "docker run x"}); err != nil {
		t.Fatal(err)
	}
	out, err := executeCommand(t, "init", "zsh")
	if err != nil {
		t.Fatalf("init zsh: %v", err)
	}
	for _, want := range []string{
		`runlike() { \command zipline run runlike -- "$@"; }`,
		"zl() {",
		"zipline init zsh --funcs-only",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("init output missing %q:\n%s", want, out)
		}
	}
}

func TestInitCustomCmd(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	out, err := executeCommand(t, "init", "bash", "--cmd", "zip")
	if err != nil {
		t.Fatalf("init bash --cmd zip: %v", err)
	}
	if !strings.Contains(out, "zip() {") {
		t.Errorf("init output missing custom wrapper:\n%s", out)
	}
}

func TestInitFuncsOnly(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	out, err := executeCommand(t, "init", "zsh", "--funcs-only")
	if err != nil {
		t.Fatalf("init --funcs-only: %v", err)
	}
	if strings.Contains(out, "zl() {") {
		t.Errorf("funcs-only output must not define wrapper:\n%s", out)
	}
}

func TestInitRejectsUnknownShell(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if _, err := executeCommand(t, "init", "fish"); err == nil {
		t.Error("init fish succeeded, want error")
	}
}
