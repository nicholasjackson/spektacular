# Statistics — Research Notes

## Claude CLI Output Analysis

Analyzed the actual Claude CLI output in `.spektacular/knowledge/architecture/example-output/run-plan.json`.

### Result Event Structure

The `result` event is the terminal event emitted by the Claude CLI. It contains comprehensive statistics:

```json
{
  "type": "result",
  "subtype": "success",
  "is_error": false,
  "duration_ms": 158492,
  "duration_api_ms": 159971,
  "num_turns": 5,
  "total_cost_usd": 0.79225225,
  "session_id": "c102b04d-...",
  "usage": {
    "input_tokens": 6,
    "cache_creation_input_tokens": 68525,
    "cache_read_input_tokens": 270116,
    "output_tokens": 9124,
    "server_tool_use": {
      "web_search_requests": 0,
      "web_fetch_requests": 0
    }
  },
  "modelUsage": {
    "claude-opus-4-6": {
      "inputTokens": 6,
      "outputTokens": 9124,
      "cacheReadInputTokens": 270116,
      "cacheCreationInputTokens": 68525,
      "costUSD": 0.79146925
    }
  }
}
```

### Key Observations

1. **Token counts**: The `usage` object in the result event provides the authoritative totals for that run. When sessions are resumed, each run produces its own result event. We must sum across all result events.

2. **Tool call counting**: Tool uses are embedded in `assistant` events as content blocks of `type: "tool_use"`. The existing `Event.ToolUses()` method already extracts these. We count them by processing every event in the stream.

3. **Duration**: `duration_ms` is the wall-clock time from the Claude CLI's perspective. For the user-facing "elapsed time", we measure wall-clock time at the Spektacular level (includes question answering pauses).

4. **JSON numeric types**: All numbers in `usage` are decoded as `float64` by Go's `encoding/json`. Must cast to `int` when summing.

5. **Accumulation across resumes**: Each resume produces a separate result event with usage for just that run. The StatsCollector must process every event to accumulate correctly.

## Design Decisions

### Decision 1: StatsCollector interface (not method on Runner)

**Problem:** The Runner interface creates a new instance per `Run()` call, but stats must accumulate across resumes.

**Solution:** Add `NewStatsCollector() StatsCollector` to the Runner interface. The collector is created once per session and processes all events, persisting across multiple runner invocations. This is cleaner than storing mutable state on the Runner itself.

### Decision 2: Where to measure elapsed time

**Chosen: Orchestration level.** The user cares about total wall-clock time including question-answering pauses, not just agent execution time. `time.Now()` is recorded at session start; `time.Since()` is computed at session end.

### Decision 3: Token aggregation

**Chosen: Single total.** Sum all four token fields (input, output, cache_creation, cache_read) into one number. The spec says "total number of tokens consumed" — a breakdown would add complexity without spec justification.

### Decision 4: Display format

**Chosen: Lipgloss bordered table in TUI, plain text in non-TUI.** The spec requires "tabular format". Lipgloss `RoundedBorder` provides clean table rendering consistent with the existing TUI aesthetic. Non-TUI path uses simple `fmt.Printf` alignment.

## Alternatives Considered

### Middleware/Observer Pattern
Wrapping the event channel with a statistics-collecting middleware. Rejected: still needs runner-specific parsing, and the spec says "built into the runner interface."

### Statistics() method on Runner (stateful)
Adding `Statistics() Statistics` directly to Runner and having each `Run()` call accumulate state. Rejected: Runner instances are created fresh on each resume in the TUI (`runner.NewRunner(cfg)` is called in `startCurrentStep` and `resumeAgentCmd`), so state would be lost. The StatsCollector factory pattern avoids this.

### Cost display
The result event includes `total_cost_usd`. Could be shown but spec doesn't request it. Left as future enhancement.
