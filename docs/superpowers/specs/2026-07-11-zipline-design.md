# Zipline Design

**Date:** 2026-07-11
**Status:** Approved

## Summary

Zipline is a CLI utility that manages long `docker run ...` invocations as named aliases, in the spirit of zoxide's approach to directories. Users store a command under an alias, hook zipline into bash/zsh, and then invoke the alias like any native command. Usage is tracked per alias and surfaces in listings.

Zipline stores **any** shell command, not just docker — the tool is docker-flavored in docs and examples, but performs no docker-specific validation (podman, nerdctl, etc. all work).

## Goals

- CRUD management of command aliases: `add`, `remove`, `show`, `list`, `edit`.
- Native-feeling execution: typing `runlike <args>` runs the stored command with `<args>` appended.
- Bash and zsh shell hooks, installed via `eval "$(zipline init <shell>)"`.
- Usage tracking (use count, last used) to sort listings.
- Database and configuration under `$XDG_CONFIG_HOME/zipline/` (default `~/.config/zipline/`).

## Non-goals

- Fuzzy/partial alias matching at run time (a wrong `docker run` can be destructive; exact-name lookup only).
- Docker-specific parsing or validation of stored commands.
- Cross-shell live sync (an alias added in shell A appears in shell B after B refreshes — same limitation zoxide accepts).

## Architecture

A single Go binary built on the existing Cobra scaffold, in three layers:

- **`cmd/`** — Cobra commands: `add`, `remove`, `show`, `list`, `edit`, `run`, `init`, `completion`.
- **`internal/store/`** — GORM + SQLite repository behind a small interface: `Add`, `Get`, `List`, `Update`, `Delete`, `Touch` (increment usage).
- **`internal/shell/`** — bash/zsh hook script generation via Go `text/template`.

### Files

| Path | Purpose |
|---|---|
| `~/.config/zipline/zipline.db` | SQLite database (created on first use, dir mode 0700) |
| `~/.config/zipline/config.toml` | Optional viper config: db path override, default short command name |

`$XDG_CONFIG_HOME` is respected when set.

### Storage choice

SQLite via GORM using the official `gorm.io/driver/sqlite` driver, as bootstrapped in `go.mod`. Known trade-off, accepted deliberately: this driver wraps `mattn/go-sqlite3` and therefore requires cgo, which complicates cross-compiled release binaries. If releases become painful, swap to a pure-Go driver (`github.com/glebarez/sqlite`) without touching the store interface.

## Data model

```go
type Alias struct {
    ID          uint      `gorm:"primarykey"`
    Name        string    `gorm:"uniqueIndex;not null"`
    Command     string    `gorm:"not null"`
    Description string
    UseCount    int       `gorm:"default:0"`
    LastUsedAt  *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

Schema is created/updated with `AutoMigrate` on store open.

### Alias name validation (on `add`)

- Must match `^[A-Za-z_][A-Za-z0-9_-]*$` (must be a legal shell function name).
- Rejected if it collides with a zipline subcommand name.
- Rejected if the alias already exists, unless `--force` overwrites.

## Execution model

Generated shell functions are **thin dispatchers**, not baked-in command strings:

```sh
runlike() { zipline run runlike -- "$@"; }
```

Rationale: the database stays the single source of truth (an `edit` takes effect immediately; function bodies never go stale — only function *names* can, when aliases are added/removed), and every invocation flows through `zipline run`, which is where usage tracking happens.

`zipline run <alias> -- [args...]`:

1. Look up the alias by exact name. Missing → exit 2 with "did you mean" suggestions.
2. Increment `UseCount`, set `LastUsedAt` (best-effort; a tracking failure must not block execution).
3. Execute via `syscall.Exec` of `sh -c '<command> "$@"' <alias> [args...]`, replacing the zipline process. This keeps interactive containers (`-it`, TTY, signal handling) behaving exactly as if the user typed the command directly, and `sh -c` lets stored commands contain env vars, `~`, and quoting without zipline parsing them.

## Shell integration

`zipline init <bash|zsh> [--cmd NAME] [--funcs-only]` prints a script intended for `eval` in `.bashrc`/`.zshrc`. It defines:

1. **One dispatcher function per stored alias** (as above).
2. **A short wrapper function** (default name `zl`, configurable with `--cmd` or config.toml). `zl` forwards all arguments to `zipline`; after a mutating subcommand (`add`, `remove`, `edit`) succeeds, it re-evals `zipline init <shell> --funcs-only` in the current shell so new aliases are immediately usable.

The short name defaults to `zl` rather than `zip` because `zip` shadows the Info-ZIP archiver at `/usr/bin/zip`; users may opt in with `--cmd zip`.

## Command surface

| Command | Behavior |
|---|---|
| `zl add <alias> "<command>" [-d desc] [--force]` | Insert; validates name; `--force` overwrites existing |
| `zl remove <alias>` (alias: `rm`) | Delete |
| `zl show <alias>` | Full detail: command, description, use count, last used, created |
| `zl list` | Table sorted by use count desc; `--sort name`; `--quiet` prints names only |
| `zl edit <alias>` | Opens the command string in `$EDITOR` via temp file; saves on clean exit |
| `zl run <alias> -- [args]` | Execute (used by dispatchers; also works directly) |
| `zipline init <shell> [--cmd NAME] [--funcs-only]` | Emit shell hook for bash or zsh |

## Error handling

- Unknown alias on `run`/`show`/`edit`/`remove`: exit code 2, one-line error with near-match suggestions.
- Store/db errors: clean one-line message, exit code 1.
- `edit` with unset `$EDITOR`: falls back to `vi`; editor non-zero exit aborts without saving.
- Usage-tracking write failures during `run` are logged to stderr but never prevent execution.

## Testing

- **Store layer:** integration-style tests against a temp-file SQLite database (CRUD, Touch, uniqueness, ordering).
- **Run flow:** `syscall.Exec` isolated behind a tiny interface; tests assert the constructed `sh -c` argv without executing anything. Name validation and near-match suggestion logic as pure unit tests.
- **Shell generation:** golden-file tests for bash and zsh output, including `--cmd` and `--funcs-only` variants.
