# Bob Runner — Implementation Context

## Key Files and Their Purpose

| File | Purpose |
|------|---------|
| `internal/runner/runner.go` | Runner interface, Event type, RunSteps orchestrator, detection helpers |
| `internal/runner/registry.go` | Runner registry (Register/NewRunner) |
| `internal/runner/claude/claude.go` | Claude runner — reference for Bob implementation |
| `internal/runner/claude/claude_test.go` | Claude runner tests — pattern to follow |
| `internal/config/config.go` | Config with AgentConfig (Command, Args, etc.) |
| `internal/tui/tui.go` | TUI that consumes Event stream |
| `cmd/root.go` | Blank imports for runner registration |
| `.spektacular/knowledge/architecture/bob-output-spec.md` | Bob event format documentation |

## Important Types and Interfaces

### `runner.Runner` (interface)
```go
type Runner interface {
    Run(opts RunOptions) (<-chan Event, <-chan error)
}
```

### `runner.Event`
```go
type Event struct {
    Type string
    Data map[string]any
}
```

Helper methods that the TUI depends on:
- `SessionID()` — reads `Data["session_id"].(string)`
- `TextContent()` — expects `Type == "assistant"` with `Data["message"]["content"]` array of `{"type":"text","text":"..."}` blocks
- `ToolUses()` — expects `Type == "assistant"` with `Data["message"]["content"]` array of `{"type":"tool_use","name":"...","input":{...}}` blocks
- `IsResult()` — checks `Type == "result"`
- `IsError()` — checks `Type == "result"` and `Data["is_error"].(bool)`
- `ResultText()` — reads `Data["result"].(string)`

### `runner.RunOptions`
```go
type RunOptions struct {
    Prompts   Prompts
    Config    config.Config
    SessionID string
    CWD       string
    LogFile   string
    Model     string
}
```

### `config.AgentConfig`
```go
type AgentConfig struct {
    Command                    string   `yaml:"command"`
    Args                       []string `yaml:"args"`
    AllowedTools               []string `yaml:"allowed_tools"`
    DisallowedTools            []string `yaml:"disallowed_tools"`
    DangerouslySkipPermissions bool     `yaml:"dangerously_skip_permissions"`
}
```

## Event Translation Map

| Bob Event | Spektacular Event.Type | Event.Data Shape |
|-----------|----------------------|------------------|
| `init` | `"system"` | `{"session_id": "uuid"}` |
| `message` (role=assistant, delta=true) | Accumulated, then `"assistant"` | `{"message": {"content": [{"type":"text","text":"..."}]}}` |
| `message` (role=user) | Skipped (echoed input) | — |
| `tool_use` | `"assistant"` | `{"message": {"content": [{"type":"tool_use","name":"...","input":{...}}]}}` |
| `tool_result` | Skipped | — |
| `result` | `"result"` | `{"result": "...", "is_error": bool}` |

## Bob CLI Command Structure

### New session
```bash
bob --output-format stream-json -p "prompt" -m premium -y [extra args...]
```

### Resume session
```bash
bob --output-format stream-json -p "answer" -m premium -y --resume SESSION_ID [extra args...]
```

## Config Example for Bob

```yaml
agent:
  command: bob
  args:
    - "--output-format"
    - "stream-json"
    - "-y"
    - "--max-coins"
    - "100"
  dangerously_skip_permissions: false
```

## Testing Patterns

- Use `testify/require` (already in go.mod)
- Compile-time interface check: `var _ runner.Runner = (*Bob)(nil)`
- Unit test the translation functions with crafted input events
- Integration test with a mock script that produces known JSONL

## Environment Requirements

- Bob CLI installed at `/home/nicj/.nvm/versions/node/v24.14.0/bin/bob`
- Go 1.25.0+
- No new dependencies needed (all parsing uses stdlib `encoding/json`, `bufio`)
