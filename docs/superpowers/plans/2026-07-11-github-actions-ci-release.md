# GitHub Actions CI/Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `linting-testing.yml` (lint + test on push/PR to main) and `build.yaml` (GoReleaser multi-platform release build on version tags) to zipline.

**Architecture:** Two independent GitHub Actions workflow files plus a new `.goreleaser.yaml` config. Because zipline's sqlite3 driver requires cgo, `build.yaml` uses a matrix of native runners (one per target OS/arch) each running `goreleaser release --clean --split`, followed by a merge job running `goreleaser continue --merge` to combine partial dists into one GitHub Release.

**Tech Stack:** GitHub Actions, `actions/checkout@v4`, `actions/setup-go@v5`, `golangci/golangci-lint-action@v6`, `goreleaser/goreleaser-action@v6` (GoReleaser v2), `actions/upload-artifact@v4` / `actions/download-artifact@v4`.

## Global Constraints

- Module: `github.com/AbsolutOD/zipline`, binary name `zipline`, main package at repo root (`main.go`).
- Go version pinned via `go-version-file: go.mod` in every job (currently declares `go 1.26.5`).
- cgo is required (`github.com/mattn/go-sqlite3`); every build/test/lint job must run with `CGO_ENABLED=1` (the default) on a runner with a C toolchain — never disable cgo to simplify cross-compilation.
- `linting-testing.yml` triggers on `push` and `pull_request` targeting `main` only.
- `build.yaml` triggers on tags matching `v*`, and publishes a GitHub Release (not build-only).
- Targets: darwin/amd64 (macos-13), darwin/arm64 (macos-14), linux/amd64 (ubuntu-latest), linux/arm64 (ubuntu-24.04-arm). No Windows.
- No `.golangci.yml` customization — use golangci-lint defaults; gofmt formatting is checked as a separate explicit step, not assumed to be covered by default linters.
- Any step that pushes to the `origin` remote, creates a branch/PR, or pushes a tag requires explicit user confirmation before running — do not execute those steps unattended.

---

### Task 1: `linting-testing.yml`

**Files:**
- Create: `.github/workflows/linting-testing.yml`

**Interfaces:**
- Produces: a `lint` job and a `test` job, both runnable independently, both gating on `main`.

- [ ] **Step 1: Write the workflow file**

Create `.github/workflows/linting-testing.yml`:

```yaml
name: Lint and Test

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Check gofmt formatting
        run: |
          unformatted=$(gofmt -l .)
          if [ -n "$unformatted" ]; then
            echo "The following files are not gofmt-formatted:"
            echo "$unformatted"
            exit 1
          fi

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Run tests
        run: go test ./... -v
```

- [ ] **Step 2: Validate YAML syntax**

Run:

```bash
python3 -c "import yaml, sys; yaml.safe_load(open('.github/workflows/linting-testing.yml')); print('OK')"
```

Expected: `OK`

- [ ] **Step 3: Verify the gofmt check locally**

Run: `gofmt -l .`
Expected: no output (empty — repo is already formatted). If any files print, run `gofmt -w <file>` on them before continuing.

- [ ] **Step 4: Verify `go vet` and tests locally**

Run:

```bash
go vet ./...
go test ./... -v
```

Expected: both exit 0; `go test` shows `PASS` for all packages.

- [ ] **Step 5: Install golangci-lint locally and run it**

Run:

```bash
brew install golangci-lint
golangci-lint run
```

Expected: exits 0 with no issues reported. If it reports real issues, fix them in the flagged source files (not the workflow) before continuing — the workflow must reflect a repo that actually passes lint.

- [ ] **Step 6: Commit**

```bash
git add .github/workflows/linting-testing.yml
git commit -m "ci: add lint and test workflow"
```

---

### Task 2: `.goreleaser.yaml`

**Files:**
- Create: `.goreleaser.yaml`

**Interfaces:**
- Produces: a GoReleaser config with build id `zipline`, producing `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64` archives — consumed by `build.yaml` in Task 3.

- [ ] **Step 1: Write the config**

Create `.goreleaser.yaml`:

```yaml
version: 2

project_name: zipline

before:
  hooks:
    - go mod tidy

builds:
  - id: zipline
    binary: zipline
    main: .
    env:
      - CGO_ENABLED=1
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64

archives:
  - id: zipline
    formats: [tar.gz]
    name_template: >-
      {{ .ProjectName }}_{{ .Os }}_{{ .Arch }}
    files:
      - LICENSE
      - README.md

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"

release:
  github:
    owner: AbsolutOD
    name: zipline
```

- [ ] **Step 2: Install GoReleaser locally**

Run: `brew install goreleaser`
Expected: `goreleaser` on PATH. Verify with `goreleaser --version`.

- [ ] **Step 3: Validate the config schema**

Run: `goreleaser check`

Expected: reports the config as valid. If it reports schema errors (e.g. a renamed key like `formats` vs `format`), fix `.goreleaser.yaml` to match what this installed GoReleaser version expects, then re-run `goreleaser check` until it passes.

- [ ] **Step 4: Dry-run a local single-target build**

Run: `goreleaser build --snapshot --clean --single-target`

Expected: exits 0 and produces a `dist/zipline_<os>_<arch>/zipline` binary for the local machine's platform (darwin/arm64). Verify it runs:

```bash
./dist/zipline_darwin_arm64_*/zipline --help
```

Expected: prints zipline's CLI help output (cobra root command usage).

- [ ] **Step 5: Clean up the local dry-run artifacts**

```bash
rm -rf dist/
```

- [ ] **Step 6: Commit**

```bash
git add .goreleaser.yaml
git commit -m "build: add GoReleaser config for multi-platform release builds"
```

---

### Task 3: `build.yaml`

**Files:**
- Create: `.github/workflows/build.yaml`

**Interfaces:**
- Consumes: `.goreleaser.yaml` from Task 2 (build id `zipline`).
- Produces: a `build` matrix job (4 legs) uploading artifacts named `release-<goos>-<goarch>` (a `.tar.gz` archive + `.sha256` file each), and a `publish` job that downloads them and publishes the GitHub Release via `gh release create`.

**Correction:** the original design used `goreleaser release --clean --split`
per leg plus `goreleaser continue --merge` in a final job. That is
**GoReleaser Pro-only** (confirmed against goreleaser.com/customization/partial/:
"This feature is exclusively available with GoReleaser Pro") — the free/OSS
`goreleaser` binary these workflows install cannot run those commands. Use
`goreleaser build --clean --single-target` (OSS) per leg instead, then
package and publish the release with plain shell + the `gh` CLI (already
present on GitHub-hosted runners).

- [ ] **Step 1: Write the workflow file**

Create `.github/workflows/build.yaml`:

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  build:
    name: Build (${{ matrix.goos }}/${{ matrix.goarch }})
    strategy:
      fail-fast: false
      matrix:
        include:
          - os: macos-13
            goos: darwin
            goarch: amd64
          - os: macos-14
            goos: darwin
            goarch: arm64
          - os: ubuntu-latest
            goos: linux
            goarch: amd64
          - os: ubuntu-24.04-arm
            goos: linux
            goarch: arm64
    runs-on: ${{ matrix.os }}
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Build binary
        uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: build --clean --single-target

      - name: Package archive
        run: |
          set -euo pipefail
          BINARY=$(find dist -type f -name zipline)
          STAGE="zipline_${{ matrix.goos }}_${{ matrix.goarch }}"
          mkdir -p "$STAGE"
          cp "$BINARY" LICENSE README.md "$STAGE/"
          tar -czf "${STAGE}.tar.gz" "$STAGE"
          sha256sum "${STAGE}.tar.gz" > "${STAGE}.tar.gz.sha256"

      - name: Upload archive
        uses: actions/upload-artifact@v4
        with:
          name: release-${{ matrix.goos }}-${{ matrix.goarch }}
          path: |
            zipline_${{ matrix.goos }}_${{ matrix.goarch }}.tar.gz
            zipline_${{ matrix.goos }}_${{ matrix.goarch }}.tar.gz.sha256
          retention-days: 1

  publish:
    name: Publish Release
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Download all release artifacts
        uses: actions/download-artifact@v4
        with:
          pattern: release-*
          path: dist
          merge-multiple: true

      - name: Combine checksums
        run: cat dist/*.sha256 > dist/checksums.txt

      - name: Publish GitHub Release
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          gh release create "${{ github.ref_name }}" \
            --repo "${{ github.repository }}" \
            --title "${{ github.ref_name }}" \
            --generate-notes \
            dist/*.tar.gz dist/checksums.txt
```

- [ ] **Step 2: Validate YAML syntax**

Run:

```bash
python3 -c "import yaml, sys; yaml.safe_load(open('.github/workflows/build.yaml')); print('OK')"
```

Expected: `OK`

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/build.yaml
git commit -m "ci: add multi-platform GoReleaser release workflow"
```

---

### Task 4: Verify both workflows actually run on GitHub Actions

**Files:** none (verification only — no new files).

**Interfaces:** none.

> **STOP before Step 1 and Step 3:** both involve pushing to the `origin` remote (a branch/PR, and a tag that triggers a real release). Confirm explicitly with the user before running either step — do not push unattended.

- [ ] **Step 1: Push a branch and open a PR to exercise `linting-testing.yml`**

After user confirmation:

```bash
git push -u origin HEAD
gh pr create --title "ci: add lint/test and release workflows" --body "Adds linting-testing.yml and build.yaml (GoReleaser multi-platform release)."
```

- [ ] **Step 2: Watch the PR checks**

```bash
gh pr checks --watch
```

Expected: both `Lint` and `Test` jobs report success. If either fails, read the failure log with `gh run view --log-failed`, fix the root cause (source code or workflow file, whichever is actually wrong), commit, push, and re-watch.

- [ ] **Step 3: Push a throwaway tag to exercise `build.yaml`**

After user confirmation, on a commit that already has the workflow (e.g. after the PR above merges, or directly on the branch):

```bash
git tag v0.0.0-test1
git push origin v0.0.0-test1
```

- [ ] **Step 4: Watch the release run**

```bash
gh run watch
```

Expected: all 4 `build` matrix legs succeed, then `publish` succeeds, and `gh release view v0.0.0-test1` shows 4 platform archives plus `checksums.txt` attached.

- [ ] **Step 5: Clean up the throwaway release and tag**

After user confirmation:

```bash
gh release delete v0.0.0-test1 --yes
git push --delete origin v0.0.0-test1
git tag -d v0.0.0-test1
```

- [ ] **Step 6: Report results to the user**

Summarize which jobs passed, any fixes made along the way, and confirm both workflows are ready for real use (merging the PR, and tagging a real release later).
