// Package shell generates the bash/zsh integration script that users
// eval from their shell rc file.
package shell

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/AbsolutOD/zipline/internal/name"
)

// Data parameterizes the generated script.
type Data struct {
	Shell     string   // "bash" or "zsh"
	Cmd       string   // wrapper function name, e.g. "zl"
	Aliases   []string // stored alias names
	FuncsOnly bool     // emit only per-alias functions (refresh mode)
}

// Each alias becomes a thin dispatcher through `zipline run`, so the
// database stays the single source of truth and usage is tracked. The
// wrapper refreshes those functions after commands that change the set
// of aliases; on remove it also unsets the stale function by name.
const scriptTmpl = `# zipline shell integration ({{ .Shell }})
{{- range .Aliases }}
{{ . }}() { \command zipline run {{ . }} -- "$@"; }
{{- end }}
{{- if not .FuncsOnly }}

{{ .Cmd }}() {
  \command zipline "$@"
  local __zipline_ret=$?
  if [ $__zipline_ret -eq 0 ]; then
    case "$1" in
      add|edit)
        eval "$(\command zipline init {{ .Shell }} --funcs-only)"
        ;;
      remove|rm)
        eval "$(\command zipline init {{ .Shell }} --funcs-only)"
        unset -f "$2" 2>/dev/null
        ;;
    esac
  fi
  return $__zipline_ret
}
{{- end }}
`

var tmpl = template.Must(template.New("script").Parse(scriptTmpl))

// Script renders the integration script, validating every name that
// becomes a shell function so the output is always safe to eval: each
// name is run through name.Validate(x, nil), which — regardless of the
// nil reserved list — always rejects shell reserved words and critical
// builtins that would otherwise break the generated script (a reserved
// word makes the whole eval fail atomically; a shadowed builtin like
// `command` or `eval` causes infinite recursion).
func Script(d Data) (string, error) {
	if d.Shell != "bash" && d.Shell != "zsh" {
		return "", fmt.Errorf("unsupported shell %q (supported: bash, zsh)", d.Shell)
	}
	if err := name.Validate(d.Cmd, nil); err != nil {
		return "", fmt.Errorf("invalid wrapper command name: %w", err)
	}
	for _, a := range d.Aliases {
		if err := name.Validate(a, nil); err != nil {
			return "", err
		}
	}
	var b strings.Builder
	if err := tmpl.Execute(&b, d); err != nil {
		return "", err
	}
	return b.String(), nil
}
