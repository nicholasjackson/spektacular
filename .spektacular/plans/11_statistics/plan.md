# Statistics — Implementation Plan

## Overview
- **Spec**: `11_statistics`
- **Complexity**: Medium
- **Dependencies**: None (builds on existing runner interface and TUI)

## Current State

### What Exists
- **Runner interface** (`internal/runner/runner.go:16-21`): Defines `Runner` with a single method `Run(opts RunOptions) (<-chan Event, <-chan error)`.
- **Event struct** (`internal/runner/runner.go:24-27`): Has `Type string` and `Data map[string]any`. Already has `ToolUses()` method that extracts `tool_use` blocks from assistant events.
- **Claude runner** (`internal/runner/claude/claude.go`): Parses JSONL output from `claude` CLI. Result events already contain rich usage data: `usage.input_tokens`, `usage.output_tokens`, `usage.cache_creation_input_tokens`, `usage.cache_read_input_tokens`, `num_turns`, `duration_ms`, `total_cost_usd`.
- **TUI** (`internal/tui/tui.go`): BubbleTea-based model. Processes events in `handleAgentEvent`. Shows completion message in `advanceStep` at line 175-208.
- **Non-TUI path**: `RunSteps()` in `internal/runner/runner.go:179-193` executes steps synchronously, used by `plan.RunPlan()` and `implement.RunImplement()`.
- **Three commands**: `plan`, `implement`, `new` — all go through either TUI or non-TUI path.

### What's Missing
- No statistics collection mechanism
- No stats extraction from runner events
- No statistics display at session completion
- Runner interface has no way to report metrics

### Key Observation — Claude CLI Result Events
Result events from `claude` CLI contain usage data that can be parsed:
```json
{
  "type": "result",
  "duration_ms": 158492,
  "num_turns": 5,
  "total_cost_usd": 0.79,
  "usage": {
    "input_tokens": 6,
    "cache_creation_input_tokens": 68525,
    "cache_read_input_tokens": 270116,
    "output_tokens": 9124
  }
}
```

Tool calls are counted from assistant events via `event.ToolUses()`.

## Implementation Strategy

The design adds a `StatsCollector` interface to the runner package that each runner implementation provides. This satisfies the spec's requirement that "statistic collection must be built into the runner interface and implemented by each individual runner." The collector is created once per session and processes every event, accumulating stats across resumes.

**Key decisions:**
1. **`StatsCollector` interface** — rather than modifying `Run()` return type, add `NewStatsCollector()` to `Runner`. This keeps the event channel pattern clean while letting each runner define how to extract stats from its own event format.
2. **Elapsed time measured at orchestration level** — the TUI model and `RunSteps` track wall-clock time. The runner's `duration_ms` is API processing time, not what the user experiences.
3. **Tokens = input + output + cache tokens** — all token types are summed into a single "tokens used" figure for simplicity.
4. **Tabular display using lipgloss** — consistent with existing TUI styling. Tool calls line is conditional (omitted when zero).

## Phase 1: Stats Types and Runner Interface

### Changes Required

- **`internal/runner/stats.go`** (new file)
  - Define `Statistics` struct with `ElapsedTime time.Duration`, `TokensUsed int`, `ToolCalls int`
  - Define `StatsCollector` interface with `ProcessEvent(event Event)` and `Stats() Statistics`
  - Add a `FormatElapsed(d time.Duration) string` helper that formats as `Xh Ym Zs` (omitting zero-value leading components)
  - Rationale: Separate file keeps stats concerns cleanly isolated from the event/question/prompt code

- **`internal/runner/runner.go`**
  - Add `NewStatsCollector() StatsCollector` to the `Runner` interface at line 16
  - Rationale: Each runner must provide its own stats extraction logic

- **`internal/runner/runner_test.go`**
  - Update `stubRunner` to implement `NewStatsCollector()` (return a no-op collector)
  - Rationale: Keeps existing tests passing after interface change

### Code Details

**`internal/runner/stats.go`:**
```go
package runner

import (
    "fmt"
    "time"
)

// Statistics holds accumulated session-wide metrics.
type Statistics struct {
    ElapsedTime time.Duration
    TokensUsed  int
    ToolCalls   int
}

// StatsCollector accumulates statistics from runner events.
// Each runner implementation provides its own collector that knows
// how to extract metrics from its agent's event format.
type StatsCollector interface {
    // ProcessEvent inspects an event and updates internal counters.
    ProcessEvent(event Event)
    // Stats returns the accumulated statistics (excluding ElapsedTime,
    // which is measured by the caller).
    Stats() Statistics
}

// noopCollector is a StatsCollector that does nothing.
// Used as a fallback when a runner doesn't support stats.
type noopCollector struct{}

func (n *noopCollector) ProcessEvent(_ Event) {}
func (n *noopCollector) Stats() Statistics    { return Statistics{} }

// NoopStatsCollector returns a collector that does not track anything.
func NoopStatsCollector() StatsCollector { return &noopCollector{} }

// FormatElapsed formats a duration as "Xh Ym Zs", omitting zero leading components.
func FormatElapsed(d time.Duration) string {
    d = d.Round(time.Second)
    h := int(d.Hours())
    m := int(d.Minutes()) % 60
    s := int(d.Seconds()) % 60

    switch {
    case h > 0:
        return fmt.Sprintf("%dh %dm %ds", h, m, s)
    case m > 0:
        return fmt.Sprintf("%dm %ds", m, s)
    default:
        return fmt.Sprintf("%ds", s)
    }
}
```

**Runner interface change (`internal/runner/runner.go`):**
```go
// Runner is the interface that all agent backends must implement.
type Runner interface {
    // Run starts the agent with the given options and returns a channel of
    // events and an error channel.
    Run(opts RunOptions) (<-chan Event, <-chan error)
    // NewStatsCollector returns a StatsCollector that can parse this
    // runner's events to extract usage metrics.
    NewStatsCollector() StatsCollector
}
```

### Testing Strategy
- Unit: `TestFormatElapsed` — table-driven tests for various durations (seconds only, minutes+seconds, hours+minutes+seconds)
- Unit: `TestNoopCollector` — verify ProcessEvent is safe to call and Stats returns zero values
- Verification: `go build ./...` passes after interface change

### Success Criteria
- [ ] `go build ./...` passes
- [ ] `Statistics` struct and `StatsCollector` interface defined
- [ ] `FormatElapsed` correctly formats all time ranges
- [ ] `stubRunner` in tests updated to satisfy new interface

---

## Phase 2: Claude StatsCollector Implementation

### Changes Required

- **`internal/runner/claude/stats.go`** (new file)
  - Implement `claudeStatsCollector` struct that counts tool calls from assistant events and extracts tokens from result events
  - Rationale: Encapsulates Claude-specific event parsing in the Claude package

- **`internal/runner/claude/claude.go`**
  - Add `NewStatsCollector()` method to `Claude` struct (line 26)
  - Rationale: Satisfy the updated Runner interface

- **`internal/runner/claude/stats_test.go`** (new file)
  - Tests for stats extraction from various event types
  - Rationale: Verify Claude-specific parsing logic

### Code Details

**`internal/runner/claude/stats.go`:**
```go
package claude

import "github.com/jumppad-labs/spektacular/internal/runner"

// claudeStatsCollector extracts statistics from Claude CLI events.
type claudeStatsCollector struct {
    toolCalls  int
    tokensUsed int
}

func (c *claudeStatsCollector) ProcessEvent(event runner.Event) {
    // Count tool uses from assistant events
    c.toolCalls += len(event.ToolUses())

    // Extract token usage from result events
    if event.IsResult() {
        c.tokensUsed += extractTokens(event)
    }
}

func (c *claudeStatsCollector) Stats() runner.Statistics {
    return runner.Statistics{
        ToolCalls:  c.toolCalls,
        TokensUsed: c.tokensUsed,
    }
}

// extractTokens sums all token fields from a Claude result event's usage data.
func extractTokens(event runner.Event) int {
    usage, ok := event.Data["usage"].(map[string]any)
    if !ok {
        return 0
    }

    var total int
    for _, key := range []string{
        "input_tokens",
        "output_tokens",
        "cache_creation_input_tokens",
        "cache_read_input_tokens",
    } {
        if v, ok := usage[key].(float64); ok {
            total += int(v)
        }
    }
    return total
}
```

**`internal/runner/claude/claude.go` addition:**
```go
// NewStatsCollector returns a StatsCollector for Claude CLI events.
func (c *Claude) NewStatsCollector() runner.StatsCollector {
    return &claudeStatsCollector{}
}
```

### Testing Strategy
- Unit: `TestClaudeStatsCollector_ToolCalls` — feed assistant events with tool_use blocks, verify count
- Unit: `TestClaudeStatsCollector_Tokens` — feed result event with usage data, verify token sum
- Unit: `TestClaudeStatsCollector_AccumulatesAcrossEvents` — feed multiple events (simulating resumes), verify accumulated totals
- Unit: `TestClaudeStatsCollector_NoUsageData` — feed result event without usage, verify zero
- Unit: `TestExtractTokens_AllFields` — verify all four token fields are summed
- Verification: `go build ./...` and `go test ./internal/runner/claude/...` pass

### Success Criteria
- [ ] `go build ./...` passes
- [ ] `go test ./internal/runner/claude/...` passes
- [ ] Claude runner satisfies updated Runner interface
- [ ] Stats correctly extracted from real Claude CLI event format

---

## Phase 3: TUI Statistics Display

### Changes Required

- **`internal/tui/tui.go`**
  - Add `statsCollector runner.StatsCollector` and `startTime time.Time` fields to `model` struct (line 61)
  - In `initialModel` (line 101): create runner, get stats collector, set start time
  - In `handleAgentEvent` (line 553): call `m.statsCollector.ProcessEvent(event)` on every event
  - In `advanceStep` (line 175): when all steps complete (the "All steps done" branch at line 184), render stats table before the "completed" line
  - Add `renderStatsTable()` method to format statistics with lipgloss
  - Rationale: The TUI model is the natural place to accumulate stats since it persists across the entire session including resumes

- **`internal/tui/tui.go` — `initialModel` change:**
  ```go
  func initialModel(wf Workflow, projectPath string, cfg config.Config) model {
      // Create runner to get a stats collector for the session.
      var statsCollector runner.StatsCollector
      if r, err := runner.NewRunner(cfg); err == nil {
          statsCollector = r.NewStatsCollector()
      } else {
          statsCollector = runner.NoopStatsCollector()
      }

      label := ""
      if len(wf.Steps) > 0 {
          label = wf.Steps[0].StatusLabel
      }
      return model{
          workflow:       wf,
          projectPath:    projectPath,
          cfg:            cfg,
          themeIdx:       0,
          followMode:     true,
          statusText:     "* thinking  " + label,
          logFile:        wf.LogFile,
          statsCollector: statsCollector,
          startTime:      time.Now(),
      }
  }
  ```

- **`internal/tui/tui.go` — stats processing in `handleAgentEvent`:**
  Add at the start of `handleAgentEvent`, before any other processing:
  ```go
  m.statsCollector.ProcessEvent(msg.event)
  ```

- **`internal/tui/tui.go` — stats display in `advanceStep`:**
  In the "All steps done" branch (after line 196), before the "completed" line:
  ```go
  // Render statistics table
  stats := m.statsCollector.Stats()
  stats.ElapsedTime = time.Since(m.startTime)
  statsLine := m.renderStatsTable(stats)
  if statsLine != "" {
      m = m.withLine(statsLine + "\n")
  }
  ```

- **`internal/tui/tui.go` — new `renderStatsTable` method:**
  ```go
  func (m model) renderStatsTable(stats runner.Statistics) string {
      p := m.currentPalette()
      headerStyle := lipgloss.NewStyle().Bold(true).Foreground(p.output)
      labelStyle := lipgloss.NewStyle().Foreground(p.faint).Width(14)
      valueStyle := lipgloss.NewStyle().Foreground(p.answer)

      var lines []string
      lines = append(lines, headerStyle.Render("Session Statistics"))
      lines = append(lines, labelStyle.Render("  Elapsed:")+valueStyle.Render(runner.FormatElapsed(stats.ElapsedTime)))
      lines = append(lines, labelStyle.Render("  Tokens:")+valueStyle.Render(formatNumber(stats.TokensUsed)))
      if stats.ToolCalls > 0 {
          lines = append(lines, labelStyle.Render("  Tool calls:")+valueStyle.Render(formatNumber(stats.ToolCalls)))
      }

      border := lipgloss.NewStyle().
          Border(lipgloss.RoundedBorder()).
          BorderForeground(p.faint).
          Padding(0, 1)
      return border.Render(strings.Join(lines, "\n"))
  }
  ```

- **`internal/tui/tui.go` — `formatNumber` helper:**
  ```go
  // formatNumber formats an integer with comma separators (e.g., 1,234,567).
  func formatNumber(n int) string {
      if n < 1000 {
          return fmt.Sprintf("%d", n)
      }
      s := fmt.Sprintf("%d", n)
      var result []byte
      for i, c := range s {
          if i > 0 && (len(s)-i)%3 == 0 {
              result = append(result, ',')
          }
          result = append(result, byte(c))
      }
      return string(result)
  }
  ```

### Testing Strategy
- Unit: `TestFormatNumber` — table-driven: 0, 999, 1000, 1234567
- Unit: `TestRenderStatsTable_ToolCallsConditional` — verify tool calls line is absent when ToolCalls == 0
- Unit: `TestStatsCollectorProcessedInHandleAgentEvent` — create model with a mock collector, feed events, verify ProcessEvent called
- Verification: `go build ./...` and `go test ./internal/tui/...` pass

### Success Criteria
- [ ] `go build ./...` passes
- [ ] `go test ./internal/tui/...` passes
- [ ] Statistics table rendered at session completion in TUI
- [ ] Tool calls line conditional on count > 0
- [ ] Elapsed time displayed in `Xh Ym Zs` format
- [ ] Stats accumulate across resumes (cross-resume accumulation)

---

## Phase 4: Non-TUI Path Statistics

### Changes Required

- **`internal/runner/runner.go`**
  - Modify `RunSteps` signature to accept and return a `StatsCollector` (or return `Statistics`)
  - In `runStep`, call `collector.ProcessEvent(event)` on each event
  - Rationale: The non-TUI path (piped output) also needs stats

  New `RunSteps` signature:
  ```go
  func RunSteps(
      r Runner,
      steps []Step,
      cfg config.Config,
      cwd string,
      onText func(string),
      onQuestion func([]Question) string,
  ) (Statistics, error) {
  ```

  Inside `runStep`, process events through the collector:
  ```go
  func runStep(
      r Runner,
      step Step,
      cfg config.Config,
      cwd string,
      onText func(string),
      onQuestion func([]Question) string,
      collector StatsCollector,
  ) error {
      // ... existing code ...
      for event := range events {
          collector.ProcessEvent(event)
          // ... rest of existing event processing ...
      }
  ```

  `RunSteps` creates the collector and measures elapsed time:
  ```go
  func RunSteps(...) (Statistics, error) {
      start := time.Now()
      collector := r.NewStatsCollector()
      for _, step := range steps {
          if err := runStep(r, step, cfg, cwd, onText, onQuestion, collector); err != nil {
              return Statistics{}, err
          }
      }
      stats := collector.Stats()
      stats.ElapsedTime = time.Since(start)
      return stats, nil
  }
  ```

- **`internal/plan/plan.go`**
  - Update `RunPlan` to capture and return stats from `RunSteps`
  - New return type: `(string, runner.Statistics, error)`

- **`internal/implement/implement.go`**
  - Update `RunImplement` to capture and return stats from `RunSteps`
  - New return type: `(string, runner.Statistics, error)`

- **`cmd/plan.go`**, **`cmd/implement.go`**
  - In the non-TUI branches, capture returned stats and print them to stdout
  - Format: simple text table (not lipgloss, since there's no terminal styling in pipe mode)

  ```go
  // After RunPlan returns:
  fmt.Println("\nSession Statistics")
  fmt.Printf("  Elapsed:    %s\n", runner.FormatElapsed(stats.ElapsedTime))
  fmt.Printf("  Tokens:     %d\n", stats.TokensUsed)
  if stats.ToolCalls > 0 {
      fmt.Printf("  Tool calls: %d\n", stats.ToolCalls)
  }
  ```

- **`cmd/new.go`**
  - The `new` command only uses TUI for interactive mode. The non-interactive path (`spec.Create`) doesn't involve an agent, so no stats to display.
  - No changes needed here.

### Testing Strategy
- Unit: Update `TestRunSteps`-style tests if any exist to handle new return value
- Integration: Verify `RunPlan` and `RunImplement` signature changes compile
- Verification: `go build ./...` and `go test ./...` pass

### Success Criteria
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes
- [ ] Non-TUI plan/implement display stats at completion
- [ ] Stats format is clean plain text for pipe-friendly output

---

## Not In Scope
- Cost display (USD) — not requested in the spec
- Per-step statistics breakdown — spec asks for session-wide only
- Persistent stats history / logging to disk
- API call count (distinct from tool calls)
- Detailed token breakdown (input vs output vs cache)

## References
- **Runner interface**: `internal/runner/runner.go:16-21`
- **Event struct with ToolUses()**: `internal/runner/runner.go:24-90`
- **Claude runner**: `internal/runner/claude/claude.go`
- **TUI model and advanceStep**: `internal/tui/tui.go:61-208`
- **handleAgentEvent**: `internal/tui/tui.go:553-631`
- **RunSteps**: `internal/runner/runner.go:179-252`
- **Claude CLI result event format**: `.spektacular/knowledge/architecture/example-output/run-plan.json` (contains `usage`, `duration_ms`, `num_turns`, `total_cost_usd`)
- **Existing test patterns**: `internal/runner/runner_test.go` (testify/require, no table-driven)
- **lipgloss styling**: `internal/tui/tui.go:659-671` (status bar), `internal/tui/theme.go` (palettes)
