# Plan: 15_implementation

<!-- Metadata -->
<!-- Created: 2026-04-13T11:03:54Z -->
<!-- Commit: 005ab0973eaea080fe73499090a5e04298e6184e -->
<!-- Branch: skills -->
<!-- Repository: git@github.com:jumppad-labs/spektacular.git -->

## Overview

Add a new `implement` workflow to spektacular that guides an agent through executing an approved plan to produce source code. The workflow mirrors the existing `spec` and `plan` command surface — markdown templates under `templates/`, mustache rendering, stateless JSON responses per step — and delegates analysis, test authoring, verification, and changelog updates to sub-agents via the existing skill-reference pattern. This plan also performs an internal refactor that lifts the step-rendering helpers currently duplicated inside `internal/steps/plan` into a new shared `internal/stepkit` package so that `plan`, `spec`, and the new `implement` workflow all consume the same rendering pipeline.

## Architecture & Design Decisions

The `implement` workflow is added as a third peer to `spec` and `plan`, sharing the same markdown-template + mustache rendering pipeline. To avoid cementing duplicate helper code across three packages, this plan extracts the step-rendering helpers (`writeStepResult`, `renderTemplate`, `getString`, `stepTitle`) out of `internal/steps/plan/steps.go` into a new shared `internal/stepkit` package — then rewrites `internal/steps/plan/steps.go`, `internal/steps/spec/steps.go`, and the new `internal/steps/implement/steps.go` to consume it. This is a larger blast radius than a scoped mirror-copy, but the repeated helpers are about to appear in a third place and a DRY extraction now is cheaper than three divergent copies later. Cross-reference: `research.md#alternatives-considered-and-rejected` for the options considered and rejected, and `research.md#chosen-approach--evidence` for the supporting file:line citations.

The shared helper package exposes a small `PathStrategy` interface so each workflow injects its own path conventions without the shared package knowing which workflow it's rendering for. Path derivation differs per workflow — `spec` writes `specs/<name>.md`, `plan` writes `plans/<name>/{plan,context,research}.md`, `implement` reads `plans/<name>/plan.md` without writing its own file — so the helper accepts a small strategy value per workflow and folds the result into the shared template-variable map. This preserves the existing `{{spec_path}}`/`{{plan_path}}`/`{{config.command}}` contract used by every template under `templates/steps/spec/` and `templates/steps/plan/` while adding implement-only vars through the extra-vars mechanism.

The implement FSM has ten steps: `new → read_plan → analyze → implement → test → verify → update_plan → update_changelog → {analyze | update_repo_changelog} → finished`. The loop back from `update_changelog` to `analyze` is encoded natively by `workflow.StepConfig.Src` accepting multiple source states — `analyze.Src = []string{"read_plan","update_changelog"}` — so `update_changelog` has two legal exits (`goto analyze` when unchecked phases remain, `goto update_repo_changelog` otherwise) and the rendered template instructs the agent which to call based on whether any `#### - [ ]` checkboxes remain unchecked in the plan file. `update_repo_changelog.Src = []string{"update_changelog"}` and `finished.Src = []string{"update_repo_changelog"}`. No new engine features are required; the existing `go-fsm` integration already supports multi-source events. Phase progress is the plan file's own checkbox state — the workflow stores only the plan name in the shared `.spektacular/state.json` (the same state file spec and plan use), making the workflow resumable via whatever plan name was last set.

The `read_plan` step is the workflow's **validation and drift gate** — it runs before any analysis or implementation work begins. Its template instructs the agent to do three things before advancing: (1) read plan.md, context.md, and research.md in full, (2) structurally validate plan.md against the scaffold shape (required sections, phase checkbox headings, resolvable `*Technical detail:*` context links), and (3) drift-check every file path, package path, and symbol named in plan.md or context.md against the current working tree and STOP on mismatch. This gate exists because plans have an observed tendency to drift from the codebase between authoring and implementation (package moves, file renames, stale line numbers); catching that drift up-front — and explicitly asking the user whether to fix the plan, proceed with corrections in memory, or abandon — is cheaper than discovering it mid-implementation when an agent has already written code against stale pointers.

The plan-inline changelog is owned end-to-end by the implement workflow and written to `plan.md` as a `## Changelog` section. The plan workflow and its scaffold do not reserve, mention, or manage the changelog in any way — the plan-scaffold template ends at `## Out of Scope`. The implement workflow's first `update_changelog` invocation for a given plan checks whether a `## Changelog` section exists and, if not, creates it (appended after `## Out of Scope`); subsequent invocations append new phase entries under the existing section. First-phase vs subsequent-phase detection is a simple "does `## Changelog` exist in `plan.md`" check. This deviates from the spec's original wording ("adjacent `changelog.md`") but satisfies its intent while keeping all implementation history in one file the reviewer already has open. Evidence and rejected alternatives for this decision live in `research.md § Alternatives considered and rejected`.

In addition to the per-phase plan-inline changelog, the implement workflow appends a user-facing summary to the repo-level `CHANGELOG.md` exactly once per plan, as the final step before `finished`. The `update_repo_changelog` step uses the plan name (the `name` key stored in workflow data) as the section header and instructs the agent to write a short user-facing summary of the overall change, creating `CHANGELOG.md` at the repo root if it does not already exist and prepending the new entry above any existing sections. This is deliberately distinct from the plan's inline `## Changelog` section: the inline changelog is the phase-by-phase implementation audit log, while the repo `CHANGELOG.md` is the human-facing release note. This capability extends beyond the spec's acceptance criteria; if the spec is treated as a hard contract, it should be amended alongside this plan change.

## Component Breakdown

- **`internal/stepkit`** (new) — A shared step-rendering helper package extracted from `internal/steps/plan/steps.go`. Owns template lookup, mustache rendering, template-variable assembly, and the standard `Result{Step, <PathField>, <NameField>, Instruction}` write pattern. Exposes a primary function that takes a per-workflow path strategy and optional extra vars, and writes a rendered instruction to the workflow's `ResultWriter`. Also houses the small utilities currently duplicated: `stepTitle`, `getString`, and `renderTemplate`. Consumed by `internal/steps/plan`, `internal/steps/spec`, and `internal/steps/implement`.

- **`internal/steps/plan`** (changed) — Loses its private helper bundle; each step callback becomes a one-line call to `stepkit` with a plan-specific path strategy that supplies `plan_path`, `context_path`, `research_path`, `plan_dir`, `plan_name`, `spec_path`. Step order and FSM transitions are unchanged. The public `PlanFilePath`/`ContextFilePath`/`ResearchFilePath` functions stay in place — they're called from outside the package by `cmd/plan.go`.

- **`internal/steps/spec`** (changed) — Same treatment as `internal/steps/plan`. Step callbacks migrate to `stepkit` with a spec-specific path strategy. The spec package's `new()` step — the one place spec diverges from plan by writing a scaffold file during `new` rather than at verification — keeps its file-writing logic alongside the stepkit call.

- **`internal/steps/implement`** (new) — Owns the implement-workflow step definitions, the ten-step FSM (`new → read_plan → analyze → implement → test → verify → update_plan → update_changelog → {analyze|update_repo_changelog} → finished`), and thin callbacks that delegate to `stepkit`. Exposes `Steps()`, `Result`, `StatusResult`, `StepEntry`, `StepsResult`, and a `PlanFilePath` helper.

- **`cmd/implement.go`** (new) — Cobra command group wiring `implement new`, `implement goto`, `implement status`, `implement steps`. Near-duplicate of `cmd/plan.go`: owns `--data`/`--stdin`/`--file`/`--schema`/`--dry-run` flag parsing, JSON input validation, and the `workflow.New` + `store.NewFileStore(dataDir)` construction. The data dir is the same `.spektacular/` used by spec and plan, and the state file is the same shared `.spektacular/state.json`. There is no per-workflow subdirectory and no `findActiveImplement` — `implement new` truncates the shared state file just like `plan new` and `spec new` do today. The shared store root lets implement-workflow instructions read plan files at `plans/<name>/plan.md` via the existing store.

- **`templates/steps/implement/`** (new) — Nine mustache templates, one per non-initialization step (including the new `update_repo_changelog` template). Picked up automatically by the existing `//go:embed all:*` directive in `templates/templates.go`. Templates reference existing and new skills via the established `{{config.command}} skill <name>` pattern.

- **`templates/skills/`** (changed) — Gains three new embedded skill markdown files: `skill_follow-test-patterns.md`, `skill_verify-implementation.md`, `skill_update-changelog.md`. Referenced by the new implement templates and loaded at runtime through the same lookup path that already serves `skill_spawn-implementation-agents.md`. No Go changes.

- **`templates/scaffold/plan.md`** (unchanged) — The plan scaffold ends at `## Out of Scope` and has no reference to changelog anywhere. The implement workflow creates the `## Changelog` section directly in each plan's `plan.md` at runtime, not via the scaffold. (The scaffold directory also contains `context.md`, `research.md`, and `spec.md` — one file per write-step output; none of them contain changelog content.)

- **Test files** (new) — `internal/stepkit/stepkit_test.go`, `internal/steps/implement/steps_test.go`, `cmd/implement_test.go`. Plan and spec step tests continue to pass unchanged as the behavior-preservation fence for the refactor.

## Data Structures & Interfaces

This plan introduces one new public contract (`stepkit.PathStrategy`), migrates two existing result-struct sets to share helper types, and adds a mirror of those result types for `internal/steps/implement`. No persisted data format changes — the workflow state file, store layout, and embedded FS contract are all unchanged.

**`stepkit.PathStrategy` (new interface)**:

```go
package stepkit

type PathStrategy interface {
    NameKey() string                                             // always "name" today
    PathVars(name string, storeRoot string) map[string]any       // workflow-specific template vars
    PrimaryPathField() string                                    // "spec_path" / "plan_path"
}
```

**`stepkit.StepRequest` (new struct)**:

```go
type StepRequest struct {
    StepName     string
    NextStep     string
    TemplatePath string
    Strategy     PathStrategy
    Extra        map[string]any
}
```

**`stepkit.WriteStepResult` (new function)**:

```go
func WriteStepResult(
    req StepRequest,
    data workflow.Data,
    out workflow.ResultWriter,
    st store.Store,
    cfg workflow.Config,
    resultBuilder func(name, primaryPath, instruction string) any,
) error
```

The `resultBuilder` closure returns a workflow-specific `Result` struct so each workflow keeps its own JSON output shape without `stepkit` needing to know about them.

**Shared helper utilities** (moved from `internal/steps/plan/steps.go`, made exported):

```go
func StepTitle(snake string) string
func GetString(data workflow.Data, key string) string
func RenderTemplate(path string, vars map[string]any) (string, error)
```

**`internal/steps/implement.Result` (new)**:

```go
type Result struct {
    Step        string `json:"step"`
    PlanPath    string `json:"plan_path"`
    PlanName    string `json:"plan_name"`
    Instruction string `json:"instruction"`
}
```

**`internal/steps/implement.StatusResult` (new)** — adds one field beyond plan's:

```go
type StatusResult struct {
    PlanName        string      `json:"plan_name"`
    PlanPath        string      `json:"plan_path"`
    CurrentStep     string      `json:"current_step"`
    CompletedSteps  []string    `json:"completed_steps"`
    TotalSteps      int         `json:"total_steps"`
    Progress        string      `json:"progress"`
    Steps           []StepEntry `json:"steps"`
    UncheckedPhases int         `json:"unchecked_phases"`
}
```

`UncheckedPhases` is computed at `status` time by grepping the plan file for unchecked phase headings. Zero-valued when the plan file can't be read.

**`cmd/implement.go` input schemas** — same `schemaObj`/`schemaProp`/`commandSchema` types already defined in `cmd/spec.go`. No new schema types.

**Template-variable contract** (extended for implement) — every implement template receives `{{step}}`, `{{title}}`, `{{next_step}}`, `{{config.command}}` from stepkit; `{{plan_path}}`, `{{plan_dir}}`, `{{plan_name}}` from the implement path strategy; and `{{changelog_section_name}}` as a literal string (`"## Changelog"`) so tests can substitute it.

**No workflow-state schema changes**. `internal/workflow.State` is unchanged. `internal/steps/implement` stores only the `name` key in `Data`, same pattern as plan/spec.

**No store interface changes**. `store.Store` / `store.FileStore` are used via existing methods only.

## Implementation Detail

**New pattern: workflow step rendering as a shared-helper contract.** Today, each workflow package owns its own copy of a `writeStepResult`/`renderTemplate`/`getString`/`stepTitle` bundle — a ~60 line helper closure over its own path conventions. This plan introduces an explicit seam: a new `internal/stepkit` package that owns template loading, mustache rendering, standard template-variable assembly, and `Result` serialization, and a thin `PathStrategy` interface that each workflow implements to inject its own path conventions. After this plan lands, adding a fourth workflow (if one ever arrives) is a 3-file task — a `steps.go` with the step list, a `result.go` with the Result types, and a handful of templates — rather than a copy-paste-and-diverge job.

**New pattern: multi-source FSM transitions used deliberately.** The existing `workflow.StepConfig.Src` field accepts `[]string` and the underlying FSM library supports multi-source events, but no current workflow actually uses this. The implement workflow is the first to use it in earnest: the `analyze` step lists both `read_plan` and `update_changelog` as sources, and `finished` lists `update_changelog` as its only source. This encodes a loop directly in the FSM declaration rather than in imperative callback code, and keeps the "phase state is derived from plan.md, not stored" invariant visible at the data-structure level. The implement templates decide which `goto` to call by telling the agent to re-read the plan and count unchecked phases.

**New pattern: agent delegation via inline template directives.** The implement workflow's `analyze`, `test`, `verify`, and `update_changelog` steps all instruct the agent to delegate work to sub-agents via the existing skill-reference pattern (`{{config.command}} skill <name>`), matching how `templates/steps/plan/02-discovery.md` and `templates/steps/plan/10-phases.md` already delegate. No new agent-delegation mechanism is introduced. The three new skill files extend the existing skill library rather than introducing a new subsystem.

**Code-shape change: `internal/steps/plan/steps.go` and `internal/steps/spec/steps.go` shrink.** After extraction, both packages lose their private helper bundle and each step callback becomes a one-liner delegation to `stepkit.WriteStepResult` with a package-local `PathStrategy` value. The public API of each package is unchanged, so `cmd/plan.go` and `cmd/spec.go` don't need edits beyond their existing call sites. The existing tests in `internal/steps/plan/steps_test.go` and `internal/steps/plan/scaffold_test.go` continue to pass without modification because they assert on rendered-template substrings, not on internal function signatures.

**Code-structure UX.** A developer looking at `internal/steps/plan/steps.go` after this lands sees a clean list of `StepConfig` entries and a clean list of one-line callbacks. A developer tracing a rendering bug follows a single path through `stepkit.WriteStepResult` rather than three parallel copies. A developer adding a new template variable edits one file and the change propagates to all three workflows.

**Following existing patterns.** Cobra command groups with `new`/`goto`/`status`/`steps` subcommands; mustache templates under `templates/steps/<workflow>/NN-<name>.md`; embedded FS via `//go:embed all:*`; stateless JSON outputs; workflow state persisted at a single shared `.spektacular/state.json` (every `<workflow> new` truncates and starts fresh — there is no per-workflow state subdirectory); tests using `stretchr/testify/require` co-located with each package. The one genuinely new pattern is `internal/stepkit`, and even that is a refactoring of code that already exists.

## Dependencies

**Internal packages (runtime)**

- **`internal/workflow`** — Provides the FSM engine, `StepConfig`, `Data`/`ResultWriter`/`Config` types, and multi-source transition support. No changes required.
- **`internal/store`** — Provides the `Store` interface and `FileStore` used to read plan files. No changes required.
- **`internal/output`** — Provides JSON output writers used by every Cobra `RunE`. No changes required.
- **`templates/skills`** — Needs three new skill markdown files added; no Go changes.
- **`internal/config`** — Provides `loadConfig()` used to derive `Config.Command`. No changes required.
- **`templates/` (embed package)** — Already covers any new subdirectory via `//go:embed all:*`; no Go changes.
- **`internal/steps/plan`** — Changed consumer. Public API unchanged.
- **`internal/steps/spec`** — Changed consumer. Public API unchanged.
- **`internal/stepkit`** — New package.
- **`internal/steps/implement`** — New package.

**External libraries (runtime)**

- **`github.com/cbroglie/mustache`** — Already a project dependency. `internal/stepkit` imports it after extraction; `internal/steps/plan` and `internal/steps/spec` no longer do. No `go.mod` changes.
- **`github.com/spf13/cobra`** — Already a project dependency. No version changes.
- **`github.com/stretchr/testify/require`** — Already a project dependency. No version changes.
- **`github.com/looplab/fsm`** (transitively via `internal/workflow`) — Already depended on. Multi-source `Src` support is assumed present in the pinned version; this assumption is flagged in Open Questions with a concrete verification step.

**Tooling dependencies**

- Existing Go toolchain per `go.mod`.
- `make test`, `make lint`, `go test ./...` — used by phase success criteria.
- `make harbor-test` — existing integration harness for the spec workflow; not extended in this plan (see Out of Scope).

**Planning dependencies**

- **Spec `15_implementation`** — already written, at `.spektacular/specs/15_implementation.md`.
- **Plan `16_plan_format`** — already landed; established the current plan scaffold shape including the inline `## Changelog` section this plan treats as a contract.
- **No upstream spec or plan must land before this plan begins.** All dependencies are in-tree and current.

## Testing Approach

Testing for this plan is almost entirely unit-level, matching the existing project convention. The plan's correctness is verified against three load-bearing guarantees: the stepkit extraction is behavior-preserving for `plan` and `spec`, the implement FSM walks the correct sequence under all branches (including the loop), and every implement template contains the directives that the spec's acceptance criteria require.

**Three tiers of tests, each matching an existing pattern:**

1. **Behavior-preservation tests for `plan` and `spec`.** The existing tests in `internal/steps/plan/steps_test.go` and `internal/steps/plan/scaffold_test.go` — which exercise rendered step output via a `renderStep(t, cb)` helper and assert substring presence — are the regression fence. They continue to pass unmodified after the stepkit extraction. This is the primary signal that the refactor was behavior-preserving. If the corresponding `internal/steps/spec` package lacks equivalent step tests today, this plan adds minimal step-rendering coverage before touching spec code so the refactor has a safety net.

2. **Stepkit unit tests.** A new `internal/stepkit/stepkit_test.go` file covers the helper directly: `StepTitle` edge cases, `RenderTemplate` success and error paths, and `WriteStepResult` correctly merging standard + strategy + extra vars, invoking the result builder, and writing to the output. Uses the same minimal-fake pattern as `internal/steps/plan/steps_test.go`.

3. **Implement FSM and template tests.** A new `internal/steps/implement/steps_test.go` covers: step-order assertions paralleling `internal/steps/plan/steps_test.go:75`; an FSM happy-path walk paralleling line 100; a dedicated loop test that walks `read_plan → analyze → … → update_changelog → analyze → update_changelog → finished`; and per-step `TestXxxStepContains…` tests mirroring lines 45-73 that assert each template contains the directive its corresponding spec acceptance criterion requires. Spec requirement → test mapping is 1:1 for every criterion in spec lines 44-65.

4. **Cobra command tests.** A new `cmd/implement_test.go` covering `new`/`goto`/`status`/`steps`/`--schema` surfaces, happy-path and error-path outputs, and JSON input validation. Pattern matches existing cmd test files.

**Coverage weighting.** The highest-risk change is the `internal/stepkit` extraction because it touches three packages; the stepkit unit tests plus the unchanged `internal/steps/plan` test suite give that change two layers of safety. The second-highest-risk change is the multi-source FSM loop because it's the one new mechanism; the loop test is the critical test and should be written before the implement FSM is wired up so the test drives the wiring.

**Deliberately out of scope for test coverage.** No harbor end-to-end test for the implement workflow (see Out of Scope). No mock-the-agent tests — the workflow produces instructions, it doesn't execute them. No skill-file content tests — those are human-reviewed content. No performance tests.

**Where tests slot into existing conventions.** Every new test file follows the same rules as the existing ones: package-co-located, `_test.go` suffix, `require` assertions, `t.TempDir()` for filesystem fixtures, helper functions like `renderStep(t, cb)` copied from `internal/steps/plan/steps_test.go:35` into new test files per the existing copy-paste-per-package convention documented in `thoughts/notes/testing.md`. No new shared test-fixture package.

## Milestones & Phases

### Milestone 1: Shared step-rendering helper lands (internal refactor)

**What changes**: Nothing user-visible. This milestone is a pure internal refactor: the step-rendering helper currently duplicated inside `internal/steps/plan/steps.go` gets lifted into a new `internal/stepkit` package, and both `internal/steps/plan` and `internal/steps/spec` are rewritten to consume it through a small `PathStrategy` interface. The `spec` and `plan` CLI commands keep their exact current behavior — same JSON shapes, same file outputs, same error messages. This cleanup milestone exists as its own deliverable because it has a blast radius across two stable workflows and needs to ship behind a green test suite before any implement-workflow code is written; rolling it together with Milestone 2 would make a regression in the refactor indistinguishable from a bug in the new workflow.

#### - [ ] Phase 1.1: Extract step-rendering helpers into `internal/stepkit`

A new package `internal/stepkit` is created and the helpers currently private to `internal/steps/plan` — `writeStepResult`, `renderTemplate`, `getString`, `stepTitle` — are lifted into it and made exported. A narrow `PathStrategy` interface lets each workflow inject its own path conventions without the shared package knowing about them. No existing behavior changes in this phase because nothing calls the new package yet; this is a pure add-only step that lets the next two phases do their migrations against a stable target.

*Technical detail:* [context.md#phase-11](./context.md#phase-11-extract-step-rendering-helpers-into-internalstepkit)

**Acceptance criteria**:

- [ ] A new `internal/stepkit` package exists with exported helpers covering template rendering, title formatting, safe workflow-data lookup, and step-result writing.
- [ ] The helper accepts a small strategy value so each calling workflow can inject its own path-variable set without stepkit knowing about plan, spec, or implement.
- [ ] A unit test file exercises the helper's contract end-to-end with a minimal fake strategy.
- [ ] `make test` and `make lint` pass. No existing tests are modified in this phase.

#### - [ ] Phase 1.2: Rewrite `internal/steps/plan` to use `stepkit`

The `internal/steps/plan` package drops its private helper bundle and each step callback becomes a one-line call through `stepkit` with a package-local strategy value that supplies the plan-specific template variables. The package's public API stays exactly as it is today, so `cmd/plan.go` and all existing tests need no changes. The existing `internal/steps/plan/steps_test.go` and `internal/steps/plan/scaffold_test.go` test suites are the behavior-preservation fence.

*Technical detail:* [context.md#phase-12](./context.md#phase-12-rewrite-internalplan-to-use-stepkit)

**Acceptance criteria**:

- [ ] Every step callback in `internal/steps/plan` delegates to `stepkit` instead of a private helper.
- [ ] `internal/steps/plan` no longer contains its own copies of the template-rendering and step-result-writing helpers.
- [ ] All existing `internal/steps/plan` tests pass without modification.
- [ ] Running `go run . plan new` and `go run . plan goto` against a fixture plan produces the same JSON output as before the refactor.

#### - [ ] Phase 1.3: Rewrite `internal/steps/spec` to use `stepkit`

The `internal/steps/spec` package gets the same treatment. The one place spec diverges from plan — the `new()` step writes a scaffold file during initialization — keeps its file-writing side effect alongside the stepkit call. Running this phase after Phase 1.2 keeps the two refactors in tight sequence so the stepkit contract is exercised by a second distinct caller immediately.

*Technical detail:* [context.md#phase-13](./context.md#phase-13-rewrite-internalspec-to-use-stepkit)

**Acceptance criteria**:

- [ ] Every step callback in `internal/steps/spec` delegates to `stepkit`.
- [ ] The `spec new` step still writes the spec scaffold file during initialization.
- [ ] `internal/steps/spec` no longer contains its own copies of the template-rendering and step-result-writing helpers.
- [ ] All existing spec tests pass without modification and an end-to-end `spec new` + `spec goto` walk produces the same JSON output as before the refactor.

### Milestone 2: `spektacular implement` command works end-to-end against a valid plan

**What changes**: Users can run `go run . implement new --data '{"name":"<plan-name>"}'` against any existing plan under `.spektacular/plans/<name>/plan.md` and receive a JSON instruction telling them to read the plan. Subsequent `implement goto --data '{"step":"<id>"}'` calls advance through the full ten-step workflow with each rendered instruction containing the exact directives the spec's acceptance criteria require. The loop works: when `update_changelog` detects remaining unchecked phases in the plan file, it transitions back to `analyze` via a multi-source FSM edge. When no unchecked phases remain, the workflow transitions to `update_repo_changelog`, which appends a user-facing summary to the repo-level `CHANGELOG.md` using the plan name as the section header, and then to `finished`.

#### - [ ] Phase 2.1: Create `internal/steps/implement` package with types and step definitions

A new `internal/steps/implement` package is added with the same public surface shape as `internal/steps/plan`: a `Steps()` function returning ten `workflow.StepConfig` entries, `Result` and `StatusResult`/`StepEntry`/`StepsResult` output types, and a `PlanFilePath` helper. Step callbacks are thin wrappers that delegate to `stepkit` with a package-local path strategy. The FSM uses multi-source transitions on `analyze`.

*Technical detail:* [context.md#phase-21](./context.md#phase-21-create-internalimplement-package-with-types-and-step-definitions)

**Acceptance criteria**:

- [ ] `internal/steps/implement` exposes `Steps()`, `Result`, `StatusResult`, `StepEntry`, `StepsResult`, and `PlanFilePath` types/functions.
- [ ] The returned `Steps()` slice contains ten steps in the order documented in the spec's technical approach section (with `update_repo_changelog` inserted between `update_changelog` and `finished`).
- [ ] The `analyze` step lists both `read_plan` and `update_changelog` as source states; `update_repo_changelog` lists only `update_changelog` as its source; and `finished` lists only `update_repo_changelog` as its source.
- [ ] Each step callback delegates rendering to `stepkit`.

#### - [ ] Phase 2.2: Add implement step templates under `templates/steps/implement/`

Nine new markdown templates are added under `templates/steps/implement/`, one per non-initialization step, each using the standard mustache variables plus implement-specific ones. Each template contains the directives the spec's acceptance criteria require. Templates are picked up automatically by the existing embed directive.

*Technical detail:* [context.md#phase-22](./context.md#phase-22-add-implement-step-templates-under-templatesstepsimplement)

**Acceptance criteria**:

- [ ] Nine template files exist under `templates/steps/implement/`, one per non-initialization step.
- [ ] Every template contains a STOP-on-mismatch directive.
- [ ] The `read_plan` template directs a full plan file read with no offset or limit; validates the plan's structural shape (every `## ` section the plan scaffold requires is present, at least one `#### - [ ] Phase` heading exists, the `*Technical detail:*` link targets for every phase resolve to headings in context.md); detects whether a `## Changelog` section exists (first-phase if absent, subsequent-phase if present); and performs a drift check against the current codebase before any implementation work starts — for each file path, package path, and symbol named in plan.md or context.md, the template instructs the agent to verify the path/symbol still exists (via `ls`, `grep`, or reading the file). On any structural failure or unresolved drift, the template instructs the agent to STOP and report the mismatches to the user, asking whether to (a) fix the plan first, (b) proceed with the corrections in memory, or (c) abandon the workflow. The template must not silently continue past a drift detection.
- [ ] The `analyze` template references the `spawn-implementation-agents` skill via the `{{config.command}} skill <name>` pattern.
- [ ] The `test`, `verify`, and `update_changelog` templates each reference the corresponding new skill from Milestone 3 via the same pattern.
- [ ] The `update_changelog` template tells the agent to create a `## Changelog` section in `plan.md` on first invocation (appending after `## Out of Scope`) and append to it on subsequent invocations.
- [ ] The `update_changelog` template tells the agent to check for remaining unchecked phases in the plan and branch between `goto analyze` and `goto update_repo_changelog`, and tells the agent to ask the user first unless previously instructed otherwise.
- [ ] The `update_plan` template tells the agent to mark plan checkboxes complete.
- [ ] The `update_repo_changelog` template instructs the agent to append a new section to the repo-level `CHANGELOG.md` using `{{plan_name}}` as the section header and a short user-facing summary of the overall change, creating `CHANGELOG.md` at the repo root if it does not exist and prepending the new entry above any existing sections. The template ends by instructing the agent to `goto finished`.

#### - [ ] Phase 2.3: Wire `cmd/implement.go` Cobra commands

A new `cmd/implement.go` file adds the `implement` command group with `new`, `goto`, `status`, and `steps` subcommands mirroring `cmd/plan.go`'s structure. The root command registers the new group alongside `specCmd` and `planCmd`.

*Technical detail:* [context.md#phase-23](./context.md#phase-23-wire-cmdimplementgo-cobra-commands)

**Acceptance criteria**:

- [ ] `go run . implement --help` lists four subcommands: new, goto, status, steps.
- [ ] `go run . implement new --data '{"name":"<valid-plan>"}'` exits 0 and emits JSON with `step`, `plan_path`, `plan_name`, and a non-empty `instruction` field.
- [ ] `go run . implement new --data '{"name":"<invalid-name>"}'` exits non-zero with an error.
- [ ] `go run . implement goto --data '{"step":"<valid-step>"}'` advances the workflow and emits the rendered JSON for that step.
- [ ] `go run . implement status` returns a progress summary including an `unchecked_phases` count.
- [ ] `go run . implement steps` lists all ten step names.
- [ ] All four subcommands accept `--schema` and emit a non-empty schema JSON.

#### - [ ] Phase 2.4: Add implement tests covering FSM, templates, and commands

A new test suite for the implement workflow lands: step-ordering and FSM-walk tests; per-step template-content tests that assert each spec acceptance criterion has a corresponding passing substring check; a dedicated loop test; and Cobra command tests.

*Technical detail:* [context.md#phase-24](./context.md#phase-24-add-implement-tests-covering-fsm-templates-and-commands)

**Acceptance criteria**:

- [ ] `TestStepsOrderMatchesExpected` in `internal/steps/implement` passes and asserts the ten-step sequence.
- [ ] `TestFSMWalkFromNewToFinished` drives the workflow through the happy path in dry-run mode (including the `update_repo_changelog` step before `finished`).
- [ ] `TestFSMLoopFromUpdateChangelogBackToAnalyze` exercises the multi-source transition at least once and then reaches `update_repo_changelog` and `finished`.
- [ ] `TestUpdateRepoChangelogTemplateContainsDirectives` asserts the rendered `update_repo_changelog` template mentions `CHANGELOG.md`, the plan name variable, and a `goto finished` transition.
- [ ] `TestReadPlanTemplateDirectsStructuralValidation` asserts the rendered `read_plan` template tells the agent to verify every required `## ` section of the plan scaffold is present, at least one `#### - [ ] Phase` heading exists, and the `*Technical detail:*` context links resolve.
- [ ] `TestReadPlanTemplateDirectsDriftCheck` asserts the rendered `read_plan` template tells the agent to check every file path, package path, and symbol named in plan.md and context.md against the working tree before advancing, and to STOP on any mismatch with a three-option (fix / proceed / abandon) user prompt.
- [ ] Each of the spec's ten acceptance criteria (spec lines 44-65) has a corresponding passing unit test.
- [ ] Cobra command tests cover `new`, `goto`, `status`, `steps`, and `--schema` for each.
- [ ] `make test` and `make lint` pass with no regressions in `internal/steps/plan`, `internal/steps/spec`, or any other existing package.

### Milestone 3: Three new delegation skills ship and are referenced by implement templates

**What changes**: Three new spektacular-native skill markdown files are added under `templates/skills/` and referenced inline by the implement templates. Users invoking `go run . skill follow-test-patterns`, `go run . skill verify-implementation`, or `go run . skill update-changelog` receive the same JSON shape they get today from `skill spawn-implementation-agents`. No Go code changes.

#### - [ ] Phase 3.1: Add three new skill markdown files

Three new markdown files are added under `templates/skills/` following the exact frontmatter and body shape of the existing `skill_spawn-implementation-agents.md`. Each file owns a single narrow agent instruction matching the spec's corresponding delegation requirement. Files are picked up automatically by the existing embed directive.

*Technical detail:* [context.md#phase-31](./context.md#phase-31-add-three-new-skill-markdown-files)

**Acceptance criteria**:

- [ ] `go run . skill follow-test-patterns` returns a non-empty JSON object with `name`, `title`, and `instructions` fields populated.
- [ ] `go run . skill verify-implementation` returns a non-empty JSON object with the same shape.
- [ ] `go run . skill update-changelog` returns a non-empty JSON object with the same shape.
- [ ] Each skill's `instructions` body describes a single narrow sub-agent task matching the spec's corresponding delegation requirement.
- [ ] The implement templates from Phase 2.2 reference all three new skills by name via the `{{config.command}} skill <name>` pattern.

## Open Questions

- **Does the pinned `looplab/fsm` version support multi-source events the way the implement loop requires?** The research assumed it does based on the presence of the `Src []string` field in `workflow.StepConfig`, but no existing workflow in the repo actually exercises a multi-source transition, so this is an unverified assumption. The first moment it becomes load-bearing is Phase 2.1. **When hit**: As the first concrete action of Phase 2.1, write a tiny fixture test in `internal/workflow` that declares three steps with a multi-source edge and asserts that `Next` and `Goto` transition correctly from both sources. If the test passes, proceed. If the test fails, STOP and ask the user — the fallback is to re-encode the loop imperatively inside the `update_changelog` callback, which means `analyze.Src` stays `[]string{"read_plan"}`.

- **Whether `internal/steps/spec` has behavior-preservation test coverage equivalent to `internal/steps/plan`'s.** The research identified `internal/steps/plan/steps_test.go` as the safety net for Phase 1.2 but did not inventory the test coverage of `internal/steps/spec`. If `internal/steps/spec` has no rendered-instruction tests today, Phase 1.3's refactor lands without a regression fence. **When hit**: At the start of Phase 1.3, before touching any spec source file, run `go test ./internal/steps/spec/... -v` and inspect the test names. If there are rendered-instruction tests comparable to `TestArchitectureStepContainsOptionsAndAgreementBeat` in the plan package, proceed. If there are not, STOP and ask the user whether to add minimal pre-refactor tests before migrating — this is a short safety net against a silent behavior change.

## Out of Scope

- **No new plan-generation capability.** Plan generation remains the job of the existing `plan` workflow. (Spec § Non-Goals.)

- **No enforcement of a specific test framework or verification command set.** The implement workflow's `test` and `verify` steps delegate to sub-agents that run whatever the plan's own acceptance criteria dictate. (Spec § Non-Goals.)

- **No git commit, branch, or pull-request management.** The implement workflow produces changes on the working tree and appends to the plan's inline `## Changelog` section but never runs any git operation. Committing stays the user's responsibility. (Spec § Non-Goals.)

- **No per-phase state persisted outside the plan file.** The workflow's `state.json` records only the plan's name. Phase progress is re-derived every call by reading the plan file's `#### - [ ]` checkbox state. (Spec § Non-Goals.)

- **No new sub-agent delegation mechanism.** Implement templates reference existing skills via the same `{{config.command}} skill <name>` pattern that plan step templates already use. (Spec § Non-Goals.)

- **No harbor end-to-end integration test for the implement workflow.** Adding a harbor job for implement is an expensive effort-to-value trade and deserves its own plan. Until then, coverage comes from FSM walk tests, template content tests, and Cobra command tests.

- **No refactor or test addition for `cmd/plan.go` or `cmd/spec.go`.** Milestone 1 touches `internal/steps/plan` and `internal/steps/spec`, not their Cobra command wiring.

- **No fix for any latent `next_step` mismatch bugs in `internal/steps/plan/steps.go`.** During the original planning pass it was noticed that a step callback passed the wrong `nextStep` value to its template — a drift-prone hazard of the hand-maintained step list. If any such mismatch is still present when implementation begins, fix it in a separate one-line commit; do not bundle it into this refactor.

- **No changes to `templates/scaffold/plan.md`.** The scaffold is treated as a stable contract; evolving it is a separate plan.
