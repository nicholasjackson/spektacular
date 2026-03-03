# Bob Runner — Research Notes

## Design Decision: Translation Layer Location

### Option A: Translate inside the runner (chosen)
Bob runner translates Bob events → Spektacular `Event` format before sending to the event channel. TUI and RunSteps remain unchanged.

**Pros:**
- Isolated change — only one new package
- TUI doesn't need to know about Bob
- Future runners follow the same pattern
- All detection helpers (questions, finished, goto) work automatically

**Cons:**
- Some information loss (tool_result details not forwarded)
- Accumulated text buffering adds complexity inside the runner

### Option B: Runner-specific event handlers in TUI
Bob runner sends raw Bob events. TUI checks runner type and handles differently.

**Pros:**
- No information loss
- Simpler runner implementation

**Cons:**
- TUI becomes coupled to runner specifics
- Every new runner requires TUI changes
- Detection helpers need runner-aware variants
- Violates runner abstraction

**Decision:** Option A is clearly better. The runner interface exists specifically to abstract backend differences.

## Delta Accumulation Strategy

Bob streams assistant text as individual token deltas:
```json
{"type":"message","role":"assistant","content":"Hello ","delta":true}
{"type":"message","role":"assistant","content":"world","delta":true}
```

We accumulate these into a buffer and flush as a single `assistant` event when:
1. A `tool_use` event arrives (flush text before tool)
2. A `result` event arrives (flush remaining text)
3. The stream ends (flush any buffered text)

This means the TUI receives fewer, larger text events rather than token-by-token updates. This is acceptable because:
- The TUI already handles complete text blocks from Claude
- Real-time streaming display is handled by Bubble Tea's event loop
- We can flush periodically (e.g., on newlines or after N characters) if real-time feel is needed later

### Thinking Block Handling

Bob includes `<thinking>` blocks in message deltas. Strategy:
- Strip `<thinking>...</thinking>` content from the accumulated buffer before flushing.
- Use a simple state machine: track whether we're inside a thinking block.
- Alternative: Keep thinking text but mark it (decided against — TUI doesn't have thinking-mode rendering).

### Tool Announcement Handling

Bob announces tools inline: `[using tool read_file: README.md]\n`
- These appear in the message delta stream before `tool_use` events.
- Strip these from accumulated text since the structured `tool_use` event provides cleaner data for the TUI's tool display.
- Regex pattern: `\[using tool [^\]]+\]\n?`

## Bob CLI Flag Mapping

| Spektacular Concept | Bob CLI Flag |
|---------------------|-------------|
| Prompt | `-p <text>` |
| Model | `-m <model>` |
| Resume | `--resume <id>` |
| Auto-approve | `-y` (or `--approval-mode yolo`) |
| Output format | `--output-format stream-json` |
| Max budget | `--max-coins <n>` |

**Note:** Bob doesn't have `--system-prompt`, `--allowedTools`, or `--disallowedTools` equivalents. These `AgentConfig` fields are Claude-specific and will be ignored by the Bob runner. The Bob runner uses `cfg.Agent.Args` for any extra flags.

## Model Tiers

Bob supports at least two model tiers based on the output spec:
- `premium` — higher capability, higher cost
- `standard` — lower capability, lower cost

The default should be `premium` for planning/implementation tasks (these are non-trivial).

## Pre-existing Test Failures

The `internal/spec` package has 3 pre-existing test failures related to auto-numbering. These are unrelated to the Bob runner work and should not block this implementation.
