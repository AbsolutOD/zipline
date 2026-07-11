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

// Validate returns an error if n is not a legal alias name or is in
// the reserved list (zipline subcommand names, wrapper name, ...).
func Validate(n string, reserved []string) error {
	if !validRE.MatchString(n) {
		return fmt.Errorf("invalid alias name %q: must start with a letter or underscore and contain only letters, digits, underscores, and hyphens", n)
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
