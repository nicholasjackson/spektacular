# Research: 16_plan_format

## Alternatives considered and rejected

### Option B: Logical grouping (~10 steps)

Fold Architecture & Design Decisions, Component Breakdown, and Data Structures & Interfaces into a single `design` step, and Dependencies + Testing Approach into a single `cross_cutting` step. Keeps tightly-coupled design work in one agent pass and shrinks the step count closer to today's seven.

**Rejected**: The coupling argument cuts both ways — grouping three sections behind one prompt makes it harder to tune section-specific guidance and harder to spot shallow-filling of any one section. The spec's success metric "reviewers can spot missing architectural patterns or design gaps from plan.md without needing to read context.md" (`.spektacular/specs/16_plan_format.md:121`) is easier to satisfy when each section has its own focused step. Cited evidence: `internal/spec/steps.go:23-36` already demonstrates that a ten-step workflow is sustainable in this codebase.

### Option C: Minimal reshape (~8 steps)

Keep the existing `approach → milestones → phases` beat and expand `approach` into a single large `design` step that asks the agent to author seven sections in one pass. Smallest Go change, fastest to ship.

**Rejected**: Concentrating seven sections of output into one agent response maximises the risk of shallow-fill, exactly the failure mode the spec is designed to prevent (`.spektacular/specs/16_plan_format.md:13`: "enough detail ... for another developer to pick up the work and implement it without needing the original planner"). Also violates the spirit of the spec requirement at line 34 ("one step per section or logical grouping") even though the letter allows grouping.

## Chosen approach — evidence

- `internal/spec/steps.go:23-36` — demonstrates a ten-step FSM using the same `StepConfig` pattern the plan package uses; proves the shape is already idiomatic in this repo.
- `templates/spec-scaffold.md:1-86` — demonstrates the rich HTML-comment inline-guidance pattern the spec (`.spektacular/specs/16_plan_format.md:47`) asks for; can be copied directly into the plan scaffold.
- `internal/plan/steps.go:117-153` — `writeStepResult` is already generic over step name, next step, and template path, so new callbacks are one-line additions.
- `templates/plan-steps/03-approach.md:8-22` — the option-presentation and user-agreement beat currently lives here and can be lifted verbatim into the new `architecture` step.
- `templates/plan-steps/04-milestones.md` and `templates/plan-steps/05-phases.md` — prove that the existing Milestones & Phases substructure is template-driven and doesn't need Go-side changes when we rewrite the scaffold.

## Files examined

- `templates/plan-scaffold.md:1-52` — current six-section plan scaffold; full rewrite target for Phase 1.1.
- `templates/spec-scaffold.md:1-86` — reference pattern for HTML comment inline guidance.
- `templates/context-scaffold.md:1-52` — confirmed untouched by this plan.
- `templates/research-scaffold.md:1-44` — confirmed untouched by this plan.
- `templates/plan-steps/01-overview.md` through `07-finished.md` — current seven step templates; renumbered in Phase 2.2.
- `internal/plan/steps.go:1-181` — current seven-step FSM, callback pattern, shared `writeStepResult` and `renderTemplate` helpers.
- `internal/spec/steps.go:1-187` — reference for the larger workflow shape.
- `internal/workflow/workflow.go` and `workflow_test.go` — FSM engine used by both packages; the new fourteen-step FSM slots into the same engine without changes.
- `internal/plan/result.go` — plan result struct for `WriteResult`; unchanged by this work.
- `.spektacular/plans/15_implementation/plan.md` — example of a current-shape plan; reference point for what "Milestones & Phases structurally unchanged" means in practice.
- `.spektacular/specs/16_plan_format.md` — the spec driving this plan.
- `cmd/plan.go` — plan CLI entry point; no changes needed.

## External references

None. This is an internal restructure.

## Prior plans / specs consulted

- `.spektacular/plans/15_implementation/plan.md` — the implement workflow plan. Confirmed that the Milestones & Phases checkbox structure is the data shape the implement workflow reads, which is why the spec requires it to remain structurally unchanged.
- `.spektacular/specs/15_implementation.md` — the implement workflow spec. Confirmed that Changelog section in plan.md is written by the implement workflow and must be preserved.

## Open assumptions

- The Changelog section at the bottom of `plan-scaffold.md` is not counted among the spec's "ten sections" and stays in place for the implement workflow. If the spec intended Changelog to be removed, the implement workflow (plan 15) would break. Implementer should verify by re-reading `.spektacular/specs/15_implementation.md` if unsure.
- The metadata block (`<!-- Created: --> <!-- Commit: --> ...`) at the top of the current scaffold is not one of the ten sections and stays unchanged. The spec does not mention it either way.
- `mustache.Render` tolerates arbitrary literal content in templates as long as `{{...}}` tokens resolve. Verified by inspecting existing templates, but new templates should avoid introducing new variables that aren't in the known set listed in context.md Phase 2.2.
- `spektacular plan steps` CLI command exists and lists step names from `plan.Steps()`. Not directly verified by reading `cmd/plan.go` — Phase 2.3 should confirm before asserting against it in a test.

## Rehydration cues

If this plan is picked up cold:

1. Re-read `.spektacular/specs/16_plan_format.md` first — it is the source of truth for the ten-section list and the high-level/open-questions rules.
2. Re-read `templates/plan-scaffold.md` and `templates/spec-scaffold.md` side by side to remind yourself what "HTML comment inline guidance style" looks like.
3. Re-read `internal/plan/steps.go` and `internal/spec/steps.go` to reconfirm the callback pattern and the shape the new fourteen-step FSM must take.
4. Run `go run . plan steps` on the current branch to see today's step output before changing anything.
5. Use the `spawn-implementation-agents` skill for Phase 2.2 — authoring eight templates in parallel is the biggest chunk of work.
