# zipline

## Goal

Running tools that ship as Docker images means typing long, fiddly
commands like:

```sh
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro assaflavie/runlike my-container
```

Most people bury these in one-off shell aliases and lose track of them.
Zipline gives them a home: it stores long `docker run ...` commands (or
any command) under short named aliases in a local database, tracks how
often you use each one, and hooks into bash/zsh so every alias works
like a native command — in the spirit of what
[zoxide](https://github.com/ajeetdsouza/zoxide) does for directories.

With zipline the command above becomes:

```sh
runlike my-container
```

## Usage

### Install and hook your shell

```sh
go install github.com/AbsolutOD/zipline@latest
```

Then add the hook to your shell rc file:

```sh
# ~/.zshrc
eval "$(zipline init zsh)"

# ~/.bashrc
eval "$(zipline init bash)"
```

This defines a short wrapper (`zl` by default — pick another with
`zipline init zsh --cmd NAME`, or set `cmd` in
`~/.config/zipline/config.toml`) plus one shell function per stored alias.

### Manage and run aliases

```sh
# Store a long docker command under an alias
zl add runlike "docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro assaflavie/runlike" \
  -d "reverse-engineer a docker run command"

# Use it like a native command; extra args are appended
runlike my-container

# Manage aliases
zl list             # sorted by most used
zl show runlike     # full details and usage stats
zl edit runlike     # edit the command in $EDITOR
zl remove runlike
```

Aliases added through `zl` are usable immediately in the same shell; other
open shells pick them up when they next source the hook.

### Files

| Path | Purpose |
|---|---|
| `~/.config/zipline/zipline.db` | alias database (SQLite) |
| `~/.config/zipline/config.toml` | optional config: `db_path`, `cmd` |

`$XDG_CONFIG_HOME` is respected when set.

### CLI reference

Generated command docs live in [docs/cli](docs/cli/). Regenerate with:

```sh
go run ./internal/tools/docgen --out ./docs/cli
```

### Caveats

- Stored commands are run as `sh -c '<command> "$@"'`, so a command
  ending in a shell metacharacter like `#` or `;` will swallow or detach
  any args you append when invoking the alias (`#` starts a comment,
  `;` starts a new statement). Avoid trailing `#`/`;` in stored commands
  if you plan to append arguments.
- The wrapper name is only "known" to `zl add`'s reserved-name check when
  it comes from `cmd` in `config.toml`. A wrapper name set solely via
  `zipline init --cmd NAME` (without also setting `cmd` in
  `config.toml`) isn't reserved, so an alias could collide with it. If
  you use a custom wrapper name, set `cmd` in `config.toml` to reserve
  it.

## Building locally

Requires Go 1.26+ and a C toolchain (the SQLite driver uses cgo; on
macOS that's Xcode Command Line Tools, on Debian/Ubuntu `build-essential`).

```sh
git clone https://github.com/AbsolutOD/zipline.git
cd zipline

# Build the binary
go build -o zipline .

# Run the test suite
go test ./...

# Try it out without touching your real config/database
export XDG_CONFIG_HOME=$(mktemp -d)
./zipline add hello "echo hello from"
./zipline run hello -- world   # prints: hello from world
./zipline list
```

To test the shell hook against your local build, put the binary on your
`PATH` (or symlink it there) before eval'ing `zipline init`.
