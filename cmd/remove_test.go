package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/AbsolutOD/zipline/internal/store"
)

func TestRemove(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	st, err := openStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Add(&store.Alias{Name: "bye", Command: "docker run x"}); err != nil {
		t.Fatal(err)
	}
	out, err := executeCommand(t, "remove", "bye")
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if !strings.Contains(out, "Removed alias \"bye\"") {
		t.Errorf("output = %q", out)
	}
	if _, err := st.Get("bye"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("alias still present after remove: %v", err)
	}
}

func TestRemoveViaRmAlias(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	st, err := openStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Add(&store.Alias{Name: "bye2", Command: "docker run x"}); err != nil {
		t.Fatal(err)
	}
	if _, err := executeCommand(t, "rm", "bye2"); err != nil {
		t.Fatalf("rm: %v", err)
	}
}

func TestRemoveMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, err := executeCommand(t, "remove", "ghost")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("remove ghost = %v, want ErrNotFound", err)
	}
}
