# Plan: 16_plan_format

<!-- Metadata -->
<!-- Created: 2026-04-13T10:33:40Z -->
<!-- Commit: 005ab0973eaea080fe73499090a5e04298e6184e -->
<!-- Branch: skills -->
<!-- Repository: git@github.com:jumppad-labs/spektacular.git -->

## Overview

Restructure Spektacular's plan generation so the output of `spektacular plan` is a self-contained technical design document with ten standardised sections. A developer handed `plan.md` alone should have almost everything needed to implement the work, so handoff between planner and implementer stops requiring a walkthrough. The change covers both the scaffold shape and the workflow steps that drive an agent through populating it.

## Implementation Approach

Rework happens in two clean milestones. Milestone 1 rewrites `templates/plan-scaffold.md` into the ten-section format, copying the rich HTML-comment inline-guidance style from `templates/spec-scaffold.md` so each section teaches the agent what content belongs in it. The existing Milestones & Phases substructure (checkbox phase headings, `*Technical detail:*` links, outcome acceptance criteria) is copied across untouched so we don't disturb the parts that already work. A focused scaffold-shape test guards the contract.

Milestone 2 then rewires the plan workflow FSM in `internal/plan/steps.go` from seven steps to fourteen, one step per new section plus the existing `discovery`, `verification`, and `finished` steps. The old `approach` step is deleted — option presentation and user agreement move into the new `architecture` step where the design decisions are actually recorded. Eight new markdown templates under `templates/plan-steps/` carry the per-section guidance; the `implementation_detail` and `testing_approach` templates explicitly mark their sections as high-level only, and the `open_questions` template narrowly restricts its bucket to impl-time uncertainties that genuinely cannot be resolved during planning.

The alternative of a single large "design" step producing seven sections in one pass was rejected because it concentrates the agent's output and makes shallow-filling hard to detect. See [research.md#alternatives-considered-and-rejected](./research.md#alternatives-considered-and-rejected) for the full comparison of options A, B, and C.

## Desired End State

- `templates/plan-scaffold.md` is a ten-section technical design document with inline HTML guidance on every section.
- Running `spektacular plan` walks the agent through fourteen steps, one per section (plus discovery, verification, finished), and produces a `plan.md` with populated content in every section.
- The Milestones & Phases block in generated plans is structurally identical to today: checkbox phase headings, `*Technical detail:*` links into `context.md`, outcome-based acceptance criteria.
- `research.md` and `context.md` scaffolds are unchanged and continue to hold research evidence and file-level technical detail respectively.
- `spektacular plan steps` lists the fourteen step names in order.

## What We're NOT Doing

- Not restructuring `research.md` or `context.md` scaffolds — they stay as they are.
- Not migrating existing plans in `.spektacular/plans/` to the new format.
- Not validating section content at step-advance time; the workflow still advances on user signal.
- Not adding new top-level CLI commands; this is entirely inside the existing `plan` workflow.
- Not changing how the implement workflow reads plans — phase checkboxes and `*Technical detail:*` links stay put.

## Milestones & Phases

<!-- 2-4 milestones. Each milestone leads with a "What changes" summary paragraph describing the user-visible difference when the milestone lands. Each phase has a 2-4 sentence summary, a *Technical detail:* link to context.md, and an **Acceptance criteria**: checkbox list with outcome statements (not shell commands). -->

### Milestone 1: New plan template format

**What changes**: When you run `spektacular plan`, the generated `plan.md` is a full technical design document — overview, architecture decisions, component breakdown, data structures, implementation sketches, dependencies, testing strategy, milestones and phases, open questions, and out of scope. Each section has inline guidance explaining what belongs in it, so a reviewer can tell at a glance whether a section is shallow-filled. Only the scaffold changes in this milestone: the workflow still drives you through the old seven steps, but those steps now write into the new template shape. A developer handed `plan.md` alone has most of what they need to start building, without needing to open `context.md` first.

#### - [ ] Phase 1.1: Rewrite `templates/plan-scaffold.md` with the ten-section format

Replace the current six-section scaffold with a ten-section technical design document: Overview, Architecture & Design Decisions, Component Breakdown, Data Structures & Interfaces, Implementation Detail, Dependencies, Testing Approach, Milestones & Phases, Open Questions, Out of Scope. Every section gets an HTML comment block in the style of `templates/spec-scaffold.md` describing what content belongs there, and the Implementation Detail / Testing Approach comments explicitly say "high-level only — per-phase detail belongs in context.md". The existing Milestones & Phases structure (metadata block, Milestone 1 scaffolding, Phase 1.1 checkbox heading with `*Technical detail:*` link and outcome acceptance criteria) is copied across untouched. The Changelog section is retained at the bottom for the implement workflow.

*Technical detail:* [context.md#phase-11](./context.md#phase-11-rewrite-plan-scaffold)

**Acceptance criteria**:

- [ ] `plan-scaffold.md` contains all ten section headings in the order specified by the spec
- [ ] Each of the ten sections is preceded by an HTML comment block explaining what belongs in that section
- [ ] The Implementation Detail and Testing Approach HTML comments explicitly say the content must be high-level and per-phase detail stays in context.md
- [ ] Milestones & Phases scaffolding (Milestone 1, Phase 1.1 checkbox heading, `*Technical detail:*` link, acceptance-criteria checkbox list) is present and structurally identical to today
- [ ] Existing metadata block and Changelog section are preserved

#### - [ ] Phase 1.2: Test that the rendered scaffold has the required shape

Add a small Go test in `internal/plan/` that renders `plan-scaffold.md` and asserts: the ten section headings are present in the correct order, every section has a preceding HTML comment, and the Milestones & Phases block contains a checkbox-style phase heading and a `*Technical detail:*` link. This is the only test the spec implicitly asks for — it protects the scaffold contract so future edits can't silently break it.

*Technical detail:* [context.md#phase-12](./context.md#phase-12-scaffold-shape-test)

**Acceptance criteria**:

- [ ] A test in `internal/plan/` renders `plan-scaffold.md` and asserts the ten headings appear in the correct order
- [ ] The same test asserts each section is preceded by an HTML comment
- [ ] The test asserts Milestones & Phases still contains a `#### - [ ] Phase` checkbox heading and a `*Technical detail:*` link

---

### Milestone 2: Workflow steps match the new sections

**What changes**: The planning workflow now walks you through the plan one section at a time. Instead of the old seven-step `overview → discovery → approach → milestones → phases → verification → finished` path, you move through fourteen steps — one per section of the new template, plus discovery, verification, and finished. `approach` is gone: option presentation and user agreement now happen inside the `architecture` step where the design decisions are actually recorded. At each step the agent is prompted for exactly the section it's authoring, with guidance tuned to that section (high-level for Implementation Detail and Testing Approach, structural for Architecture and Components, narrow for Open Questions). When the workflow finishes you get a fully populated plan.md with every section filled in.

#### - [ ] Phase 2.1: Redefine the plan FSM with fourteen steps

Replace the current seven-step FSM in `internal/plan/steps.go` with a fourteen-step FSM: `new → overview → discovery → architecture → components → data_structures → implementation_detail → dependencies → testing_approach → milestones → phases → open_questions → out_of_scope → verification → finished`. The old `approach` step is deleted and its responsibilities (option presentation, user agreement on chosen direction) move into the new `architecture` step. Each new step has its own callback that routes to the matching step template. The `verification` and `finished` steps are unchanged other than now being reached from `out_of_scope` instead of `phases`.

*Technical detail:* [context.md#phase-21](./context.md#phase-21-plan-fsm-steps)

**Acceptance criteria**:

- [ ] `internal/plan/steps.go` defines the fourteen steps in the order specified above
- [ ] The `approach` step and its callback are removed
- [ ] `spektacular plan steps` lists the fourteen step names in order
- [ ] Each new step callback renders its matching template under `templates/plan-steps/`

#### - [ ] Phase 2.2: Author the new step templates and renumber the existing ones

Under `templates/plan-steps/`, renumber the existing templates to fit the new order and author eight new templates (`architecture`, `components`, `data_structures`, `implementation_detail`, `dependencies`, `testing_approach`, `open_questions`, `out_of_scope`). The `architecture` template absorbs the option-presentation-and-agreement beat from the deleted `approach` step. The `implementation_detail` and `testing_approach` templates explicitly instruct the agent that this section is high-level only: sketches of new patterns, major code-shape changes, overall testing strategy — not per-phase detail. The `open_questions` template explicitly narrows the bucket to "things that genuinely cannot be answered until implementation begins" so the agent isn't tempted to stash resolvable uncertainties there. The `out_of_scope` template is a short step where the agent and user lock in the exclusions.

*Technical detail:* [context.md#phase-22](./context.md#phase-22-step-templates)

**Acceptance criteria**:

- [ ] Eight new step templates exist under `templates/plan-steps/`, one per new step
- [ ] Existing templates are renumbered to match the new step order and still render correctly
- [ ] The `architecture` template includes the option-presentation and user-agreement beat previously in `approach`
- [ ] The `implementation_detail` and `testing_approach` templates explicitly state that the content must be high-level and direct per-phase detail to context.md
- [ ] The `open_questions` template explicitly restricts the section to impl-time uncertainties

#### - [ ] Phase 2.3: Test that each new step renders the directive strings that make it correct

Add focused callback tests for the new steps. For each of `architecture`, `implementation_detail`, `testing_approach`, `open_questions`, and `out_of_scope`, assert that the rendered instruction contains the directive strings that carry the contract — the `implementation_detail` rendered instruction must contain a phrase enforcing "high-level only", the `open_questions` rendered instruction must contain the impl-time restriction, and the `architecture` rendered instruction must contain the option-presentation and user-agreement beat. Also add a regression test that walks the fourteen-step FSM from `new` to `finished` via the existing workflow engine.

*Technical detail:* [context.md#phase-23](./context.md#phase-23-workflow-contract-tests)

**Acceptance criteria**:

- [ ] Each of the five contract-bearing steps has a test asserting its required directive strings appear in the rendered instruction
- [ ] A regression test walks the fourteen-step FSM from `new` to `finished` and asserts each transition succeeds
- [ ] `spektacular plan steps` output matches the expected fourteen-step ordered list

## Changelog

<!-- Left empty — appended during implementation by the implement workflow. -->
