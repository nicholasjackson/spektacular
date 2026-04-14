# Plan: 18_release_workflow

<!-- Metadata -->
<!-- Created: 2026-04-14T15:15:21Z -->
<!-- Commit: 39a5378 -->
<!-- Branch: skills -->
<!-- Repository: git@github.com:jumppad-labs/spektacular.git -->

<!--
  OVERVIEW
  A concise 2-3 sentence summary of the plan. Answer:
    1. What is being built?
    2. What problem does it solve?
    3. Who benefits?
  No file paths, no commands, no implementation detail. A reviewer should be
  able to decide whether the plan is worth reading in full from this section
  alone.
-->
## Overview

A GitHub Actions release workflow for spektacular that reuses jumppad's Dagger module to build cross-platform binaries (darwin/linux on amd64/arm64), sign and notarize macOS artifacts via Quill, and publish archives as assets on GitHub Releases, along with Homebrew formula updates and GemFury package distribution. This eliminates the current need for end users to clone and `go build` spektacular themselves, providing signed, installable binaries via direct download, Homebrew, or GemFury.

<!--
  ARCHITECTURE & DESIGN DECISIONS
  The chosen design direction in 2-4 short paragraphs. Explain the shape of
  the solution, the key decisions and their trade-offs, and why the chosen
  direction beats the alternatives. Cross-reference
  research.md#alternatives-considered-and-rejected so readers can drill into
  the evidence for rejected options. This is plan.md's load-bearing section —
  a reviewer should be able to spot missing architectural patterns or design
  gaps from this section without needing to read context.md.
-->
## Architecture & Design Decisions

The release workflow will reuse jumppad's proven Dagger module by importing it as a dependency rather than porting ~800 lines of code. Spektacular will create a thin wrapper Dagger module that imports `github.com/jumppad-labs/jumppad/dagger` and calls its functions with spektacular-specific parameters (binary name "spektacular", repo "jumppad-labs/spektacular", formula name "spektacular.rb"). This approach leverages Dagger's module system to inherit jumppad's battle-tested build, signing, and publishing logic while maintaining spektacular-specific configuration. The wrapper module will be invoked from a new `.github/workflows/build_and_deploy.yaml` workflow that runs builds on every push and gates releases to the `main` branch only.

**Key design decisions**:

1. **Module reuse over porting**: Importing jumppad's Dagger module as a dependency eliminates the need to port and maintain duplicate code. If direct reuse proves infeasible due to tight coupling, the implementation will fall back to creating a thin wrapper or selective porting. This trade-off prioritizes maintainability and proven patterns over custom implementation, but introduces a dependency on jumppad's module stability.

2. **PR-label-driven versioning**: Version resolution uses the Dagger GitHub module to read `release/patch|minor|major` labels from the associated PR, returning `0.0.0` when no label is found and failing the release. This enforces explicit version bumps on every merge to `main` and prevents accidental releases, but requires discipline to label every PR before merge.

3. **Mandatory macOS notarization on every build**: The build workflow will fail if Quill credentials are missing, ensuring every pushed commit produces release-ready signed binaries. This eliminates the "forgot to sign" failure mode at release time but requires contributors to have access to signing secrets or use a fork without the signing step.

4. **Shared Homebrew tap with jumppad**: The release function will create a new `Formula/spektacular.rb` file in the existing `jumppad-labs/homebrew-repo` tap, allowing users to `brew install jumppad-labs/repo/spektacular` alongside the existing `jumppad` formula. This reuses infrastructure but requires careful testing to ensure the tap remains functional for both tools.

This direction beats the alternatives (full port, minimal scope, hybrid approach) because it provides the fastest path to production-grade releases through proven architecture, ensures release artifacts are always signed and notarized through mandatory signing, and maintains consistency with jumppad's release process through shared patterns and infrastructure. See `research.md#alternatives-considered-and-rejected` for the options that were considered but rejected due to higher maintenance burden or incomplete spec coverage.

<!--
  COMPONENT BREAKDOWN
  The components (new or changed) that make up the solution, with their
  responsibilities and how they interact. One bullet or short paragraph per
  component. Name the component, state what it owns, and describe its
  relationship to the other components. Do not list file paths or line
  numbers here — component responsibilities, not implementation sites.
-->
## Component Breakdown

**Spektacular Dagger Wrapper Module** — A thin wrapper module in `dagger/` that imports jumppad's Dagger module and exposes `All()` and `Release()` functions with spektacular-specific configuration. Owns the mapping between spektacular's needs (binary name, repo paths, formula name) and jumppad's generic build functions. Delegates all build, signing, and publishing logic to the imported jumppad module.

**Jumppad Dagger Module (Imported)** — The upstream `github.com/jumppad-labs/jumppad/dagger` module that provides the core build pipeline. Handles cross-platform Go compilation, Quill-based macOS signing/notarization, archive packaging, GitHub release creation, Homebrew formula updates, and GemFury package publishing. Consumed as a Dagger module dependency.

**Build Configuration** — Spektacular-specific parameters passed to jumppad's module: binary name ("spektacular"), version string (from PR labels), repository ("jumppad-labs/spektacular"), formula name ("spektacular.rb"), and platform matrix (darwin/linux on amd64/arm64). Defined in the wrapper module's function signatures.

**GitHub Actions Workflow** — The `.github/workflows/build_and_deploy.yaml` file that triggers the pipeline. The `dagger_build` job runs on every push, invokes `dagger call all`, and uploads archives as a workflow artifact. The `release` job runs only on `main`, downloads the artifact, invokes `dagger call release`, and exposes the version as a job output. Uses pinned Dagger versions (CLI `0.19.8`, action `v8.2.0`).

<!--
  DATA STRUCTURES & INTERFACES
  The types, interface signatures, and serialization boundaries introduced or
  changed by the plan. Show type shapes in pseudocode or a short code block
  where useful. Focus on the contract between components, not internal
  representation detail.
-->
## Data Structures & Interfaces

**Spektacular Dagger Module** — The wrapper module struct:
```go
type SpektacularCI struct {
    jumppadCI *JumppadCI // Imported from github.com/jumppad-labs/jumppad/dagger
}
```

**Build Configuration** — Parameters passed to jumppad's module:
- `binaryName`: "spektacular"
- `version`: Semver string from PR labels (e.g., "1.2.3")
- `repository`: "jumppad-labs/spektacular"
- `formulaName`: "spektacular.rb"
- `platforms`: `["darwin/amd64", "darwin/arm64", "linux/amd64", "linux/arm64"]`

**Archive Metadata** — Structure inherited from jumppad's module:
- Filename pattern: `spektacular_{version}_{os}_{arch}.tar.gz`
- Internal structure: Single binary named `spektacular`
- Checksum file: `checksums.txt` with SHA256 hashes

**Quill Credentials** — Required secrets for macOS signing (inherited from jumppad):
- `QUILL_SIGN_P12`: PKCS#12 certificate file
- `QUILL_SIGN_PASSWORD`: Certificate password
- `QUILL_NOTORY_KEY`: Apple notary API key
- `QUILL_NOTARY_KEY_ID`: Key identifier
- `QUILL_NOTARY_ISSUER`: Issuer identifier

**Release Inputs** — Parameters for the Release function (inherited from jumppad):
- `archives`: Directory containing all `.tar.gz` files
- `githubToken`: GitHub personal access token
- `gemfuryToken`: GemFury API token
- Returns: Version string on success

**GitHub Actions Artifact** — Workflow artifact structure:
- Name: `archives`
- Contents: All four `.tar.gz` files from the build output directory
- Consumed by: `release` job when running on `main` branch

No new Go interfaces are introduced — the wrapper module delegates to jumppad's existing Dagger functions. The Dagger SDK provides all necessary abstractions (`Directory`, `File`, `Secret`, `Container`).

<!--
  IMPLEMENTATION DETAIL
  High-level only. Sketch new patterns being introduced, major code-shape
  changes, and code-structure UX — enough for a reviewer to spot missing
  patterns or design gaps. This is NOT per-phase file:line work — that
  belongs in context.md. If you find yourself writing "in file X at line Y",
  stop and move it to context.md.
-->
## Implementation Detail

**Dagger Module Reuse Pattern** — The implementation follows Dagger's module import pattern where spektacular's wrapper module declares a dependency on `github.com/jumppad-labs/jumppad/dagger` in its `dagger.json` and imports the `JumppadCI` struct. The wrapper module's methods instantiate the imported module with spektacular-specific configuration and delegate all operations. This pattern is idiomatic Dagger and allows spektacular to benefit from jumppad's improvements without manual syncing.

**Configuration Mapping Layer** — The wrapper module translates spektacular's domain concepts (binary name, repo, formula) into the parameters jumppad's module expects. This mapping happens at the function boundary where spektacular's `All()` and `Release()` methods call jumppad's corresponding functions with transformed inputs. The mapping is explicit and centralized in the wrapper module, making it easy to adjust if jumppad's interface changes.

**Fallback Strategy** — If direct module import proves infeasible (e.g., jumppad's module is too tightly coupled to jumppad-specific paths), the implementation will fall back to one of two approaches: (1) create a thin adapter layer that wraps jumppad's functions and translates parameters, or (2) selectively port only the necessary functions into spektacular's module. The decision point is during Phase 1.1 when attempting the import.

**Workflow Artifact Passing** — The GitHub Actions workflow uses the artifact upload/download pattern to pass build outputs between jobs. The `dagger_build` job uploads the entire archive directory as a single artifact, and the `release` job downloads it before invoking the Dagger release function. This decouples build from release and allows the release job to be conditionally skipped on non-main branches.

**Error Handling Inheritance** — The wrapper module inherits jumppad's error chaining pattern where operations check for previous errors before proceeding. This provides fail-fast behavior where the first error in a pipeline stops all subsequent operations. The wrapper module does not need to implement its own error handling — it simply propagates errors from the imported module.

The implementation reuses jumppad's proven patterns (error chaining, build matrix, Quill integration, multi-channel publishing) rather than introducing new abstractions. The only spektacular-specific logic is in the configuration mapping layer that translates spektacular's parameters into jumppad's expected format.

<!--
  DEPENDENCIES
  The internal packages, external libraries, upstream specs, or prior plans
  this work depends on. One bullet per dependency with a one-line note on
  what it provides and whether it needs any changes.
-->
## Dependencies

**Jumppad Dagger Module** — Provides the complete build, signing, and publishing pipeline. Imported as `github.com/jumppad-labs/jumppad/dagger`. No changes needed to jumppad's module.

**Dagger Go SDK** — Provides the container orchestration and module system. Pinned to version `0.19.8` to match jumppad. No changes needed.

**Dagger GitHub Module** — Used by jumppad's module for version resolution and GitHub release creation. Consumed transitively through jumppad's module. No changes needed.

**Quill (via Dagger)** — Apple code signing and notarization tool, invoked by jumppad's module. Requires P12 certificate, password, and notary credentials as secrets. No changes needed.

**GitHub Actions** — Workflow execution platform. Requires `dagger/dagger-for-github@v8.2.0` action pinned to match jumppad. No changes needed.

**GitHub Secrets** — Seven org-level secrets must exist: `GH_TOKEN`, `FURY_TOKEN`, `QUILL_SIGN_P12`, `QUILL_SIGN_PASSWORD`, `QUILL_NOTORY_KEY`, `QUILL_NOTARY_KEY_ID`, `QUILL_NOTARY_ISSUER`. These must be configured before the workflow can run successfully.

**jumppad-labs/homebrew-repo** — Existing Homebrew tap repository where `Formula/spektacular.rb` will be added. No changes to the tap structure needed, but write access via `GH_TOKEN` is required.

**GemFury APT Repository** — External package hosting service for deb/rpm distribution. Requires `FURY_TOKEN` for authentication. No changes needed.

**fpm or nfpm** — Package building tool for creating deb/rpm packages. Jumppad's module already uses this; spektacular inherits the dependency. No changes needed.

**Go toolchain** — Required for cross-compilation. Uses the version specified in `go.mod` (currently `1.25.0`). No changes needed.

No prior specs or plans block this work — the release workflow is independent of other spektacular features. The existing `.github/workflows/build.yml` will be replaced, so there are no compatibility constraints with the current workflow.

<!--
  TESTING APPROACH
  High-level overview of the testing strategy: what kinds of tests
  (unit, integration, contract, regression), which components get the most
  coverage, and what the load-bearing assertions are. Per-phase testing
  detail — which specific tests live in which specific files — stays in
  context.md.
-->
## Testing Approach

**Manual Dagger Function Testing** — Each Dagger function (`All`, `Release`) will be tested manually via `dagger call` commands during development. This allows verifying behavior with real inputs before committing changes.

**Manual Archive Verification** — After running the build, manually verify that:
- Archives follow the naming pattern `spektacular_{version}_{os}_{arch}.tar.gz`
- Each archive extracts to a single binary named exactly `spektacular`
- Archive count matches the expected four platforms

**Manual Signing Verification** — For macOS archives, manually verify that:
- `codesign -dv <binary>` shows valid Apple Developer signature
- `spctl -a -vv <binary>` confirms successful notarization
- Binaries run without Gatekeeper warnings on a clean macOS system

**Manual Release Verification** — After the first release to `main`, manually verify that:
- GitHub Release exists with all four archives as assets
- `jumppad-labs/homebrew-repo` has a new commit updating `Formula/spektacular.rb`
- GemFury repository shows the new packages
- `brew install jumppad-labs/repo/spektacular` works on macOS

**No Automated Tests** — No automated tests are added for the Dagger module or GitHub Actions workflow. The release infrastructure is tested through actual releases, with manual verification of each publishing channel. The existing `make test` suite for the Go CLI remains unchanged.

<!--
  MILESTONES & PHASES
  2-4 milestones. Each milestone leads with a "What changes" summary
  paragraph describing the user-visible difference when the milestone lands.
  Each phase has a 2-4 sentence plain-language summary, a *Technical detail:*
  link to context.md, and an **Acceptance criteria**: checkbox list with
  outcome statements (not shell commands). No file:line references in
  plan.md phase content — those live in context.md.
-->
## Milestones & Phases

### Milestone 1: Cross-Platform Build Infrastructure

**What changes**: Developers can run `dagger call all --src=. --version=1.0.0` to produce spektacular binaries for all four target platforms (darwin/linux on amd64/arm64) in a structured output directory. The Dagger module handles Go cross-compilation with version injection via ldflags, producing static binaries ready for distribution. This establishes the foundation for the release pipeline without requiring any external credentials or publishing infrastructure.

#### - [ ] Phase 1.1: Dagger Module Scaffolding

Create the basic Dagger module structure that imports jumppad's module and exposes wrapper functions. This establishes the foundation for reusing jumppad's proven build pipeline.

*Technical detail:* [context.md#phase-11-dagger-module-scaffolding](./context.md#phase-11-dagger-module-scaffolding)

**Acceptance criteria**:
- [ ] `dagger/dagger.json` exists and declares a dependency on `github.com/jumppad-labs/jumppad/dagger`
- [ ] `dagger/main.go` contains a `SpektacularCI` struct that imports and wraps `JumppadCI`
- [ ] Running `dagger functions` lists the wrapper functions without errors
- [ ] The module can be instantiated and called locally via `dagger call`

#### - [ ] Phase 1.2: Cross-Platform Build Function

Implement the `All` wrapper function that calls jumppad's build pipeline with spektacular-specific configuration.

*Technical detail:* [context.md#phase-12-cross-platform-build-function](./context.md#phase-12-cross-platform-build-function)

**Acceptance criteria**:
- [ ] `dagger call all --src=. --version=1.0.0` produces four binaries in the expected directory structure
- [ ] Each binary reports the correct version when executed with `--version`
- [ ] Linux binaries are statically linked (no dynamic dependencies)
- [ ] Darwin binaries are statically linked (no dynamic dependencies beyond system frameworks)
- [ ] Archives are created with the naming pattern `spektacular_{version}_{os}_{arch}.tar.gz`

### Milestone 2: Archive Packaging and macOS Signing

**What changes**: The Dagger module produces distribution-ready archives via `dagger call all`, creating `.tar.gz` files for each platform with the correct naming pattern. macOS archives are automatically signed and notarized via Quill, ensuring binaries pass Gatekeeper checks on user machines. This milestone delivers the complete artifact creation pipeline, making spektacular installable via direct download without triggering security warnings.

#### - [ ] Phase 2.1: macOS Signing and Notarization

Configure the wrapper to pass Quill credentials to jumppad's signing function, ensuring macOS binaries are signed and notarized.

*Technical detail:* [context.md#phase-21-macos-signing-and-notarization](./context.md#phase-21-macos-signing-and-notarization)

**Acceptance criteria**:
- [ ] Running `dagger call all` with valid Quill credentials produces signed macOS archives
- [ ] Extracted macOS binaries pass `codesign -dv` verification
- [ ] Extracted macOS binaries pass `spctl -a -vv` verification
- [ ] Running without credentials fails with a clear error message
- [ ] Linux archives are not modified by the signing process

### Milestone 3: Multi-Channel Release Publishing

**What changes**: The Dagger module can publish a complete release via `dagger call release`, creating a GitHub Release with all archives as assets, updating the Homebrew formula in `jumppad-labs/homebrew-repo`, and pushing packages to GemFury. Users can now install spektacular via `brew install jumppad-labs/repo/spektacular` on macOS or download signed binaries directly from GitHub Releases. This completes the distribution pipeline, making spektacular available through standard package managers.

#### - [ ] Phase 3.1: Release Orchestration

Implement the `Release` wrapper function that calls jumppad's release pipeline with spektacular-specific configuration.

*Technical detail:* [context.md#phase-31-release-orchestration](./context.md#phase-31-release-orchestration)

**Acceptance criteria**:
- [ ] `dagger call release` creates a GitHub Release tagged with the resolved version
- [ ] All four archives are attached as release assets
- [ ] A new commit appears in `jumppad-labs/homebrew-repo` updating `Formula/spektacular.rb`
- [ ] The formula contains download URLs for both darwin platforms with correct SHA256 checksums
- [ ] Packages appear in the GemFury repository after upload
- [ ] Function returns the released version string on success
- [ ] Function fails when version is `0.0.0` (no PR label)

### Milestone 4: GitHub Actions Automation

**What changes**: Pushing any branch triggers an automated build that produces signed, notarized archives uploaded as workflow artifacts. Merging to `main` automatically publishes a new release to all distribution channels (GitHub, Homebrew, GemFury) with version determined by the PR's release label. Developers no longer need to run Dagger commands manually — the entire release process is automated and gated to prevent accidental releases from non-main branches.

#### - [ ] Phase 4.1: Build Workflow

Create the GitHub Actions workflow that runs Dagger builds on every push.

*Technical detail:* [context.md#phase-41-build-workflow](./context.md#phase-41-build-workflow)

**Acceptance criteria**:
- [ ] `.github/workflows/build_and_deploy.yaml` exists
- [ ] Pushing any branch triggers the `dagger_build` job
- [ ] Job runs `dagger call all` with Quill credentials from secrets
- [ ] Job uploads an `archives` artifact containing four `.tar.gz` files
- [ ] Job shows green status in the Actions UI

#### - [ ] Phase 4.2: Release Workflow

Add the release job that publishes to all channels on `main` merges.

*Technical detail:* [context.md#phase-42-release-workflow](./context.md#phase-42-release-workflow)

**Acceptance criteria**:
- [ ] `release` job has `if: ${{ github.ref == 'refs/heads/main' }}` condition
- [ ] Job downloads the `archives` artifact from the build job
- [ ] Job runs `dagger call release` with `GH_TOKEN` and `FURY_TOKEN` secrets
- [ ] Job exposes the released version as an output
- [ ] Non-main pushes show the job as skipped in the Actions UI

#### - [ ] Phase 4.3: Cleanup Old Workflow

Remove the existing `.github/workflows/build.yml` file.

*Technical detail:* [context.md#phase-43-cleanup-old-workflow](./context.md#phase-43-cleanup-old-workflow)

**Acceptance criteria**:
- [ ] `.github/workflows/build.yml` is deleted
- [ ] No references to the old workflow remain in documentation
- [ ] The new workflow provides equivalent or better functionality

<!--
  OPEN QUESTIONS
  Strictly for questions that genuinely cannot be resolved until
  implementation begins. Anything resolvable by asking the user, reading the
  code, or running a quick experiment must be resolved now — not parked
  here. If this section is empty, that is the expected outcome of a healthy
  planning pass.
-->
## Open Questions

**Dagger module reuse vs. custom implementation** — Dagger supports importing modules as dependencies (e.g., `dagger install github.com/jumppad-labs/jumppad/dagger`). The implementer should first attempt to reuse jumppad's Dagger module directly by importing it and calling its functions with spektacular-specific parameters (binary name, repo paths, formula name). If direct reuse works, the implementation becomes configuration rather than porting ~800 lines of code. If jumppad's module is too tightly coupled to jumppad-specific paths or naming, the implementer should STOP and ask the user whether to:
1. Fork jumppad's module and adapt it for spektacular
2. Create a thin wrapper module that delegates to jumppad's module
3. Port the necessary functions into a new spektacular-specific module

**GemFury package metadata** — The exact package metadata (maintainer, description, license, dependencies) and fpm/nfpm invocation will be discovered when examining jumppad's `UpdateGemFury` implementation. If the ported or reused function fails to produce valid packages, the implementer should STOP and ask the user for guidance on package metadata or scope adjustment.

**Homebrew tap write permissions** — The `GH_TOKEN` secret must have write access to `jumppad-labs/homebrew-repo`. If permissions are insufficient or branch protection blocks automated commits, the implementer should STOP and ask the user to verify token permissions.

<!--
  OUT OF SCOPE
  Explicit exclusions agreed during planning. Each bullet states what is NOT
  being done and, where useful, where it is tracked instead. This is as
  important as the requirements — it prevents scope creep and sets clear
  expectations for reviewers.
-->
## Out of Scope

**Windows builds and distribution** — No Windows binaries, winget publishing, or `sync_winget` job. Windows support is not planned for spektacular at this time.

**Functional/integration tests in the release pipeline** — No container-based tests or jumppad-style blueprint tests. The release workflow focuses on artifact creation and distribution, not runtime behavior validation.

**Chat notifications** — No Discord, Slack, or other chat integrations for build/release status. Notifications remain in GitHub Actions UI only.

**Website version updates** — No spektacular website exists, so no `UpdateWebsite` step or version file publishing.

**Alternative package managers** — No apt PPA, snap, chocolatey, Scoop, or AUR support beyond Homebrew and GemFury.

**Container/OCI image publishing** — No Docker Hub, GHCR, or other container registry publishing. Spektacular is distributed as binaries only.

**Linux binary signing** — Only macOS binaries are signed and notarized. Linux binaries remain unsigned.

**Alternative release triggers** — No `workflow_dispatch`, scheduled releases, or tag-push triggers. Releases only occur on merges to `main` with PR labels.

**Backporting releases** — No support for releasing from non-main branches or maintaining multiple release branches.

**Graceful migration from old workflow** — The existing `.github/workflows/build.yml` is replaced outright with no backwards compatibility or transition period.
