# Spec Verify -- Context

## Key Files and Their Purpose

| File | Purpose |
|------|---------|
| `internal/steps/new.go` | Defines `SpecCreatorWorkflow` and all section steps. **Primary file to modify.** |
| `internal/spec/spec.go` | Spec file operations and agent prompt loading. Add `LoadVerifyAgentSystemPrompt()` here. |
| `internal/defaults/files/agents/verify.md` | **New file.** Embedded system prompt for the verification agent. |
| `internal/defaults/files/agents/spec.md` | Existing spec creator agent prompt. Reference for question format rules. |
| `internal/defaults/defaults.go` | Embeds `files/` directory. New file auto-included by `//go:embed files`. |
| `internal/tui/tui.go` | TUI framework. Handles `WorkflowStep`, GOTO, FINISHED, and question detection. **No changes needed.** |
| `internal/runner/runner.go` | Runner interface, event types, marker detection. **No changes needed.** |
| `internal/steps/plan.go` | Contains `feedbackStep` -- reference implementation for GOTO loop pattern. |

## Important Types and Interfaces

```go
// tui.WorkflowStep -- one step in the multi-step TUI workflow
type WorkflowStep struct {
    Name            string  // unique name, GOTO target
    StatusLabel     string  // shown in status bar
    BuildRunOptions func(cfg config.Config, cwd string) (runner.RunOptions, error)
}

// runner.RunOptions -- parameters for a single agent invocation
type RunOptions struct {
    Prompts   Prompts   // User + System prompt
    Config    config.Config
    SessionID string    // resume session (carried by TUI automatically)
    CWD       string
    LogFile   string
    Model     string    // model override
}
```

## Agent Communication Protocol

- **Questions**: `<!--QUESTION:{"questions":[...]}-->` in agent text output
- **Step complete**: `<!-- FINISHED -->` in agent text output
- **Loop back**: `<!-- GOTO: verify -->` in agent text output (name must match `WorkflowStep.Name`)
- **No AskUserQuestion tool** -- agents use HTML comments only

## Model Hierarchy

| Step | Model | Reason |
|------|-------|--------|
| Spec collection (7 steps) | `claude-haiku-4-5-20251001` | Simple Q&A, low cost |
| Spec verification (new) | `claude-sonnet-4-6` | Needs analytical depth for validation |
| Planning | `claude-opus-4-6` | Complex reasoning for architecture |
| Plan feedback | `claude-sonnet-4-6` | Incorporates user feedback |
| Implementation | `claude-sonnet-4-6` | Code generation |

## GOTO Loop Pattern (from `internal/steps/plan.go`)

The verification step follows the same pattern as `feedbackStep`:

1. Agent runs, validates, asks question
2. User answers via TUI
3. Agent resumes in same session (session ID carried automatically)
4. Agent makes edits, outputs `<!-- GOTO: verify -->` to re-validate
5. TUI detects GOTO, jumps to step with matching `Name`
6. Agent runs again with fresh prompt but same session context
7. When satisfied, agent outputs `<!-- FINISHED -->`

## Environment

- Go 1.25.0
- Testing: `testify/require`
- CLI: `cobra`
- TUI: `bubbletea` + `bubbles` + `lipgloss` + `glamour`
- Agent: Claude CLI subprocess with stream-JSON output
