# Statistics — Implementation Context

## Key Files

| File | Purpose |
|------|---------|
| `internal/runner/runner.go` | Runner interface, Event struct, RunSteps orchestration |
| `internal/runner/claude/claude.go` | Claude CLI runner implementation |
| `internal/tui/tui.go` | BubbleTea TUI — model, Update, View |
| `internal/tui/theme.go` | Palette definitions and theme cycling |
| `internal/steps/plan.go` | PlanWorkflow — builds TUI workflow for plan command |
| `internal/steps/implement.go` | ImplementWorkflow — builds TUI workflow for implement command |
| `internal/steps/new.go` | SpecCreatorWorkflow — builds TUI workflow for new command |
| `internal/plan/plan.go` | RunPlan — non-TUI plan execution |
| `internal/implement/implement.go` | RunImplement — non-TUI implement execution |
| `cmd/plan.go` | Plan cobra command — TUI and non-TUI branches |
| `cmd/implement.go` | Implement cobra command — TUI and non-TUI branches |
| `cmd/new.go` | New cobra command — interactive (TUI) and non-interactive branches |

## Important Types

- `runner.Runner` — Interface: `Run(RunOptions) (<-chan Event, <-chan error)`. Must add `NewStatsCollector()`.
- `runner.Event` — `{Type string, Data map[string]any}`. Has `ToolUses()`, `IsResult()`, `TextContent()` methods.
- `tui.model` — BubbleTea model. Has `statsCollector` and `startTime` (to be added).
- `tui.Workflow` / `tui.WorkflowStep` — Multi-step agent pipeline definition.
- `config.Config` — Top-level config, contains `Agent.Command` used to select runner.

## New Files to Create

| File | Purpose |
|------|---------|
| `internal/runner/stats.go` | `Statistics` struct, `StatsCollector` interface, `FormatElapsed()` |
| `internal/runner/stats_test.go` | Tests for FormatElapsed and NoopCollector |
| `internal/runner/claude/stats.go` | `claudeStatsCollector` — Claude-specific event parsing |
| `internal/runner/claude/stats_test.go` | Tests for Claude stats extraction |

## Claude CLI Result Event Token Fields

From `event.Data["usage"].(map[string]any)`:
- `input_tokens` — direct input tokens
- `output_tokens` — generated output tokens
- `cache_creation_input_tokens` — tokens used for cache creation
- `cache_read_input_tokens` — tokens read from cache

All are `float64` after JSON unmarshaling. Sum all four for total tokens.

## Event Flow (TUI)

```
initialModel() — create StatsCollector, record startTime
    |
startCurrentStep() -> spawns runner -> events channel
    |
handleAgentEvent() — process each event:
    1. statsCollector.ProcessEvent(event)  <- NEW
    2. existing: session ID, tool uses, text content, questions
    3. if result/finished -> advanceStep()
    |
advanceStep() — when all steps done:
    1. render stats table  <- NEW
    2. call OnDone
    3. show "completed" message
```

## Event Flow (Non-TUI)

```
RunSteps() — create StatsCollector, record startTime
    |
runStep() — for each event:
    1. collector.ProcessEvent(event)  <- NEW
    2. existing: session ID, text callback, question callback, finish detection
    |
return Statistics <- NEW
```

## Testing Patterns

- Uses `testify/require` (not `assert`)
- Individual test functions (not table-driven in most files)
- Compile-time interface checks: `var _ runner.Runner = (*Claude)(nil)`
- `stubRunner` in `runner_test.go` implements Runner interface for registry tests
