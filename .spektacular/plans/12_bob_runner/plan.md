# Bob Runner — Implementation Plan

## Overview
- **Spec**: 12_bob_runner
- **Complexity**: Complex
- **Dependencies**: existing runner interface, config system, TUI, runner registry

## Current State

### What Exists
- **Runner interface** (`internal/runner/runner.go`): Defines `Runner` with `Run(opts RunOptions) (<-chan Event, <-chan error)`, plus `Event`, `RunOptions`, `RunSteps`, detection helpers (`DetectQuestions`, `DetectFinished`, `DetectGoto`, `StripMarkers`), and prompt builders.
- **Claude runner** (`internal/runner/claude/claude.go`): Implements `Runner` for the Claude CLI. Registered via `init()` as `"claude"`.
- **Runner registry** (`internal/runner/registry.go`): Map-based registry with `Register(name, constructor)` and `NewRunner(cfg)` that resolves by `cfg.Agent.Command`.
- **Config** (`internal/config/config.go`): `AgentConfig` has `Command`, `Args`, `AllowedTools`, `DisallowedTools`, `DangerouslySkipPermissions`. Default command is `"claude"`.
- **TUI** (`internal/tui/tui.go`): Consumes `runner.Event` via `handleAgentEvent`. Uses `event.SessionID()`, `event.TextContent()`, `event.ToolUses()`, `event.IsResult()`, `DetectFinished`, `DetectGoto`, `DetectQuestions`.
- **Bob output spec** (`.spektacular/knowledge/architecture/bob-output-spec.md`): Documents Bob's `stream-json` JSONL format with event types: `init`, `message`, `tool_use`, `tool_result`, `result`.

### What's Missing
- No Bob runner package exists.
- No model list or default model for Bob.
- Config doesn't support runner-specific model configuration.
- Bob events have a fundamentally different structure from Claude events — they need translation to Spektacular's `Event` format for the TUI to work.

### Key Constraints
- Bob events are **flat JSONL** with separate events for each type, while Claude wraps tool calls inside assistant message content blocks.
- Bob streams assistant text as individual `delta: true` message events (token-by-token), while Claude sends complete assistant messages with structured content blocks.
- The TUI relies on `Event.TextContent()` (expects `type: "assistant"` with `message.content` array) and `Event.ToolUses()` (expects `tool_use` blocks in content).
- The existing detection helpers (`DetectQuestions`, `DetectFinished`, `DetectGoto`) operate on text strings, so they work with any runner that provides text via `TextContent()`.

## Implementation Strategy

The Bob runner will **translate Bob's flat event stream into Spektacular's `Event` format** inside the runner, so the TUI and `RunSteps` work without modification. This is the cleanest approach because:

1. It keeps the translation logic isolated in one package.
2. The TUI, `RunSteps`, and all detection helpers remain unchanged.
3. Future runners follow the same pattern — translate into canonical `Event` format.

The translation strategy:
- **`init`** → `Event{Type: "system", Data: {"session_id": "..."}}` (matches Claude's system/init pattern, picked up by `SessionID()`)
- **`message` (assistant, delta)** → Accumulate deltas into a buffer. When a non-assistant event arrives (tool_use, tool_result, result, or next user message), flush the buffer as `Event{Type: "assistant", Data: {message: {content: [{type: "text", text: "accumulated"}]}}}`.
- **`tool_use`** → Emit a synthetic `Event{Type: "assistant"}` with a `tool_use` content block, so `ToolUses()` picks it up.
- **`tool_result`** → Skip (internal to Bob; not displayed by TUI currently).
- **`result`** → `Event{Type: "result", Data: {"result": status, "is_error": status=="error"}}`.

For **model support**, the existing `RunOptions.Model` field and `AgentConfig` already support model selection. We add a Bob-specific default model and a model list for Bob.

## Phase 1: Bob Runner Package

### Changes Required

- **`internal/runner/bob/bob.go`** (new file)
  - Create `Bob` struct implementing `runner.Runner`.
  - Register as `"bob"` via `init()` with `runner.Register("bob", ...)`.
  - Build the Bob CLI command: `bob --output-format stream-json -p <prompt> -m <model> -y [--resume <sessionID>]`.
  - Support `RunOptions.Prompts.User`, `RunOptions.SessionID`, `RunOptions.Model`, `RunOptions.CWD`, `RunOptions.LogFile`.
  - Map Bob CLI flags from config: `cfg.Agent.Args` (extra args), `cfg.Agent.DangerouslySkipPermissions` (maps to `-y`).
  - Parse Bob's JSONL output line by line.
  - Implement event translation (Bob events → Spektacular `Event` format):
    - Accumulate assistant deltas into a text buffer.
    - Flush accumulated text on tool_use, result, or stream end.
    - Emit synthetic `assistant` events with properly-shaped `message.content` arrays.
  - Handle `<thinking>` blocks by stripping them from accumulated text before flushing.
  - Handle `[using tool ...]` announcements by stripping them from accumulated text (the structured `tool_use` event provides the real data).
  - Rationale: Isolates all Bob-specific parsing in one package; rest of codebase unchanged.

- **`internal/runner/bob/bob.go`** — Event translation detail:

  ```go
  // Bob init → Spektacular system event
  // {"type":"init","session_id":"uuid","model":"premium"}
  // → Event{Type: "system", Data: map[string]any{"session_id": "uuid"}}

  // Bob message (assistant, delta) → accumulate in buffer
  // When flushing:
  // → Event{Type: "assistant", Data: map[string]any{
  //     "message": map[string]any{
  //       "content": []any{
  //         map[string]any{"type": "text", "text": accumulated},
  //       },
  //     },
  //   }}

  // Bob tool_use → Spektacular assistant event with tool_use block
  // {"type":"tool_use","tool_name":"read_file","tool_id":"tool-1","parameters":{...}}
  // → Event{Type: "assistant", Data: map[string]any{
  //     "message": map[string]any{
  //       "content": []any{
  //         map[string]any{"type": "tool_use", "name": "read_file", "input": params},
  //       },
  //     },
  //   }}

  // Bob result → Spektacular result event
  // {"type":"result","status":"success","stats":{...}}
  // → Event{Type: "result", Data: map[string]any{
  //     "result": "completed",
  //     "is_error": false,
  //   }}
  ```

### Testing Strategy
- **Unit tests** (`internal/runner/bob/bob_test.go`):
  - Compile-time interface check: `var _ runner.Runner = (*Bob)(nil)`
  - `TestNew_ReturnsNonNil` — basic construction
  - `TestTranslateInitEvent` — verifies `init` → `system` event with session_id
  - `TestTranslateAssistantDeltas` — feed multiple delta messages, verify accumulated text produces correct `assistant` event
  - `TestTranslateToolUse` — verifies `tool_use` → assistant event with tool_use content block
  - `TestTranslateToolResult_Skipped` — verifies `tool_result` events are not forwarded
  - `TestTranslateResultSuccess` — verifies `result` → result event with `is_error: false`
  - `TestTranslateResultError` — verifies error result mapping
  - `TestThinkingBlocksStripped` — assistant deltas with `<thinking>` tags produce text without them
  - `TestToolAnnouncementStripped` — `[using tool ...]` lines are stripped from accumulated text
  - `TestFlushOnToolUse` — accumulated text is flushed as an assistant event before the tool_use event
  - `TestFlushOnResult` — accumulated text is flushed before result event
  - Use testify/require, follow table-driven patterns where appropriate.
- **Verification**: `go build ./...` and `go test ./internal/runner/bob/...`

### Success Criteria
- [ ] `go build ./...` passes
- [ ] `go test ./internal/runner/bob/...` passes
- [ ] Bob runner registered as `"bob"` in the registry
- [ ] `runner.NewRunner(cfg)` returns a Bob runner when `cfg.Agent.Command == "bob"`

## Phase 2: Register Bob Runner and Update Config Defaults

### Changes Required

- **`cmd/root.go`**
  - Add blank import for the Bob runner package: `_ "github.com/jumppad-labs/spektacular/internal/runner/bob"`
  - Rationale: Follows the same pattern as the Claude runner import. This ensures the Bob runner's `init()` runs and registers itself.

- **`internal/config/config.go`**
  - No structural changes needed. The existing `AgentConfig` structure already supports Bob via `Command: "bob"` and `Args`.
  - Add a comment documenting that `Command` selects the runner (e.g., `"claude"`, `"bob"`).
  - Rationale: Config is already runner-agnostic by design.

- **Documentation**: Update the `.spektacular/knowledge/architecture/bob-output-spec.md` to note that the Bob runner is now implemented and reference the package path.

### Testing Strategy
- `TestNewRunner_ReturnsBobRunner` in `internal/runner/runner_test.go` — verify that after the bob package is imported, `NewRunner` with `Command: "bob"` returns a Bob runner.
- Alternatively, test within `bob_test.go` since the `init()` self-registers.
- Verification: `go build ./...` and `go test ./...`

### Success Criteria
- [ ] `go build ./...` passes with both Claude and Bob runners registered
- [ ] `go test ./...` passes (runner + bob packages)
- [ ] Config with `agent.command: bob` creates a working Bob runner

## Phase 3: Model Support for Bob

### Changes Required

- **`internal/runner/bob/bob.go`**
  - Define a default Bob model constant: `DefaultModel = "premium"`
  - Define a model list: `var Models = []string{"premium", "standard"}` (based on Bob's documented model tiers)
  - In `Run()`, if `opts.Model` is empty, use `DefaultModel`.
  - Pass `-m <model>` to the Bob CLI command.
  - Rationale: The spec requires model selection, a model list, and a default fallback. The model list is Bob-specific.

- **`internal/runner/bob/models.go`** (new file, optional — can be in `bob.go`)
  - Export `DefaultModel` and `Models` for use by other packages if needed.
  - Rationale: Clean separation if the model list grows or needs runtime updates.

### Testing Strategy
- `TestDefaultModel_UsedWhenEmpty` — verify that `Run()` with empty `opts.Model` builds command with `-m premium`
- `TestModelOverride_UsedWhenSet` — verify that `Run()` with `opts.Model = "standard"` builds command with `-m standard`
- `TestModels_ContainsExpected` — verify the model list has expected entries

### Success Criteria
- [ ] When no model is specified, Bob uses `"premium"` as default
- [ ] When a model is specified in `RunOptions.Model`, it is passed to the Bob CLI
- [ ] `Models` list is exported and contains `"premium"` and `"standard"`

## Phase 4: Session Management (Resume + Session ID Capture)

### Changes Required

- **`internal/runner/bob/bob.go`**
  - When `opts.SessionID != ""`, add `--resume <sessionID>` to the Bob CLI command.
  - When processing `init` events, extract `session_id` and emit it in the system event so `Event.SessionID()` returns it.
  - Rationale: The existing `RunSteps` and TUI already handle session propagation via `Event.SessionID()` and `RunOptions.SessionID`. Bob's `--resume` flag mirrors Claude's.

- **Session lifecycle** (already handled by `RunSteps`/TUI):
  1. First run: No session ID. Bob emits `init` with `session_id`. Captured by TUI/RunSteps.
  2. Question detected: TUI collects answer. Calls `resumeAgentCmd` with captured session ID.
  3. Resume run: Session ID passed as `--resume` to Bob CLI.
  4. GOTO/FINISHED: Detected by existing helpers, session ID preserved.

### Testing Strategy
- `TestSessionID_ExtractedFromInit` — feed an init event, verify session_id is present in emitted system event
- `TestResume_IncludedInCommand` — verify `--resume <id>` is in the CLI args when `opts.SessionID` is set
- `TestResume_OmittedWhenEmpty` — verify `--resume` is not in args when `SessionID` is empty

### Success Criteria
- [ ] Session ID from Bob's `init` event is captured by `Event.SessionID()`
- [ ] `--resume <sessionID>` is passed to Bob CLI when resuming
- [ ] `RunSteps` loop correctly resumes Bob sessions after questions

## Phase 5: Integration Testing and TUI Compatibility

### Changes Required

- **No code changes** — this phase verifies end-to-end integration.
- Ensure the translated events work correctly with:
  - `handleAgentEvent` in `internal/tui/tui.go` — text display, tool display, question detection, step advancement
  - `RunSteps` in `internal/runner/runner.go` — question loop, session resume, FINISHED/GOTO handling
  - `DetectQuestions`, `DetectFinished`, `DetectGoto` — work on the text from translated events

### Testing Strategy
- **Integration test** (`internal/runner/bob/bob_integration_test.go`):
  - Create a mock Bob CLI script (shell script or Go test binary) that outputs known JSONL.
  - Point the Bob runner at the mock script.
  - Verify the full event stream translation end-to-end.
  - Test scenarios:
    1. Simple run: init → message deltas → result
    2. Tool use: init → message deltas → tool_use → tool_result → message deltas → result
    3. Question flow: init → message deltas with `<!--QUESTION:...-->` → result
    4. FINISHED flow: init → message deltas with `<!-- FINISHED -->` → result
    5. GOTO flow: init → message deltas with `<!-- GOTO: stepname -->` → result
- **Manual verification**: Run `spektacular plan` with `agent.command: bob` against a real Bob CLI session.
- Verification: `go test ./...`

### Success Criteria
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes (excluding pre-existing spec test failures)
- [ ] JSON-stream output from Bob CLI is parsed and events reach the TUI
- [ ] Questions in Bob output are detected and handled
- [ ] FINISHED and GOTO markers in Bob output trigger correct step transitions
- [ ] Messages and tool calls from Bob render in the TUI
- [ ] Session ID is captured from first run and reused on resume
- [ ] Model selection works with default fallback

## Out of Scope
- Changes to the TUI event handling logic
- Changes to `RunSteps` orchestration
- New CLI commands or flags
- Config file schema changes (existing `agent.command` + `agent.args` are sufficient)
- Bob MCP server management
- Bob extension management
- Bob budget control (can be passed via `agent.args` if needed)

## References
- **Runner interface**: `internal/runner/runner.go:16-21` — `Runner` interface definition
- **Event type**: `internal/runner/runner.go:24-27` — `Event` struct
- **Event helpers**: `internal/runner/runner.go:30-90` — `SessionID()`, `TextContent()`, `ToolUses()`, `IsResult()`
- **RunSteps**: `internal/runner/runner.go:191-264` — step execution loop with question/resume handling
- **Claude runner**: `internal/runner/claude/claude.go` — reference implementation for Bob runner
- **Registry**: `internal/runner/registry.go` — `Register()` and `NewRunner()` pattern
- **Config**: `internal/config/config.go:57-63` — `AgentConfig` struct
- **TUI event handler**: `internal/tui/tui.go:570-652` — `handleAgentEvent` (consumes translated events)
- **Bob CLI docs**: `.spektacular/knowledge/architecture/bob-output-spec.md` — full Bob event format specification
- **Bob sample output**: `.spektacular/knowledge/architecture/bob_json/bob_stream.json` — real Bob JSONL output
- **Root command**: `cmd/root.go:10-11` — blank import pattern for runner registration
