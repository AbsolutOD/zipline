package cmd

import (
	"strings"
	"testing"
)

func TestAdd(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	out, err := executeCommand(t, "add", "runlike", "docker run --rm assaflavie/runlike", "-d", "reverse docker run")
	if err != nil {
		t.Fatalf("add: %v (%s)", err, out)
	}
	if !strings.Contains(out, "Added alias \"runlike\"") {
		t.Errorf("output = %q", out)
	}
	st, err := openStore()
	if err != nil {
		t.Fatal(err)
	}
	a, err := st.Get("runlike")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if a.Command != "docker run --rm assaflavie/runlike" || a.Description != "reverse docker run" {
		t.Errorf("stored %+v", a)
	}
}

func TestAddRejectsInvalidName(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if _, err := executeCommand(t, "add", "9bad", "docker run x"); err == nil {
		t.Error("add 9bad succeeded, want error")
	}
	if _, err := executeCommand(t, "add", "list", "docker run x"); err == nil {
		t.Error("add list (reserved) succeeded, want error")
	}
}

func TestAddDuplicateNeedsForce(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if _, err := executeCommand(t, "add", "dup", "docker run old"); err != nil {
		t.Fatalf("first add: %v", err)
	}
	if _, err := executeCommand(t, "add", "dup", "docker run new"); err == nil {
		t.Fatal("second add succeeded, want error")
	}
	out, err := executeCommand(t, "add", "dup", "docker run new", "--force")
	if err != nil {
		t.Fatalf("forced add: %v (%s)", err, out)
	}
	st, _ := openStore()
	a, err := st.Get("dup")
	if err != nil {
		t.Fatal(err)
	}
	if a.Command != "docker run new" {
		t.Errorf("Command = %q, want %q", a.Command, "docker run new")
	}
}
