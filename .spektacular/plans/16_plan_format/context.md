# Context: 16_plan_format

## Current State Analysis

- `templates/plan-scaffold.md:1-52` — the current six-section plan scaffold. Sections: Overview, Implementation Approach, Desired End State, What We're NOT Doing, Milestones & Phases, Changelog. Each section has a one-line HTML comment. The Milestones & Phases block starts at line 29 and includes a Milestone 1 stub, a Phase 1.1 checkbox heading (`#### - [ ] Phase 1.1: <short title>`), a `*Technical detail:*` link into `context.md`, and an outcome-based acceptance criteria checkbox list — all of which must be preserved verbatim by Phase 1.1 of this plan.
- `templates/spec-scaffold.md:1-86` — the reference pattern for rich HTML-comment inline guidance. Each section gets a multi-line HTML comment explaining purpose, format rules, and examples. This is the style Phase 1.1 copies into plan-scaffold.md.
- `internal/plan/steps.go:30-41` — the current seven-step FSM: `new → overview → discovery → approach → milestones → phases → verification → finished`. Each step has its own callback and maps to a template under `templates/plan-steps/NN-<name>.md`.
- `internal/plan/steps.go:80-102` — the `verification` step callback. It renders `plan-scaffold.md`, `context-scaffold.md`, and `research-scaffold.md` with the plan name, then passes all three as `{{plan_template}}`, `{{context_template}}`, `{{research_template}}` into the step template. Phase 2.1 does not need to change this — only the FSM wiring above it.
- `internal/plan/steps.go:117-153` — `writeStepResult` is the shared helper for rendering a step template and writing the result. New callbacks in Phase 2.1 can reuse this as-is.
- `internal/spec/steps.go:23-36` — reference for a ten-step workflow. The new fourteen-step plan FSM follows the same shape: one `StepConfig` entry per step, each wired to its own named callback.
- `templates/plan-steps/` — currently contains `01-overview.md`, `02-discovery.md`, `03-approach.md`, `04-milestones.md`, `05-phases.md`, `06-verification.md`, `07-finished.md`. After Phase 2.2 this directory holds fourteen templates in the new order.
- `templates/plan-steps/03-approach.md:1-34` — the soon-to-be-deleted `approach` step. Its key beats — presenting 2-3 design options, getting user agreement on direction and out-of-scope — must migrate to the new `architecture` step in Phase 2.2.
- `templates/plan-steps/05-phases.md:34` — contains the "NO open questions — resolve any uncertainties now" rule, which stays in force for phase-level decisions even though a new Open Questions section is introduced at the plan level. The new `open_questions` template in Phase 2.2 must be clear that its bucket is strictly for uncertainties that genuinely cannot be resolved until implementation begins.
- `internal/plan/` has no existing `_test.go` files, confirmed by `find . -name "*_test.go"`. Phase 1.2 and Phase 2.3 create the first tests in the package.

## Per-Phase Technical Notes

### Phase 1.1: Rewrite `plan-scaffold.md`

- `templates/plan-scaffold.md` — full rewrite. Preserve lines 1-7 (title and metadata block) and lines 49-51 (Changelog section). Replace everything between with the ten new sections in order: Overview, Architecture & Design Decisions, Component Breakdown, Data Structures & Interfaces, Implementation Detail, Dependencies, Testing Approach, Milestones & Phases, Open Questions, Out of Scope.
- Each new section is preceded by an HTML comment block modelled on `templates/spec-scaffold.md:3-10` (multi-line, explains purpose, lists rules or example bullets). Keep the comment block tight — aim for 4-8 lines per section.
- Implementation Detail and Testing Approach HTML comments must contain an explicit "high-level only — per-phase detail belongs in `context.md`" phrase so the agent reading the comment cannot miss it.
- Open Questions HTML comment must restrict the section to "items that genuinely cannot be resolved until implementation" — mirror the language the user gave in the approach step.
- Milestones & Phases content — copy lines 29-47 of the current scaffold verbatim. Do not rewrite the Milestone 1 stub, the Phase 1.1 checkbox heading, the `*Technical detail:*` link, or the acceptance criteria checkboxes. Phase 1.2's test asserts on their presence.

**Complexity**: Low
**Token estimate**: ~5k
**Agent strategy**: Single agent, sequential. Pure template rewrite.

### Phase 1.2: Scaffold shape test

- New file `internal/plan/scaffold_test.go`. Renders `plan-scaffold.md` via the existing `renderTemplate` helper at `internal/plan/steps.go:164-170` (or inlines an equivalent — it's two lines of `templates.FS.ReadFile` + `mustache.Render`).
- Test assertions:
  1. Substring match for each of the ten section headings in document order (use `strings.Index` and require monotonically increasing indices).
  2. For each section heading, assert that an HTML comment (`<!--`) appears somewhere between the previous section and the current heading.
  3. Substring match for `#### - [ ] Phase` and `*Technical detail:*` anywhere in the rendered output.
- Heading list (in order): `## Overview`, `## Architecture & Design Decisions`, `## Component Breakdown`, `## Data Structures & Interfaces`, `## Implementation Detail`, `## Dependencies`, `## Testing Approach`, `## Milestones & Phases`, `## Open Questions`, `## Out of Scope`.
- Use standard `testing.T` — the project's Go tests in `internal/workflow/workflow_test.go` and `internal/store/store_test.go` both use the standard library, no extra framework.

**Complexity**: Low
**Token estimate**: ~4k
**Agent strategy**: Single agent, sequential.

### Phase 2.1: Plan FSM steps

- `internal/plan/steps.go:30-41` — replace the `Steps()` return with the new fourteen entries in this order:
  ```
  new, overview, discovery, architecture, components, data_structures,
  implementation_detail, dependencies, testing_approach, milestones,
  phases, open_questions, out_of_scope, verification, finished
  ```
  Each entry follows the existing `{Name, Src, Dst, Callback}` pattern. `Src` for each new step is the previous step's `Name`. `verification`'s `Src` changes from `phases` to `out_of_scope`.
- Delete the `approach()` function (currently `internal/plan/steps.go:62-66`).
- Add eight new callback functions mirroring the existing one-line pattern: each returns `writeStepResult(<name>, <next>, "plan-steps/NN-<name>.md", ...)`. Template numbering: `03-architecture.md`, `04-components.md`, `05-data_structures.md`, `06-implementation_detail.md`, `07-dependencies.md`, `08-testing_approach.md`, `09-milestones.md`, `10-phases.md`, `11-open_questions.md`, `12-out_of_scope.md`, `13-verification.md`, `14-finished.md`.
- Update `overview()` next-step to `discovery` (unchanged) and `discovery()` next-step from `approach` to `architecture`.
- Update `milestones()` template path from `04-milestones.md` to `09-milestones.md`; `phases()` from `05-phases.md` to `10-phases.md`; `verification()` from `06-verification.md` to `13-verification.md`; `finished()` from `07-finished.md` to `14-finished.md`. These callbacks are defined at `internal/plan/steps.go:68-113`.
- `stepTitle` at `internal/plan/steps.go:172-180` handles underscore-to-space conversion automatically — no change needed for `data_structures`, `implementation_detail`, `testing_approach`, `open_questions`, `out_of_scope`.

**Complexity**: Medium
**Token estimate**: ~8k
**Agent strategy**: Single agent, sequential. Needs to be landed together with Phase 2.2 or the FSM will reference templates that don't exist.

### Phase 2.2: Step templates

- Rename existing templates under `templates/plan-steps/`:
  - `01-overview.md` → stays `01-overview.md`
  - `02-discovery.md` → stays `02-discovery.md`
  - `03-approach.md` → **delete**
  - `04-milestones.md` → `09-milestones.md`
  - `05-phases.md` → `10-phases.md`
  - `06-verification.md` → `13-verification.md`
  - `07-finished.md` → `14-finished.md`
- Create eight new templates:
  - `03-architecture.md` — leads with the option-presentation and user-agreement beat from the old `03-approach.md:8-22`, then prompts the agent to author the Architecture & Design Decisions section. Key directive strings: "present 2-3 design options", "get the user's agreement", "record the chosen direction".
  - `04-components.md` — prompts the agent to author the Component Breakdown section. New components, their responsibilities, and how they interact.
  - `05-data_structures.md` — prompts the agent to author the Data Structures & Interfaces section. Type shapes, interface signatures, serialization boundaries.
  - `06-implementation_detail.md` — prompts the agent to author the Implementation Detail section. **Must contain** the directive "high-level only" and "per-phase detail stays in context.md".
  - `07-dependencies.md` — prompts the agent to author the Dependencies section. Internal packages, external libraries, upstream specs.
  - `08-testing_approach.md` — prompts the agent to author the Testing Approach section. **Must contain** the directive "high-level only" and "per-phase testing detail stays in context.md".
  - `11-open_questions.md` — prompts the agent to author the Open Questions section. **Must contain** the directive "items that genuinely cannot be resolved until implementation" so the agent does not stash resolvable uncertainties here.
  - `12-out_of_scope.md` — short step where agent and user lock in the exclusions that go into the Out of Scope section.
- Each template follows the existing header pattern at the top of every current plan-steps template: `## Step {{step}}: {{title}}` then body content, ending with `{{config.command}} plan goto --data '{"step":"{{next_step}}"}'`. See `templates/plan-steps/02-discovery.md:1-2,46` for the shape.
- Every template must render cleanly through `mustache.Render` — no stray `{{...}}` that isn't in the known variable set (`step`, `title`, `plan_path`, `context_path`, `research_path`, `plan_dir`, `plan_name`, `spec_path`, `next_step`, `config.command`).

**Complexity**: Medium
**Token estimate**: ~15k
**Agent strategy**: 2-3 parallel agents — the eight new templates are independent of each other and can be authored in parallel, with a final sequential pass to check directive strings.

### Phase 2.3: Workflow contract tests

- New file `internal/plan/steps_test.go`. Uses `testing.T` and the existing `workflow` package to drive the FSM.
- For each of `architecture`, `implementation_detail`, `testing_approach`, `open_questions`, `out_of_scope`: construct a minimal `workflow.Data` with `name: "test"`, call the step's callback via a test `ResultWriter` that captures the rendered `Instruction`, and assert the required directive substrings are present:
  - `architecture`: must contain "design options" and "user" (for the agreement beat).
  - `implementation_detail`: must contain "high-level" and "context.md".
  - `testing_approach`: must contain "high-level" and "context.md".
  - `open_questions`: must contain "implementation" (in the sense of "cannot be resolved until implementation").
  - `out_of_scope`: must contain "Out of Scope" or "exclusions".
- FSM regression test: construct a `workflow.Workflow` from `plan.Steps()`, start from `new`, call `Advance()` fourteen times, assert the final step is `finished` and each intermediate transition succeeds.
- `spektacular plan steps` CLI test: either add to `cmd/` or assert via an in-package call that the ordered step names match the expected list. Simplest form is a table-driven test in `internal/plan/steps_test.go` that iterates `plan.Steps()` and asserts `[]string{"new", "overview", ..., "finished"}`.
- The `ResultWriter` mock can be a tiny struct with a `WriteResult(r Result) error` method that stores the `Result` for inspection — same pattern the workflow package already uses if any helpers exist; otherwise inline it.

**Complexity**: Medium
**Token estimate**: ~10k
**Agent strategy**: 2 parallel agents — one writes the directive-string tests, one writes the FSM regression test. Merge is trivial (no shared file contention beyond a single test file).

## Testing Strategy

- **Scaffold shape** (Phase 1.2): one Go test asserts the ten headings appear in order and every section has a preceding HTML comment. Protects the scaffold contract from silent regression.
- **Step contract** (Phase 2.3): per-step callback tests that assert required directive substrings in the rendered instruction. The contract of each step lives in its markdown template, so directly testing the rendered output is the right level.
- **FSM regression** (Phase 2.3): end-to-end walk of the fourteen-step FSM via the existing `internal/workflow` engine. Guards against accidental edge removal or step reordering.
- No integration test against a real spec — the existing `.spektacular/plans/15_implementation/` plan is the functional reference for the old shape, and Phase 1.2 / Phase 2.3 together cover the new shape.

## Project References

- `thoughts/notes/commands.md` — does not yet exist; create via `discover-project-commands` skill during implementation.
- `thoughts/notes/testing.md` — does not yet exist; create via `discover-test-patterns` skill during implementation.
- `templates/spec-scaffold.md` — reference for HTML comment inline guidance style.
- `internal/spec/steps.go` — reference for a ten-step workflow shape.

## Token Management Strategy

| Tier | Token Budget | Agent Strategy |
|------|-------------|----------------|
| Low | ~10k | Single agent, sequential |
| Medium | ~25k | 2-3 parallel agents |
| High | ~50k+ | Parallel analysis, sequential integration |

Phase 1.1 (Low), Phase 1.2 (Low), Phase 2.1 (Medium), Phase 2.2 (Medium), Phase 2.3 (Medium). Total work sits comfortably in the Medium tier.

## Migration Notes

No migration of existing plans. `.spektacular/plans/15_implementation/plan.md` stays in the old format and is not rewritten. Only new plans produced after this work lands use the new template.

## Performance Considerations

None. All changes are template and FSM wiring; no hot-path code.
