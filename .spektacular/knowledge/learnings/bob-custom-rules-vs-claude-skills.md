# Bob Custom Rules: Concept, Comparison to Claude Skills, and Technical Reference

---

## What Are Custom Rules?

Custom rules are discrete, markdown-based instruction files that Bob loads into its context to shape how it responds — influencing coding style, documentation standards, decision-making processes, and team conventions. Unlike modes (which are mutually exclusive operational contexts), rules are **additive and composable**: multiple rule files from multiple sources are combined and loaded simultaneously, making them the fundamental building block for packaging and distributing reusable AI behaviors.

Rules do not change *what* Bob can do — they change *how* Bob does it. They are persistent behavioral guidelines that apply automatically without requiring any prompt-level instruction from the user.

---

## Comparison to Claude Skills

Claude skills are discrete, reusable units of capability or behavior that can be loaded into a plan independently and in combination with other skills. A plan can load many skills at once, each contributing a focused domain of knowledge or behavioral instruction.

Bob custom rules map directly to this concept:

| Claude Concept | Bob Equivalent |
|---|---|
| Skill | A single rule file (e.g., `typescript.md`, `testing.md`) |
| Skill library / skill set | A rules directory (`.bob/rules/`) |
| Plan loading multiple skills | Bob loading all files from `rules/` and `rules-{mode}/` simultaneously |
| Globally installed skill | Global rule in `~/.bob/rules/` |
| Project-scoped skill | Workspace rule in `.bob/rules/` |
| Shared/team skill | Version-controlled `.bob/rules/` committed to a repo, or `AGENTS.md` |

The critical parallel is **composability**. Just as a plan loads multiple skills and they all contribute to the AI's behavior simultaneously, Bob reads *all* files in the rules directories and stacks them together into a single combined context. Each file is an independent, focused unit — a skill — and they compose without conflict.

The main difference is scope: Claude skills can also package tool permissions and external integrations, whereas Bob rules are purely instruction-level. To replicate a fully featured skill that includes tool access, you would pair rules (behavioral instructions) with MCP servers (external tool capabilities) and optionally custom modes (tool permission scoping).

---

## Technical Reference

### Rule Scopes

Bob supports two scopes that determine where rules apply:

**Global rules** apply automatically across all projects and all workspaces. They are the equivalent of a globally installed skill — always active regardless of what project is open. Store these in:
- Linux/macOS: `~/.bob/rules/`
- Windows: `%USERPROFILE%\.bob\rules\`

**Workspace rules** apply only within the current project. They are project-scoped skills, appropriate for standards, conventions, or workflows specific to one codebase. Store these in `.bob/rules/` at the project root.

### Configuration Methods

**File-based (simple)** — place a single file directly in your workspace root:
- `.bobrules` — general rules for all modes
- `.bobrules-code` — rules scoped to Code mode only
- `.bobrules-{modeSlug}` — rules scoped to any named mode

**Directory-based (recommended)** — use structured directories for organization and multi-file composability:

```
.bob/
├── rules/             # General rules, all modes
│   ├── coding-style.md
│   ├── testing.md
│   └── documentation.md
├── rules-code/        # Code mode only
│   └── typescript.md
├── rules-plan/        # Plan mode only
│   └── architecture.md
└── rules-{modeSlug}/  # Any custom mode
    └── custom-skill.md
```

### Rule Loading Priority

Bob combines rules from all sources in this order, with later entries able to override earlier ones:

1. Global rules (`~/.bob/rules/`)
2. Workspace rules (`.bob/rules/`)
3. Within each level, mode-specific rules load before general rules

Within a directory, files are processed in **alphabetical order by filename**, which allows you to control load sequence by prefixing filenames (e.g., `01-base.md`, `02-overrides.md`).

### AGENTS.md

`AGENTS.md` is a special auto-loaded file placed in the workspace root. It functions as a pre-installed shared skill — automatically loaded for all team members without any configuration. It sits in the loading order after mode-specific rules but before general workspace rules.

Disable it per-user with `"bob-cline.useAgentRules": false` in settings if needed.

### File Behavior

- Bob reads all files in rules directories **recursively**, including subdirectories
- Files are processed in **alphabetical order** by filename
- Cache and system files are automatically excluded (`.DS_Store`, `*.bak`, `*.cache`, `*.log`, `*.tmp`, `Thumbs.db`)
- Symbolic links are supported up to a maximum depth of 5
- Empty files are silently skipped

---

## Best Practices

**Treat each file as a single-purpose skill.** Rather than putting all your rules in one file, split them by domain — one file for code style, one for testing standards, one for documentation requirements. This makes individual rules easy to maintain, override, or share independently, exactly like discrete skills.

**Be specific and actionable.** Rules that are vague produce inconsistent results. State exactly what you want:
- Good: `"Use 4 spaces for indentation in JavaScript files"`
- Avoid: `"Format code nicely"`

**Use clear internal structure.** Organize each rule file with markdown headings to group related directives. This improves both human readability and Bob's ability to parse intent:

```markdown
# Code Style
- Use camelCase for variables
- Use PascalCase for classes

# Testing
- Write unit tests for all public functions
- Use Jest as the testing framework

# Documentation
- Add JSDoc comments for all public APIs
```

**Use alphabetical prefixes to control load order.** If some rules are foundational and others are overrides, prefix filenames numerically to make the sequence explicit: `01-base-style.md`, `02-project-overrides.md`.

**Version-control your workspace rules.** Committing `.bob/rules/` to your repository ensures every team member gets the same set of skills automatically, and rule changes are tracked like code changes:

```bash
git add .bob/rules/
git commit -m "Add Bob custom rules"
```

**Use AGENTS.md for zero-configuration team skills.** For standards that every team member must always have active, `AGENTS.md` is the lowest-friction distribution mechanism — no clone step, no setup required.

**Use global rules for personal or org-wide standards, workspace rules for project specifics.** This hybrid model mirrors the skill library pattern: a shared global foundation that every project inherits, with project-specific skills layered on top and overriding where needed.

**Avoid rule conflicts through naming conventions.** Project-level workspace rules override global rules of the same concern, so document which rules are intentional overrides and why, particularly in team environments where global rules may be centrally managed.

---

## Summary

Bob custom rules are the functional equivalent of Claude skills — discrete, composable, markdown-based instruction units that stack together to define the AI's behavior for a given context. A well-organized rules directory is a skill library: each file is a skill, the directory is the library, and Bob loads them all simultaneously into every session, just as a plan loads multiple skills at once.
