package execute_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/AbsolutOD/zipline/internal/execute"
)

type fakeExecer struct {
	argv0 string
	argv  []string
	env   []string
}

func (f *fakeExecer) Exec(argv0 string, argv, env []string) error {
	f.argv0, f.argv, f.env = argv0, argv, env
	return nil
}

func TestArgv(t *testing.T) {
	got := execute.Argv("docker run --rm img", "runlike", []string{"my-container", "-p"})
	want := []string{"sh", "-c", `docker run --rm img "$@"`, "runlike", "my-container", "-p"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Argv = %#v, want %#v", got, want)
	}
}

func TestArgvNoArgs(t *testing.T) {
	got := execute.Argv("docker run img", "d", nil)
	want := []string{"sh", "-c", `docker run img "$@"`, "d"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Argv = %#v, want %#v", got, want)
	}
}

func TestRun(t *testing.T) {
	fe := &fakeExecer{}
	if err := execute.Run(fe, "docker run img", "d", []string{"x"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.HasSuffix(fe.argv0, "/sh") {
		t.Errorf("argv0 = %q, want absolute path to sh", fe.argv0)
	}
	want := []string{"sh", "-c", `docker run img "$@"`, "d", "x"}
	if !reflect.DeepEqual(fe.argv, want) {
		t.Errorf("argv = %#v, want %#v", fe.argv, want)
	}
	if len(fe.env) == 0 {
		t.Error("env is empty, want os.Environ()")
	}
}
