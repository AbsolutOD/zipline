package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/AbsolutOD/zipline/internal/store"
)

func TestShow(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	st, err := openStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Add(&store.Alias{Name: "runlike", Command: "docker run --rm assaflavie/runlike", Description: "reverse docker run"}); err != nil {
		t.Fatal(err)
	}
	out, err := executeCommand(t, "show", "runlike")
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	for _, want := range []string{"runlike", "docker run --rm assaflavie/runlike", "reverse docker run", "never"} {
		if !strings.Contains(out, want) {
			t.Errorf("show output missing %q:\n%s", want, out)
		}
	}
}

func TestShowMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, err := executeCommand(t, "show", "ghost")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("show ghost = %v, want ErrNotFound", err)
	}
}
