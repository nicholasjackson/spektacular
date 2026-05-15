# Feature: 17_plan_testing

<!--
  OVERVIEW
  A concise 2-3 sentence summary of the feature. Answer three questions:
    1. What is being built?
    2. What problem does it solve?
    3. Who benefits and why does it matter?
  Avoid implementation details — this should be readable by any stakeholder.
-->
## Overview

A harbor test that exercises the `plan` workflow end-to-end, mirroring the existing `tests/harbor/spec-workflow` test. It guards against regressions where plan code or plan-step templates drift away from their documented behaviour — in particular that sub-agent orchestration and skill invocations prescribed in the step templates actually happen at runtime. Maintainers of the plan workflow benefit by catching breakages in CI instead of in production.

<!--
  REQUIREMENTS
  Specific, testable behaviours the feature must deliver.
  Format: bold title on the checkbox line, detail indented below.
  Rules:
    - Use active voice: "Users can...", "The system must..."
    - Each requirement should be independently verifiable
    - Focus on WHAT, not HOW — avoid prescribing implementation
    - Keep each item atomic — one behaviour per line
-->
## Requirements

- **The plan workflow is exercised step by step**
  A harbor test drives the full plan workflow from start to terminal step and verifies each individual step was completed, rather than only checking the end state.

- **Steps are executed in the correct order**
  The test asserts the sequence of completed steps matches the canonical order defined by the plan state machine.

- **The agent calls the plan CLI for each step**
  For every step, the test asserts the agent invoked the corresponding `plan new` / `plan goto` command.

- **Skills referenced in a step are retrieved**
  Where a step template references a named skill, the test asserts the agent retrieved that skill during that step.

- **Sub-agents referenced in a step are called**
  Where a step template directs the agent to spawn sub-agents, the test asserts at least one sub-agent invocation occurred during that step.

- **Each step produces its expected artefact**
  Where a step is expected to produce an output file or section, the test asserts that artefact exists and contains substantive, on-topic content.

<!--
  CONSTRAINTS
  Hard boundaries the solution must operate within. These are non-negotiable.
  Examples:
    - Must integrate with the existing authentication system
    - Cannot introduce breaking changes to the public API
    - Must support the current minimum supported runtime versions
  Leave blank if there are no constraints.
-->
## Constraints


<!--
  ACCEPTANCE CRITERIA
  The specific, binary conditions that define "done".
  Format: bold title on the checkbox line, verifiable detail indented below.
  Each criterion must be:
    - Independently verifiable (pass/fail, not subjective)
    - Traceable back to a requirement above
    - Testable by someone who didn't write the code
-->
## Acceptance Criteria

- **Per-step pass/fail reporting**
  Running the harbor test produces a distinct pass/fail result for each step in the plan state machine, not a single aggregate assertion.

- **Step order matches state machine**
  The test reads `state.json`'s `completed_steps` and fails unless the list exactly equals the order returned by the plan state machine's `Steps()`.

- **Plan CLI invoked for every step**
  For each expected step, the test parses the agent transcript and fails unless it finds a `plan new` call (for the first step) or a `plan goto` call with `"step":"<step_name>"` before that step is marked complete.

- **Skills referenced in a step are retrieved during that step**
  For every plan step template under `templates/plan-steps/`, the test extracts referenced skill names and fails unless the transcript shows the agent retrieving each one (via a `Skill` tool_use or a `<cmd> skill <name>` Bash call) while that step is active.

- **Sub-agents referenced in a step are spawned during that step**
  For every plan step template that directs the agent to spawn sub-agents, the test fails unless the transcript contains at least one `Task`/`Agent` tool_use block during that step.

- **Expected artefacts exist with substantive content**
  For each step that the template says produces an output file (e.g. `research.md`), the test fails unless the file exists, contains at least a minimum character count of real content, and has no placeholder markers (TODO/TBD/FIXME).

<!--
  TECHNICAL APPROACH
  High-level technical direction to guide the planning agent. Include:
    - Key architectural decisions already made
    - Preferred patterns or technologies if known
    - Integration points with existing systems
    - Known risks or areas of uncertainty
  Leave blank if you want the planner to propose the approach.
-->
## Technical Approach

Follow the same approach as `tests/harbor/spec-workflow/` and re-use anything appropriate:

- Mirror the directory layout: `task.toml`, `instruction.md`, `environment/`, `solution/`, `tests/`.
- Re-use the pytest style, transcript-parsing helpers, and assertion patterns from `tests/harbor/spec-workflow/tests/test_spec_workflow.py` (e.g. `extract_tool_calls`, `find_spektacular_calls`, `parse_sections`), extending them where needed for plan-specific concerns (sub-agent and skill detection).
- Derive the canonical step list from the plan state machine source of truth (the equivalent of `internal/spec/steps.go` for plans) so the test stays in sync with code changes.
- Parse skill and sub-agent references directly from `templates/plan-steps/*.md` so new or renamed steps are picked up without editing the test.

<!--
  SUCCESS METRICS
  How you will know the feature is working well after delivery. Be specific:
    - Quantitative: "p99 latency < 200ms", "error rate < 0.1%"
    - Behavioural: "users complete the flow without support intervention"
  Leave blank if not applicable.
-->
## Success Metrics


<!--
  NON-GOALS
  Explicitly state what this spec does NOT cover. This is as important as
  the requirements — it prevents scope creep and sets clear expectations.
  Examples:
    - "Mobile support is out of scope (tracked in #456)"
    - "Internationalisation will be addressed in a follow-up spec"
  Leave blank if there are no explicit exclusions to call out.
-->
## Non-Goals
