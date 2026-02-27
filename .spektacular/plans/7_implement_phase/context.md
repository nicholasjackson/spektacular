# Implement Phase - Context

## Quick Summary
Add a `spektacular implement` command that executes implementation plans by loading the executor agent, constructing a prompt from plan files, and running Claude via the same TUI used by the plan command.

## Key Files & Locations

### Existing (to modify)
- `cmd/root.go:25-30` — Command registration (add `implementCmd`)
- `internal/tui/tui.go:39-71` — Model struct (replace `specPath` with `Workflow`)
- `internal/tui/tui.go:73-82` — `initialModel` (accept `Workflow`)
- `internal/tui/tui.go:88-90` — `Init()` (use `m.workflow.Start`)
- `internal/tui/tui.go:282,338` — Status text (use `m.workflow.StatusLabel`)
- `internal/tui/tui.go:384-411` — Result handling (use `m.workflow.OnResult`)
- `internal/tui/tui.go:616-637` — `RunPlanTUI` (refactor to use `RunAgentTUI`)
- `internal/runner/runner.go:125-135` — `BuildPrompt` (refactor, add `BuildPromptWithHeader`)

### New (to create)
- `cmd/implement.go` — Implement command definition
- `internal/implement/implement.go` — Implement orchestration package
- `internal/implement/implement_test.go` — Tests for implement package

### Reference (read-only)
- `internal/defaults/files/agents/executor.md` — Executor agent prompt (already exists)
- `internal/defaults/files/agents/planner.md` — Planner agent prompt (for comparison)
- `cmd/plan.go` — Plan command (pattern to follow)
- `internal/plan/plan.go` — Plan orchestration (pattern to follow)

## Dependencies

### Internal
- `internal/config` — Config loading (reused)
- `internal/defaults` — Embedded agent prompts (reused, executor.md)
- `internal/plan` — `LoadKnowledge()` reused by implement
- `internal/runner` — `RunClaude()`, `BuildPromptWithHeader()`, `DetectQuestions()`
- `internal/tui` — `RunImplementTUI()` (new), `RunAgentTUI()` (new generic)

### External (no new dependencies)
- `github.com/spf13/cobra` — CLI framework
- `github.com/charmbracelet/bubbletea` — TUI framework
- `golang.org/x/term` — TTY detection

## Module Mapping

| Component | Plan | Implement |
|-----------|------|-----------|
| Agent prompt | `agents/planner.md` | `agents/executor.md` |
| Content source | Spec file (`.md`) | Plan directory (`plan.md` + `context.md` + `research.md`) |
| Prompt header | "Specification to Plan" | "Implementation Plan" |
| Output validation | `plan.WritePlanOutput()` — checks `plan.md` exists | None — executor writes code directly |
| TUI entry | `tui.RunPlanTUI()` | `tui.RunImplementTUI()` |
| Non-interactive | `plan.RunPlan()` | `implement.RunImplement()` |
| Command | `spektacular plan <spec-file>` | `spektacular implement <plan-dir>` |
