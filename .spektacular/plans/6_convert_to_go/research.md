# Convert Spektacular to Go - Research Notes

## Specification Analysis
- **Original Requirements**: Convert Python CLI to Go, preserve all functionality, use Cobra + Bubble Tea, full test coverage
- **Implicit Requirements**: Embedded file support (agent prompts, templates), YAML config compatibility, cross-platform binary distribution, stream-JSON parsing for Claude CLI output
- **Constraints Identified**: Must maintain `.spektacular/` directory format compatibility, config YAML format must remain identical

## Research Process

### Sub-agents Spawned
1. **Codebase Explorer** — Complete inventory of all Python source files, tests, and defaults
2. **Architecture Analyst** — Traced data flow from CLI → config → runner → plan → TUI

### Files Examined

| File | Lines | Purpose |
|---|---|---|
| `src/spektacular/__init__.py` | 3 | Version definition |
| `src/spektacular/cli.py` | 87 | CLI entry point (Click framework) |
| `src/spektacular/config.py` | 132 | Pydantic config with YAML I/O |
| `src/spektacular/runner.py` | 178 | Claude subprocess + stream-JSON parser |
| `src/spektacular/plan.py` | 116 | Plan orchestration loop |
| `src/spektacular/tui.py` | 545 | Textual TUI with themes + Q&A |
| `src/spektacular/init.py` | 79 | Project scaffolding |
| `src/spektacular/spec.py` | 72 | Spec template creation |
| `tests/test_config.py` | 67 | Config tests |
| `tests/test_plan.py` | 55 | Plan function tests |
| `tests/test_runner.py` | 210 | Runner + event parsing tests |
| `tests/test_tui.py` | 540 | Comprehensive TUI tests |
| `pyproject.toml` | 34 | Project metadata + deps |
| `defaults/agents/planner.md` | 287 | Planning agent prompt |
| `defaults/agents/executor.md` | 309 | Execution agent prompt |
| `defaults/spec-template.md` | 28 | Spec template |
| `defaults/conventions.md` | 15 | Coding standards |
| `defaults/.gitignore` | 8 | Default gitignore |

**Total Python source**: ~1,212 lines
**Total tests**: ~872 lines

### Patterns Discovered

#### CLI Pattern (Click → Cobra)
- Python uses Click's `@click.group()` / `@click.command()` decorator pattern
- Go equivalent: Cobra's `cobra.Command` structs with `RunE` functions
- Both support subcommands, flags, and arguments similarly

#### Config Pattern (Pydantic → Go structs + yaml tags)
- Python: Nested Pydantic `BaseModel` classes with `Field(default=...)` and validators
- Go: Nested structs with `yaml:"field_name"` tags, constructor function for defaults
- Key behavior to preserve: `${VAR_NAME}` env var expansion in YAML values
- Python expands after parsing; Go should expand in raw YAML string before parsing

#### Process Runner Pattern (subprocess → os/exec)
- Python: `subprocess.Popen` with `stdout=PIPE`, line-by-line iteration, generator yield
- Go: `exec.Command` with `StdoutPipe()`, `bufio.Scanner`, channel-based event delivery
- Key difference: Python generators are synchronous pull; Go channels are async push
- Stderr handling: Python uses background thread; Go can use `io.Discard` or similar

#### TUI Pattern (Textual → Bubble Tea)
- **Textual** (Python): Widget-based composition, CSS styling, async workers, message passing
- **Bubble Tea** (Go): Elm architecture (Model/Update/View), lipgloss styling, tea.Cmd for async

Critical mapping differences:
| Feature | Textual | Bubble Tea |
|---|---|---|
| Architecture | Widget tree + CSS | Single model + View string |
| Async work | `@work(thread=True)` decorator | `tea.Cmd` returning `tea.Msg` |
| Styling | CSS selectors | lipgloss inline styles |
| Scrolling | `VerticalScroll` container | `viewport.Model` from bubbles |
| Rich text | `RichLog` + Rich markdown | `glamour` markdown renderer |
| Events | Textual `Message` classes | Go types implementing `tea.Msg` |
| Key bindings | `BINDINGS` list | `tea.KeyMsg` pattern matching |

#### Embedded Files Pattern (importlib.resources → embed.FS)
- Python: `importlib.resources.files("spektacular").joinpath("defaults/file.md")`
- Go: `//go:embed files/*` directive + `embed.FS`
- Go's embed is simpler and more reliable — no package resolution issues

#### Question Detection Pattern
- Both use regex: `<!--QUESTION:(.*?)-->` with DOTALL flag
- Go's `regexp` supports this via `(?s)` flag or using `FindAllStringSubmatch`
- JSON parsing: Python `json.loads` → Go `json.Unmarshal`

#### Testing Pattern
- Python: pytest + pytest-asyncio, `unittest.mock.patch`, `tmp_path` fixture
- Go: `testing` package + testify/require, `t.TempDir()`, interface-based mocking
- TUI testing: Textual has built-in `app.run_test()` pilot; Bubble Tea has `teatest` package

## Key Findings

### Architecture Insights
The application follows a clean layered architecture:
```
CLI layer (cli.py) → Business logic (plan.py, init.py, spec.py)
                   → Infrastructure (runner.py, config.py)
                   → Presentation (tui.py)
```

This maps well to Go's package organization:
```
cmd/ → internal/plan, internal/project, internal/spec
     → internal/runner, internal/config
     → internal/tui
```

### Data Flow: Plan Command
1. CLI parses args, loads config
2. TUI app mounts, reads spec file and knowledge files
3. Builds combined prompt (agent prompt + knowledge + spec)
4. Spawns Claude subprocess with stream-JSON output
5. Parses events: text → display, questions → interactive panel, result → write file
6. Session resumption: answers sent as new prompt with `--resume SESSION_ID`
7. Final result written to `.spektacular/plans/{name}/plan.md`

### Reusable Components
- Agent prompt files are pure markdown — copy directly
- Spec template is pure markdown — copy directly
- Config YAML format is identical between Python and Go
- `.spektacular/` directory structure is language-agnostic

## Design Decisions

### Decision: Use `map[string]interface{}` for ClaudeEvent.Data
- **Options**: Typed structs vs dynamic map
- **Rationale**: Claude's stream-JSON has variable structure per event type. Using `map[string]interface{}` matches the Python approach (`dict`) and avoids needing to define structs for every possible event shape. Accessor methods provide type-safe access to common fields.
- **Trade-offs**: Less type safety, but more flexible and matches the dynamic nature of the data.

### Decision: Channel-based event streaming
- **Options**: Callback functions, channel-based, iterator pattern
- **Rationale**: Go channels are the idiomatic way to stream data from goroutines. They map naturally to the Python generator pattern but work asynchronously. Bubble Tea's `tea.Cmd` pattern integrates well with channels.
- **Trade-offs**: Slightly more complex error handling (separate error channel), but cleaner separation of concerns.

### Decision: Glamour for markdown rendering
- **Options**: `glamour`, `goldmark` + custom renderer, raw ANSI
- **Rationale**: Glamour is part of the Charm ecosystem (same as Bubble Tea and lipgloss), so it integrates seamlessly. It handles terminal-width-aware markdown rendering out of the box.
- **Trade-offs**: Theme customization is different from Rich — may need custom glamour styles to match the 5 existing themes.

### Decision: `internal/` package layout
- **Options**: Flat package, `pkg/` convention, `internal/`
- **Rationale**: `internal/` enforces that packages are not importable by external consumers. This is appropriate since Spektacular is a CLI tool, not a library. Follows Go standard project layout recommendations.

### Decision: `embed.FS` for default files
- **Options**: `embed.FS`, bundled as string constants, runtime file loading
- **Rationale**: `embed.FS` is the standard Go approach for bundling static assets. It's compile-time safe, doesn't require external files at runtime, and supports directory structure.

### Decision: Env var expansion on raw YAML string
- **Options**: Expand before parsing (string replacement), expand after parsing (walk struct)
- **Rationale**: Expanding `${VAR}` patterns in the raw YAML string before `yaml.Unmarshal` is simpler and handles all value types uniformly. The Python version does post-parse expansion on the dict, but pre-parse string expansion is equivalent for the `${VAR}` pattern used.

## Code Examples & Patterns

### Python Generator → Go Channel Pattern
```python
# Python (runner.py:116-168)
def run_claude(prompt, config, session_id=None, cwd=None, command="unknown"):
    process = subprocess.Popen(cmd, stdout=subprocess.PIPE, ...)
    for line in process.stdout:
        data = json.loads(line)
        yield ClaudeEvent(type=data.get("type"), data=data)
```

```go
// Go equivalent
func RunClaude(opts RunOptions) (<-chan ClaudeEvent, <-chan error) {
    events := make(chan ClaudeEvent, 64)
    errc := make(chan error, 1)
    go func() {
        defer close(events)
        scanner := bufio.NewScanner(stdout)
        for scanner.Scan() {
            var data map[string]interface{}
            json.Unmarshal(scanner.Bytes(), &data)
            events <- ClaudeEvent{Type: data["type"].(string), Data: data}
        }
    }()
    return events, errc
}
```

### Pydantic Model → Go Struct Pattern
```python
# Python (config.py:10-13)
class ApiConfig(BaseModel):
    anthropic_api_key: str = Field(default="${ANTHROPIC_API_KEY}")
    timeout: int = Field(default=60)
```

```go
// Go equivalent
type APIConfig struct {
    AnthropicAPIKey string `yaml:"anthropic_api_key"`
    Timeout         int    `yaml:"timeout"`
}
// Defaults set in NewDefault() constructor
```

### Textual Message → Bubble Tea Msg Pattern
```python
# Python (tui.py:184-225)
class AgentOutput(Message):
    def __init__(self, text: str) -> None:
        super().__init__()
        self.text = text
```

```go
// Go equivalent
type agentOutputMsg struct{ text string }
// Used in Update() via type switch:
// case agentOutputMsg: ...
```

## Open Questions (All Resolved)

All questions resolved through code analysis:

1. **Q**: Should the Go binary replace the Python package or coexist?
   **A**: Coexist initially. The Go binary is a new `main.go` at the project root. The Python source remains untouched. Distribution switches to Go binary.

2. **Q**: How to handle the Textual → Bubble Tea TUI differences?
   **A**: Elm architecture pattern in Bubble Tea. Viewport for scrollable output. Glamour for markdown. Lipgloss for styling. The mapping is well-defined (see design decisions above).

3. **Q**: How to embed default files?
   **A**: Go's `//go:embed files/*` directive in `internal/defaults/defaults.go`. Files copied to `internal/defaults/files/` directory.

4. **Q**: How to handle the `.gitignore` default that references Python?
   **A**: Update the embedded `.gitignore` to be language-agnostic (remove `__pycache__`, add Go binary name).

5. **Q**: What Go testing library to use?
   **A**: Standard `testing` package + `github.com/stretchr/testify` for assertions. `t.TempDir()` replaces pytest's `tmp_path`. For TUI testing, use `charmbracelet/x/exp/teatest`.
