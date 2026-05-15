# Convert Spektacular to Go - Context

## Quick Summary
Rewrite the Spektacular CLI from Python to Go for single-binary distribution. Uses Cobra for CLI, Bubble Tea for TUI, and Go embed for default files. All existing functionality (init, new, plan, run) is preserved with identical `.spektacular/` directory format.

## Key Files & Locations

### Go Implementation (new)
- **Entry point**: `main.go`
- **CLI commands**: `cmd/root.go`, `cmd/init.go`, `cmd/new.go`, `cmd/plan.go`, `cmd/run.go`
- **Configuration**: `internal/config/config.go`
- **Claude runner**: `internal/runner/runner.go`
- **Plan orchestration**: `internal/plan/plan.go`
- **TUI**: `internal/tui/tui.go`, `internal/tui/theme.go`, `internal/tui/question.go`
- **Project init**: `internal/project/init.go`
- **Spec creation**: `internal/spec/spec.go`
- **Embedded defaults**: `internal/defaults/defaults.go`, `internal/defaults/files/`
- **Tests**: `internal/*/.*_test.go`

### Python Reference (existing, read-only)
- **CLI**: `src/spektacular/cli.py`
- **Config**: `src/spektacular/config.py`
- **Runner**: `src/spektacular/runner.py`
- **Plan**: `src/spektacular/plan.py`
- **TUI**: `src/spektacular/tui.py`
- **Init**: `src/spektacular/init.py`
- **Spec**: `src/spektacular/spec.py`
- **Tests**: `tests/test_*.py`

### Shared (unchanged)
- **Config format**: `.spektacular/config.yaml`
- **Agent prompts**: `agents/planner.md`, `agents/executor.md`
- **Spec template**: `spec-template.md`
- **Directory structure**: `.spektacular/{plans,specs,knowledge}/`

## Dependencies

### Go Dependencies
- **CLI**: `github.com/spf13/cobra` v1.8+
- **TUI**: `github.com/charmbracelet/bubbletea` v1.2+
- **TUI components**: `github.com/charmbracelet/bubbles` v0.20+
- **Styling**: `github.com/charmbracelet/lipgloss` v1.0+
- **Markdown**: `github.com/charmbracelet/glamour` v0.8+
- **YAML**: `gopkg.in/yaml.v3`
- **Testing**: `github.com/stretchr/testify` v1.9+

### Python Dependencies (reference)
- click, textual, pydantic, pyyaml, anthropic, jinja2, markdown, rich

### External Dependencies
- `claude` CLI must be installed and on PATH (unchanged)

## Module Mapping

| Python → Go | Description |
|---|---|
| `cli.py` (Click) → `cmd/` (Cobra) | CLI commands |
| `config.py` (Pydantic) → `internal/config/` | YAML config |
| `runner.py` → `internal/runner/` | Claude subprocess |
| `plan.py` → `internal/plan/` | Plan orchestration |
| `tui.py` (Textual) → `internal/tui/` (Bubble Tea) | Interactive TUI |
| `init.py` → `internal/project/` | Project scaffolding |
| `spec.py` → `internal/spec/` | Spec creation |
| `defaults/` → `internal/defaults/` (embed.FS) | Embedded files |

## Build & Test Commands
- **Build**: `go build -o spektacular .`
- **Test**: `go test ./...`
- **Lint**: `go vet ./...`
- **Cross-compile**: `GOOS=darwin GOARCH=arm64 go build -o spektacular-darwin-arm64 .`
