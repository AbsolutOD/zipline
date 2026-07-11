package cmd

import (
	"strings"
	"testing"

	"github.com/AbsolutOD/zipline/internal/store"
)

func seedAliases(t *testing.T) {
	t.Helper()
	st, err := openStore()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range []store.Alias{
		{Name: "dive", Command: "docker run --rm -it wagoodman/dive"},
		{Name: "runlike", Command: "docker run --rm assaflavie/runlike"},
	} {
		a := a
		if err := st.Add(&a); err != nil {
			t.Fatal(err)
		}
	}
	if err := st.Touch("runlike"); err != nil {
		t.Fatal(err)
	}
}

func TestListSortedByUsage(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	seedAliases(t)
	out, err := executeCommand(t, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out, "ALIAS") || !strings.Contains(out, "USES") {
		t.Errorf("missing header:\n%s", out)
	}
	if ri, di := strings.Index(out, "runlike"), strings.Index(out, "dive"); ri == -1 || di == -1 || ri > di {
		t.Errorf("runlike (1 use) should list before dive:\n%s", out)
	}
}

func TestListSortByName(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	seedAliases(t)
	out, err := executeCommand(t, "list", "--sort", "name")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if ri, di := strings.Index(out, "runlike"), strings.Index(out, "dive"); di > ri {
		t.Errorf("dive should list before runlike with --sort name:\n%s", out)
	}
	if _, err := executeCommand(t, "list", "--sort", "bogus"); err == nil {
		t.Error("--sort bogus succeeded, want error")
	}
}

func TestListQuiet(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	seedAliases(t)
	out, err := executeCommand(t, "list", "-q")
	if err != nil {
		t.Fatalf("list -q: %v", err)
	}
	if want := "runlike\ndive\n"; out != want {
		t.Errorf("quiet output = %q, want %q", out, want)
	}
}

func TestListEmpty(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	out, err := executeCommand(t, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out, "No aliases stored") {
		t.Errorf("empty-list output = %q", out)
	}
}
