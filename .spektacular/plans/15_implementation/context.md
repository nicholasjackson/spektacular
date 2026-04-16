# Context: 15_implementation

> **Note on drift:** The original research for this plan was written against an earlier codebase layout. The package paths (`internal/steps/plan`, `internal/steps/spec`), template paths (`templates/steps/plan`, `templates/steps/spec`, `templates/scaffold/plan.md`), and skill location (`templates/skills/`) have been corrected, but **line numbers and line ranges throughout this file are stale** and should be treated as approximate pointers. Re-discover precise locations via `grep`/`Read` at the start of each phase rather than trusting the numeric ranges below. The step list in `internal/steps/plan/steps.go` now has 18 entries (not 15) — the extra steps are `write_plan`, `write_context`, `write_research` which take the filled-in plan/context/research contents from `--file` or `--stdin` and persist them via the store.
>
> **State model drift:** When this plan was written, `plan` and `implement` were each expected to own a per-workflow data directory (`.spektacular/plan-<name>/`, `.spektacular/implement-<name>/`) with its own `state.json`. The current codebase uses a **single shared `.spektacular/state.json`** that is truncated on every `<workflow> new`. The `findActivePlan`/`findActiveImplement` discovery helpers and `planDataDir`/`implementDataDir` builders referenced below no longer exist (and the implement workflow should not re-introduce them). Every workflow reads and writes the same state file, same store root, same data dir.

## Current State Analysis

The repository has two working workflow commands — `spec` and `plan` — that share the same architectural shape and still each own a private copy of the step-rendering helper bundle:

- **`cmd/spec.go`** — Cobra command group with `new`/`goto`/`status`/`steps` subcommands. Uses `store.NewFileStore(dataDir)` as the store root where `dataDir` is `.spektacular/`. State file is `.spektacular/state.json`.
- **`cmd/plan.go`** — Cobra command group with the same subcommand shape and the same store-root construction. Same shared `.spektacular/state.json`. No per-plan subdirectory.
- Neither `cmd/spec.go` nor `cmd/plan.go` has an "active workflow discovery" helper — there is only one state file and `plan new` / `spec new` truncate it outright (`os.Remove(statePath)` before rewriting).
- **`internal/steps/plan/steps.go` — path helpers** — `PlanFilePath`, `ContextFilePath`, `ResearchFilePath` return store-relative paths like `plans/<name>/plan.md`, `plans/<name>/context.md`, `plans/<name>/research.md`.
- **`internal/steps/plan/steps.go` — `Steps()`** — Returns 18 `workflow.StepConfig` entries in this order: `new → overview → discovery → architecture → components → data_structures → implementation_detail → dependencies → testing_approach → milestones → phases → open_questions → out_of_scope → verification → write_plan → write_context → write_research → finished`. Every entry uses a single-source `Src` slice today; no existing workflow exercises multi-source transitions.
- **`internal/steps/plan/steps.go` — `new()` callback** — Auto-advances to `overview` without writing anything. Contrast with spec's `new()` which writes the spec scaffold.
- **`internal/steps/plan/steps.go` — step callbacks** — Most are one-liners calling `writeStepResult(name, nextStep, templatePath, data, out, st, cfg, extra...)`. Exceptions: `verification()` renders the plan/context/research scaffolds (`templates/scaffold/plan.md`, `templates/scaffold/context.md`, `templates/scaffold/research.md`) and passes them to the template as `plan_template`, `context_template`, `research_template` extra vars. `writePlan()`, `writeContext()`, `writeResearch()` pull their key out of workflow data (populated via `--file` or `--stdin`) and call `st.Write(...)` to persist the filled document.
- **`internal/steps/plan/steps.go` — private helper bundle** — `writeStepResult` (assembles standard template vars, renders via mustache, emits a `Result`), `getString` (safe `workflow.Data` string lookup), `renderTemplate` (mustache render from `templates.FS`), `stepTitle` (snake_case → Title Case). These four are the extraction target for Phase 1.1.
- **`internal/steps/plan/result.go`** — Declares `Result`, `StepEntry`, `StatusResult`, `StepsResult`.
- No `scaffold.go` file exists in `internal/steps/plan/` — scaffolding happens inline inside `verification()`.
- **`internal/steps/spec/steps.go`** — Mirrors plan's shape but with its own private copies of `writeStepResult` / `renderTemplate` / `getString` / `stepTitle`. The spec `new()` callback writes the spec scaffold into the store via `st.Write(SpecFilePath(name), ...)` as its side effect. The spec step file does not currently have a `_test.go` sibling (see Open Question in plan.md).
- **`cmd/spec.go` + `cmd/plan.go` — shared `--file`/`--stdin` input plumbing** — Both commands use `readInputIntoWorkflow(cmd, wf)` to pull a workflow-data key from either stdin or a relative file path (basename-minus-extension becomes the key). The `write_plan` / `write_context` / `write_research` steps rely on this to receive their rendered content.
- **`internal/workflow/workflow.go`** — Defines `Config{Command, DryRun}`, `StepConfig{Name, Src, Dst, Callback}`, `Data`, `ResultWriter`, `Workflow`. `New()` wires each step's callback into an `after_<stepName>` hook and registers an `enter_state` callback for auto-save and completion tracking. Persistence is JSON to the single shared state file unless `DryRun` is set. `Next()` fires the first available FSM transition; if the callback returns a non-empty next step, `Goto()` is called recursively to advance further. `StepConfig.Src` is already typed as `[]string` — multi-source support is the plan assumption that needs the fixture test in Phase 2.1.
- **`internal/store/store.go`** — `Store` interface and `FileStore` struct. All methods validate paths against `..` traversal. `FileStore.Root()` returns the absolute store root.
- **`templates/templates.go`** — `//go:embed all:*` covers every file and subdirectory in `templates/`. No changes needed to add `templates/steps/implement/`.
- **`templates/scaffold/plan.md`** — The plan scaffold ends at `## Out of Scope`. Has no `## Changelog` section and no mention of changelog anywhere. The implement workflow creates the `## Changelog` section directly in each plan's `plan.md` at runtime on its first `update_changelog` invocation.
- **`templates/steps/plan/02-discovery.md`** — The canonical pattern for referencing skills in a template: `` `{{config.command}} skill <name>` ``.
- **`templates/steps/plan/10-phases.md`** — Same pattern, for `spawn-implementation-agents`.
- **`templates/skills/skill_spawn-implementation-agents.md`** — Reference skill file showing frontmatter and body shape. Loaded at runtime through `cmd/skill.go` from an embedded FS rooted at `templates/skills/`.
- **`internal/steps/plan/steps_test.go`** — Contains the `renderStep` helper pattern (a tiny `testData`/`captureWriter` pair satisfying `workflow.Data`/`workflow.ResultWriter`), the per-step substring assertion tests, `TestStepsOrderMatchesExpected`, and `TestFSMWalkFromNewToFinished` (which drives a real `workflow.Workflow` with `DryRun: true` through every step sequentially and asserts `wf.Current()` at each transition). This is the behavior-preservation fence for Phase 1.2.
- **`internal/steps/plan/scaffold_test.go`** — Template rendering test that reads the scaffold via `templates.FS.ReadFile`, renders with `mustache.Render`, and asserts heading presence and order.
- **`thoughts/notes/commands.md`** — Project command reference (re-run `go run . skill discover-project-commands` if missing).
- **`thoughts/notes/testing.md`** — Testing patterns reference (re-run `go run . skill discover-test-patterns` if missing).

A **latent bug** was noticed during original discovery: one of the plan step callbacks passes a wrong `nextStep` string to its template (the original finding was `discovery()` passing `"approach"` when the real next step is `architecture`). Whether this specific bug is still present or has been fixed in a prior commit is unknown after drift — re-verify with a quick grep at implementation time. This is out of scope for this plan — if still present, fix it in a separate one-line commit.

## Per-Phase Technical Notes

### Phase 1.1: Extract step-rendering helpers into `internal/stepkit`

**File changes**:

- `internal/stepkit/stepkit.go` (new, ~150 lines) — Contains:
  - `PathStrategy` interface with `NameKey() string`, `PathVars(name, storeRoot string) map[string]any`, `PrimaryPathField() string`.
  - `StepRequest` struct with fields `StepName`, `NextStep`, `TemplatePath`, `Strategy`, `Extra`.
  - `WriteStepResult(req StepRequest, data workflow.Data, out workflow.ResultWriter, st store.Store, cfg workflow.Config, resultBuilder func(name, primaryPath, instruction string) any) error` — the primary entry point. Assembles the standard template-variable map (`step`, `title`, `next_step`, `config.command`), calls `req.Strategy.PathVars(name, st.Root())` and merges the result, merges `req.Extra`, renders the template via `RenderTemplate`, invokes `resultBuilder(req.StepName, vars[req.Strategy.PrimaryPathField()].(string), instruction)`, and calls `out.WriteResult` on the returned value.
  - `StepTitle(snake string) string` — lifted verbatim from `internal/steps/plan/steps.go:221-229`; exported.
  - `GetString(data workflow.Data, key string) string` — lifted verbatim from `internal/steps/plan/steps.go:204-211`; exported.
  - `RenderTemplate(templatePath string, data map[string]any) (string, error)` — lifted verbatim from `internal/steps/plan/steps.go:213-219`; exported. Imports `github.com/cbroglie/mustache` and `github.com/jumppad-labs/spektacular/templates`.

- `internal/stepkit/stepkit_test.go` (new, ~180 lines) — Unit tests covering:
  - `TestStepTitle_SnakeCase` — `"data_structures"` → `"Data Structures"`, `"overview"` → `"Overview"`, `""` → `""`, `"a"` → `"A"`, `"multi_word_snake"` → `"Multi Word Snake"`.
  - `TestRenderTemplate_Success` — renders an existing template (e.g. `steps/plan/01-overview.md`) with known vars and asserts substring presence.
  - `TestRenderTemplate_MissingTemplate` — returns a wrapped `"loading template"` error for a non-existent path.
  - `TestWriteStepResult_StandardVars` — a test `PathStrategy` implementation returns a known `PathVars` map; the test asserts the rendered instruction reflects both standard and strategy vars.
  - `TestWriteStepResult_ExtraOverridesStrategy` — the `Extra` map takes precedence over strategy vars for the same key.
  - `TestWriteStepResult_ResultBuilderInvoked` — a recording `resultBuilder` captures the arguments; the test asserts they match.
  - `TestWriteStepResult_TemplateError` — missing template path propagates the wrapped error.
  - Uses `t.TempDir()` + `store.NewFileStore` per the pattern in `internal/steps/plan/steps_test.go:35`.

**Complexity**: Low

**Token estimate**: ~8k

**Agent strategy**: Single agent, sequential. Write `stepkit.go` first, then `stepkit_test.go`, run `go test ./internal/stepkit/...`, iterate. No cross-package changes in this phase.

### Phase 1.2: Rewrite `internal/steps/plan` to use `stepkit`

**File changes**:

- `internal/steps/plan/steps.go` (modified):
  - **Delete** lines 165-202 (private `writeStepResult`), 204-211 (`getString`), 213-219 (`renderTemplate`), 221-229 (`stepTitle`).
  - **Add** an import of `github.com/jumppad-labs/spektacular/internal/stepkit`.
  - **Remove** the imports that become unused after the deletion (`path/filepath`, `strings`, `github.com/cbroglie/mustache`, `github.com/jumppad-labs/spektacular/templates`) — but keep `filepath` if any remaining code needs it (the `new()` and `finished()` steps use `st.Exists` with store-relative paths that don't need `filepath`).
  - **Modify** each step callback at lines 57-163 to delegate to `stepkit.WriteStepResult`. Pattern for a typical one-liner (replacing `overview()`):
    ```go
    func overview() workflow.StepCallback {
        return func(data workflow.Data, out workflow.ResultWriter, st store.Store, cfg workflow.Config) (string, error) {
            return "", stepkit.WriteStepResult(
                stepkit.StepRequest{
                    StepName:     "overview",
                    NextStep:     "discovery",
                    TemplatePath: "steps/plan/01-overview.md",
                    Strategy:     planStrategy{},
                },
                data, out, st, cfg,
                func(name, primaryPath, instruction string) any {
                    return Result{
                        Step:        name,
                        PlanPath:    primaryPath,
                        PlanName:    stepkit.GetString(data, "name"),
                        Instruction: instruction,
                    }
                },
            )
        }
    }
    ```
  - **Modify** the `verification()` callback at lines 129-151 to continue rendering the three scaffold templates via `stepkit.RenderTemplate` and pass them through `StepRequest.Extra`:
    ```go
    Extra: map[string]any{
        "plan_template":     planScaffold,
        "context_template":  contextScaffold,
        "research_template": researchScaffold,
    },
    ```
  - **Modify** the `finished()` callback at lines 153-163 to keep its three-file existence check (the loop over `PlanFilePath`/`ContextFilePath`/`ResearchFilePath` calling `st.Exists`) before calling `stepkit.WriteStepResult`.
  - **Keep** the `new()` callback at lines 51-55 as a simple auto-advance (no stepkit call, no template rendering).
  - **Keep** `PlanFilePath`, `ContextFilePath`, `ResearchFilePath` (lines 14-27) exported for cross-package callers.
  - **Keep** `Steps()` (lines 29-48) unchanged.

- `internal/steps/plan/path_strategy.go` (new, ~30 lines):
  ```go
  package plan

  import (
      "path/filepath"
      "github.com/jumppad-labs/spektacular/internal/stepkit"
  )

  type planStrategy struct{}

  func (planStrategy) NameKey() string            { return "name" }
  func (planStrategy) PrimaryPathField() string   { return "plan_path" }
  func (planStrategy) PathVars(name, root string) map[string]any {
      planPath := filepath.Join(root, PlanFilePath(name))
      contextPath := filepath.Join(root, ContextFilePath(name))
      researchPath := filepath.Join(root, ResearchFilePath(name))
      specPath := filepath.Join(root, "specs", name+".md")
      return map[string]any{
          "plan_path":     planPath,
          "context_path":  contextPath,
          "research_path": researchPath,
          "plan_dir":      filepath.Dir(planPath),
          "plan_name":     name,
          "spec_path":     specPath,
      }
  }

  var _ stepkit.PathStrategy = planStrategy{}
  ```

- `internal/steps/plan/steps_test.go` — **no edits**. The existing `renderStep` helper (lines 13-43), `TestArchitectureStepContainsOptionsAndAgreementBeat` (line 45), `TestImplementationDetailStepIsHighLevelOnly` (line 51), `TestTestingApproachStepIsHighLevelOnly` (line 57), `TestOpenQuestionsStepRestrictsToImplTimeUncertainties` (line 63), `TestOutOfScopeStepCoversExclusions` (line 69), `TestStepsOrderMatchesExpected` (line 75), and `TestFSMWalkFromNewToFinished` (line 100) must all continue passing unmodified.
- `internal/steps/plan/scaffold_test.go` — **no edits**. `TestPlanScaffoldShape` (line 12) continues asserting scaffold heading order.

**Complexity**: Medium

**Token estimate**: ~12k

**Agent strategy**: Single agent, sequential. Create `path_strategy.go` first, then rewrite each step callback in `steps.go` one at a time (keeping the package buildable between edits), then delete the private helpers, then run `go test ./internal/steps/plan/...`. Iterate until the full existing test suite passes. The behavior-preservation guarantee is tight: if any substring test fails, the rendered instruction has silently changed and the strategy's `PathVars` output needs to match the original map exactly.

### Phase 1.3: Rewrite `internal/steps/spec` to use `stepkit`

**File changes**:

- `internal/steps/spec/steps.go` (modified): Same pattern as Phase 1.2 but for the spec package. Delete the spec-package private copies of `writeStepResult`, `renderTemplate`, `getString`, `stepTitle` (exact line numbers confirmed at implementation time). Rewrite each step callback to delegate to `stepkit.WriteStepResult` with a `specStrategy` value. The `new()` callback keeps its `st.Write(SpecFilePath(name), ...)` side effect — call it before the stepkit delegation so the file exists by the time the instruction is emitted.

- `internal/steps/spec/path_strategy.go` (new, ~25 lines): A `specStrategy` type whose `PathVars` returns `spec_path` (= `filepath.Join(root, SpecFilePath(name))`) and `spec_name` (= name). `PrimaryPathField` returns `"spec_path"`.

- `internal/steps/spec/steps_test.go` (modified only if necessary) — **verify at phase start** that equivalent rendered-instruction tests exist. If not, follow the Open Questions guidance: STOP and ask the user whether to add minimal pre-refactor tests.

**Complexity**: Medium

**Token estimate**: ~10k

**Agent strategy**: Single agent, sequential. Runs after Phase 1.2 so the stepkit contract is already exercised by a second consumer. Same build-and-iterate pattern.

### Phase 2.1: Create `internal/steps/implement` package with types and step definitions

**File changes**:

- `internal/steps/implement/result.go` (new, ~60 lines):
  ```go
  package implement

  type Result struct {
      Step        string `json:"step"`
      PlanPath    string `json:"plan_path"`
      PlanName    string `json:"plan_name"`
      Instruction string `json:"instruction"`
  }

  type StepEntry struct {
      Name   string `json:"name"`
      Status string `json:"status"`
  }

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

  type StepsResult struct {
      Steps []string `json:"steps"`
  }
  ```

- `internal/steps/implement/plan_path.go` (new, ~8 lines):
  ```go
  package implement

  func PlanFilePath(name string) string {
      return "plans/" + name + "/plan.md"
  }
  ```
  Deliberately a copy of `internal/steps/plan.PlanFilePath` rather than an import, to avoid a cross-package dependency for a single constant.

- `internal/steps/implement/path_strategy.go` (new, ~30 lines): An `implementStrategy` value whose `PathVars` returns `plan_path`, `plan_dir`, `plan_name`, and `changelog_section_name` (literal `"## Changelog"`). `PrimaryPathField` returns `"plan_path"`. `NameKey` returns `"name"`.

- `internal/steps/implement/steps.go` (new, ~150 lines):
  ```go
  package implement

  import (
      "github.com/jumppad-labs/spektacular/internal/stepkit"
      "github.com/jumppad-labs/spektacular/internal/store"
      "github.com/jumppad-labs/spektacular/internal/workflow"
  )

  func Steps() []workflow.StepConfig {
      return []workflow.StepConfig{
          {Name: "new", Src: []string{"start"}, Dst: "new", Callback: newStep()},
          {Name: "read_plan", Src: []string{"new"}, Dst: "read_plan", Callback: readPlan()},
          {Name: "analyze", Src: []string{"read_plan", "update_changelog"}, Dst: "analyze", Callback: analyze()},
          {Name: "implement", Src: []string{"analyze"}, Dst: "implement", Callback: implement()},
          {Name: "test", Src: []string{"implement"}, Dst: "test", Callback: testStep()},
          {Name: "verify", Src: []string{"test"}, Dst: "verify", Callback: verify()},
          {Name: "update_plan", Src: []string{"verify"}, Dst: "update_plan", Callback: updatePlan()},
          {Name: "update_changelog", Src: []string{"update_plan"}, Dst: "update_changelog", Callback: updateChangelog()},
          {Name: "finished", Src: []string{"update_changelog"}, Dst: "finished", Callback: finished()},
      }
  }

  func newStep() workflow.StepCallback {
      return func(data workflow.Data, out workflow.ResultWriter, st store.Store, cfg workflow.Config) (string, error) {
          return "read_plan", nil
      }
  }

  // readPlan, analyze, implement, testStep, verify, updatePlan, updateChangelog, finished
  // — each is a one-liner delegating to stepkit.WriteStepResult with StepRequest{…}
  // and a resultBuilder that returns a Result{Step, PlanPath, PlanName, Instruction}.
  ```

- No `internal/steps/implement/steps_test.go` in this phase — tests land in Phase 2.4.

**Complexity**: Medium

**Token estimate**: ~15k

**Agent strategy**: Single agent, sequential. Write `result.go`, then `plan_path.go`, then `path_strategy.go`, then `steps.go`. Run `go build ./internal/steps/implement/...` after each file. The multi-source transition on `analyze` is the critical piece — if `go-fsm` doesn't accept it, this phase STOPs per the Open Question guidance.

### Phase 2.2: Add implement step templates under `templates/steps/implement/`

**File changes**:

- `templates/steps/implement/01-read_plan.md` (new, ~90 lines) — First instruction. This template is the workflow's gate: nothing else runs until it passes. Structured in four parts:

  1. **Full plan read.** Read `{{plan_path}}` in full (emphasize "no offset, no limit"). Also read the sibling `{{context_path}}` and `{{research_path}}` in full. These three are the source of truth for every downstream step.

  2. **Structural validation.** Verify the plan file has every `## ` section the plan scaffold requires (Overview, Architecture & Design Decisions, Component Breakdown, Data Structures & Interfaces, Implementation Detail, Dependencies, Testing Approach, Milestones & Phases, Open Questions, Out of Scope — matching the list in `templates/scaffold/plan.md`). Verify at least one `#### - [ ] Phase N.M:` heading exists. For every phase heading, verify its `*Technical detail:* [context.md#phase-NM](./context.md#...)` link target resolves to a matching heading inside `{{context_path}}`. If any structural check fails, STOP and report to the user.

  3. **Drift check against the current codebase.** For every file path, package path, function name, type name, command path, and template path named in `{{plan_path}}` or `{{context_path}}` (including inside code blocks and in file:line references), verify the target still exists in the working tree. Method: for file/dir paths, use `ls` or check that the file opens; for Go symbols and package paths, use `grep -rn` or a codebase-locator sub-agent; for CLI commands, check `cmd/` wiring. Collect every mismatch into a list. If the list is non-empty, STOP and report the mismatches to the user with a three-option prompt: (a) fix the plan first and restart this step, (b) proceed with the mismatches noted in memory (agent will adapt during implementation), (c) abandon the workflow. Do not silently continue.

  4. **Changelog mode detection.** Detect whether a `{{changelog_section_name}}` section exists in the plan file. If it does, this is a subsequent-phase invocation and the agent should pick up at the first unchecked `#### - [ ]` phase; if not, this is a first-phase invocation and the `update_changelog` step will later create the section.

  Contains the STOP-and-ask rule for any unexpected state. Ends with the `{{config.command}} implement goto --data '{"step":"{{next_step}}"}'` advance line.

- `templates/steps/implement/02-analyze.md` (new, ~50 lines) — Instructs the agent to identify the current phase (first `#### - [ ]` in the plan), then delegate codebase analysis to sub-agents via `{{config.command}} skill spawn-implementation-agents`. The analysis should cover the files mentioned in the phase's `*Technical detail:* context.md#phase-NM` link, the integration points, and any patterns to follow. Contains the STOP-and-ask rule. Ends with the advance line.

- `templates/steps/implement/03-implement.md` (new, ~40 lines) — Tells the agent to write code for the current phase guided by the analysis summaries from the previous step. Forbids inline test authoring (that's the next step). Includes a STOP-and-ask rule for plan/reality mismatches. Ends with the advance line.

- `templates/steps/implement/04-test.md` (new, ~40 lines) — Directs delegation of test authoring to a sub-agent via `{{config.command}} skill follow-test-patterns`. Explicitly forbids writing tests in the main context. Ends with the advance line.

- `templates/steps/implement/05-verify.md` (new, ~40 lines) — Directs delegation of success-criteria command execution to a sub-agent via `{{config.command}} skill verify-implementation` and instructs the agent to receive back a concise pass/fail summary. Ends with the advance line.

- `templates/steps/implement/06-update_plan.md` (new, ~30 lines) — Directs the agent to mark the current phase's acceptance-criteria checkboxes complete in `{{plan_path}}` — changing `[ ]` to `[x]` for every criterion that passed verification. Ends with the advance line.

- `templates/steps/implement/07-update_changelog.md` (new, ~90 lines) — Directs the agent to write a phase entry to `{{plan_path}}`'s `{{changelog_section_name}}` section with the four fields from spec line 59: (1) what was done, (2) deviations from the plan, (3) files changed, (4) discoveries. **Create-if-missing logic**: the template instructs the agent to first check whether a `{{changelog_section_name}}` section already exists in `{{plan_path}}`; if not, append a new `## Changelog\n\n` heading after the existing `## Out of Scope` section (or at the very end of the file if `## Out of Scope` is missing) and write the first phase entry beneath it; if the section already exists, append the new phase entry below any existing entries. References `{{config.command}} skill update-changelog` for the exact per-entry format. Tells the agent: check for remaining unchecked phases and branch — if any remain and the user has not previously said "run without asking", prompt the user whether to continue; if yes or if autonomous, call `{{config.command}} implement goto --data '{"step":"analyze"}'`; if no remaining phases, prepend a FINAL SUMMARY block to the `## Changelog` section and call `{{config.command}} implement goto --data '{"step":"finished"}'`. Contains the STOP-and-ask rule.

- `templates/steps/implement/08-finished.md` (new, ~20 lines) — Terminal confirmation. Lists the completed phases by re-reading `{{plan_path}}` and pointing the user at the inline changelog. Does not emit a `goto` — this is the terminal state.

**Note on step numbering**: The templates are named `01-07` + `finished` because there are 8 non-initialization steps. The `new` step (step 0) has no template. Numbering matches the order in `Steps()` after stripping `new`.

**Complexity**: Medium

**Token estimate**: ~20k

**Agent strategy**: 2-3 parallel agents each taking 2-3 templates to draft in parallel; each uses `{{config.command}} skill spawn-implementation-agents` for orchestration guidance. After parallel authoring, a single integration pass verifies: every template has the STOP-and-ask rule; every template ends with the `{{config.command}} implement goto` advance line except `08-finished.md`; every spec acceptance criterion L44-65 maps to at least one directive in at least one template. Cross-reference the matrix inline in the integration pass so no criterion gets missed.

### Phase 2.3: Wire `cmd/implement.go` Cobra commands

**File changes**:

- `cmd/implement.go` (new, ~220 lines) — Near-verbatim copy of `cmd/plan.go` with these substitutions:
  - `plan` → `implement` in command strings, variable names, function names, error messages.
  - `planCmd` → `implementCmd` and likewise for each subcommand.
  - `planResultOutputSchema` → `implementResultOutputSchema`; `planStatusOutputSchema` → `implementStatusOutputSchema` (add `unchecked_phases` to the latter).
  - Data directory construction is **unchanged** from plan/spec — the helper `dataDir()` in `cmd/root.go` already returns `.spektacular/` and is shared. Do **not** introduce `implementDataDir(name)` or a per-workflow subdirectory — implement uses the same shared state file as spec and plan.
  - Store construction is `store.NewFileStore(dataDir)` (same as plan/spec), so the implement workflow reads plan files at `plans/<name>/plan.md` through the shared store root.
  - `runPlanNew` → `runImplementNew`. Validates the name matches `nameRegexp` (shared with plan/spec). Importantly, before initializing the workflow, it **verifies the plan file exists** at `filepath.Join(dataDir, implement.PlanFilePath(input.Name))` and returns a non-zero error if not — this is the spec L45 negative case.
  - `runPlanGoto` → `runImplementGoto`. Reads the name from the single shared state file (via `wf.GetData("name")`) — same pattern as `runPlanGoto` today. No active-workflow discovery needed.
  - `runPlanStatus` → `runImplementStatus`. Additionally reads the plan file and counts unchecked `#### - [ ]` phases via a simple regex scan (`regexp.MustCompile(`(?m)^#### - \[ \] Phase \d+\.\d+:`)`), populates `StatusResult.UncheckedPhases`. Handles plan file read errors gracefully (zero count).
  - `runPlanSteps` → `runImplementSteps` (unchanged body beyond the type rename).
  - Do **not** port any `findActivePlan`-style helper — it does not exist in `cmd/plan.go` today.
  - `init()` registers `implementCmd` flags (`--data`, `--stdin`, `--file`, `--schema`, `--dry-run`) and subcommands identically to plan.

- `cmd/root.go` — **Add** `rootCmd.AddCommand(implementCmd)` alongside the existing `rootCmd.AddCommand(specCmd, planCmd)`. One-line change.

**Caveat about shared state**: because `implement new` truncates the same `.spektacular/state.json` that `spec new` and `plan new` use, running `implement new` while a plan or spec workflow is mid-flight will clobber it. This is consistent with existing spec/plan behavior (they already clobber each other). If a future plan wants per-workflow state isolation, that's a separate change; do not introduce it here.

**Complexity**: Medium

**Token estimate**: ~12k

**Agent strategy**: Single agent, sequential. Copy `cmd/plan.go` as the starting scaffold; do the substitutions in order; add the plan-file-exists check; add the `UncheckedPhases` regex scan; register in `root.go`. Build-and-iterate until `go build ./...` succeeds. Tests in Phase 2.4.

### Phase 2.4: Add implement tests covering FSM, templates, and commands

**File changes**:

- `internal/steps/implement/steps_test.go` (new, ~300 lines):
  - `renderStep(t, cb)` helper copy-pasted from `internal/steps/plan/steps_test.go:13-43` (per the copy-paste-per-package convention in `thoughts/notes/testing.md`). Swap the `Result` type reference for `implement.Result`.
  - `TestStepsOrderMatchesExpected` — asserts the nine-step sequence mirroring `internal/steps/plan/steps_test.go:75`.
  - `TestFSMWalkFromNewToFinished` — mirrors `internal/steps/plan/steps_test.go:100` but walks: `new → read_plan → analyze → implement → test → verify → update_plan → update_changelog → finished`. Uses `DryRun: true`. Asserts `wf.Current()` at each transition.
  - `TestFSMLoopFromUpdateChangelogBackToAnalyze` — the critical test for the multi-source transition. Walks through the happy path to `update_changelog`, then calls `wf.Goto("analyze")` and asserts the transition succeeds (current state = `analyze`), walks forward again to `update_changelog`, then calls `wf.Goto("finished")` and asserts success.
  - `TestReadPlanStepContainsFullReadDirective` — asserts the rendered `read_plan` template contains strings like `"no offset"`, `"no limit"`, and `{{plan_path}}` resolved to the test plan path. Also asserts the template directs reads of `{{context_path}}` and `{{research_path}}`. Satisfies spec L46.
  - `TestReadPlanStepMentionsChangelog` — asserts the rendered template contains `"## Changelog"` and references first-phase vs subsequent-phase logic. Satisfies spec L48-49.
  - `TestReadPlanTemplateDirectsStructuralValidation` — asserts the rendered `read_plan` template mentions every required `## ` section name from the plan scaffold (Overview, Architecture & Design Decisions, Component Breakdown, Data Structures & Interfaces, Implementation Detail, Dependencies, Testing Approach, Milestones & Phases, Open Questions, Out of Scope), the `#### - [ ] Phase` pattern, and the `*Technical detail:*` link-resolution directive. New test, not tied to an existing spec line.
  - `TestReadPlanTemplateDirectsDriftCheck` — asserts the rendered `read_plan` template contains a drift-check directive covering file paths, package paths, and symbols named in plan.md/context.md, uses words like "verify", "working tree", "STOP", and offers a three-option (fix / proceed / abandon) prompt. New test, not tied to an existing spec line — this is a post-drift learning, added because the 15_implementation plan itself drifted between authoring and implementation and we want the workflow to catch similar drift automatically.
  - `TestAnalyzeStepReferencesSpawnImplementationAgents` — asserts the rendered `analyze` template contains `"skill spawn-implementation-agents"`. Satisfies spec L51.
  - `TestTestStepForbidsInlineTests` — asserts the rendered `test` template contains `"follow-test-patterns"` and forbids main-context test authoring. Satisfies spec L53.
  - `TestVerifyStepReferencesVerifyImplementation` — asserts the rendered `verify` template contains `"verify-implementation"` and references a concise pass/fail summary. Satisfies spec L55.
  - `TestUpdatePlanStepDirectsCheckboxMarking` — asserts the rendered `update_plan` template contains `[x]` marking instructions. Satisfies spec L57.
  - `TestUpdateChangelogStepSpecifiesFields` — asserts the rendered `update_changelog` template mentions "what was done", "deviations", "files changed", "discoveries", and "FINAL SUMMARY". Satisfies spec L59.
  - `TestStopOnMismatchDirectiveInstructionsPresent` — iterates over every implement step template and asserts a STOP-and-ask directive is present. Satisfies spec L61.
  - `TestUpdateChangelogStepLoopBranching` — asserts the rendered template contains both `goto --data '{"step":"analyze"}'` and `goto --data '{"step":"finished"}'` with branching logic. Satisfies spec L63.
  - `TestFinishedStepIsTerminalMarker` — asserts the rendered `finished` template emits no further `goto` command. Satisfies spec L65.

- `cmd/implement_test.go` (new, ~250 lines):
  - `TestImplementNewReturnsValidInstruction` — invokes `implementNewCmd` with `--data '{"name":"fixture-plan"}'` in a `t.TempDir()`-based environment where the plan file exists; asserts exit code 0 and JSON output contains `step`, `plan_path`, `plan_name`, and a non-empty `instruction`. Satisfies spec L45 positive.
  - `TestImplementNewFailsOnMissingPlan` — invokes with a name that has no plan file; asserts non-zero error and an informative error message. Satisfies spec L45 negative.
  - `TestImplementGotoAdvancesThroughSteps` — drives a full walk via `Goto` calls and asserts each emits the expected JSON shape.
  - `TestImplementStatusReturnsUncheckedPhasesCount` — creates a fixture plan file with three unchecked `#### - [ ]` phases and one checked, asserts `UncheckedPhases` == 3.
  - `TestImplementStepsListsAllNineSteps` — invokes `steps` and asserts the returned list contains all nine step names in order.
  - `TestImplementNewSchemaOutput` — invokes with `--schema` and asserts a non-empty schema JSON. Repeats for `goto`, `status`, `steps`.

- No edits to `internal/stepkit/stepkit_test.go`, `internal/steps/plan/steps_test.go`, `internal/steps/plan/scaffold_test.go`, or `internal/steps/spec/*_test.go`. The lack of edits there is a positive signal — the refactor in Milestone 1 did not change observable behavior.

**Complexity**: Medium

**Token estimate**: ~18k

**Agent strategy**: 2 parallel agents, one writing `internal/steps/implement/steps_test.go` and one writing `cmd/implement_test.go`. The per-spec-criterion tests should be authored directly from the spec file (`.spektacular/specs/15_implementation.md` lines 44-65) — one test per criterion, making the mapping trivially auditable. After parallel authoring, a single agent runs `make test` and iterates on failures.

### Phase 3.1: Add three new skill markdown files

**File changes**:

- `templates/skills/skill_follow-test-patterns.md` (new, ~40 lines) — Frontmatter: `name: follow-test-patterns`, `title: Follow Test Patterns`. Body: "Write tests for the code just implemented in phase N, matching the conventions documented in `thoughts/notes/testing.md`. Use `stretchr/testify/require` for assertions; co-locate `*_test.go` next to the package under test; use `t.TempDir()` and `store.NewFileStore` for fixtures. Do not mock; use minimal in-test fakes instead. Return the written test file paths and a one-line summary of what each test asserts."

- `templates/skills/skill_verify-implementation.md` (new, ~40 lines) — Frontmatter: `name: verify-implementation`, `title: Verify Implementation`. Body: "Run the plan's success-criteria commands (from `thoughts/notes/commands.md`: typically `make test`, `make lint`, and any phase-specific commands listed under the current phase's acceptance criteria). Capture exit codes and a short excerpt of each failure. Return a concise pass/fail summary — one line per command — not the full test output. If everything passes, return a single 'all green' line. If anything fails, return the failing command and a 5-10 line excerpt of the failure."

- `templates/skills/skill_update-changelog.md` (new, ~60 lines) — Frontmatter: `name: update-changelog`, `title: Update Changelog`. Body: describes both the lifecycle and the per-entry format the implement-workflow agent should follow. **Lifecycle**: on first invocation for a given plan, create a new `## Changelog` section by appending `## Changelog\n\n` after the plan's existing `## Out of Scope` section (or at the end of the file if absent); on subsequent invocations, append new phase entries below any existing entries under the existing section; on the final invocation (no more unchecked phases), prepend a `### FINAL SUMMARY` block at the top of the section. **Per-entry format**: a dated heading per phase (e.g. `### 2026-04-13 — Phase 1.1: <title>`); a 'What was done' paragraph; a 'Deviations' paragraph listing anything that didn't match the plan (or "None" explicitly); a 'Files changed' bullet list; a 'Discoveries' paragraph capturing anything the next phase or future maintainer should know. **FINAL SUMMARY format**: a `### FINAL SUMMARY` heading placed immediately below the `## Changelog` heading (above all per-phase entries), with a 2-4 sentence overall summary and a "Total phases: N/M" line.

**Complexity**: Low

**Token estimate**: ~6k

**Agent strategy**: Single agent, sequential. Write each file top-to-bottom against the existing `skill_spawn-implementation-agents.md` as a format reference. Verify each file via `go run . skill <name>` before moving to the next.

## Testing Strategy

Testing ownership per package:

- **`internal/stepkit`** — New unit test file exercising helper contracts with a minimal fake `PathStrategy`. ~7 tests covering standard vars, strategy vars, extras override, result builder invocation, and error paths.
- **`internal/steps/plan`** — Existing test files unchanged. Their unchanged passing status is the behavior-preservation signal for Phase 1.2.
- **`internal/steps/spec`** — Existing test files unchanged (assuming they exist per Open Question). Their unchanged passing status is the behavior-preservation signal for Phase 1.3.
- **`internal/steps/implement`** — New unit test file with ~13 tests: step ordering, FSM walk, loop, and per-spec-criterion substring assertions.
- **`cmd/`** — New `cmd/implement_test.go` covering the Cobra command surface with ~6 tests.
- **`templates/steps/implement/*.md`** — No direct tests. Coverage comes from the `internal/steps/implement` per-spec-criterion substring tests which render each template through `renderStep` and assert on the output.
- **`templates/skills/`** — No direct tests for the new skill files. Their content is exercised through integration with the implement templates.

Total new tests: ~26. Total changed existing tests: 0. The refactor in Milestone 1 is invisible at the test level by design.

## Project References

- `thoughts/notes/commands.md` — Project commands reference.
- `thoughts/notes/testing.md` — Testing patterns reference.
- `.spektacular/specs/15_implementation.md` — Source spec driving this plan.
- `.spektacular/plans/16_plan_format/plan.md` — Prior plan that established the current plan scaffold shape.
- `templates/scaffold/plan.md` — Plan scaffold. Ends at `## Out of Scope`; contains no `## Changelog` section (the implement workflow creates it at runtime).
- `templates/steps/plan/02-discovery.md` — Canonical skill-reference pattern (`{{config.command}} skill <name>`).
- `templates/steps/plan/10-phases.md` — Secondary example of the skill-reference pattern.
- `internal/steps/plan/steps.go` — The helper bundle (`writeStepResult`, `getString`, `renderTemplate`, `stepTitle`) being extracted into `internal/stepkit`. Lives at the bottom of the file below the step callbacks.
- `internal/steps/plan/steps_test.go` — `renderStep` helper pattern and `TestFSMWalkFromNewToFinished` pattern to mirror.
- `cmd/plan.go` — Command wiring pattern to mirror in `cmd/implement.go`.
- `internal/workflow/workflow.go` — FSM engine and `Config`/`Data`/`StepConfig` types.
- `internal/store/store.go` — Store interface.

## Token Management Strategy

| Tier | Token Budget | Agent Strategy |
|------|-------------|----------------|
| Low | ~10k | Single agent, sequential |
| Medium | ~25k | 2-3 parallel agents |
| High | ~50k+ | Parallel analysis, sequential integration |

All phases in this plan are Low or Medium complexity. No High-complexity phases exist — the work is well-scoped enough to stay under ~20k tokens per phase with a single agent or two parallel agents.

## Migration Notes

No runtime migration is required. The refactor in Milestone 1 is behavior-preserving; the existing shared `.spektacular/state.json` file is read and written by the new code without modification. The new `implement` workflow uses the same `.spektacular/` data dir and the same `.spektacular/state.json` state file — running `implement new` truncates the state file the same way `plan new` and `spec new` do today. No new directory family is introduced.

No user-visible CLI changes for `spec` and `plan` — their command surface, flags, JSON output, and error messages are preserved byte-for-byte where the refactor touches them. The only net-new user-visible addition is `spektacular implement` and its four subcommands.

## Performance Considerations

None. The workflow is I/O bound against a few kilobytes of markdown per step and runs once per agent step; performance is a non-issue at the spektacular CLI layer. The `UncheckedPhases` regex scan in `implement status` reads the plan file once per invocation (a few KB) — trivially fast. The stepkit extraction adds one extra function-call indirection per rendered step — unmeasurable overhead.
