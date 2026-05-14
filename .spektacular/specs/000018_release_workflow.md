# Feature: 18_release_workflow

## Overview

A GitHub Actions release workflow for spektacular that mirrors the jumppad release pipeline — a Dagger module builds cross-platform binaries (darwin/linux on amd64/arm64), signs and notarizes the macOS artifacts via Quill, and publishes archives as assets on a GitHub Release, along with a Homebrew formula update and a Gemfury package push. This removes the current need for end users to clone and `go build` spektacular themselves, giving them a signed, installable binary via direct download, Homebrew, or Gemfury.

## Requirements

- **Dagger module present at repo root**
  A `dagger/` directory exists containing `dagger.json`, `main.go`, and supporting files, exposing Dagger functions to build, package, and release spektacular modeled on `jumppad-labs/jumppad/dagger/main.go`.
- **Cross-platform archive production**
  The Dagger build function produces archives for darwin/amd64, darwin/arm64, linux/amd64, and linux/arm64, named `spektacular_<version>_<os>_<arch>.tar.gz`, written to an output directory.
- **Mandatory macOS notarization**
  The Dagger build function signs and notarizes macOS binaries via Quill on every run, consuming `notorize-cert`, `notorize-cert-password`, `notorize-key`, `notorize-id`, and `notorize-issuer` inputs, and fails the build if notarization fails.
- **PR-label-driven versioning**
  The Dagger module determines the release version by reading the next-version label from the associated PR via the Dagger GitHub module (same mechanism as `JumppadCI.getVersion`), returning `0.0.0` when no label is found and failing the release when the version is `0.0.0`.
- **Release function with side-effects**
  The Dagger module exposes a `Release` function that takes `src`, `archives`, `githubToken`, and `gemfuryToken`, creates a GitHub Release with the archives as assets, publishes a `spektacular.rb` formula to `jumppad-labs/homebrew-repo`, pushes packages to Gemfury, and returns the released version string.
- **No website step and no Windows logic**
  The Dagger module does not include an `UpdateWebsite` step or any Windows / winget logic.
- **Build job runs on every push**
  A GitHub Actions workflow at `.github/workflows/build_and_deploy.yaml` runs on push to any branch, executes the Dagger build job, signs/notarizes on every run, and uploads the archive directory as a workflow artifact named `archives`.
- **Release job gated on main**
  The workflow's `release` job runs only when `github.ref == 'refs/heads/main'`, downloads the `archives` artifact, invokes the Dagger `Release` function with `GH_TOKEN` and `FURY_TOKEN` secrets, and exposes the released version as a job output.
- **No excluded jobs**
  The workflow does not include `sync_winget`, `winget`, functional test, or Discord notification jobs.

## Constraints

- The workflow must consume exactly these secret names so existing org-level secrets are reused: `GH_TOKEN`, `FURY_TOKEN`, `QUILL_SIGN_P12`, `QUILL_SIGN_PASSWORD`, `QUILL_NOTORY_KEY`, `QUILL_NOTARY_KEY_ID`, `QUILL_NOTARY_ISSUER`.
- The workflow must pin `dagger/dagger-for-github@v8.2.0` and Dagger CLI `0.19.8` (matching jumppad) so no Dagger upgrades are required.
- The extracted binary inside each archive must be named exactly `spektacular` (no OS/arch suffix).
- The existing `.github/workflows/build.yml` may be replaced or removed outright; no backwards compatibility is required.
- License headers on new files are not required.

## Acceptance Criteria

- **Dagger module exists and is callable**
  `dagger/dagger.json`, `dagger/main.go`, and `dagger/go.mod` are present; `dagger functions` run against `./dagger` lists the `all` and `release` functions.
- **Archive directory contains exactly the expected files**
  After `dagger call all --src=. --output=./output ...`, `./output` contains exactly four files matching `spektacular_<version>_{linux,darwin}_{amd64,arm64}.tar.gz` and no other archives; each archive extracts to a binary named exactly `spektacular`.
- **Notarization is enforced**
  Running the Dagger build without valid Quill inputs exits non-zero; with valid inputs, `codesign -dv` and `spctl -a -vv` on the extracted darwin binaries report a signed, notarized binary.
- **Version resolution pass/fail**
  Calling the Dagger release function against a commit whose associated PR has a `release/patch` | `release/minor` | `release/major` label returns the bumped semver string; running it against a commit with no labeled PR exits non-zero with the message `no version to release, did you tag the PR?`.
- **Release function side-effects observable**
  After a successful `dagger call release` on `main`: a GitHub Release tagged with the resolved version exists on `jumppad-labs/spektacular` and contains all four archives as assets; `jumppad-labs/homebrew-repo` has a commit on `main` updating `Formula/spektacular.rb` to the resolved version; a package matching the resolved version appears in the Gemfury repo; the function returns the version string to stdout.
- **No website update**
  `grep -r UpdateWebsite dagger/` returns no matches.
- **Build job green and artifact uploaded**
  `.github/workflows/build_and_deploy.yaml` exists; pushing any branch triggers the `dagger_build` job, which finishes green and uploads a workflow artifact named `archives` containing the four tar.gz files.
- **Release job gating wired correctly**
  The `release` job has `if: ${{ github.ref == 'refs/heads/main' }}`, downloads the `archives` artifact, invokes the Dagger `Release` verb with `GH_TOKEN` and `FURY_TOKEN` env vars, writes the returned version to `./version.txt`, and exposes it via `outputs.version`; non-main pushes show the job as skipped in the Actions UI.
- **Excluded jobs absent**
  `grep -E 'sync_winget|winget|discord|Discord|functional_test' .github/workflows/build_and_deploy.yaml` returns no matches.

## Technical Approach

- Port `jumppad-labs/jumppad/dagger/main.go` to `spektacular/dagger/main.go`, renaming `JumppadCI` to `SpektacularCI` and replacing jumppad-specific paths, binary names, repo owner/name, and formula/package identifiers with spektacular equivalents.
- Remove the `UpdateWebsite` function and its call from `Release`; remove any Windows-targeted build branches from the archive loop; remove Winget-related helpers.
- `UpdateBrew` updates a new `Formula/spektacular.rb` file on `jumppad-labs/homebrew-repo`, reusing the existing token/workflow pattern so a single tap hosts both `jumppad.rb` and `spektacular.rb`.
- `getVersion` is reused verbatim — PR-label-driven bumps via the Dagger GitHub module against `jumppad-labs/spektacular`.
- Build uses the same Quill-based signing flow that jumppad uses, invoked for darwin archives only; linux archives skip signing.
- `.github/workflows/build_and_deploy.yaml` is created new, copied from jumppad's workflow minus the functional-test, podman, discord, sync_winget, and winget jobs; the existing `.github/workflows/build.yml` is deleted.
- Dagger CLI pinned to `0.19.8` and `dagger/dagger-for-github@v8.2.0`, same as jumppad, so no Dagger upgrades are required.
- **Known risks**:
  - Gemfury packaging expects deb/rpm; jumppad's `UpdateGemFury` likely builds these via fpm/nfpm — spektacular will need matching package metadata (maintainer, description, license).
  - The shared `jumppad-labs/homebrew-repo` tap must keep working for existing `jumppad` consumers while a new `spektacular.rb` is added alongside.
  - PR-label versioning requires every merged PR to `main` to carry a `release/*` label; without enforcement, releases on `main` will fail with `no version to release`.

## Success Metrics

No formal success metrics defined — the user will manually verify the workflow after delivery.

## Non-Goals

- Windows builds and distribution, including winget publishing and any `sync_winget` job.
- Functional / integration tests in the release pipeline (e.g. container-based `jumppad test` equivalents).
- Discord or any other chat notifications on build/release status.
- Website version updates — there is no spektacular website repo.
- Alternative package managers beyond Homebrew and Gemfury (no apt PPA, snap, chocolatey, Scoop, AUR).
- Container / OCI image publishing.
- Signing of linux binaries — only macOS binaries are signed/notarized.
- Release automation triggers other than push to `main` (no `workflow_dispatch`, no scheduled releases, no tag-push triggers).
- Backporting releases to non-main branches.
- Graceful retirement of `.github/workflows/build.yml` — it is replaced outright.
