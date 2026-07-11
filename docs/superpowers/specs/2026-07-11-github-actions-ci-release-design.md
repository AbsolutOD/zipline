# GitHub Actions: lint/test CI and multi-platform release build

## Goal

Add two GitHub Actions workflows to zipline:

1. `linting-testing.yml` — lint (gofmt + vet + staticcheck etc. via golangci-lint)
   and run the full test suite on every push/PR to `main`.
2. `build.yaml` — use GoReleaser to build release binaries for macOS Intel,
   macOS Apple Silicon, Linux amd64, and Linux arm64 when a version tag is
   pushed, and publish them as a GitHub Release.

## Constraint driving the design

zipline depends on `github.com/mattn/go-sqlite3` (via `gorm.io/driver/sqlite`),
which requires cgo. Cgo binaries cannot be reliably cross-compiled to a
different OS from a single Linux runner (macOS targets need Apple's SDK/linker;
cross-compiling with `CGO_ENABLED=0` would silently drop sqlite3 support).
This rules out a single-runner GoReleaser job for all 4 platforms and drives
the matrix-of-native-runners approach below.

## `linting-testing.yml`

- Triggers: `push` and `pull_request` targeting `main`.
- Two independent jobs, both on `ubuntu-latest`:
  - **lint**: `actions/checkout` → `actions/setup-go` (`go-version-file: go.mod`,
    `cache: true`) → `golangci-lint-action` (default linter set, which
    includes formatting/gofmt-equivalent checks, `go vet`, staticcheck,
    unused, etc.). No repo-specific `.golangci.yml` exists yet — default
    config is used.
  - **test**: `actions/checkout` → `actions/setup-go` → `go test ./...`.
    Runs with cgo enabled (default) since `ubuntu-latest` ships a C toolchain.

## `build.yaml`

- Trigger: `push` of tags matching `v*`.
- Permissions: `contents: write` (needed to publish a GitHub Release).
- New file `.goreleaser.yaml` (none exists yet) defines 4 builds
  (darwin/amd64, darwin/arm64, linux/amd64, linux/arm64) with
  `CGO_ENABLED: 1`, plus archive naming, checksums, and changelog config.

### Matrix build job

One job, matrixed over 4 native runners so each cgo build happens on its
actual target platform:

| Runner | Target |
|---|---|
| `macos-13` | darwin/amd64 (Intel) |
| `macos-14` | darwin/arm64 (Apple Silicon) |
| `ubuntu-latest` | linux/amd64 |
| `ubuntu-24.04-arm` | linux/arm64 |

Each leg: checkout (`fetch-depth: 0`, needed for GoReleaser's changelog) →
`actions/setup-go` → `goreleaser/goreleaser-action` with
`args: release --clean --split`, which builds and packages only that
runner's platform into a partial `dist/`. Upload that partial `dist/` as a
build artifact named by platform (e.g. `dist-darwin-amd64`).

### Merge/release job

Runs after all matrix legs complete, on `ubuntu-latest`:
- Checkout, setup-go.
- Download all `dist-*` artifacts back into a combined `dist/`.
- `goreleaser/goreleaser-action` with `args: continue --merge`, which
  stitches the partial dists together, produces the combined checksums
  file, and publishes a single GitHub Release with all 4 binaries attached.
- Requires `GITHUB_TOKEN` (from `secrets.GITHUB_TOKEN`, default permissions
  are sufficient given `contents: write` above).

## Verification

`--split` / `continue --merge` is GoReleaser's documented mechanism for
splitting a cgo build across native runners, but it's a less common path
than a single-runner release. After implementation, push a real or throwaway
`vX.Y.Z-test` tag on a branch/fork to confirm the full matrix-build →
merge → release flow works end-to-end before considering this done.

## Out of scope

- No Windows target (not requested).
- No Homebrew tap / package manager publishing (not requested).
- No Docker image builds (not requested).
- No `.golangci.yml` customization — defaults are used unless the user
  wants specific linters enabled/disabled later.
