# Context: 19_codex_integration

## Current State Analysis

Today's `init` pipeline is a single hardcoded dispatcher with an inline commands table. The critical sites are:

- `cmd/init.go:14-19` — `initCmd` declares `Use: "init <claude|bob>"` and takes one positional argument.
- `cmd/init.go:21-25` — `runInit` validates the agent with a direct string comparison: `if agent != "claude" && agent != "bob"`. This is the place the new registry dispatch replaces.
- `cmd/init.go:48-56` — inline table of three workflow commands, each with hardcoded `claudeDest` and `bobDest` paths. Gets deleted; the equivalent data lives inside each agent implementation after the refactor.
- `cmd/init.go:61-89` — the per-command render-and-write loop. Mustache `{{command}}` rendering happens here; the new shared helpers preserve that rendering contract.

Templates:

- `templates/commands/spek-new.md`, `templates/commands/spek-plan.md`, `templates/commands/spek-implement.md` — the three full-content command files. Their bodies move into SKILL.md files; these three files are deleted.
- `templates/skills/skill_*.md` — seven on-demand helper skills. Left untouched by this plan; retrieval via `cmd/skill.go` stays.
- `templates/templates.go:7` — `//go:embed all:*` already picks up anything under `templates/`, so the new `templates/skills/workflows/` tree is exposed automatically.

Config and scaffolding:

- `internal/config/config.go:19-23` — `Config` struct with `Command`, `Agent`, `Debug`. `Agent` is already persisted by today's init; no change needed.
- `internal/project/init.go:16-76` — `project.Init` creates `.spektacular/` subdirectories and writes `config.yaml`, `.gitignore`, conventions and READMEs. Agent-agnostic; called unchanged.

Tests:

- `cmd/init_test.go` — five tests (`TestInit_Claude`, `TestInit_Bob`, `TestInit_InvalidAgent`, `TestInit_CustomCommand`, `TestInit_Idempotent`). All use `t.TempDir()` + `t.Chdir()` + `testify/require`. These are the tests we extend; no new testing framework is introduced.

CLI plumbing:

- `cmd/root.go:55-62` — `init()` wires `initCmd` via `rootCmd.AddCommand(initCmd)`. Unchanged.
- `cmd/root.go:33-43` — `loadConfig` is reused by the dispatcher; unchanged.

Skill retrieval that we are NOT touching:

- `cmd/skill.go:27-65` — `runSkill` reads from the embedded FS under `skills/skill_<name>.md`. Stays exactly as-is.
- `cmd/skill.go:68-80` — `listSkills()` enumerates the embedded helpers. Stays as-is.

## Per-Phase Technical Notes

### Phase 1.1: Introduce `internal/agent` package

New package at `internal/agent/`. Create the following files:

- `internal/agent/agent.go` (~25 lines) — Defines the `Agent` interface, the `registry` map (initially empty), `Lookup(name string) (Agent, error)`, `Supported() []string`, and an exported `ErrUnknownAgent` sentinel. Registrations happen in per-agent files via `init()` functions.
- `internal/agent/skills.go` (~60 lines) — Defines the `workflowSkill` descriptor slice covering the three entry skills, and exports `installWorkflowSkills(projectPath, targetSkillsDir string, cfg config.Config, out io.Writer) error`. The function reads each skill's template via `templates.FS`, renders `{{command}}` through mustache, creates `<targetSkillsDir>/<skill-name>/` under `projectPath`, and writes `SKILL.md`. Emits one `  Skill:  <path>` line per file.
- `internal/agent/commands.go` (~40 lines) — Exports `installCommandWrappers(projectPath, targetCommandsDir string, filename func(skillName string) string, cfg config.Config, out io.Writer) error`. Reads `templates/commands/wrapper.md` once, renders it for each of the three workflow skills (passing `{{command}}` plus the skill name so the wrapper can name what it invokes), creates the target dir, writes each file using `filename(skill)` for the on-disk name. Emits one `  Command:  <path>` line per file.
- `internal/agent/agent_test.go` (~80 lines) — Unit tests: `Lookup("unknown")` returns `ErrUnknownAgent` with supported names in the message; `Supported()` returns the registered names in stable order; `installWorkflowSkills` against a `t.TempDir()` writes exactly three files under `<target>/spek-{new,plan,implement}/SKILL.md` with `{{command}}` fully rendered; `installCommandWrappers` writes the expected names via the provided `filename` function.

No other package imports `internal/agent` in this phase — registration files for Claude/Bob/Codex land in later phases.

**Complexity**: Medium
**Token estimate**: ~12k
**Agent strategy**: Single agent, sequential — the files are tightly coupled by interface shape. Write `agent.go` first (the contract), then `skills.go` and `commands.go` (the helpers), then the test file last.

### Phase 1.2: Convert workflow templates to SKILL.md

Move workflow content. For each of `spek-new`, `spek-plan`, `spek-implement`:

- Create `templates/skills/workflows/<name>/SKILL.md` with YAML frontmatter (`name`, `description`) followed by the body from the current `templates/commands/<name>.md`.
- The `description` values:
  - `spek-new` — "Create a new Specification for a feature."
  - `spek-plan` — "Create a new Plan from an approved Specification."
  - `spek-implement` — "Execute an approved Plan to implement the feature."
- The body preserves the `{{command}}` mustache placeholders currently present (e.g. `templates/commands/spek-plan.md:28` — `{{command}} plan new ...`).
- Delete `templates/commands/spek-new.md`, `templates/commands/spek-plan.md`, `templates/commands/spek-implement.md`.

Create `templates/commands/wrapper.md` (~20 lines). Content: YAML frontmatter matching Claude Code's existing command format (`description`, `argument-hint`), then a short body directing the agent to invoke the named skill. The template is rendered per-skill by `installCommandWrappers`, so it uses `{{command}}` and `{{skill}}` placeholders. A minimal body example:

```markdown
---
description: {{description}}
argument-hint: <name>
---

Run the `{{skill}}` skill for Spektacular:

    {{command}} skill {{skill}}

Pass `$ARGUMENTS` as input to the skill.
```

The wrapper's description is agent-side (Claude Code / Bob show it in their slash-command menu), so it describes the workflow the skill wraps; pass a per-skill description through the render context in `installCommandWrappers`.

No Go code changes in this phase. `templates/templates.go:7`'s `//go:embed all:*` directive picks up the new layout automatically. Only the contents of `templates/` change.

**Complexity**: Low
**Token estimate**: ~8k (mostly moving text)
**Agent strategy**: Single agent, sequential. Three near-identical content moves plus one new wrapper file.

### Phase 1.3: Register Claude and Bob and wire up init

Create `internal/agent/claude.go` (~40 lines) — implements `claudeAgent`:

- `Name() string` returns `"claude"`.
- `Install(projectPath, cfg, out)` calls `installWorkflowSkills(projectPath, ".claude/skills", cfg, out)` and `installCommandWrappers(projectPath, ".claude/commands/spek", func(s string) string { return strings.TrimPrefix(s, "spek-") + ".md" }, cfg, out)`. The filename function produces `.claude/commands/spek/new.md`, `plan.md`, `implement.md` — matching today's layout at `cmd/init.go:53-55`.
- `init()` registers into `registry["claude"]`.

Create `internal/agent/bob.go` (~40 lines) — implements `bobAgent`:

- `Name() string` returns `"bob"`.
- `Install(projectPath, cfg, out)` calls `installWorkflowSkills(projectPath, ".bob/skills", cfg, out)` and `installCommandWrappers(projectPath, ".bob/commands", func(s string) string { return s + ".md" }, cfg, out)`. Filenames: `.bob/commands/spek-new.md`, `spek-plan.md`, `spek-implement.md` — matching today's layout at `cmd/init.go:53-55`.
- `init()` registers into `registry["bob"]`.

Rewrite `cmd/init.go:14-92` — the entire `initCmd` and `runInit` function:

- `Use: "init <agent>"`, `Short` mentions supported agents sourced from `agent.Supported()`.
- `runInit` is reduced to: `a, err := agent.Lookup(args[0])`; error handling; get cwd; `project.Init(cwd, true)`; `loadConfig`; persist `cfg.Agent = a.Name()` with `cfg.ToYAMLFile`; print header; call `a.Install(cwd, cfg, cmd.OutOrStdout())`. All of `cmd/init.go:48-89` (inline table + render loop + switch) is deleted — that behaviour now lives in `installWorkflowSkills` / `installCommandWrappers`.
- Remove the unused `templates` and `mustache` imports if nothing else in the file needs them after the rewrite.

Update `cmd/init_test.go`:

- `TestInit_Claude` (current lines 11-29): add assertions for three `SKILL.md` files under `.claude/skills/spek-{new,plan,implement}/` with rendered `{{command}}`, and keep the existing command-file assertion at `.claude/commands/spek/new.md` (now thin wrapper content).
- `TestInit_Bob` (current lines 31-49): same shape for `.bob/skills/spek-*/SKILL.md` and `.bob/commands/spek-*.md`.
- `TestInit_InvalidAgent` (current lines 51-59): strengthen to assert the error names every registered agent.
- `TestInit_CustomCommand` (current lines 61-82): keep semantics — after setting `command: go run .` in config, re-init and assert the rendered SKILL.md (not the command wrapper) contains `go run . spec new`.
- `TestInit_Idempotent` (current lines 84-103): unchanged in shape, but extend the preserved sibling file check to also cover `.claude/skills/other/SKILL.md` surviving a re-init.

**Complexity**: Medium
**Token estimate**: ~18k
**Agent strategy**: 2 parallel agents — one writes the two per-agent files (`claude.go`, `bob.go`), the other rewrites `cmd/init.go` and updates `cmd/init_test.go`. Main agent integrates.

### Phase 2.1: Add the Codex agent

Create `internal/agent/codex.go` (~30 lines) — implements `codexAgent`:

- `Name() string` returns `"codex"`.
- `Install(projectPath, cfg, out)` calls `installWorkflowSkills(projectPath, ".agents/skills", cfg, out)` only. No command installation — Codex has no per-repo slash-command mechanism.
- `init()` registers into `registry["codex"]`.

Update `cmd/init_test.go` — add `TestInit_Codex`: runs `init codex` in a temp dir, asserts three SKILL.md files under `.agents/skills/spek-{new,plan,implement}/` with rendered `{{command}}`, and asserts that `.agents/commands`, `.claude/`, and `.bob/` do not exist. Update `TestInit_InvalidAgent` to check the error message names all three agents (`claude`, `bob`, `codex`).

Add a frontmatter validity helper (if not already added in Phase 1.3) — `validateSkillFrontmatter(t *testing.T, path string)` — that parses the YAML frontmatter of an installed SKILL.md and asserts `name` matches the parent directory, `name` passes agentskills.io naming (`^[a-z0-9]+(-[a-z0-9]+)*$`, 1-64 chars), and `description` is non-empty. Call it from each of the three agent tests once. This satisfies "SKILL.md validity" from the testing approach.

**Complexity**: Low
**Token estimate**: ~6k
**Agent strategy**: Single agent, sequential. Small, well-scoped addition.

## Testing Strategy

All new tests are Go unit tests using `testify/require`, living alongside the code they test:

- `internal/agent/agent_test.go` — interface-level tests: `Lookup` not-found, `Supported` ordering, helper-function behaviour with a `t.TempDir()`.
- `internal/agent/claude_test.go`, `internal/agent/bob_test.go`, `internal/agent/codex_test.go` — each runs its agent's `Install` against a `t.TempDir()` and asserts the exact file set. Each calls the shared `validateSkillFrontmatter` helper on every installed SKILL.md.
- `cmd/init_test.go` — CLI-level tests covering dispatcher behaviour: per-agent success path, unknown-agent error, custom `{{command}}` rendering, idempotency with sibling files.

The shared frontmatter helper lives in `internal/agent/agent_test.go` (or a small `testhelp_test.go` if needed). It uses `strings.SplitN(raw, "---", 3)` to extract the YAML block and `gopkg.in/yaml.v3` (already in `go.sum` via `internal/config`) to parse it.

Manual test for the final phase: `go build -o /tmp/spektacular .; cd /tmp/test-codex; /tmp/spektacular init codex`, then open a Codex session in `/tmp/test-codex` and type `$spek-plan`. The workflow should launch.

## Project References

- Specification: `.spektacular/specs/19_codex_integration.md` — the source of truth for scope and acceptance criteria.
- agentskills.io specification: https://agentskills.io/specification — SKILL.md format.
- Codex skills documentation: https://developers.openai.com/codex/skills — install path and invocation.
- Bob skills documentation: https://bob.ibm.com/docs/ide/features/skills — install path and invocation.
- Claude Code skills documentation: https://code.claude.com/docs/en/skills — install path.
- Prior plan establishing the on-demand skill retrieval pattern: `.spektacular/plans/15_implementation/`.
- Prior learnings on Bob: `.spektacular/knowledge/learnings/bob-custom-rules-vs-claude-skills.md` (partially superseded by Bob's native skills feature used here).

## Token Management Strategy

| Tier | Token Budget | Agent Strategy |
|------|-------------|----------------|
| Low | ~10k | Single agent, sequential |
| Medium | ~25k | 2-3 parallel agents |
| High | ~50k+ | Parallel analysis, sequential integration |

This plan's total context is ~44k tokens across four phases — comfortably under the High tier. Phase 1.3 is the only phase that benefits from parallelism (two agents). The other three phases are tightly scoped and run single-agent.

## Migration Notes

Existing Claude and Bob users who ran a prior `spektacular init` will have the old command files at the old paths (`.claude/commands/spek/<name>.md` containing the full workflow body, and `.bob/commands/spek-<name>.md` containing the full workflow body). After this plan lands:

- Re-running `init claude` or `init bob` overwrites those files with the new thin wrappers and additionally creates `.claude/skills/spek-*/SKILL.md` or `.bob/skills/spek-*/SKILL.md`.
- The old content is replaced in-place at the same path; no orphan files are left behind from the rename perspective (the command filenames stay the same; only their contents shrink).
- No migration tooling is supplied. Users just run `init` again, as noted in plan.md § Out of Scope.

## Performance Considerations

None. `init` runs interactively at project setup time and installs ~six small files. There is no runtime overhead in the workflow state machine or the skill retrieval path. The refactor does not touch any hot path.
