package shell_test

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AbsolutOD/zipline/internal/shell"
)

// update regenerates the golden files from the current Script() output.
// Run: go test ./internal/shell/ -run TestGolden -update
var update = flag.Bool("update", false, "update golden files")

// goldenCases covers the representative Script() shapes the spec calls
// out: a full bash script, a full zsh script, the funcs-only refresh
// variant, and a non-default --cmd wrapper name.
var goldenCases = []struct {
	name string // also the testdata file basename
	data shell.Data
}{
	{
		name: "bash_full",
		data: shell.Data{
			Shell:   "bash",
			Cmd:     "zl",
			Aliases: []string{"dive", "runlike"},
		},
	},
	{
		name: "zsh_full",
		data: shell.Data{
			Shell:   "zsh",
			Cmd:     "zl",
			Aliases: []string{"dive", "runlike"},
		},
	},
	{
		name: "funcs_only",
		data: shell.Data{
			Shell:     "bash",
			Cmd:       "zl",
			Aliases:   []string{"dive", "runlike"},
			FuncsOnly: true,
		},
	},
	{
		name: "cmd_variant",
		data: shell.Data{
			Shell:   "zsh",
			Cmd:     "zip",
			Aliases: []string{"dive"},
		},
	},
}

func goldenPath(name string) string {
	return filepath.Join("testdata", name+".golden")
}

func TestGolden(t *testing.T) {
	for _, tc := range goldenCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := shell.Script(tc.data)
			if err != nil {
				t.Fatalf("Script: %v", err)
			}
			path := goldenPath(tc.name)
			if *update {
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatalf("writing golden file: %v", err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading golden file %s: %v (run with -update to create it)", path, err)
			}
			if got != string(want) {
				t.Errorf("Script() output does not match %s\n--- got ---\n%s\n--- want ---\n%s", path, got, string(want))
			}
		})
	}
}

// TestGeneratedScriptIsSyntacticallyValid runs the shell's own syntax
// checker (`bash -n` / `zsh -n`) over a full generated script, so a
// blocklist gap or template bug that produces invalid shell surfaces
// here even if it doesn't happen to break the substring/golden
// assertions above. It skips when the corresponding shell binary isn't
// installed rather than failing the suite.
func TestGeneratedScriptIsSyntacticallyValid(t *testing.T) {
	cases := []struct {
		shell string
		bin   string
		flag  string
	}{
		{"bash", "bash", "-n"},
		{"zsh", "zsh", "-n"},
	}
	for _, tc := range cases {
		t.Run(tc.shell, func(t *testing.T) {
			binPath, err := exec.LookPath(tc.bin)
			if err != nil {
				t.Skipf("%s not found in PATH: %v", tc.bin, err)
			}
			got, err := shell.Script(shell.Data{
				Shell:   tc.shell,
				Cmd:     "zl",
				Aliases: []string{"dive", "runlike"},
			})
			if err != nil {
				t.Fatalf("Script: %v", err)
			}
			dir := t.TempDir()
			file := filepath.Join(dir, "hook."+tc.shell)
			if err := os.WriteFile(file, []byte(got), 0o644); err != nil {
				t.Fatalf("writing script to temp file: %v", err)
			}
			cmd := exec.Command(binPath, tc.flag, file)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Errorf("%s %s %s failed: %v\n%s", tc.bin, tc.flag, file, err, out)
			}
		})
	}
}
