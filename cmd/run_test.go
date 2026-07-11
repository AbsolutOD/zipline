package cmd

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/AbsolutOD/zipline/internal/store"
)

type fakeExecer struct {
	argv0 string
	argv  []string
}

func (f *fakeExecer) Exec(argv0 string, argv, env []string) error {
	f.argv0, f.argv = argv0, argv
	return nil
}

func TestRunAlias(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	st, err := openStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Add(&store.Alias{Name: "runlike", Command: "docker run --rm img"}); err != nil {
		t.Fatal(err)
	}
	fe := &fakeExecer{}
	if err := runAlias(st, fe, "runlike", []string{"my-container", "-p"}, io.Discard); err != nil {
		t.Fatalf("runAlias: %v", err)
	}
	want := []string{"sh", "-c", `docker run --rm img "$@"`, "runlike", "my-container", "-p"}
	if !reflect.DeepEqual(fe.argv, want) {
		t.Errorf("argv = %#v, want %#v", fe.argv, want)
	}
	a, err := st.Get("runlike")
	if err != nil {
		t.Fatal(err)
	}
	if a.UseCount != 1 || a.LastUsedAt == nil {
		t.Errorf("usage not recorded: count=%d lastused=%v", a.UseCount, a.LastUsedAt)
	}
}

func TestRunAliasNotFound(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	st, err := openStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Add(&store.Alias{Name: "runlike", Command: "docker run img"}); err != nil {
		t.Fatal(err)
	}
	fe := &fakeExecer{}
	err = runAlias(st, fe, "runlik", nil, io.Discard)
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	if !strings.Contains(err.Error(), "did you mean") {
		t.Errorf("error %q missing suggestions", err)
	}
	if fe.argv != nil {
		t.Error("Exec was called for a missing alias")
	}
}

func TestRunCommandStripsDashDash(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	st, err := openStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Add(&store.Alias{Name: "d", Command: "docker run img"}); err != nil {
		t.Fatal(err)
	}
	fe := &fakeExecer{}
	testExecer = fe
	defer func() { testExecer = nil }()
	if _, err := executeCommand(t, "run", "d", "--", "-it", "alpine"); err != nil {
		t.Fatalf("run: %v", err)
	}
	want := []string{"sh", "-c", `docker run img "$@"`, "d", "-it", "alpine"}
	if !reflect.DeepEqual(fe.argv, want) {
		t.Errorf("argv = %#v, want %#v", fe.argv, want)
	}
}
