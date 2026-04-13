# Feature: 15_implementation

## Overview

Add an implementation workflow to spektacular that guides an agent through executing an approved plan to produce source code. It mirrors the existing `spec` and `plan` workflows — a sequence of markdown templates rendered via mustache and returned as JSON instructions — so any LLM agent can systematically work through a plan's phases, loop over unchecked items, and update the plan and changelog as it goes. Developers using spektacular benefit by getting repeatable, token-efficient, resumable plan execution without bespoke per-project orchestration.

## Requirements

- **Invoke the workflow against a plan file**
  Users can run `spektacular implement new --data '{"plan":"<path>"}'` to start the workflow, and `spektacular implement goto --data '{"step":"<id>"}'` to advance.
- **Read the plan fully before acting**
  The system must instruct the agent to read the plan file with no limit or offset.
- **Detect and read adjacent changelog**
  The system must check for `changelog.md` next to the plan and adapt instructions for first-phase vs subsequent-phase execution.
- **Delegate research via existing skill references**
  The system must instruct the agent to use its agent-orchestration capability (referencing existing `spawn-implementation-agents` and related skills) to analyze the codebase before implementing.
- **Delegate test generation**
  The system must instruct the agent to delegate test writing to a separate agent rather than authoring tests in the main context.
- **Delegate verification**
  The system must instruct the agent to delegate running the plan's success-criteria commands to a separate agent and receive a concise pass/fail summary.
- **Update plan checkboxes as work completes**
  The system must instruct the agent to mark phase checkboxes complete in the plan file as each criterion passes.
- **Append changelog entries per phase**
  The system must instruct the agent to append a phase entry (what was done, deviations, files changed, discoveries) after each phase, and a FINAL SUMMARY after the last phase.
- **Update repo-level CHANGELOG.md on completion**
  After all phases are complete, the system must instruct the agent to append a new section to the repo-level `CHANGELOG.md` using the plan name as the section header and a short user-facing summary of the overall change. If `CHANGELOG.md` does not exist, the agent creates it. This happens exactly once per plan, as the final step before the terminal state.
- **Stop on mismatch**
  The system must include an explicit STOP-and-ask rule when the plan conflicts with reality.
- **Loop over phases using plan state**
  The workflow must loop through phases by re-reading the plan's checkboxes as the source of truth for progress, with no duplicate per-phase state.
- **Confirmation by default, autonomous on user instruction**
  The end-of-phase instruction must tell the agent to check for remaining unchecked phases and ask the user whether to continue — unless the user has already told the agent to run without prompting, in which case it proceeds directly.
- **Advance via next_step in JSON responses**
  Each step's JSON output must contain the literal next command for the agent to invoke.

## Constraints

- Must follow the existing spec-workflow pattern: markdown templates under `templates/`, mustache rendering, stateless JSON responses per step.
- Must reuse `internal/workflow`, `internal/store`, and the existing `templates.FS` embed — no new dependencies.
- Must align with the existing `plan new` / `plan goto` command surface conventions.
- Must not introduce a new sub-agent mechanism; templates reference existing skills via `{{config.command}} skill <name>` exactly like `plan-steps/discovery.md` and `plan-steps/phases.md`.
- Must not require persistent session state beyond the workflow state file.

## Acceptance Criteria

- **Invocation returns a valid instruction**
  Given a valid plan path, `implement new` exits 0 and emits JSON containing `step`, `plan_path`, and a non-empty `instruction`; given an invalid/missing plan path, it exits non-zero with an error.
- **Instruction directs full plan read**
  The rendered instruction text for the plan-read step contains a directive to read the entire plan file with no limit or offset.
- **Changelog awareness**
  When `changelog.md` exists next to the plan, the rendered instruction references it; when it does not, the instruction states this is the first phase.
- **Analysis delegation referenced**
  The implementation step instruction tells the agent to use agent orchestration and references the existing `spawn-implementation-agents` skill via `{{config.command}} skill spawn-implementation-agents`.
- **Test delegation referenced**
  The test step instruction forbids writing tests in the main context and directs the agent to spawn a dedicated test-writing agent.
- **Verify delegation referenced**
  The verify step instruction directs the agent to spawn a verification agent and return a concise pass/fail summary.
- **Checkbox update directive**
  The post-verify instruction tells the agent to mark plan checkboxes complete as criteria pass.
- **Changelog append directive**
  The changelog step instruction specifies the exact fields to append (what was done, deviations, files changed, discoveries) and a FINAL SUMMARY on the last phase.
- **Stop-on-mismatch directive**
  Instructions contain an explicit STOP-and-ask rule when the plan conflicts with reality.
- **Loop back on remaining phases**
  The end-of-phase instruction tells the agent to check for unchecked phases in the plan, prompt the user (unless previously told not to), and either call `goto --data '{"step":"analyze"}'` again or advance to `update_repo_changelog`.
- **Repo CHANGELOG.md append directive**
  The rendered `update_repo_changelog` instruction text references `CHANGELOG.md` at the repo root, directs the agent to use the plan name as the section header, directs a short user-facing summary of the overall change, instructs creation of the file if it does not exist, and instructs the agent to prepend the new entry above any existing sections. The instruction ends by directing the agent to call `goto --data '{"step":"finished"}'`.
- **Terminal state**
  After the final step, `goto` returns a terminal marker so the agent knows to stop.

## Technical Approach

**Command surface**: New `implement` command group mirroring `plan` and `spec`:
- `spektacular implement new --data '{"plan":"<path>"}'` — validates the plan path, initializes state, returns the first instruction.
- `spektacular implement goto --data '{"step":"<id>"}'` — advances to the named step.

**Code layout** (mirror `internal/plan/`):
- `internal/implement/steps.go` — step definitions using existing `workflow.StepConfig`.
- `internal/implement/result.go` — result struct (Step, PlanPath, Instruction, ...).
- `cmd/implement.go` — Cobra command wiring modeled on `cmd/plan.go`.
- `templates/implement-steps/*.md` — mustache templates, one per step, registered in the `templates.FS` embed.

**Steps**:
1. `new` — initializes state with plan path; transitions to `read_plan`.
2. `read_plan` — instruction to read the plan file fully and detect the plan's inline changelog section.
3. `analyze` — instruction to use agent orchestration (referencing `spawn-implementation-agents` skill) to analyze the codebase for the current phase.
4. `implement` — instruction to write code for the current phase guided by analysis summaries.
5. `test` — instruction to delegate test authoring to a separate agent.
6. `verify` — instruction to delegate running success-criteria commands to a separate agent and return a concise pass/fail summary.
7. `update_plan` — instruction to mark plan checkboxes complete.
8. `update_changelog` — instruction to append phase entry (what was done, deviations, files changed, discoveries) and, if the last phase, a FINAL SUMMARY.
9. `update_repo_changelog` — instruction to append a new section to the repo-level `CHANGELOG.md` using the plan name as the section header and a short user-facing summary of the overall change. Runs exactly once per plan, after all phases are complete.
10. `finished` — terminal step.

**Looping**: `update_changelog` can transition to either `analyze` (next phase) or `update_repo_changelog` (all phases complete). The template instruction tells the agent to check for remaining unchecked phases in the plan and, if any, ask the user whether to continue — unless previously told to run without prompting. `workflow.StepConfig` already supports multiple `Src` entries, so `analyze` lists both `read_plan` and `update_changelog` as sources. `update_repo_changelog` lists only `update_changelog` as its source, and `finished` lists only `update_repo_changelog`.

**State**: The state file holds only the plan path. Phase progress is derived by re-reading the plan's checkboxes each iteration, making the workflow resumable and auto-correcting.

**Template variables**: `{{step}}`, `{{title}}`, `{{plan_path}}`, `{{plan_dir}}`, `{{next_step}}`, `{{config.command}}`.

## Success Metrics

- **Confirmation by default**: end-of-phase instruction prompts the user before continuing to the next phase.
- **Respects prior instruction**: if the user has said "don't ask, just implement," the agent proceeds without prompting.
- **Course correction**: user feedback on a completed phase is accepted, applied, and the loop resumes.
- **Correctness**: each checked phase in the plan has a corresponding changelog entry.
- **Release-note visibility**: on completion, the repo-level `CHANGELOG.md` contains a section keyed by the plan name with a short user-facing summary of the change.
- **Resumability**: re-invoking the workflow mid-run picks up at the first unchecked phase.

## Non-Goals

- Not generating plans — that remains in the `plan` workflow.
- Not enforcing a specific test framework or verification command — the plan's own success criteria dictate commands.
- Not managing git commits or PRs — committing remains the user's (or another command's) responsibility.
- Not tracking per-phase state outside the plan file itself; plan checkboxes are the source of truth.
- Not introducing a new sub-agent mechanism — templates reference existing skills via `{{config.command}} skill <name>`, matching `plan-steps/discovery.md` and `plan-steps/phases.md`. Agents without native sub-agent support follow the "start a new agent with this prompt" guidance from the referenced skills.
