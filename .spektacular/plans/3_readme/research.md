# 3 Readme - Research Notes

## Specification Analysis

### Original Requirements
- Simple README for the project
- Explains what Spektacular is, how to use it, how to contribute
- Clear and concise language
- At least one screenshot of the TUI in action
- Simple example of CLI usage for plan mode
- Reference `.spektacular/knowledge/architecture/initial-idea.md` for content

### Implicit Requirements
- Installation instructions (Python 3.12+, uv/pip)
- Prerequisites (Claude CLI must be installed)
- Project status indicator (v0.1, early/alpha)
- License information (Apache 2.0)
- Links to related tools/inspiration

### Constraints
- Simple language, no jargon
- Concise and to the point

## Research Process

### Files Examined

| File | Purpose | Relevance |
|------|---------|-----------|
| `README.md` | Existing README | 2-line placeholder to be replaced |
| `pyproject.toml` | Project metadata | Version (0.1.0), dependencies, Python version |
| `LICENSE` | License file | Apache License 2.0 |
| `src/spektacular/cli.py` | CLI commands | 4 commands: init, new, plan, run |
| `src/spektacular/tui.py` | TUI implementation | Textual-based, 5 themes, interactive Q&A |
| `src/spektacular/__init__.py` | Package init | Version = "0.1.0" |
| `src/spektacular/config.py` | Configuration | Model tiers, complexity thresholds |
| `.spektacular/config.yaml` | Default config | Example configuration |
| `spektacular-v0.1-spec.md` | Bootstrap spec | MVP scope definition |
| `.spektacular/knowledge/architecture/initial-idea.md` | Architecture doc | Core concepts, vision, differentiators |
| `src/spektacular/defaults/spec-template.md` | Spec template | Shows spec format |

### Key Findings

#### What Spektacular Is (from architecture doc)
- Multi-surface orchestration platform for spec-driven development
- Takes markdown specs as input, orchestrates AI coding agents
- Pipeline: analyse -> plan -> execute -> validate
- Agent-agnostic: works with Claude Code, Aider, Cursor, etc.
- Intelligent model routing based on complexity scoring
- "Control plane for AI coding" - like Kubernetes for coding agents

#### Current State (v0.1)
- Python CLI built with Click
- Commands: `init`, `new`, `plan`, `run` (run is stub)
- TUI built with Textual (5 color themes)
- Uses Claude Code as execution agent
- Generates plan.md, research.md, context.md artifacts
- Configuration via `.spektacular/config.yaml`
- Knowledge base in `.spektacular/knowledge/`

#### How to Install
- Python 3.12+ required
- Package manager: uv (preferred) or pip
- Claude CLI must be installed separately
- `uv pip install -e .` for development

#### How to Use
1. `spektacular init` - Create project structure
2. `spektacular new my-feature` - Create spec from template
3. Edit the spec file
4. `spektacular plan .spektacular/specs/my-feature.md` - Generate plan via TUI

#### No Existing Screenshots
- No images in the repository
- TUI uses Textual framework with Rich markdown rendering
- Need to either capture a screenshot or create a placeholder

## Design Decisions

### README Structure
- **Decision**: Use a standard open-source README format with: tagline, overview, features, quickstart, usage, contributing, license
- **Rationale**: Familiar structure for developers, easy to scan
- **Trade-offs**: Could be more creative, but clarity > creativity for a dev tool

### Content Depth
- **Decision**: Focus on current v0.1 capabilities, mention future vision briefly
- **Rationale**: README should reflect what works today, not aspirational features
- **Trade-offs**: Less impressive than showing the full vision, but more honest

### Screenshot Approach
- **Decision**: Include a placeholder reference for TUI screenshot
- **Rationale**: No screenshots exist yet; README should be ready for one when captured
- **Trade-offs**: README will have a gap until screenshot is added

### Tone
- **Decision**: Developer-friendly, concise, practical
- **Rationale**: Spec requires "simple language, no jargon" and "concise"
- **Trade-offs**: Won't include the enterprise pitch from architecture doc

## Open Questions (Resolved)

- **Q**: Should we include the enterprise vision?
- **A**: Brief mention only. Focus on what v0.1 does today.

- **Q**: How to handle the `run` command which is a stub?
- **A**: Don't document it. Only document working commands.

- **Q**: Should we include API/programmatic usage?
- **A**: No. CLI-only for now per v0.1 scope.
