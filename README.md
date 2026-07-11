# zipline

Zipline stores long `docker run ...` commands (or any command) under short
aliases, tracks how often you use them, and hooks into bash/zsh so each
alias works like a native command — in the spirit of what
[zoxide](https://github.com/ajeetdsouza/zoxide) does for directories.

## Install

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

## Usage

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

## Files

| Path | Purpose |
|---|---|
| `~/.config/zipline/zipline.db` | alias database (SQLite) |
| `~/.config/zipline/config.toml` | optional config: `db_path`, `cmd` |

`$XDG_CONFIG_HOME` is respected when set.

## CLI reference

Generated command docs live in [docs/cli](docs/cli/). Regenerate with:

```sh
go run ./internal/tools/docgen --out ./docs/cli
```
