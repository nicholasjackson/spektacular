# Plan: 17_plan_testing

<!-- Metadata -->
<!-- Created: 2026-04-13T14:35:01Z -->
<!-- Commit: 4a94cf8b114091436a57f0f2ae81b6fa9298d1dc -->
<!-- Branch: skills -->
<!-- Repository: git@github.com:jumppad-labs/spektacular.git -->

## Overview

A harbor integration test that exercises the `plan` workflow end-to-end and guards against regressions in the Go state machine and the step templates. It mirrors `tests/harbor/spec-workflow/` in structure, hand-maintains its expectation maps as independent behavioural oracles (step order, skills per step, sub-agent spawn steps, plan.md sections), and asserts per-step that the agent followed the rendered instructions, retrieved each expected skill, spawned sub-agents where required, and produced substantive artefact files. Maintainers of the plan workflow benefit by catching drift in CI instead of in production runs.

## Architecture & Design Decisions

A new harbor task `tests/harbor/plan-workflow/` mirrors the directory layout and pytest conventions of `tests/harbor/spec-workflow/`: `task.toml`, `instruction.md`, `environment/Dockerfile`, `solution/solve.sh`, `tests/test.sh`, `tests/test_plan_workflow.py`. The agent runs `spektacular init claude`, then drives the full plan workflow against a seeded spec from `plan new` through `finished`. The verifier is a pytest suite that parses the agent's JSONL transcript, loads the persisted workflow state, reads the produced artefact files, and asserts per-step behaviour against **hand-maintained expectation maps** that live at the top of the verifier module and serve as the independent behavioural oracle.

Three load-bearing decisions shape this design:

1. **All expectation maps are hand-maintained, not derived at runtime.** The verifier carries `EXPECTED_STEP_ORDER`, `EXPECTED_SKILLS_PER_STEP`, `EXPECTED_SPAWN_STEPS`, and `EXPECTED_PLAN_SECTIONS` as literal constants — exactly the pattern `test_spec_workflow.py:21-31` uses. Deriving any of them at runtime (from `spektacular plan steps`, from `templates/plan-steps/*.md`, or from any other source under test) would make the verifier tautological: "the agent did what the subject said" is a closed loop, not a behavioural check. When a legitimate change lands in the state machine or a step template, the maintainer updates the corresponding map in the same commit — that update is the behavioural confirmation.

2. **Step order bugs surface through `completed_steps`, not a separate introspection test.** Because the agent follows the state machine's order at runtime, any reorder in `Steps()` surfaces as a diff between `completed_steps` and `EXPECTED_STEP_ORDER` in the `test_steps_executed_in_order` assertion. No separate `plan steps` cross-check is needed — that would be a second assertion catching the same bug class.

3. **Transcript parsing detects `Bash`, `Skill`, and sub-agent tool uses with step-scoped windowing.** The existing spec test only extracts `Bash` `tool_use` blocks whole-transcript. The plan test extends extraction to also capture `Skill` and `Task`/`Agent` blocks, and partitions the transcript into per-step windows keyed by the step name each `plan new`/`plan goto` Bash call transitions *into*. A step "retrieves skill X" if the transcript contains a matching `Skill` or `spektacular skill X` Bash call within that step's window; a step "spawns a sub-agent" if a `Task`/`Agent` block falls in that window. The per-step attribution is what makes "retrieved skill X *during that step*" a meaningful assertion rather than a whole-transcript hunt.

## Component Breakdown

- **`plan-workflow` harbor task directory** — a new `tests/harbor/plan-workflow/` sibling to the existing `spec-workflow/` task. Owns the task definition, agent instruction, environment Dockerfile, reference solution, and verifier entry script. Its structure mirrors `spec-workflow/` so the harbor CLI and the Makefile can invoke it without bespoke tooling.

- **Seed spec fixture** — a deterministic fixture spec created inside the container during the test so the agent has a concrete spec to plan against. Belongs to the environment; drives which topic keywords the verifier expects to see in drafted plan sections.

- **Hand-maintained expectation maps** — module-level constants at the top of `test_plan_workflow.py`: `EXPECTED_STEP_ORDER`, `EXPECTED_SKILLS_PER_STEP`, `EXPECTED_SPAWN_STEPS`, `EXPECTED_PLAN_SECTIONS`. The independent behavioural oracle the verifier asserts against. Updated by the maintainer in the same commit as any legitimate state-machine or template change.

- **Transcript extractor (extended from `test_spec_workflow.py`)** — generalises the existing `extract_tool_calls` helper so it captures `Bash`, `Skill`, and `Task`/`Agent` tool_use blocks in a flat ordered list instead of `Bash` only. Downstream helpers project it through filters: `bash_calls_only`, `skill_calls`, `agent_spawn_calls`.

- **Step window resolver** — maps each step name to a `[start_index, end_index)` slice of the transcript so "retrieved skill X during step Y" and "spawned agent during step Y" can be asserted unambiguously. Windows are derived from the positions of `spektacular plan new` / `plan goto --data '{"step":"<name>"}'` calls in the transcript.

- **`EXPECTED_STEP_ORDER` constant** — a hand-maintained list of plan step names in canonical order at the top of `test_plan_workflow.py`. Serves as the behavioural oracle for the `completed_steps` order check. Mirrors the `test_spec_workflow.py:21-31` convention.

- **Per-step pytest class family** — one pytest class per discovered step, plus a `TestWorkflow` class for cross-cutting invariants (workflow finished, all steps completed, steps in order, no placeholder markers, rendered `next_step` validity). Step classes are generated via `@pytest.mark.parametrize` over the state-machine introspector output so adding a step to `Steps()` adds test cases automatically.

- **Artefact readers** — small helpers that read `plan.md`, `context.md`, and `research.md` from `.spektacular/plans/<name>/` after the workflow has reached `finished`, and split each file into sections by `##`/`###` heading (reusing the spec test's `parse_sections` approach extended for subsections).

- **Verifier entry script (`tests/test.sh`)** — unchanged in shape from the spec-workflow equivalent: runs pytest, writes `1`/`0` to `/logs/verifier/reward.txt`.

- **Makefile target** — a new `plan-harbor-test` target that cross-compiles the linux/amd64 binary into the new task's environment directory and runs `harbor run`, mirroring the existing `harbor-test` target.

## Data Structures & Interfaces

This plan introduces no new Go types. It adds a small set of Python dataclasses inside the verifier module:

```python
@dataclass(frozen=True)
class StepExpectations:
    name: str
    skills: frozenset[str]
    spawns_agents: bool
    template_path: Path

@dataclass(frozen=True)
class ToolCall:
    index: int
    type: str        # "Bash" | "Skill" | "Task" | "Agent"
    name: str
    input: dict

@dataclass(frozen=True)
class StepWindow:
    step: str
    start: int       # inclusive
    end: int         # exclusive
```

The canonical step list is a plain `list[str]` literal declared at the top of the verifier module (the `EXPECTED_STEP_ORDER` constant). The verifier does not consume `spektacular plan steps` — the agent follows the state machine's order at runtime, so any reorder surfaces as a diff between `completed_steps` and `EXPECTED_STEP_ORDER` without needing a second introspection path. The artefact sections map is a `dict[str, str]` keyed by lowercase heading, identical in shape to the spec test's `parse_sections` return value, extended to accept `###` subsections so the phases block under plan.md can be checked per-phase.

No changes to any existing Go struct. The verifier depends on the current JSON shapes of `workflow.State`, `plan.Result`, `plan.StatusResult`, and `plan.StepsResult` as contracts.

## Implementation Detail

This plan introduces one new pattern and extends one existing pattern; no major code-shape change elsewhere.

**Pattern followed: hand-maintained expectation maps as the oracle.** Like `test_spec_workflow.py`, the verifier hard-codes its per-step expectations (step order, skill retrievals, sub-agent spawn steps, plan.md section names) as literal constants. When a legitimate change lands in the state machine or a template, the maintainer updates the constant in the same commit. Parameterised tests iterate over the constants to produce per-step test cases so pytest -v lists each step's assertions separately.

**Extended pattern: step-window attribution of transcript events.** The existing spec verifier checks "did the agent make call X?" as a whole-transcript question. This plan introduces the tighter question "did the agent make call X *while step Y was active?*", which is needed to assert "skills referenced in a step are retrieved during that step". The shape: walk the transcript once, identify every `spektacular plan new` / `plan goto --data '{"step":"<name>"}'` bash call, and partition the transcript into `[entry_call, next_entry_call)` windows keyed by the step name the call transitions into.

**Code-structure UX.** A developer reading `test_plan_workflow.py` sees three clearly named sections: a top block of constants and helper functions (transcript extractor, step-window resolver, artefact readers), a middle block of lazy module-level caches so transcript parsing runs once per session, and a bottom block of pytest classes organised by behaviour. This matches the shape of `test_spec_workflow.py` and should feel familiar.

**Existing patterns followed.** Harbor task layout, pytest verifier entry script writing reward to `/logs/verifier/reward.txt`, Dockerfile base image and `uv`/pytest install, agent transcript path at `/logs/agent/claude-code.txt`, HTML-comment-stripping section parser. All reused verbatim.

**No new module boundary and no refactors to existing Go code.** The verifier is pure Python and the only Go-side touchpoint is the existing `spektacular plan steps` command. The `discovery → approach` drift bug is a prerequisite fix but not part of this plan's implementation scope beyond the one-line Phase 1.1 change.

## Dependencies

- **`harbor` CLI** — external; orchestrates the agent and verifier containers. Already a dependency of `make harbor-test`. No version change.
- **`uv` + Python 3.12 + `pytest==8.4.1`** — external; installed via the existing `spec-workflow/environment/Dockerfile` pattern. No new versions.
- **`spektacular` binary** — internal; cross-compiled into the new task's environment directory by the Makefile, exactly as `spec-workflow` does.
- **`spektacular plan steps`** — internal, defined at `cmd/plan.go:253`; contract dependency on its `{"steps": [...]}` JSON shape.
- **`spektacular plan status`** — internal, defined at `cmd/plan.go:218`; secondary cross-check on state fields.
- **Claude Code agent transcript format** — external contract at `/logs/agent/claude-code.txt`. Depends on `Bash`, `Skill`, and `Task`/`Agent` `tool_use` block types.
- **Spec `.spektacular/specs/17_plan_testing.md`** — source of truth; no changes required.
- **Existing `tests/harbor/spec-workflow/` task** — reference implementation whose helpers and structure are ported and extended. Not modified.

**Must-land-first prerequisite.** Fix `internal/plan/steps.go:65` to pass `"architecture"` as the discovery step's next_step instead of `"approach"`. Bundled into this plan as Phase 1.1 so the harbor test is not blocked on an unrelated bug.

**No new external dependencies.** No new Python packages beyond `pytest`, no new Go modules, no new container base images.

## Testing Approach

This plan **is** a test — the deliverable itself is a harbor integration test. One new end-to-end harbor task sits alongside `tests/harbor/spec-workflow/` and is the only test added. No new Go unit tests.

**Load-bearing assertions, in plain language.**

- The workflow reaches `finished`. If not, fail with the stuck step name.
- Every step in the verifier's hand-maintained `EXPECTED_STEP_ORDER` appears in `completed_steps` exactly once, in the same order. Because the agent follows the state machine's order at runtime, this single assertion also catches reordering bugs in `Steps()`.
- For every step, the agent invoked the corresponding plan CLI call during that step's window.
- For every skill named in a step's template, the agent retrieved it during that step's transcript window.
- For every step whose template expects sub-agent spawning, the agent spawned at least one `Task`/`Agent` tool_use during that step's window.
- The three artefact files exist with substantive on-topic content and no placeholder markers.
- Every plan.md section implied by the templates exists with content above a minimum character count.
- Cross-cutting invariant: every rendered `instruction`'s `next_step` directive references a step that `plan goto` actually accepts. This is the formal guard for the class of bug that caused Phase 1.1 to exist.

**Pass/fail granularity.** Per-step. `pytest -v` lists each step-class test separately; `/logs/verifier/reward.txt` reflects the overall exit code. A failure in one step does not mask failures elsewhere.

**Coverage distribution.** Most coverage is concentrated on the steps with non-trivial template content: `discovery` (multiple skills + sub-agent spawn), `phases` (skill reference), `verification` and `finished` (artefact file creation). Trivial steps get the baseline assertions (step completed, CLI call made, plan.md section present).

**Existing conventions followed.** Harbor task layout, pytest module layout, `/logs/verifier/reward.txt` contract, transcript parsing at `/logs/agent/claude-code.txt`, section parser stripping HTML comments. The test reads as a sibling of `test_spec_workflow.py`, not a rewrite.

**Deliberate gaps.** No cross-model runs. No unit tests for the verifier's Python helpers. No semantic-quality judgement of drafted content beyond length and topic keywords. No auto-repair of detected drift.

## Milestones & Phases

### Milestone 1: Template-drift prep fix

**What changes**: The one-line typo in `internal/plan/steps.go` where the discovery step passes `"approach"` as the next step name — a step that does not exist in the state machine — is corrected to `"architecture"`. After this milestone, any maintainer running the plan workflow end-to-end by copying the `next_step` command from the rendered instructions reaches `finished` without needing to know the correct step name. Internal but user-visible in the sense that the workflow stops silently misdirecting its own operator, and lands first so the harbor test is not blocked on an unrelated bug.

#### - [ ] Phase 1.1: Fix discovery → architecture `next_step`

The discovery step in the plan state machine instructs the agent to advance to a step named `"approach"`, which does not exist. The valid next step is `"architecture"`. This single-token typo in the Go source prevents any happy-path run of the plan workflow from completing without the operator manually knowing the correct step name. The fix is a one-line string change; no templates, no tests, no other call sites are touched.

*Technical detail:* [context.md#phase-11](./context.md#phase-11-fix-discovery-next_step)

**Acceptance criteria**:

- [ ] Running the plan workflow from `plan new` through to `finished` by copying the `next_step` command printed in each rendered instruction succeeds without errors.
- [ ] The discovery step's rendered instruction ends with a `plan goto --data '{"step":"architecture"}'` call.
- [ ] Existing `internal/plan` Go tests continue to pass.

### Milestone 2: Harbor plan-workflow test exists and runs

**What changes**: A new harbor task exists under `tests/harbor/plan-workflow/` and can be run via a new Makefile target. The test drives the full plan workflow end-to-end against a seeded spec using a reference `solve.sh`, and a baseline pytest verifier asserts the workflow reached `finished`, `completed_steps` matches the state machine's canonical step order, and every step's `plan new`/`plan goto` CLI call appears in the agent transcript. After this milestone a maintainer running the new Makefile target sees a passing pytest run for the plan workflow the same way they run the spec-workflow test today.

#### - [ ] Phase 2.1: Harbor task scaffold

Create the new `tests/harbor/plan-workflow/` task directory with the five standard harbor files, modelled as a sibling of the existing `spec-workflow` task. At this phase the files exist, the Dockerfile builds a container with `spektacular` on `PATH` and the `templates/plan-steps/` directory mounted at a known location, and a stub pytest verifier runs and fails loudly with a clear "not implemented" message. The reference `solve.sh` drives the plan workflow end-to-end to `finished` using the happy-path CLI sequence.

*Technical detail:* [context.md#phase-21](./context.md#phase-21-harbor-task-scaffold)

**Acceptance criteria**:

- [ ] `tests/harbor/plan-workflow/` contains `task.toml`, `instruction.md`, `environment/Dockerfile`, `solution/solve.sh`, `tests/test.sh`, and a `tests/test_plan_workflow.py` module that imports cleanly under pytest.
- [ ] Running the reference solution against the built container produces three artefact files under `.spektacular/plans/<name>/`.
- [ ] The Dockerfile copies `templates/plan-steps/` into the verifier container at a documented path the verifier reads from.

#### - [ ] Phase 2.2: Baseline pytest verifier

Implement the verifier's foundation: declare a hand-maintained `EXPECTED_STEP_ORDER` constant at the top of the module, load `.spektacular/plan-<name>/state.json`, extract `Bash` tool calls from the agent transcript, and assert that the workflow reached `finished`, that `completed_steps` equals `EXPECTED_STEP_ORDER` in order, and that for every step the agent invoked the correct `plan new` or `plan goto --data '{"step":"<name>"}'` Bash call. Because the agent follows the state machine's order at runtime, the `completed_steps == EXPECTED_STEP_ORDER` assertion catches any reordering bug in `Steps()` without a separate introspection test. These assertions mirror the existing spec test's baseline layer.

*Technical detail:* [context.md#phase-22](./context.md#phase-22-baseline-pytest-verifier)

**Acceptance criteria**:

- [ ] Running the harbor task with the reference solution produces a passing pytest run where every plan step has its own named test case in the output.
- [ ] Reordering two entries in `internal/plan/steps.go` Steps() causes `test_steps_executed_in_order` to fail with a clear diff naming the divergence from `EXPECTED_STEP_ORDER`.
- [ ] Removing one `plan goto` call from the reference solution causes exactly that step's CLI-call test to fail without masking the others.

#### - [ ] Phase 2.3: Makefile wiring

Add a new Makefile target that cross-compiles the linux/amd64 binary into the new task's environment directory, runs `harbor run` against the new task, and prints the verifier's stdout — matching the shape of the existing `harbor-test` target. The new target is independent so maintainers can run the two harbor tests separately.

*Technical detail:* [context.md#phase-23](./context.md#phase-23-makefile-wiring)

**Acceptance criteria**:

- [ ] `make plan-harbor-test` builds the binary, runs the harbor task, and prints the pytest output.
- [ ] The existing `make harbor-test` target continues to work unchanged.
- [ ] Running both targets from a clean checkout produces independent job directories under `tests/harbor/jobs/`.

### Milestone 3: Skill retrieval and sub-agent spawn assertions

**What changes**: The verifier gains per-step skill-retrieval and sub-agent-spawn assertions, driven by hand-maintained `EXPECTED_SKILLS_PER_STEP` and `EXPECTED_SPAWN_STEPS` constants. For every (step, skill) pair in the map, the verifier asserts the agent retrieved that skill during the step's transcript window; for every step in `EXPECTED_SPAWN_STEPS`, it asserts at least one sub-agent was spawned during that step's window. After this milestone the maintainer sees per-step skill and sub-agent tests in the pytest output, and any regression where the agent stops retrieving an expected skill or stops spawning a sub-agent fails the corresponding assertion with a clear message.

#### - [ ] Phase 3.1: Step-window resolver

Add the `resolve_step_windows` helper that walks the flat transcript tool-call list, identifies every `spektacular plan new` / `plan goto --data '{"step":"<name>"}'` Bash call, and partitions the transcript into `[entry_call, next_entry_call)` windows keyed by the step name each call transitions into. The implicit `overview` transition inside `plan new` shares `new`'s window so assertions against `overview` still resolve.

*Technical detail:* [context.md#phase-31](./context.md#phase-31-step-window-resolver)

**Acceptance criteria**:

- [ ] The window resolver produces a window for every step in `EXPECTED_STEP_ORDER` with monotonically increasing start indexes.
- [ ] Windows for unseen steps are absent from the map; tests that query an unseen step produce a clear "step not entered" failure message.

#### - [ ] Phase 3.2: Skill retrieval assertions

Extend the transcript extractor to capture `Skill` tool_use blocks and `spektacular skill <name>` Bash calls, then add a parameterised test that iterates over each `(step, expected_skill)` pair from `EXPECTED_SKILLS_PER_STEP` and asserts the transcript contains a matching retrieval within that step's window. Failures name both the step and the missing skill.

*Technical detail:* [context.md#phase-32](./context.md#phase-32-skill-retrieval-assertions)

**Acceptance criteria**:

- [ ] A reference solution that exercises every skill in `EXPECTED_SKILLS_PER_STEP` passes every skill assertion.
- [ ] Removing one skill from the reference solution causes exactly the matching `(step, skill)` test to fail without masking the others.
- [ ] Adding a new skill to `EXPECTED_SKILLS_PER_STEP` without the agent retrieving it causes a new failing assertion on the next run.

#### - [ ] Phase 3.3: Sub-agent spawn assertions

Extend the extractor to capture `Task` and `Agent` tool_use blocks, then add a parameterised test that iterates over every step in `EXPECTED_SPAWN_STEPS` and asserts at least one such block falls in that step's window.

*Technical detail:* [context.md#phase-33](./context.md#phase-33-sub-agent-spawn-assertions)

**Acceptance criteria**:

- [ ] The reference agent run spawns at least one sub-agent during the discovery step and passes the sub-agent assertion.
- [ ] Forcing the agent to skip sub-agent spawning during discovery causes the discovery sub-agent test to fail with a clear "expected sub-agent spawn but found none" message.
- [ ] Adding a step to `EXPECTED_SPAWN_STEPS` that doesn't actually spawn causes a new failing assertion to appear.

### Milestone 4: Artefact and instruction validity assertions

**What changes**: The final layer of assertions lands. The verifier checks that the three artefact files exist with substantive on-topic content, every expected plan.md section is present with enough content, no placeholder markers appear anywhere, and every rendered `next_step` instruction references a valid state-machine step. After this milestone the test covers every spec acceptance criterion and can be trusted as CI-grade signal against the plan workflow.

#### - [ ] Phase 4.1: Artefact content assertions

After the workflow reaches `finished`, read `plan.md`, `context.md`, and `research.md` from `.spektacular/plans/<name>/`, split them into sections, and assert each expected section exists with content above a minimum length and without placeholder markers. Expected section names come from the step templates, not a hand-maintained list. Per-section topic keywords are checked where unambiguous.

*Technical detail:* [context.md#phase-41](./context.md#phase-41-artefact-content-assertions)

**Acceptance criteria**:

- [ ] The reference solution produces a plan.md that passes every per-section assertion.
- [ ] Removing the Component Breakdown section from the reference solution's plan.md causes exactly that test to fail.
- [ ] Leaving a `TODO` in any artefact file causes the placeholder test to fail with a message naming the offending file and marker.

#### - [ ] Phase 4.2: Instruction `next_step` validity invariant

Add a single cross-cutting test that captures the `instruction` field from each `plan new`/`plan goto` JSON response during the run and asserts the `{"step":"<X>"}` target of the `next_step` command literally printed in it matches a step the state machine accepts. This is the formal guard for the Phase 1.1 class of bug and fires once per step in the canonical list so a regression is located precisely.

*Technical detail:* [context.md#phase-42](./context.md#phase-42-instruction-next_step-validity-invariant)

**Acceptance criteria**:

- [ ] The reference solution passes this invariant.
- [ ] Temporarily restoring `nextStep="approach"` in `internal/plan/steps.go` causes the invariant to fail with a message naming the discovery step and the invalid target.
- [ ] The invariant fires once per step, so regressions are located precisely.

## Open Questions

This section is intentionally empty — every question surfaced during planning was resolved by reading the code, confirming a contract, or running a quick experiment. An empty Open Questions section is the expected healthy outcome of a completed planning pass.

## Out of Scope

- **Cross-model harness runs.** The new harbor test hard-codes `claude-sonnet-4-6`, matching the existing spec-workflow test. Running against other models is a separate concern.
- **Gating CI on the plan-workflow test.** The spec requires the test to exist and produce per-step pass/fail signal. When and where it runs in CI is deferred to a follow-up discussion.
- **Unit tests for the verifier's Python helpers.** Exercised only through their end-to-end use inside the harbor run. A standalone helper test file is noted as a possible follow-up.
- **Semantic validation of draft section content beyond length + topic keywords.** Deeper LLM-as-judge evaluation is out of scope.
- **Automated fixing of template drift.** When the `next_step` validity invariant fires, the test reports but does not repair.
- **Deriving expectation maps from templates or state machine at runtime.** Explicitly rejected — see Architecture decision #1. The maintenance cost of hand-maintained oracles is the point.
- **A `spek:implement` harbor test.** A separate test against a separate workflow. Helpers introduced by this plan are deliberately workflow-agnostic so they can be reused.
- **Changes to the plan Go code beyond the one-line `next_step` typo fix.** The harbor test is a pure verifier; it does not touch the state machine or the CLI contract beyond Phase 1.1.
- **Shared templates mount path standardisation across harbor tasks.** Future consistency concern, not tackled here.
- **Parallel or sharded harbor runs.** A single serial container run is sufficient for CI-grade signal.
