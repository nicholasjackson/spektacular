# Research: 15_implementation

> **Note on drift:** This research was conducted against an earlier codebase layout. Paths have been corrected (`internal/steps/plan`, `internal/steps/spec`, `templates/steps/plan/`, `templates/steps/spec/`, `templates/scaffold/plan.md`, `templates/skills/`), but **line numbers throughout are stale** and the state-model claims below (per-workflow directories, `findActivePlan`) describe an older architecture. The current architecture uses a **single shared `.spektacular/state.json`** with no per-workflow subdirectory; `findActivePlan`, `planDataDir`, and the `filepath.Join(dataDir, "..")` store-root trick no longer exist. See plan.md and context.md for the corrected architectural picture.

## Alternatives considered and rejected

### Option A: Pure mirror of the `plan` package with scoped duplication

**Description**: Create `cmd/implement.go`, `internal/steps/implement/`, and `templates/steps/implement/` as near-perfect copies of their `plan` counterparts, copy-pasting the ~60 lines of helper code (`writeStepResult`, `renderTemplate`, `getString`, `stepTitle`) into the new `internal/steps/implement/steps.go` without extracting a shared package. Leaves `internal/steps/plan` and `internal/steps/spec` untouched at the helper-code level.

**Rejected**: Leaves the helpers duplicated across three packages, and the user has a standing preference for DRY extraction when a new feature would copy >~30 lines of adjacent-package utilities (recorded in `~/.claude/projects/-home-nicj-code-github-com-jumppad-labs-spektacular/memory/feedback_dry_refactor_preference.md`). Cited evidence: the duplicated helper bundle at `internal/steps/plan/steps.go:165-229` would gain a third copy after this plan, putting three concrete call sites in play and tripping the "rule of three" threshold for abstraction. A minimal-diff option was presented and the user explicitly chose the refactor path instead. See also: the `internal/steps/spec/steps.go` package already contains near-verbatim copies of the same helpers, which means the duplication is two-wide today and the implement work would make it three-wide without intervention.

### Option B: Script-inlined instructions (no templates)

**Description**: Drop the `templates/steps/implement/*.md` pattern entirely. Put instruction strings directly in `internal/steps/implement/steps.go` as Go string literals. Each step's callback returns the full instruction text as a hard-coded string.

**Rejected**: Explicitly forbidden by the spec's constraint at `.spektacular/specs/15_implementation.md:36`: "Must follow the existing spec-workflow pattern: markdown templates under `templates/`, mustache rendering, stateless JSON responses per step." This option also loses the syntax highlighting and markdown preview benefits of external templates (the existing plan-steps templates are 40-50 lines each, and embedding them as Go string literals would make them painful to author and review). A non-starter.

### Option C: Adjacent `changelog.md` file instead of inline `## Changelog` section

**Description**: Write changelog entries to a separate `changelog.md` file in the plan directory rather than appending to an inline `## Changelog` section of `plan.md`. This matches the spec's original wording at `.spektacular/specs/15_implementation.md:14`.

**Rejected**: Writing to a separate file would split the plan's history across two documents and force reviewers to open both to see the full story. Keeping the changelog inline in `plan.md` means the plan's author, reviewer, and implement agent all read from (and write to) a single file. A spec-wording deviation is the right call here because the benefit of file locality outweighs the ergonomic cost of a larger `plan.md`. The implement workflow owns the `## Changelog` section entirely: it creates the section on its first `update_changelog` invocation (appended after `## Out of Scope`) and appends subsequent phase entries under it. The plan workflow and its scaffold have no involvement — the scaffold ends at `## Out of Scope` and no plan-workflow step template, Go file, or command mentions changelog. This "implement owns it end-to-end" model was confirmed by a repo-wide grep showing zero changelog references in `templates/steps/plan/`, `internal/steps/plan/`, or `cmd/plan.go` after the scaffold's `## Changelog` section was removed.

### Option D: Imperative loop logic inside the `update_changelog` callback

**Description**: Instead of using multi-source FSM transitions, make the `update_changelog` callback compute whether unchecked phases remain and return either `"analyze"` or `"finished"` as the next step. `analyze.Src` stays `[]string{"read_plan"}` (single source).

**Rejected** as the primary design, **retained** as a fallback. This option works but pushes the loop logic into Go rather than the FSM declaration, making the loop structure less visible at the data-structure level. The multi-source transition via `analyze.Src = []string{"read_plan", "update_changelog"}` is more declarative and gives the FSM engine a chance to validate transitions statically. However, multi-source transitions are unused elsewhere in the codebase today, so there's a risk the pinned `looplab/fsm` version doesn't support them — captured in Open Questions. If the verification fails, fall back to this option.

### Option E: Cross-package import of `internal/steps/plan.PlanFilePath` from `internal/steps/implement`

**Description**: Instead of copying the `PlanFilePath` constant into `internal/steps/implement/plan_path.go`, have the implement package import `internal/steps/plan` and call `plan.PlanFilePath` directly.

**Rejected**: Creates a package dependency between `internal/steps/implement` and `internal/steps/plan` for a single 10-line constant function. The direction of the dependency is correct (implement operates on plans, so depending on plan is natural) but introduces coupling where none is needed — the function is pure and its logic is unlikely to change. Copying it is the cheaper move. If both packages' path conventions ever need to stay in sync (they might not — implement doesn't care about context.md or research.md), a follow-up plan can consolidate them.

### Option F: ~~Extending `findActivePlan` to also find implement state dirs~~ (no longer applicable)

**Original description**: Instead of adding a separate `findActiveImplement` function, modify `cmd/plan.go:findActivePlan` to scan for both `plan-*` and `implement-*` directories.

**No longer applicable**: The current codebase has no `findActivePlan` helper and no per-workflow data directory — spec, plan, and (future) implement all share a single `.spektacular/state.json`. There is nothing to "find" because there is only one active workflow at a time. The implement command should follow the same pattern: read `name` from the shared state file via `wf.GetData("name")`, as `runPlanGoto` and `runSpecGoto` already do.

## Chosen approach — evidence

The chosen approach — a shared `internal/stepkit` package extraction plus an `internal/steps/implement` mirror of plan with multi-source FSM loop — is supported by the following evidence:

- **`internal/steps/plan/steps.go:165-202`** — The `writeStepResult` helper bundle to be extracted. Clean, self-contained, 38 lines. Takes `data workflow.Data`, `out workflow.ResultWriter`, `st store.Store`, `cfg workflow.Config`, and variadic `extra ...map[string]any`. Assembles standard + extra vars, renders mustache, writes a Result. The variadic `extra` pattern is preserved in the new `StepRequest.Extra` field.

- **`internal/steps/plan/steps.go:204-229`** — The small helper bundle (`getString`, `renderTemplate`, `stepTitle`) — another 26 lines. All three are pure functions with no state, making the extraction trivially behavior-preserving.

- **`internal/steps/spec/steps.go`** (package body inspected by research agent, not quoted in full here) — Contains near-verbatim copies of the same four helpers. Extracting once now means rewriting two packages; waiting until later would mean rewriting three.

- **`workflow.StepConfig`** at `internal/workflow/workflow.go` — Already has `Src []string` as the source field, not a single string. The FSM engine threads this through `fsm.EventDesc{Name, Src, Dst}` per the research agent's read of lines 94-113, and the underlying `looplab/fsm` library has supported multi-source events since its early versions. The first real-world usage in this codebase will be the `analyze.Src = []string{"read_plan", "update_changelog"}` declaration in Phase 2.1 — a point where the assumption becomes load-bearing and needs a one-test verification (see Open Questions).

- **`templates/steps/plan/02-discovery.md:9-22`** — The canonical pattern for referencing skills in a rendered template: `` `{{config.command}} skill <name>` ``. Used twice on lines 9-10 for `discover-project-commands` and `discover-test-patterns`, and once on line 22 for `spawn-planning-agents`. The implement templates follow this pattern for every skill reference.

- **`templates/steps/plan/10-phases.md:30`** — Second example of the same pattern, this time for `spawn-implementation-agents`. Confirms the pattern is stable across templates.

- **`templates/scaffold/plan.md`** — Ends at `## Out of Scope`. Contains no `## Changelog` section and no reference to changelog anywhere. The plan workflow never instructs an agent to draft or include changelog content; the implement workflow creates the section in each plan's `plan.md` at runtime on its first `update_changelog` invocation.

- **`cmd/plan.go`** — The store-root construction today: `store.NewFileStore(dataDir)` where `dataDir` is `.spektacular/`. The same construction applies to implement: `cmd/implement.go:runImplementNew` uses `store.NewFileStore(dataDir)` so the implement workflow reads plan files via the shared `.spektacular/` root at `plans/<name>/plan.md`. (Historical note: an earlier version used `filepath.Join(dataDir, "..")` with per-workflow subdirectories; that model no longer exists.)

- **`cmd/plan.go`** — No `findActivePlan` helper exists in the current codebase. `runPlanGoto` reads the active name from the shared state file via `wf.GetData("name")` and errors with "no active plan found — run 'plan new' first" if the key is absent. `cmd/implement.go` should follow the same pattern.

- **`internal/steps/plan/steps_test.go:100-136`** — `TestFSMWalkFromNewToFinished` uses `workflow.Config{DryRun: true}` to drive the FSM forward without persisting state, asserting `wf.Current()` at each transition. The implement workflow's `TestFSMWalkFromNewToFinished` test is a direct copy of this pattern with different expected step names. The `TestFSMLoopFromUpdateChangelogBackToAnalyze` test extends the pattern with a `wf.Goto("analyze")` call after reaching `update_changelog`.

- **`internal/steps/plan/steps_test.go:13-43`** — The `testData`/`captureWriter`/`renderStep` fake trio. Copy-pasted into each test package per `thoughts/notes/testing.md`'s documented convention. The implement test file reuses this pattern.

- **`templates/templates.go` `//go:embed all:*` directive** — Covers every file and subdirectory under `templates/`. Adding `templates/steps/implement/` requires no code changes. Confirmed by research agent 1 in its report.

- **`templates/skills/` `//go:embed all:*.md` directive** — Confirmed by research agent 2. Adding three new `skill_*.md` files requires no Go changes; they're picked up at compile time.

- **`.spektacular/specs/15_implementation.md:44-65`** — The ten acceptance criteria that drive the per-step substring test assertions in Phase 2.4. Each criterion maps 1:1 to a test in `internal/steps/implement/steps_test.go`.

## Files examined

- `.spektacular/specs/15_implementation.md:1-111` — Source spec driving the plan. Requirements, constraints, acceptance criteria, technical approach, non-goals.
- `.spektacular/plans/16_plan_format/plan.md` — Prior plan that landed the current plan-scaffold format including the inline `## Changelog` section.
- `Makefile:1-35` — Build/test/lint/cross/harbor-test targets. No `implement`-specific target needed — `make test` and `make lint` cover it.
- `CLAUDE.md` and `AGENTS.md` — Project instructions; pointed at `.tessl/RULES.md` which was empty or informational only.
- `cmd/plan.go` — Cobra command wiring pattern to mirror in `cmd/implement.go`. Sections (re-locate by name at implementation time): command declarations, `runPlanNew`, `runPlanGoto`, `runPlanStatus`, `runPlanSteps`, `init`. No `findActivePlan` helper. Uses `readInputIntoWorkflow(cmd, wf)` for `--stdin`/`--file` plumbing.
- `cmd/spec.go` — Peer Cobra command for comparison. Structurally identical to plan; both use `store.NewFileStore(dataDir)` against the same shared `.spektacular/` data dir.
- `cmd/root.go` — Config loading pattern via `loadConfig()`. Used unchanged by the new `cmd/implement.go`.
- `internal/steps/plan/steps.go:1-229` — The package being refactored. Key line references: 14-27 (path helpers), 29-48 (`Steps`), 51-55 (`new` — auto-advance), 57-163 (step callbacks — one-liners wrapping `writeStepResult`), 129-151 (`verification` — renders scaffolds), 153-163 (`finished` — file existence check), 165-202 (private `writeStepResult`), 204-229 (small helpers to extract).
- `internal/steps/plan/result.go` — `Result`, `StepEntry`, `StatusResult`, `StepsResult` type declarations. `internal/steps/implement/result.go` mirrors this file.
- `internal/steps/plan/steps_test.go:1-137` — Test patterns to mirror. Key line references: 13-43 (fake trio + `renderStep`), 45-73 (per-step substring assertions), 75-98 (`TestStepsOrderMatchesExpected`), 100-136 (`TestFSMWalkFromNewToFinished`).
- `internal/steps/plan/scaffold_test.go:1-53` — Template rendering test pattern (mustache render + heading order assertion). Not mirrored by implement tests — implement has no own scaffold file.
- `internal/steps/spec/steps.go` (full file inspected by research agent, not directly read here) — Mirrors plan's shape with spec-specific path helpers. Contains its own private copies of the helpers being extracted.
- `internal/workflow/workflow.go` — `Config`, `StepConfig`, `Data`, `ResultWriter`, `Workflow` types. Key research finding: `StepConfig.Src []string` supports multi-source events; the FSM engine threads this to `fsm.EventDesc`.
- `internal/workflow/workflow_test.go:1-174` — FSM engine tests. `TestGotoForward` (line 59), `TestCompletedStepsTracked` (line 163), `TestNextStepName` (line 128) are the relevant patterns for the implement loop test.
- `internal/store/store.go` — `Store` interface and `FileStore`. Path traversal rejection via `abs()`.
- `internal/store/store_test.go:1-96` — Store test patterns.
- `templates/steps/plan/02-discovery.md:1-46` — Canonical skill-reference template. Quoted extensively in the Implementation Detail section.
- `templates/steps/plan/10-phases.md:1-41` — Second skill-reference example.
- `templates/scaffold/plan.md:1-146` — Plan scaffold template. Line 144-146 is the inline `## Changelog` contract point.
- `templates/steps/plan/` directory listing — Confirmed 14 files (01-overview.md through 14-finished.md) for the plan workflow. Implement will have 8 templates (01-read_plan.md through 08-finished.md — the `new` step has no template).
- `templates/skills/skill_spawn-implementation-agents.md` — Reference skill file format. The three new skill files mirror its frontmatter + body structure.
- `templates/skills/skill_spawn-planning-agents.md` — Second reference skill file. Confirmed format stability.
- `templates/skills/skill_discover-test-patterns.md` — Reference for small content-only skill files.
- `templates/skills/skill_discover-project-commands.md` — Reference for small content-only skill files.
- `thoughts/notes/commands.md` — Created at the start of this planning session documenting `make` targets and `go run .` commands.
- `thoughts/notes/testing.md` — Created at the start of this planning session documenting `stretchr/testify/require` + `t.TempDir()` pattern and the copy-paste-per-package helper convention.
- `.spektacular/plans/15_implementation/` directory listing — Confirmed empty before this planning session; `state.json` was created when `plan new` ran.

## External references

None. This plan is entirely internal to the spektacular repository — no external libraries are added, no external APIs are consulted, no RFCs or academic papers drive the design. The only external dependency touched is `github.com/cbroglie/mustache`, which is already pinned and already used by `internal/steps/plan` and `internal/steps/spec`.

## Prior plans / specs consulted

- **`.spektacular/specs/15_implementation.md`** — The source spec. Drives every requirement and acceptance criterion in the plan. Read in full during the `overview` step; referenced throughout subsequent steps when deriving test cases and template directives.
- **`.spektacular/plans/16_plan_format/plan.md`** — Prior plan that established the current plan scaffold format. Learned: the current scaffold emits an inline `## Changelog` section at line 144-146 of `templates/scaffold/plan.md`, which is reserved for the implement workflow. This informed the inline-vs-adjacent changelog decision (see Alternatives § Option C).
- **No other spektacular plans consulted.** The `.spektacular/plans/` directory contains a handful of earlier plans that are either superseded or not relevant to workflow infrastructure.

## Open assumptions

- **`looplab/fsm` supports multi-source events in the pinned version.** Assumed based on the presence of `Src []string` in `workflow.StepConfig` and the research agent's read of `internal/workflow/workflow.go`. Unverified because no existing workflow uses multi-source transitions in anger. **If wrong**: the implement workflow must fall back to the imperative loop in Alternative Option D. **Verification**: Phase 2.1's first action writes a three-step fixture test in `internal/workflow/` that asserts multi-source transitions work.

- **`internal/steps/spec` has behavior-preservation test coverage equivalent to `internal/steps/plan`.** Assumed because both packages look structurally identical, but not directly confirmed during research. **If wrong**: Phase 1.3 lands without a regression fence for spec. **Verification**: Phase 1.3's first action runs `go test ./internal/steps/spec/... -v` and inspects the test names. If no rendered-instruction tests exist, STOP and ask the user whether to add them before the refactor.

- **`templates/scaffold/plan.md` stays changelog-free through to implementation.** The scaffold was edited during this planning session to remove its pre-reserved `## Changelog` section. If a concurrent plan reintroduces any changelog reference to the scaffold or to `templates/steps/plan/` before this plan is implemented, the "implement owns changelog end-to-end" design assumption breaks. **If wrong**: STOP at the start of Phase 2.2 (implement templates) and ask the user how to reconcile the ownership model. **Verification**: a cheap `grep -rni changelog templates/steps/plan/ templates/scaffold/plan.md internal/steps/plan/ cmd/plan.go` at the start of Phase 2.2 catches any regression — zero matches expected.

- **`templates.FS` via `//go:embed all:*` covers new subdirectories automatically.** Research agent 1 confirmed this by reading `templates/templates.go` directly. High confidence; included here for completeness.

- **`templates/skills` via `//go:embed all:*.md` covers new markdown files automatically.** Research agent 2 confirmed this. High confidence.

- **The existing `internal/steps/plan/steps_test.go` tests are tight enough to catch byte-level changes in rendered instructions.** The tests assert on specific substrings (`"option"`, `"agreement"`, `"high-level"`, `"context.md"`, etc.) rather than on full instruction equality, so small whitespace changes would pass silently. **If wrong** (i.e. the refactor introduces a subtle whitespace change that breaks the rendered output in a way the substring tests miss): the bug would surface during manual end-to-end testing in Phase 1.2's acceptance criteria ("Running `go run . plan new` and `go run . plan goto` produces the same JSON output as before"). Acceptable mitigation.

- **The `{{config.command}}` template variable is always set to the same literal value at runtime.** Confirmed: `cmd/plan.go:133` sets `wfCfg := workflow.Config{Command: cfg.Command, …}` from `loadConfig()`, which returns the config's `Command` field. Tests use `workflow.Config{Command: "spektacular"}` (see `internal/steps/plan/steps_test.go:40`). Stable assumption.

## Rehydration cues

If context is lost mid-implementation, regenerate it as follows:

1. **Read the spec**: `.spektacular/specs/15_implementation.md` in full. This is the source of truth for requirements and acceptance criteria. No substitute.

2. **Read this plan**: `.spektacular/plans/15_implementation/plan.md` in full. No offset, no limit. The `## Architecture & Design Decisions` section is the load-bearing one; the rest flows from it.

3. **Read the per-phase context**: `.spektacular/plans/15_implementation/context.md` at the section matching the current phase (e.g. `### Phase 1.1: …`). Each phase section contains the file:line changes and agent strategy.

4. **Read the reference helpers**: `internal/steps/plan/steps.go:165-229` to remind yourself of the exact helper bundle being extracted. This is the master copy; the new `internal/stepkit/stepkit.go` is a near-verbatim lift with exported names.

5. **Read the pattern references**: `cmd/plan.go:1-319` (command wiring to mirror), `internal/steps/plan/steps.go:14-48` (path helpers and `Steps()` pattern), `internal/steps/plan/steps_test.go:13-136` (test patterns to mirror), `templates/steps/plan/02-discovery.md` (skill-reference pattern for templates).

6. **Read the scaffold's end**: `templates/scaffold/plan.md` ends at `## Out of Scope` and has no `## Changelog` section. The implement workflow creates the section inline in each plan's `plan.md` on its first `update_changelog` invocation.

7. **Re-run project-context skills if `thoughts/notes/` files are missing**:
   - `go run . skill discover-project-commands` → writes `thoughts/notes/commands.md`
   - `go run . skill discover-test-patterns` → writes `thoughts/notes/testing.md`

8. **For agent orchestration guidance during implementation**:
   - `go run . skill spawn-implementation-agents` — how to delegate codebase work to parallel sub-agents by complexity tier.
   - `go run . skill spawn-planning-agents` — if you find yourself needing more research mid-implementation.

9. **Verify the multi-source FSM assumption at the start of Phase 2.1**: write a fixture test in `internal/workflow/` with three steps and a multi-source edge. If it fails, STOP and re-read `## Open Questions` in plan.md for the fallback.

10. **When updating the plan's changelog during implementation**, the format lives in `go run . skill update-changelog` (once Phase 3.1 has landed) or, before that, in the `## Per-Phase Technical Notes § Phase 3.1` section of context.md.
