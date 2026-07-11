package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/AbsolutOD/zipline/internal/store"
)

// fakeEditor writes a script that replaces the temp file's content.
func fakeEditor(t *testing.T, script string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fake-editor.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EDITOR", path)
}

func TestEdit(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	st, err := openStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Add(&store.Alias{Name: "runlike", Command: "docker run old"}); err != nil {
		t.Fatal(err)
	}
	fakeEditor(t, `printf 'docker run new\n' > "$1"`)
	if _, err := executeCommand(t, "edit", "runlike"); err != nil {
		t.Fatalf("edit: %v", err)
	}
	a, err := st.Get("runlike")
	if err != nil {
		t.Fatal(err)
	}
	if a.Command != "docker run new" {
		t.Errorf("Command = %q, want %q", a.Command, "docker run new")
	}
}

func TestEditEmptyAborts(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	st, err := openStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Add(&store.Alias{Name: "keep", Command: "docker run keep"}); err != nil {
		t.Fatal(err)
	}
	fakeEditor(t, `: > "$1"`)
	if _, err := executeCommand(t, "edit", "keep"); err == nil {
		t.Error("edit to empty succeeded, want error")
	}
	a, _ := st.Get("keep")
	if a.Command != "docker run keep" {
		t.Errorf("Command changed to %q, want unchanged", a.Command)
	}
}

func TestEditEditorFailureAborts(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	st, err := openStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Add(&store.Alias{Name: "keep2", Command: "docker run keep"}); err != nil {
		t.Fatal(err)
	}
	fakeEditor(t, `exit 1`)
	if _, err := executeCommand(t, "edit", "keep2"); err == nil {
		t.Error("edit with failing editor succeeded, want error")
	}
	a, _ := st.Get("keep2")
	if a.Command != "docker run keep" {
		t.Errorf("Command changed to %q, want unchanged", a.Command)
	}
}

func TestEditMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	fakeEditor(t, `:`)
	_, err := executeCommand(t, "edit", "ghost")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("edit ghost = %v, want ErrNotFound", err)
	}
}
