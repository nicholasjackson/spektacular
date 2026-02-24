# Adhoc JSON Protocol - Research Notes

## Specification Analysis

### Original Requirements
1. Add `spektacular adhoc --output <cli|json>` output mode selection
2. Support optional persisted default output mode via config
3. In JSON mode, use JSONL over stdin/stdout for bidirectional messaging
4. Define common message envelope with versioning and correlation IDs
5. Define inbound message types: `run.start`, `run.input`, `run.cancel`
6. Define outbound message types: `run.started`, `run.progress`, `run.question`, `run.artifact`, `run.completed`, `run.failed`, `run.cancelled`
7. Define tool-agnostic question schema for interactive prompts
8. Ensure exactly one terminal event per run
9. Define compatibility behavior for unknown fields/types and version mismatch
10. Define safe handling for secrets and diagnostics

### Implicit Requirements
- The `run` command needs to load specs, build prompts, and run Claude — reusing existing infrastructure
- CLI mode should feel like a simpler version of existing `plan` command output
- JSON mode must be machine-parseable with no human-formatted text on stdout
- The protocol must bridge between Claude's `<!--QUESTION:...-->` format and the normalized question schema

### Constraints Identified
- stdout is exclusively for JSONL frames in JSON mode (no mixing with human text)
- stderr is for diagnostics only (not protocol)
- Must not break existing commands when `--output` is omitted
- Protocol must work regardless of which agent backend is used (Claude, OpenAI, etc.)
- Single run lifecycle per process invocation

## Research Process

### Sub-agents Spawned
1. **Codebase Explorer** — Full project structure inventory, all source files
2. **CLI Command Analyst** — How Click commands are registered, existing patterns
3. **Test Infrastructure Analyst** — Testing frameworks, mocking patterns, fixtures

### Files Examined

| File | Lines | Key Findings |
|------|-------|-------------|
| `src/spektacular/cli.py` | 1-88 | Click-based CLI with 4 commands. `run` is placeholder (TODO at line 43). Pattern: load config, call handler function. |
| `src/spektacular/config.py` | 1-133 | Pydantic BaseModel hierarchy. `OutputConfig` at line 41-44 has `format` and `include_metadata`. YAML round-trip via `from_yaml_file()`/`to_yaml_file()`. |
| `src/spektacular/runner.py` | 1-179 | `run_claude()` generator at line 116. Spawns subprocess, yields `ClaudeEvent`. `detect_questions()` at line 71 parses `<!--QUESTION:...-->`. `Question` dataclass at line 60. |
| `src/spektacular/plan.py` | 1-117 | `run_plan()` at line 64 shows the complete question-answer loop pattern. `load_knowledge()`, `load_agent_prompt()`, `build_prompt()` utilities. |
| `src/spektacular/tui.py` | 1-546 | Textual TUI with message-passing. `AgentOutput`, `AgentQuestion`, `AgentComplete`, `AgentError` messages. Shows how events flow from runner to UI. |
| `tests/test_runner.py` | 1-210 | Subprocess mocking pattern with `MagicMock` for `Popen`. Tests event parsing, question detection, debug logging. |
| `tests/test_config.py` | 1-67 | Config model tests with YAML round-trip. Uses `tmp_path` fixture. |
| `pyproject.toml` | 1-31 | Python 3.12+, Click 8.3.1+, Pydantic 2.0+, pytest with asyncio support. |

### Patterns Discovered

#### CLI Command Pattern (`cli.py`)
```python
@cli.command()
@click.argument("spec_file", type=click.Path(exists=True, path_type=Path))
def command(spec_file):
    project_path = Path.cwd()
    config_path = project_path / ".spektacular" / "config.yaml"
    if config_path.exists():
        config = SpektacularConfig.from_yaml_file(config_path)
    else:
        config = SpektacularConfig()
    # ... call handler
```

#### Runner Event Loop Pattern (`plan.py:88-109`)
```python
current_prompt = prompt
while True:
    questions_found = []
    for event in run_claude(current_prompt, config, session_id, project_path, command="plan"):
        if event.session_id:
            session_id = event.session_id
        if text := event.text_content:
            # process text
            questions_found.extend(detect_questions(text))
        if event.is_result:
            if event.is_error:
                raise RuntimeError(...)
            final_result = event.result_text
    if questions_found:
        answer = get_answer(questions_found)
        current_prompt = answer
        continue
    break
```

#### Subprocess Mocking Pattern (`test_runner.py:121-130`)
```python
mock_process = MagicMock()
mock_process.stdout = iter(events)
mock_process.stderr.read.return_value = ""
mock_process.returncode = 0
with patch("spektacular.runner.subprocess.Popen", return_value=mock_process):
    result = list(run_claude("prompt", config))
```

#### Question Detection Pattern (`runner.py:68-85`)
```python
QUESTION_PATTERN = re.compile(r"<!--QUESTION:(.*?)-->", re.DOTALL)
# Parses JSON from HTML comment markers
# Returns list[Question] with question, header, options fields
```

## Key Findings

### Architecture Insights
1. The existing `run_claude()` generator is the core execution engine — all commands use it
2. Question detection happens by scanning text content for HTML comment markers
3. The question-answer loop is a `while True` pattern: run agent -> detect questions -> get answers -> resume
4. Session IDs enable Claude process resumption across question rounds
5. All output currently goes through Click's `echo()` or Textual's message system

### Existing Implementations
- `plan.py:run_plan()` (line 64-116): Non-TUI plan generation with CLI Q&A — closest model for the CLI output mode
- `tui.py:PlanTUI._run_agent()` (line 416-450): TUI plan generation with message-passing — shows event-to-UI bridging
- `tui.py:AgentOutput/AgentQuestion/AgentComplete/AgentError` messages: Internal event protocol for TUI — analogous to JSONL protocol events

### Reusable Components
- `run_claude()` — Core subprocess runner (no changes needed)
- `detect_questions()` — Question extraction from text
- `build_prompt()` — Prompt construction from spec + agent + knowledge
- `load_knowledge()`, `load_agent_prompt()` — Resource loading utilities
- `SpektacularConfig.from_yaml_file()` — Config loading with env var expansion
- `Question` dataclass — Internal question model (needs mapping to protocol model)

### Testing Infrastructure
- pytest with `tmp_path` fixture for filesystem operations
- `unittest.mock.patch` for subprocess mocking
- `MagicMock` for process simulation
- `pytest.raises` for error case testing
- `capsys` available for stdout/stderr capture in non-TUI tests

## Questions & Answers

### Q: Should the run command be separate (`adhoc`) or replace the existing `run` placeholder?
**A (user decision)**: Replace `run`. The placeholder at `cli.py:38-44` is a TODO that was always meant to be implemented. Adding `--output` to it is cleaner than creating a parallel command.

### Q: Should we wire to `run_claude()` immediately or stub the executor?
**A (user decision)**: Full integration. Wire directly to `run_claude()` from day one. No stub executor — the existing subprocess runner is mature enough.

### Q: Should `spektacular config set/get` be implemented in this phase?
**A (user decision)**: Defer. Only support `--output` CLI flag and manual config.yaml editing. The `config` command group can be added later.

### Q: Where should the protocol models live?
**Decision**: New file `src/spektacular/protocol.py`. The models are substantial enough to warrant their own module rather than being crammed into `runner.py` or the handler module. This also makes them importable by external tools that want to construct/parse protocol messages.

### Q: How to handle the bridge between Claude's question format and the protocol format?
**Decision**: A `_question_from_runner()` conversion function in `adhoc.py` that maps `Question(question, header, options)` to `QuestionPayload(question_id, kind, text, header, options, ...)`. This keeps the conversion logic close to where it's used and doesn't pollute the protocol models with runner-specific knowledge.

### Q: Should JSON mode require `run.start` on stdin or auto-start?
**Decision**: Require `run.start`. The spec mandates a lifecycle: receive `run.start` -> emit `run.started` -> work -> terminal event. This enables orchestrators to control timing and pass config overrides.

### Q: How to ensure exactly one terminal event?
**Decision**: A `terminal_sent` boolean flag in `run_adhoc_json()`. Set to `True` after emitting any terminal event. The `finally` block checks this flag and emits `run.failed` if no terminal event was sent. This handles all edge cases including exceptions.

## Design Decisions

### Decision: Pydantic for Protocol Models
**Options Considered**: Plain dataclasses, TypedDict, Pydantic BaseModel
**Rationale**: Pydantic provides JSON serialization (`.model_dump_json()`), validation, and schema generation. Already a project dependency. TypedDict would require manual serialization. Dataclasses would need custom JSON handling.
**Trade-offs**: Slightly heavier than dataclasses, but the validation and serialization benefits outweigh the cost.

### Decision: Separate `protocol.py` Module
**Options Considered**: Protocol models in `runner.py`, in `adhoc.py`, or in separate module
**Rationale**: Protocol models are a standalone concern that may be imported by external tools, test utilities, and future modules (e.g., MCP server). Keeping them separate maintains single-responsibility.
**Trade-offs**: One more file to maintain, but cleaner architecture.

### Decision: Replace `run` Instead of New `adhoc` Command
**Options Considered**: New `adhoc` command, replace `run`, subcommand `run --adhoc`
**Rationale**: User decision. The existing `run` command is a TODO placeholder — replacing it is the cleanest path. No need for a separate command that duplicates the concept.
**Trade-offs**: Changes the `run` command signature (adds `--output` flag), but since `run` was unimplemented, there's no backwards compatibility concern.

### Decision: `_emit()` and `_read_event()` as Module-Level Functions
**Options Considered**: Class-based protocol handler, standalone functions, context manager
**Rationale**: Simple functions are easier to test (patch `sys.stdout`/`sys.stdin`), understand, and compose. A class would add unnecessary ceremony for what is essentially "write JSON line" and "read JSON line".
**Trade-offs**: Less encapsulation than a class, but more Pythonic for this scope.

### Decision: CLI Flag Overrides Config (No `config` command)
**Resolution order**: `--output` flag > `config.output.adhoc_mode` > default `"cli"`
**Rationale**: User chose to defer `config set/get`. Users can manually set `adhoc_mode` in config.yaml. Standard CLI convention: flags override persistent config.

## Code Examples & Patterns

### JSONL Emission Pattern
```python
# File: src/spektacular/adhoc.py
def _emit(envelope: Envelope) -> None:
    sys.stdout.write(envelope.to_jsonl() + "\n")
    sys.stdout.flush()
```

### Terminal Event Safety Pattern
```python
terminal_sent = False
try:
    # ... work ...
    _emit(make_envelope(EventType.RUN_COMPLETED, run_id, payload))
    terminal_sent = True
except Exception as e:
    if not terminal_sent:
        _emit(make_envelope(EventType.RUN_FAILED, run_id, {"error": str(e)}))
        terminal_sent = True
finally:
    if not terminal_sent:
        _emit(make_envelope(EventType.RUN_FAILED, run_id, {"error": "Unexpected"}))
```

### Question Bridge Pattern
```python
# Internal: Question(question="Which?", header="H", options=[{"label": "A"}])
# Protocol: QuestionPayload(question_id="q_0", kind="select", text="Which?", ...)
def _question_from_runner(q: Question, idx: int) -> QuestionPayload:
    return QuestionPayload(
        question_id=f"q_{idx}",
        kind=QuestionKind.SELECT if q.options else QuestionKind.TEXT,
        text=q.question,
        header=q.header,
        options=[QuestionOption(label=o["label"], description=o.get("description")) for o in q.options],
    )
```

## Open Questions (All Resolved)

All questions have been resolved during the research phase and through user input. No blockers remain.
