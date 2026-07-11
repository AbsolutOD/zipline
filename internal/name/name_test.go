package name_test

import (
	"reflect"
	"testing"

	"github.com/AbsolutOD/zipline/internal/name"
)

func TestValidate(t *testing.T) {
	reserved := []string{"add", "list", "zl"}
	valid := []string{"runlike", "run_like", "run-like", "_x", "R2", "a"}
	for _, n := range valid {
		if err := name.Validate(n, reserved); err != nil {
			t.Errorf("Validate(%q) = %v, want nil", n, err)
		}
	}
	invalid := []string{"", "9lives", "-x", "has space", "has/slash", "has$dollar", "add", "zl"}
	for _, n := range invalid {
		if err := name.Validate(n, reserved); err == nil {
			t.Errorf("Validate(%q) = nil, want error", n)
		}
	}
}

func TestSuggest(t *testing.T) {
	candidates := []string{"runlike", "dive", "ctop", "runlike2"}
	tests := []struct {
		in   string
		max  int
		want []string
	}{
		{"runlik", 3, []string{"runlike", "runlike2"}},  // distance 1 and 2
		{"RUNLIKE", 3, []string{"runlike", "runlike2"}}, // case-insensitive
		{"run", 3, []string{"runlike", "runlike2"}},     // substring match
		{"dvie", 3, []string{"dive"}},                   // transposition = distance 2
		{"zzzzz", 3, nil},                               // nothing close
		{"runlik", 1, []string{"runlike"}},              // max truncates
	}
	for _, tc := range tests {
		got := name.Suggest(tc.in, candidates, tc.max)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Suggest(%q, max=%d) = %v, want %v", tc.in, tc.max, got, tc.want)
		}
	}
}
