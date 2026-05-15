# Feature: Adhoc JSON Protocol

## Overview
Add an adhoc execution mode to Spektacular that supports both human terminal output and machine-readable JSON output.

This feature defines a tool-agnostic, bidirectional stdin/stdout protocol so OpenClaw (and other orchestrators) can run one-off tasks through Spektacular without going through the full spec -> plan -> integrate workflow.

The protocol must normalize provider-specific interaction patterns (Claude CLI, OpenAI CLI, Gemini CLI, etc.) into one abstract question/answer interface.

## Requirements
- [ ] Add `spektacular adhoc --output <cli|json>` output mode selection
- [ ] Support optional persisted default output mode via config
- [ ] In JSON mode, use JSONL over stdin/stdout for bidirectional messaging
- [ ] Define a common message envelope with versioning and correlation IDs
- [ ] Define standardized inbound message types: `run.start`, `run.input`, `run.cancel`
- [ ] Define standardized outbound message types: `run.started`, `run.progress`, `run.question`, `run.artifact`, `run.completed`, `run.failed`, `run.cancelled`
- [ ] Define tool-agnostic question schema for interactive prompts
- [ ] Ensure exactly one terminal event per run
- [ ] Define compatibility behavior for unknown fields/types and version mismatch
- [ ] Define safe handling for secrets and diagnostics

## Constraints
- Must be provider/tool agnostic (no provider-specific protocol coupling)
- Must not break existing CLI UX when `--output` is omitted (default remains `cli`)
- JSON protocol frames must only be emitted on stdout (stderr reserved for non-protocol diagnostics)
- Protocol framing must be line-delimited JSON (single object per line)
- Terminal events must be deterministic and unambiguous

## Acceptance Criteria
- [ ] `spektacular adhoc --output json` emits valid JSONL frames only
- [ ] `run.start` receives `run.started` within a reasonable timeout
- [ ] Interactive question flow works end-to-end (`run.question` <-> `run.input`)
- [ ] Exactly one terminal event is emitted per run (`run.completed|run.failed|run.cancelled`)
- [ ] Version mismatch emits `run.failed` with `unsupported_version`
- [ ] Exit codes are consistent with terminal state

## Technical Approach

### Command Interface
- Add output selector:
  - `spektacular adhoc --output cli`
  - `spektacular adhoc --output json`
- Add optional persisted config:
  - `spektacular config set output cli`
  - `spektacular config set output json`
- CLI flag overrides config value for current invocation.

### Transport and Framing
- stdin: inbound control/input events
- stdout: outbound protocol events
- stderr: diagnostics only (not protocol)
- Framing: JSON Lines (JSONL), UTF-8, one message per line

### Message Envelope (all events)
```json
{
  "v": "1",
  "id": "msg_123",
  "ts": "2026-02-20T10:00:00Z",
  "type": "run.started",
  "run_id": "run_abc",
  "payload": {}
}
```

### Inbound Event Types (stdin)
- `run.start` : start adhoc execution
- `run.input` : answer to `run.question`
- `run.cancel` : request cancellation

### Outbound Event Types (stdout)
- `run.started` : acknowledged start + resolved config
- `run.progress` : status/log update
- `run.question` : normalized interaction request
- `run.artifact` : generated artifact metadata
- `run.completed` : terminal success
- `run.failed` : terminal failure
- `run.cancelled` : terminal cancellation

### Question Abstraction
Normalize provider-specific prompts to:
- `question_id`
- `kind` (`confirm|select|text|secret|file_path`)
- `text`
- `options` (optional)
- `default` (optional)
- `validation` (optional)
- `required` (default true)
- `timeout_s` (optional)

### Lifecycle
1. receive `run.start`
2. emit `run.started`
3. emit zero or more `run.progress|run.question|run.artifact`
4. emit exactly one terminal event (`run.completed|run.failed|run.cancelled`)

### Exit Codes
- `0` for completed/cancelled runs
- non-zero for failed runtime/protocol conditions (emit `run.failed` first when possible)

### Compatibility Rules
- Unknown fields: ignore
- Unknown event types: ignore with warning
- Unsupported version: fail with `unsupported_version`

### Security
- Never emit secrets in progress/failure logs
- Redact `secret` answers from logs/artifacts
- Keep provider/internal details out of protocol surface

## Success Metrics
- Protocol interoperability across at least 3 providers (Claude/OpenAI/Gemini adapters)
- Stable orchestration in OpenClaw without provider-specific branching
- Reduced integration complexity for external orchestrators (single protocol implementation)
- Zero ambiguous terminal states in integration tests

## Non-Goals
- Full spec-driven planning/integration workflow (handled by existing plan mode)
- Defining provider-specific implementation internals
- Building network transport protocol in this phase (stdio only)
- Multi-run multiplexing in one process (single run lifecycle is sufficient for initial release)
