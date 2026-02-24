# Adhoc JSON Protocol - Context

## Quick Summary
Replace the placeholder `spektacular run` command with a full implementation supporting dual output modes: human-readable CLI (default) and machine-readable JSONL protocol for orchestrator integration. Defines a bidirectional stdin/stdout protocol with versioned message envelopes, normalized question schema, and deterministic run lifecycle. Wires directly to the existing `run_claude()` subprocess runner.

## Key Files & Locations

### New Files (to create)
- **Protocol Models**: `src/spektacular/protocol.py` — Pydantic models for all message types (~180 lines)
- **Run Handlers**: `src/spektacular/adhoc.py` — CLI and JSON mode execution handlers (~180 lines)
- **Protocol Tests**: `tests/test_protocol.py` — Message model unit tests (~80 lines)
- **Handler Tests**: `tests/test_adhoc.py` — Handler integration tests (~200 lines)

### Modified Files
- **CLI Entry Point**: `src/spektacular/cli.py:38-44` — Replace placeholder `run` with full impl + `--output` flag
- **Config Models**: `src/spektacular/config.py:41-44` — Add `adhoc_mode` to `OutputConfig`
- **Config Tests**: `tests/test_config.py` — Add `TestOutputConfig` class

### Referenced Files (read-only)
- **Runner**: `src/spektacular/runner.py:116-179` — `run_claude()` subprocess runner
- **Question Detection**: `src/spektacular/runner.py:71-85` — `detect_questions()` parser
- **Plan Orchestration**: `src/spektacular/plan.py:64-116` — `run_plan()` event loop pattern (template for adhoc CLI mode)
- **Knowledge Loading**: `src/spektacular/plan.py:12-29` — `load_knowledge()`, `load_agent_prompt()`

## Dependencies

### Code Dependencies (internal)
- `spektacular.runner` — `run_claude()`, `ClaudeEvent`, `Question`, `detect_questions()`, `build_prompt()`
- `spektacular.plan` — `load_knowledge()`, `load_agent_prompt()`, `prompt_user_for_answer()`
- `spektacular.config` — `SpektacularConfig`, `OutputConfig`

### External Dependencies (already installed)
- `pydantic>=2.0.0` — Message model validation and JSON serialization
- `click>=8.3.1` — CLI command registration and option handling
- `pyyaml>=6.0.3` — Config persistence

### No New External Dependencies Required

## Environment Requirements

### Configuration Variables
- `output.adhoc_mode` in `.spektacular/config.yaml` — Optional persisted default ("cli" or "json")
- No new environment variables needed
- No `spektacular config set/get` command — manual config.yaml editing only (deferred)

### Feature Flags
- None — Feature is gated by the `--output` CLI flag

## Integration Points

### CLI Interface
- `spektacular run <spec_file>` — Default CLI mode (replaces TODO placeholder)
- `spektacular run --output json <spec_file>` — JSON protocol mode
- `spektacular run --output cli <spec_file>` — Explicit CLI mode

### JSONL Protocol (stdin/stdout)
- **Inbound** (stdin): `run.start`, `run.input`, `run.cancel`
- **Outbound** (stdout): `run.started`, `run.progress`, `run.question`, `run.artifact`, `run.completed`, `run.failed`, `run.cancelled`

### Exit Codes
- `0` — completed or cancelled
- Non-zero — failed or protocol error

## Design Decisions Summary
1. **Replace `run`** — Not a new `adhoc` command; implement the existing placeholder
2. **Full Integration** — Wires to `run_claude()` from day one, no stub executor
3. **Defer `config` command** — Only `--output` flag + manual config.yaml editing
4. **Separate `protocol.py`** — Protocol models as standalone importable module
5. **`terminal_sent` flag** — Guarantees exactly one terminal event via try/except/finally
6. **Flag > config > default** — Output mode resolution: `--output` > `config.output.adhoc_mode` > `"cli"`

## Message Envelope Reference

```json
{
  "v": "1",
  "id": "msg_<12-char-hex>",
  "ts": "2026-02-20T10:00:00+00:00",
  "type": "run.<event>",
  "run_id": "run_<id>",
  "payload": {}
}
```
