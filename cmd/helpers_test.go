package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AbsolutOD/zipline/internal/store"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// executeCommand runs the root command with args, capturing combined
// stdout+stderr. Flag values are reset to defaults afterwards so tests
// don't leak state into each other.
func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	rootCmd.SetArgs(nil)
	resetFlags(rootCmd)
	return buf.String(), err
}

func resetFlags(c *cobra.Command) {
	c.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	for _, sub := range c.Commands() {
		resetFlags(sub)
	}
}

func TestRootHelp(t *testing.T) {
	out, err := executeCommand(t, "--help")
	if err != nil {
		t.Fatalf("--help: %v", err)
	}
	if !strings.Contains(out, "zipline") || !strings.Contains(out, "alias") {
		t.Errorf("help output missing expected text:\n%s", out)
	}
}

func TestGetAliasSuggests(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	st, err := openStore()
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	if err := st.Add(&store.Alias{Name: "runlike", Command: "docker run x"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	_, err = getAlias(st, "runlik")
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("getAlias = %v, want ErrNotFound", err)
	}
	if !strings.Contains(err.Error(), "did you mean") || !strings.Contains(err.Error(), "runlike") {
		t.Errorf("error %q missing suggestion", err)
	}
}

func TestReservedNamesIncludesSubcommandsAndWrapper(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	got := reservedNames()
	set := map[string]bool{}
	for _, n := range got {
		set[n] = true
	}
	for _, want := range []string{"zipline", "zl", "help", "completion"} {
		if !set[want] {
			t.Errorf("reservedNames() missing %q (got %v)", want, got)
		}
	}
}

func TestHumanTime(t *testing.T) {
	if got := humanTime(nil); got != "never" {
		t.Errorf("humanTime(nil) = %q, want never", got)
	}
	cases := []struct {
		ago  time.Duration
		want string
	}{
		{30 * time.Second, "just now"},
		{5 * time.Minute, "5m ago"},
		{2 * time.Hour, "2h ago"},
		{3 * 24 * time.Hour, "3d ago"},
	}
	for _, tc := range cases {
		ts := time.Now().Add(-tc.ago)
		if got := humanTime(&ts); got != tc.want {
			t.Errorf("humanTime(-%v) = %q, want %q", tc.ago, got, tc.want)
		}
	}
	old := time.Now().Add(-90 * 24 * time.Hour)
	if got := humanTime(&old); got != old.Format("2006-01-02") {
		t.Errorf("humanTime(90d) = %q, want date form", got)
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("short", 10); got != "short" {
		t.Errorf("truncate(short,10) = %q", got)
	}
	if got := truncate("abcdefghij", 8); got != "abcde..." {
		t.Errorf("truncate = %q, want abcde...", got)
	}
}
