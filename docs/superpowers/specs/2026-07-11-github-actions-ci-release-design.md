# GitHub Actions: lint/test CI and multi-platform release build

## Goal

Add two GitHub Actions workflows to zipline:

1. `linting-testing.yml` — lint (gofmt + vet + staticcheck etc. via golangci-lint)
   and run the full test suite on every push/PR to `main`.
2. `build.yaml` — use GoReleaser to build release binaries for macOS
   Apple Silicon, Linux amd64, and Linux arm64 when a version tag is
   pushed, and publish them as a GitHub Release. (macOS Intel was originally
   in scope too — see the Task 4 correction below for why it was dropped.)

## Constraint driving the design

zipline depends on `github.com/mattn/go-sqlite3` (via `gorm.io/driver/sqlite`),
which requires cgo. Cgo binaries cannot be reliably cross-compiled to a
different OS from a single Linux runner (macOS targets need Apple's SDK/linker;
cross-compiling with `CGO_ENABLED=0` would silently drop sqlite3 support).
This rules out a single-runner GoReleaser job for all target platforms and
drives the matrix-of-native-runners approach below.

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
- New file `.goreleaser.yaml` (none exists yet) defines builds for
  darwin/arm64, linux/amd64, linux/arm64 (darwin/amd64 explicitly
  `ignore`d — see Task 4 correction below) with `CGO_ENABLED: 1`, plus
  archive naming, checksums, and changelog config.

### Matrix build job

One job, matrixed over 3 native runners so each cgo build happens on its
actual target platform:

| Runner | Target |
|---|---|
| `macos-latest` | darwin/arm64 (Apple Silicon) |
| `ubuntu-latest` | linux/amd64 |
| `ubuntu-24.04-arm` | linux/arm64 |

Each leg: checkout → `actions/setup-go` → `goreleaser/goreleaser-action` with
`args: build --clean --single-target`, which builds only that runner's
native `GOOS`/`GOARCH` (inferred from the host — no explicit env needed)
into `dist/<id>_<goos>_<goarch>*/zipline`. The leg then packages that binary
into a `.tar.gz` archive (with `LICENSE`/`README.md`) and a `.sha256`
checksum file via plain shell (`find`/`tar`/`sha256sum`), and uploads both
as a build artifact named `release-<goos>-<goarch>`.

**Correction (post-implementation):** the original design used
`goreleaser release --clean --split` per leg plus `goreleaser continue
--merge` in a final job. That is a **GoReleaser Pro-only** feature
(confirmed against goreleaser.com/customization/partial/: "This feature is
exclusively available with GoReleaser Pro") — the free/OSS `goreleaser`
binary these workflows install cannot run those commands. The design below
replaces that mechanism with a plain `goreleaser build --single-target`
(OSS) per leg plus manual archiving/publishing, so the whole pipeline stays
free/OSS while keeping the native-runner-per-platform cgo reliability.

**Correction #2 (found during Task 4 CI verification): no free Intel-macOS
runner exists anymore.** The original design used `macos-13` for
darwin/amd64. `macos-13` was fully retired by GitHub in December 2025 —
jobs targeting it queue forever rather than failing outright, which is why
this surfaced as a silent hang during verification rather than an error.
Its replacement, `macos-15-intel`, is a "larger runner" label that requires
a paid plan with larger-runner access enabled, not the standard free
GitHub-hosted pool. Cross-compiling darwin/amd64 from the arm64 runner
(`GOARCH=amd64 CC="clang -arch x86_64"`) was considered — Xcode's clang
supports targeting x86_64 from Apple Silicon — but the decision was made to
drop Intel macOS support entirely rather than add that complexity. The
`darwin/amd64` target is removed from `.goreleaser.yaml` (via `ignore:`)
and from `build.yaml`'s matrix. Separately, `macos-14` (used for
darwin/arm64 in the original design) began its own deprecation cycle on
2026-07-06 (full retirement 2026-11-02) — the darwin/arm64 leg was moved to
`macos-latest` to avoid hitting the same problem again shortly.

### Publish job

Runs after all matrix legs complete, on `ubuntu-latest`:
- Download all `release-*` artifacts into a combined `dist/` (`merge-multiple: true`).
- Concatenate the per-platform `.sha256` files into one `dist/checksums.txt`.
- `gh release create "${{ github.ref_name }}" --repo "${{ github.repository }}"
  --title "${{ github.ref_name }}" --generate-notes dist/*.tar.gz
  dist/checksums.txt`, publishing a single GitHub Release with all 3
  archives plus the combined checksums file attached.
- Requires `GH_TOKEN` (from `secrets.GITHUB_TOKEN`) for the `gh` CLI, and
  `contents: write` (already set at the workflow level) for `gh release create`.

## Verification

After implementation, push a real or throwaway `vX.Y.Z-test` tag on a
branch/fork to confirm the full matrix-build → publish flow works
end-to-end (all 3 archives + checksums.txt attached to a real GitHub
Release) before considering this done.

**Correction #3 (found during Task 4 CI verification):** the packaging
step's `sha256sum` command doesn't exist on the macOS runner — only Linux
ships GNU coreutils' `sha256sum`; macOS has `shasum -a 256` instead. Fixed
by falling back to `shasum -a 256` when `sha256sum` isn't on `PATH`.

## Out of scope

- No Windows target (not requested).
- No Intel macOS (darwin/amd64) target — dropped after Task 4 CI
  verification showed no free GitHub-hosted Intel-macOS runner exists
  anymore (see Correction #2). Revisit if the user later gets access to a
  paid "larger runners" plan, or wants cross-compilation from the arm64
  runner via `clang -arch x86_64`.
- No Homebrew tap / package manager publishing (not requested).
- No Docker image builds (not requested).
- No `.golangci.yml` customization — defaults are used unless the user
  wants specific linters enabled/disabled later.
