# Feature: 16_plan_format

<!--
  OVERVIEW
  A concise 2-3 sentence summary of the feature. Answer three questions:
    1. What is being built?
    2. What problem does it solve?
    3. Who benefits and why does it matter?
  Avoid implementation details — this should be readable by any stakeholder.
-->
## Overview

Restructure Spektacular's plan generation so that generated plans follow a standard technical design document format used in professional software development. This ensures plans contain enough detail (architecture, components, data structures, implementation sketches, testing, dependencies, open questions, out-of-scope) for another developer to pick up the work and implement it without needing the original planner. Developers handing off or inheriting implementation work benefit from a consistent, complete, and actionable plan format.

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

- [ ] **plan.md contains the bulk of the implementation information**
  A developer reading plan.md alone must have most of what they need to implement the work. context.md and research.md are supporting documents, not primary sources.

- [ ] **plan.md template must include ten sections in order**
  Overview, Architecture & Design Decisions, Component Breakdown, Data Structures & Interfaces, Implementation Detail, Dependencies, Testing Approach, Milestones & Phases, Open Questions, Out of Scope.

- [ ] **The plan workflow steps must map to these sections**
  The agent must populate each section in sequence, one step per section (or logical grouping), following the same goto-based workflow used today.

- [ ] **Milestones & Phases must retain their existing structure**
  Milestone → phase → acceptance criteria checkboxes, with `*Technical detail:*` links into context.md preserved.

- [ ] **Implementation Detail in plan.md must be high-level**
  Sketches of new patterns, major code-shape changes, and code-structure UX — enough for a reviewer to spot missing patterns. Not per-phase file:line work (that stays in context.md).

- [ ] **Testing Approach in plan.md must be a high-level overview**
  Overall testing strategy and test types. Per-phase testing detail remains in context.md.

- [ ] **Each plan section must include inline guidance**
  HTML comments explaining what content belongs in that section, following the pattern used by the spec template.

- [ ] **research.md must be preserved**
  Continues to hold raw planning analysis — papers read, packages analysed, alternatives considered, evidence for the chosen approach.

- [ ] **context.md must be preserved**
  Continues to hold tribal knowledge (working patterns, idiomatic code details not inferable from plan.md) and file-level detail that sharpens the implementation approach.

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

- [ ] **Plan template sections present**
  `templates/plan-scaffold.md` contains all ten section headings in order (Overview, Architecture & Design Decisions, Component Breakdown, Data Structures & Interfaces, Implementation Detail, Dependencies, Testing Approach, Milestones & Phases, Open Questions, Out of Scope), each preceded by an HTML comment explaining its purpose.

- [ ] **Workflow step mapping**
  `go run . plan goto --data '{"step":"<name>"}'` can be invoked for each section, and each invocation returns an instruction targeting the matching section in plan.md.

- [ ] **Milestones & Phases structurally unchanged**
  A newly generated plan.md contains Milestone 1 scaffolding with phase, acceptance-criteria checkbox list, and a `*Technical detail:*` link into context.md.

- [ ] **Implementation Detail guidance is high-level**
  The inline HTML comment in plan.md's Implementation Detail section explicitly directs the agent to sketch new patterns, major code-shape changes, and code structure — not per-phase file:line changes.

- [ ] **Testing Approach guidance is high-level**
  The inline HTML comment in plan.md's Testing Approach section directs the agent to describe overall strategy and test types, explicitly excluding per-phase testing detail.

- [ ] **Supporting docs preserved**
  Running the plan workflow still produces `research.md` and `context.md` alongside `plan.md`, each populated from their existing scaffold templates.

- [ ] **Self-contained handoff**
  A generated plan.md has no required `see context.md for …` references in the Overview, Architecture, Component Breakdown, Data Structures, Dependencies, Open Questions, or Out of Scope sections. Cross-references are only permitted from Milestones & Phases `*Technical detail:*` lines, matching current behaviour.

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


<!--
  SUCCESS METRICS
  How you will know the feature is working well after delivery. Be specific:
    - Quantitative: "p99 latency < 200ms", "error rate < 0.1%"
    - Behavioural: "users complete the flow without support intervention"
  Leave blank if not applicable.
-->
## Success Metrics

- A developer unfamiliar with the spec can read plan.md alone and produce a working implementation (handoff test).
- Generated plans from real specs include populated content in every section, with no skipped or empty sections.
- Reviewers can spot missing architectural patterns or design gaps from plan.md without needing to read context.md.

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

- Restructuring `research.md` or `context.md` scaffolds — they remain as they are today.
- Migrating existing plans to the new format — only newly generated plans use the new structure.
- Validation of section content (e.g., refusing to advance if a section is empty) — the workflow advances on user signal, as today.
- Adding new top-level CLI commands — only changes to the existing `plan` workflow templates and step definitions.
