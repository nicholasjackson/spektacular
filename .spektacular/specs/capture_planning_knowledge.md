# Feature: capture_planning_knowledge

<!--
  OVERVIEW
  A concise 2-3 sentence summary of the feature. Answer three questions:
    1. What is being built?
    2. What problem does it solve?
    3. Who benefits and why does it matter?
  Avoid implementation details — this should be readable by any stakeholder.
-->
## Overview

During the `plan` workflow, whenever discovery or research surfaces a piece of non-obvious knowledge — a gotcha, an architectural surprise, an undocumented constraint, or a learning a future planning agent would benefit from — the agent pauses and asks the user whether to save it into `.spektacular/knowledge/`. If the user agrees, the agent files the entry into the appropriate subdirectory (gotchas, learnings, architecture, or conventions). Today, these discoveries exist only in the current conversation and are lost when the session ends, forcing future planning agents to rediscover the same dead ends; capturing them at the moment of discovery — when context is freshest — keeps the knowledge base alive without relying on the user to remember to document things manually.

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

- **The planning agent collects candidate knowledge during discovery and research.**
  As the agent investigates during the `plan` workflow's discovery and research phases, it identifies non-obvious knowledge worth preserving — gotchas, architectural surprises, undocumented constraints, rejected alternatives with non-obvious reasons, and learnings a future cold session would benefit from. No user prompts or disk writes happen at this stage.

- **Candidates are appended to `research.md` under a dedicated `## Knowledge Candidates` section.**
  Each candidate is a short entry containing a title, the content, and the proposed target subdirectory in `.spektacular/knowledge/` (`gotchas/`, `learnings/`, `architecture/`, or `conventions.md`). Storing in `research.md` means candidates survive session crashes and cold resumes, and rehydration is automatic on restart.

- **A new `review_knowledge` step runs at the end of the `plan` workflow.**
  The step is inserted after `verification` and before `finished`, acting as the single consolidation point for knowledge capture. On entry, the agent reads the `## Knowledge Candidates` section from `research.md`.

- **The review step presents all collected candidates to the user in one pass.**
  Each candidate is shown with its title, content, and proposed target subdirectory. The user can accept, reject, or edit each entry individually.

- **Accepted candidates are written into `.spektacular/knowledge/`.**
  The agent files each accepted entry into the correct subdirectory, creating the subdirectory and file if missing. Every entry links back to the originating plan by name.

- **Before presenting a candidate, the agent checks whether `.spektacular/knowledge/` already covers it.**
  Duplicates — entries with the same title or substantially overlapping content — are dropped from the review list so the user is not asked about knowledge already captured.

- **The user can accept-all, reject-all, or skip the whole review in one response.**
  Knowledge-rich plans must not turn the review into an approval marathon; bulk actions complete the step in a single turn.

- **The review step must not block plan completion.**
  Declining or skipping advances the workflow cleanly to `finished` with no lingering state or partial writes.

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

- **Must integrate with the existing plan FSM.** The new `review_knowledge` step is added as a new FSM state via `internal/workflow` / `go-fsm`, with `Src = []string{"verification"}` and a transition to `finished`. No new FSM engine features may be introduced.

- **Must use the existing markdown-template + mustache rendering pipeline.** The new step gets its own numbered template file under `templates/steps/plan/` (e.g. `14-review_knowledge.md`) and is rendered through the shared step helpers. No bespoke rendering code.

- **Must not change the shape of JSON responses returned by `plan goto`.** The existing result type (`step`, `plan_path`, `plan_name`, `instruction`) stays the same; the only new observable thing is a new legal value for the `step` field.

- **Must not require new config keys.** The feature works with defaults from `.spektacular/config.yaml`; no new top-level keys or user-facing configuration is introduced.

- **Must be backwards-compatible with plans in progress.** A plan state file created before this feature shipped, currently parked in any existing step, must still reach `finished` without error. If no `## Knowledge Candidates` section exists in `research.md`, the review step produces an empty review and advances cleanly.

- **Must not depend on network or external services.** All candidate detection, duplicate checking, and writes happen against the local filesystem only.

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

- **Candidate collection is observable in `research.md`.**
  After a plan workflow run where discovery encountered at least one non-obvious finding, `research.md` contains a `## Knowledge Candidates` section with at least one entry. Each entry has a title, content, and target subdirectory. Traces to: "The planning agent collects candidate knowledge during discovery and research."

- **No disk writes to `.spektacular/knowledge/` happen before the review step.**
  Inspecting the directory's mtimes and file list at any point between `discovery` and `verification` shows no new or modified files attributable to the planning session. Traces to: "Candidates are appended to `research.md` under a dedicated `## Knowledge Candidates` section."

- **The `review_knowledge` step exists in the plan FSM between `verification` and `finished`.**
  Running `go run . plan goto --data '{"step":"review_knowledge"}'` from source state `verification` is accepted; running it from any other source state is rejected by the FSM. Traces to: "A new `review_knowledge` step runs at the end of the `plan` workflow."

- **The review step renders all non-duplicate candidates in a single instruction.**
  When the agent enters `review_knowledge`, the returned instruction contains every candidate from `research.md` that is not already represented in `.spektacular/knowledge/`. Traces to: "The review step presents all collected candidates to the user in one pass."

- **Accepted candidates end up on disk in the correct subdirectory.**
  When the user accepts a candidate targeted at `gotchas/`, a file exists under `.spektacular/knowledge/gotchas/` after the step completes, containing the candidate's title, content, and a reference back to the plan name. Traces to: "Accepted candidates are written into `.spektacular/knowledge/`."

- **Rejected candidates leave no trace on disk.**
  After a review where the user rejects a candidate, no file corresponding to that candidate exists anywhere under `.spektacular/knowledge/`. Traces to: "Accepted candidates are written into `.spektacular/knowledge/`."

- **Duplicate detection removes candidates from the review list.**
  When a candidate's title matches an existing file name or its content substantially overlaps an existing entry, running the step produces an instruction that omits that candidate. Traces to: "Before presenting a candidate, the agent checks whether `.spektacular/knowledge/` already covers it."

- **Bulk-action commands terminate the review in one turn.**
  When the user responds with "accept all", "reject all", or "skip", the next `goto` call advances to `finished` with no follow-up prompts. Traces to: "The user can accept-all, reject-all, or skip the whole review in one response."

- **Skipping the review still reaches `finished`.**
  Running the `review_knowledge` step with the user declining or skipping, followed by `goto finished`, leaves the FSM in `finished` with `.spektacular/knowledge/` unchanged. Traces to: "The review step must not block plan completion."

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

- **Add the review as a new FSM state inside the existing `plan` workflow**, not a separate workflow. It enters from `verification` and exits to `finished`.

- **Candidate collection is driven by template wording, not Go plumbing.** The discovery/research template instructs the planning agent to append candidates to `research.md` under a dedicated section as it investigates — no new code paths for capture.

- **The review step's logic lives in the rendered instruction and is executed by the agent.** No new Go beyond registering the FSM state and its template; the agent reads candidates from `research.md`, dedups against `.spektacular/knowledge/`, and writes accepted entries.

- **Knowledge-base file layout follows the existing pattern** — one-per-entry markdown files in `gotchas/`, `learnings/`, and `architecture/`, with `conventions.md` as a single appended file. Each written entry carries a link back to the originating plan.

- **Known risks**: duplicate detection is fuzzy beyond exact-title matches; adding a new plan step may touch test fixtures that reference step filenames or FSM source lists; existing `research.md` files may already use a heading the candidates section would collide with.

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

- **Knowledge capture in the `spec` workflow is out of scope.** This spec only addresses capture during `plan`. If spec-time capture is wanted later, it will be handled by a separate spec.
- **Knowledge capture in the (future) `implement` workflow is out of scope.** The `implement` workflow is being designed in plan `15_implementation` and may add its own knowledge-capture step; that design is not this spec's problem.
- **No back-filling from past plans.** This feature only captures knowledge from new plan runs going forward; it does not retroactively scan existing `research.md` files to mine candidates.
- **No standalone UI or CLI for editing the knowledge base.** Users edit `.spektacular/knowledge/` with their normal editor; this feature only writes from the review step and only reads for the duplicate check.
- **No quality validation of captured content.** The user is the gatekeeper during review; the agent does not grade, score, or rewrite what it proposes.
- **No deletion or pruning of existing knowledge.** The review step only adds entries; removing stale entries is a separate concern handled outside this spec.
