// Package name validates alias names and suggests near matches.
package name

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Alias names must be legal shell function names.
var validRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)

// blocklist holds names that always break the generated hook, regardless
// of the caller-supplied reserved list: bash/zsh reserved words (defining
// a shell function named after one, e.g. `time() { ... }`, is a syntax
// error that makes the whole `eval` fail atomically) plus builtins the
// generated script itself relies on (`\command` only suppresses aliases,
// not functions, so shadowing these causes infinite recursion).
var blocklist = map[string]bool{
	// bash/zsh reserved words
	"if": true, "then": true, "else": true, "elif": true, "fi": true,
	"case": true, "esac": true, "for": true, "while": true, "until": true,
	"do": true, "done": true, "function": true, "select": true,
	"time": true, "coproc": true, "repeat": true, "in": true,
	"foreach": true, "end": true,
	// builtins the generated hook depends on; shadowing them with a
	// function would break `\command` dispatch or the wrapper itself
	"command": true, "eval": true, "unset": true, "local": true,
	"return": true, "builtin": true, "exec": true, "exit": true,
	"set": true, "declare": true, "typeset": true, "readonly": true,
	"export": true, "trap": true, "source": true,
}

// Validate returns an error if n is not a legal alias name, is in the
// caller-supplied reserved list (zipline subcommand names, wrapper name,
// ...), or is a shell reserved word / critical builtin that would break
// the generated hook script no matter what the caller allows.
func Validate(n string, reserved []string) error {
	if !validRE.MatchString(n) {
		return fmt.Errorf("invalid alias name %q: must start with a letter or underscore and contain only letters, digits, underscores, and hyphens", n)
	}
	if blocklist[n] {
		return fmt.Errorf("alias name %q is a shell reserved word or builtin that would break the generated hook", n)
	}
	for _, r := range reserved {
		if n == r {
			return fmt.Errorf("alias name %q is reserved", n)
		}
	}
	return nil
}

// Suggest returns up to max candidates that are close to n: levenshtein
// distance <= 2 (case-insensitive) or containing n as a substring.
// Closest first, ties alphabetical.
func Suggest(n string, candidates []string, max int) []string {
	type scored struct {
		name string
		dist int
	}
	lower := strings.ToLower(n)
	var matches []scored
	for _, c := range candidates {
		cl := strings.ToLower(c)
		d := levenshtein(lower, cl)
		if d <= 2 || strings.Contains(cl, lower) {
			matches = append(matches, scored{c, d})
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].dist != matches[j].dist {
			return matches[i].dist < matches[j].dist
		}
		return matches[i].name < matches[j].name
	})
	if len(matches) > max {
		matches = matches[:max]
	}
	if len(matches) == 0 {
		return nil
	}
	out := make([]string, len(matches))
	for i, m := range matches {
		out[i] = m.name
	}
	return out
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur := make([]int, len(rb)+1)
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(rb)]
}
