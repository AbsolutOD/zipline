package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AbsolutOD/zipline/internal/config"
	"github.com/AbsolutOD/zipline/internal/name"
	"github.com/AbsolutOD/zipline/internal/store"
)

// openStore opens the alias database at the configured path.
func openStore() (*store.Store, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return store.Open(cfg.DBPath)
}

// reservedNames lists names an alias may not take: the binary name,
// the configured wrapper name, and every subcommand name and alias.
func reservedNames() []string {
	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultCompletionCmd()
	names := []string{"zipline", "help", "completion"}
	if cfg, err := config.Load(); err == nil && cfg.Cmd != "" {
		names = append(names, cfg.Cmd)
	}
	for _, c := range rootCmd.Commands() {
		names = append(names, c.Name())
		names = append(names, c.Aliases...)
	}
	return names
}

// getAlias fetches an alias by exact name, decorating a not-found
// error with near-match suggestions.
func getAlias(st *store.Store, aliasName string) (*store.Alias, error) {
	a, err := st.Get(aliasName)
	if err == nil {
		return a, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}
	if all, listErr := st.List("name"); listErr == nil {
		candidates := make([]string, len(all))
		for i, al := range all {
			candidates[i] = al.Name
		}
		if sugg := name.Suggest(aliasName, candidates, 3); len(sugg) > 0 {
			return nil, fmt.Errorf("alias %q not found (did you mean: %s?): %w",
				aliasName, strings.Join(sugg, ", "), store.ErrNotFound)
		}
	}
	return nil, fmt.Errorf("alias %q not found: %w", aliasName, store.ErrNotFound)
}

// humanTime renders a timestamp as a compact relative age.
func humanTime(t *time.Time) string {
	if t == nil {
		return "never"
	}
	d := time.Since(*t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}

// truncate shortens s to max characters, ending in "..." when cut.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
