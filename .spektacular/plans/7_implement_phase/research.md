# Implement Phase - Research Notes

## Specification Analysis

### Original Requirements
- Add `spektacular implement` command that takes a plan directory as argument
- Interactive TUI support
- Use executor agent from defaults
- Validate plan exists before executing

### Implicit Requirements
- Non-TTY mode (piped/background execution) should work, matching plan command pattern
- Multi-turn conversation support (questions from executor agent)
- Debug logging support (matching plan command pattern)
- Plan directory resolution should be flexible (full path, relative path, or plan name)

### Constraints Identified
- Must reuse TUI and agent execution logic from plan (spec requirement)
- Must not break existing plan command behavior
- Executor agent prompt already exists — no need to create it

## Research Process

### Files Examined

| File | Lines | Summary |
|------|-------|---------|
| `cmd/plan.go` | 1-67 | Plan command: loads config, detects TTY, calls TUI or RunPlan |
| `cmd/run.go` | 1-21 | Stub command, not yet implemented |
| `cmd/root.go` | 1-31 | Command registration, version 0.1.0 |
| `internal/plan/plan.go` | 1-168 | Plan orchestration: LoadKnowledge, LoadAgentPrompt, PreparePlanDir, WritePlanOutput, RunPlan |
| `internal/tui/tui.go` | 1-638 | Full TUI: model, events, view, question handling, agent start/resume |
| `internal/tui/theme.go` | 1-75 | 5 themes with palette definitions |
| `internal/runner/runner.go` | 1-244 | Claude subprocess, event streaming, question detection, BuildPrompt |
| `internal/defaults/defaults.go` | 1-30 | Embedded file system for agent prompts |
| `internal/defaults/files/agents/executor.md` | 1-309 | Executor agent definition |
| `internal/defaults/files/agents/planner.md` | 1-287 | Planner agent definition (for comparison) |
| `internal/config/config.go` | 1-161 | Config with AgentConfig struct |
| `internal/plan/plan_test.go` | 1-67 | Tests: LoadKnowledge, WritePlanOutput, LoadAgentPrompt |
| `internal/tui/tui_test.go` | 1-138 | Tests: helpers, readNext, initialModel |
| `internal/runner/runner_test.go` | 1-141 | Tests: event parsing, questions, BuildPrompt |

### Patterns Discovered

**Command Pattern** (`cmd/plan.go:20-64`):
1. Get working directory
2. Load config from `.spektacular/config.yaml` (fallback to defaults)
3. Detect TTY → route to TUI or non-interactive
4. Print result path

**TUI Agent Flow** (`internal/tui/tui.go:88-136`):
1. `startAgentCmd()` — builds prompt, prepares output dir, spawns Claude
2. `resumeAgentCmd()` — resumes with user answer via session ID
3. Events flow through `agentEventMsg` carrying open channels

**Plan-Specific Code in TUI** — 5 locations:
1. `startAgentCmd` (line 94): Calls `plan.LoadAgentPrompt()` and `plan.LoadKnowledge()`
2. `startAgentCmd` (line 102): Calls `runner.BuildPrompt()` with plan-specific header
3. `startAgentCmd` (line 104): Derives plan dir from spec path
4. Status text (lines 282, 338): Uses `filepath.Base(m.specPath)`
5. Result handling (lines 396-411): Derives plan dir from spec path, calls `plan.WritePlanOutput()`

**Non-Interactive Flow** (`internal/plan/plan.go:82-159`):
- Identical structure: build prompt → spawn Claude → stream events → handle questions → validate output
- The implement non-interactive flow will mirror this exactly

## Key Findings

### Architecture Insights
- The TUI is the primary interactive interface; the non-interactive path in `plan.go` duplicates prompt construction logic
- Channels flow through tea messages (not stored in model) to respect Bubble Tea's copy-on-update semantics
- Session management enables multi-turn conversations — critical for executor agents that may need clarification

### Existing Implementations
- The executor agent definition (`executor.md`) is comprehensive and ready to use
- It expects to read `context.md` → `plan.md` → `research.md` in that order
- The prompt structure (agent + knowledge + content) is the same for both plan and implement

### Reusable Components
- `plan.LoadKnowledge()` — used by both plan and implement
- `runner.RunClaude()` — generic Claude subprocess runner
- `runner.DetectQuestions()` — question parsing from agent output
- TUI: viewport, themes, question panel, tool status — all generic

### Testing Infrastructure
- `testify/require` for assertions
- `t.TempDir()` for isolated filesystem tests
- No mocking needed — tests are mostly unit/pure-function tests
- No integration tests with real Claude subprocess

## Design Decisions

### Decision: Refactor TUI with Workflow struct vs duplicate TUI code
- **Options**: (A) Workflow abstraction, (B) Copy entire TUI, (C) Parameterize with flags
- **Chosen**: A — Workflow struct with Start/OnResult callbacks
- **Rationale**: Cleanest separation. The TUI display logic (viewport, themes, questions) is identical; only prompt construction and result handling differ. A Workflow struct captures exactly this difference.
- **Trade-offs**: Slightly more abstract than direct code, but avoids 600+ lines of duplication.

### Decision: Extend BuildPrompt vs create separate function
- **Options**: (A) Add header parameter to BuildPrompt, (B) Create BuildImplementPrompt
- **Chosen**: A via new `BuildPromptWithHeader`, with `BuildPrompt` as backwards-compatible wrapper
- **Rationale**: The prompt structure is identical except for the content section header. A parameter is simpler than a separate function. The wrapper maintains backwards compatibility.

### Decision: Plan directory resolution strategy
- **Options**: (A) Require exact path, (B) Support multiple resolution strategies
- **Chosen**: B — try direct path, relative path, then `.spektacular/plans/{name}/`
- **Rationale**: Better UX. Users can type `spektacular implement my-feature` instead of the full path. Still supports exact paths for scripts.

### Decision: No output validation for implement
- **Options**: (A) Validate specific output files, (B) Just check result event
- **Chosen**: B — the executor agent produces code changes directly, not a specific output file
- **Rationale**: Unlike the planner (which must produce plan.md), the executor modifies project files directly via its tools. There's no single output file to validate. The result event from Claude indicates success/failure.

## Open Questions (All Resolved)
- None — all design decisions have been made based on codebase analysis.
