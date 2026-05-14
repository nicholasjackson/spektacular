# Research: 19_codex_integration

## Alternatives considered and rejected

### Option B: Declarative struct table of agent descriptors

A single Go struct capturing every per-agent knob (`SkillDir`, `CommandDir`, `CommandPrefix`, `HasCommands`, etc.), with one `install(AgentDef)` function iterating all three.

**Rejected**: conflates behaviour with configuration. Every new per-agent variation would add a field to the struct, turning it into a schema rather than a clean contract. The spec text at `.spektacular/specs/19_codex_integration.md:31` explicitly says "Define an `Agent` interface in Go that each integration implements", not a struct-table. Also weaker isolation: a field change for Codex risks nudging Claude's behaviour because they share the same install function. Cited evidence: today's `cmd/init.go:48-89` already suffers from exactly this problem — one shared render-and-write loop with agent-specific switch cases at `cmd/init.go:72-78`. The new design removes this class of coupling.

### Option C: YAML-defined agents under `templates/agents/*.yaml`

Agent descriptors live as data files loaded by the Go side at init time. Adding an agent becomes "add a YAML file", no Go code.

**Rejected**: directly contradicts the spec (`spec 19_codex_integration.md:31`): "Define an `Agent` interface in Go that each integration implements". Adds complexity (YAML schema, loading, validation) for three known integrations that barely differ. Testing becomes indirect — you test a data-driven runtime rather than per-agent code. A future agent with quirky install semantics would force the YAML schema to grow exception fields.

### Install runtime helper skills (`spawn-planning-agents`, `verify-implementation`, …) to disk as well

Extend the plan to install all 10 skills (3 workflow + 7 helpers) to `.<agent>/skills/` during `init`.

**Rejected**: the user confirmed "keep on demand" during architecture. The seven helper skills are already retrievable via `spektacular skill <name>` (see `cmd/skill.go:27-65`), and that pattern was explicitly established in `.spektacular/plans/15_implementation/`. Promoting them to installed files without user confirmation would expand scope beyond the spec. Noted as out-of-scope in plan.md.

### Use Codex's `.codex/skills/` path as named in the spec

The spec at `.spektacular/specs/19_codex_integration.md:35` says "installed into the path Codex expects (`.codex/skills/`)".

**Rejected**: the official Codex skills documentation (https://developers.openai.com/codex/skills) specifies `.agents/skills/` at repo scope, not `.codex/skills/`. The spec text was written before this was verified. Treated as a spec error and the plan uses `.agents/skills/`. Confirmed with the user during architecture.

### Install Codex custom prompts as a repo-level slash-command equivalent

Place `~/.codex/prompts/*.md` wrappers so Codex users get `/prompts:spek-plan`.

**Rejected**: the Codex custom-prompts directory is **user-scoped**, not per-project (https://developers.openai.com/codex/custom-prompts). Installing to a user's home directory from a project-init command is outside the established `project.Init` contract (`internal/project/init.go:16-76` targets `projectPath` only) and would surprise users. Skills at `.agents/skills/` already deliver the workflow invocation requirement (`spec 19_codex_integration.md:27`) via `$spek-plan`, so the custom-prompts path adds no user-visible capability.

### Put Bob skills in `.bob/rules/`

Suggested by the older `.spektacular/knowledge/learnings/bob-custom-rules-vs-claude-skills.md` note, which maps Claude skills → Bob rules.

**Rejected**: Bob now has a native skills feature that uses the agentskills.io SKILL.md format and a distinct directory (`.bob/skills/<name>/SKILL.md`, per https://bob.ibm.com/docs/ide/features/skills). Rules remain a separate, orthogonal mechanism for persistent behaviour rather than skill packaging. The learnings note is partially superseded.

## Chosen approach — evidence

Each supporting claim, with its citation:

- **All three agents share the same SKILL.md format** — https://agentskills.io/specification defines `name`/`description` required frontmatter, 1-64 char lowercase-hyphen names, directory-per-skill layout. Both Codex (https://developers.openai.com/codex/skills) and Bob (https://bob.ibm.com/docs/ide/features/skills) explicitly adopt this spec; Claude Code's skills system (https://code.claude.com/docs/en/skills) uses the same file structure.
- **Codex scans `.agents/skills/`** — https://developers.openai.com/codex/skills: "Repository: `$CWD/.agents/skills` … Repository Root: `$REPO_ROOT/.agents/skills`".
- **Bob scans `.bob/skills/`** — https://bob.ibm.com/docs/ide/features/skills: "workspace-level skills directory is located at `<project>/.bob/skills/`".
- **Claude scans `.claude/skills/`** — https://code.claude.com/docs/en/skills: "`.claude/skills/` (relative to a project root) is the project skills directory".
- **Codex invocation uses `$skill-name`** — https://developers.openai.com/codex/skills: "explicit invocation in CLI/IDE, type `$` to mention a skill".
- **The embedded FS already exposes new paths automatically** — `templates/templates.go:7` uses `//go:embed all:*`, so adding `templates/skills/workflows/<name>/SKILL.md` requires no build config change.
- **Mustache `{{command}}` rendering is the project's existing templating pattern** — `cmd/init.go:67` (current) already renders templates with `mustache.Render`. The new helpers preserve this exact pattern for both skill bodies and command wrappers.
- **`config.Config.Agent` already exists** — `internal/config/config.go:21` has `Agent string` in the struct. `cmd/init.go:42` already persists it. No config schema change required.
- **Existing tests use `t.TempDir()` + `t.Chdir()` + `testify/require`** — `cmd/init_test.go:11-103` is the reference shape the new tests follow.
- **The spec accepts breaking changes** — `.spektacular/specs/19_codex_integration.md:18`: "No constraints. Breaking changes are acceptable."
- **Interface-based extensibility matches spec language** — `.spektacular/specs/19_codex_integration.md:31`: "Define an `Agent` interface in Go that each integration implements."

## Files examined

- `cmd/init.go:14-92` — current init dispatcher; the rewrite target. Hardcoded agent list at line 23, inline commands table at lines 48-56, render loop at lines 61-89.
- `cmd/init_test.go:1-103` — five existing tests; the shape all new CLI-level tests copy.
- `cmd/skill.go:27-80` — on-demand skill retrieval path; confirmed untouched by this plan.
- `cmd/root.go:33-62` — `loadConfig` and `rootCmd` wiring; unchanged.
- `cmd/plan.go`, `cmd/spec.go`, `cmd/implement.go` — workflow CLI entry points; unchanged. Scanned to confirm no cross-dependency on command-file layout.
- `cmd/file.go:1-98` — shows the `spec file` subcommand; unrelated to this plan, noted only to confirm no collision.
- `internal/config/config.go:14-61` — `Config` struct (`Command`, `Agent`, `Debug`), `NewDefault`, `FromYAMLFile`, `ToYAMLFile`. `Agent` already persisted by today's init.
- `internal/project/init.go:16-76` — `.spektacular/` scaffolding; confirmed agent-agnostic and called unchanged by the new dispatcher.
- `templates/templates.go:1-8` — the `embed.FS`; `go:embed all:*` picks up new paths.
- `templates/commands/spek-new.md`, `spek-plan.md`, `spek-implement.md` — full workflow bodies; content moves verbatim into SKILL.md bodies in Phase 1.2.
- `templates/skills/skill_*.md` — seven helper skills; confirmed not installed to disk by this plan.
- `templates/steps/implement/02-analyze.md:21` — example reference to `{{config.command}} skill spawn-implementation-agents` showing the established on-demand skill retrieval pattern.
- `main.go:1` — confirms `go run .` is the single CLI entry point.
- `go.mod`, `go.sum` — confirm `github.com/cbroglie/mustache` and `gopkg.in/yaml.v3` already vendored; no new dependency needed.

## External references

- https://agentskills.io/specification — SKILL.md format, frontmatter fields, `name` naming rules, directory layout. Governs all three agent install formats uniformly.
- https://developers.openai.com/codex/skills — Codex install path (`.agents/skills/`), `$skill-name` invocation. Overrides the spec's `.codex/skills/` text.
- https://developers.openai.com/codex/custom-prompts — Codex custom prompts are user-scoped only; explains why Codex has no per-repo command support.
- https://bob.ibm.com/docs/ide/features/skills — Bob install path (`.bob/skills/`) and SKILL.md conformance.
- https://code.claude.com/docs/en/skills — Claude Code project skills at `.claude/skills/`.

## Prior plans / specs consulted

- `.spektacular/specs/19_codex_integration.md` — source of truth. Note that its `.codex/skills/` path at line 35 is incorrect; the plan uses `.agents/skills/` instead, confirmed with the user.
- `.spektacular/plans/15_implementation/` — established the `{{command}} skill <name>` on-demand skill retrieval pattern that this plan preserves for the seven helper skills.
- `.spektacular/plans/16_plan_format/` — earlier plan format refactor; scanned for relevant patterns, nothing reused.
- `.spektacular/knowledge/learnings/bob-custom-rules-vs-claude-skills.md` — outdated Bob rules-vs-skills mapping. Partially superseded by Bob's native skills feature at https://bob.ibm.com/docs/ide/features/skills. Noted here so future agents don't re-adopt the rules-based path.
- `.spektacular/knowledge/conventions.md` — general coding standards; no specific constraints on this work.
- `.spektacular/knowledge/architecture/initial-idea.md` — original Spektacular vision document; scanned for Codex-related notes, none beyond the general multi-agent intent.

## Open assumptions

- **`.agents/skills/` is the correct Codex install path today.** Verified against https://developers.openai.com/codex/skills. If Codex changes the scanned path, the `codexAgent` implementation is the single point of change.
- **Codex's `$skill-name` invocation remains the canonical way to trigger a skill.** If Codex drops this syntax in a future version, the spec's acceptance criterion 7 becomes untestable — the implement workflow must STOP and ask the user.
- **The existing mustache `{{command}}` placeholder is the only variable substitution needed in SKILL.md bodies.** If any workflow body grows a new placeholder (e.g. `{{plan_path}}`), the rendering context in `installWorkflowSkills` must be extended.
- **Bob's `.bob/skills/` layout matches agentskills.io.** If Bob diverges (e.g. requires an extra manifest), the `bobAgent` implementation is the change point.
- **Users re-run `init` manually after upgrading.** The plan does not auto-detect stale artefacts. If silent staleness becomes a support issue, a future plan can add a `status`/`doctor` command.

If any assumption fails during implementation, STOP and ask the user before proceeding.

## Rehydration cues

A future agent picking up this plan cold should:

1. Read `.spektacular/specs/19_codex_integration.md` for scope.
2. Read `.spektacular/plans/19_codex_integration/plan.md` for direction, then `context.md` for per-phase file:line detail.
3. Run `go run . skill spawn-implementation-agents` and `go run . skill verify-implementation` for orchestration guidance during implementation.
4. Re-fetch the four external documentation pages listed above — install paths and invocation syntax for all three agents can change.
5. Re-read `cmd/init.go` and `cmd/init_test.go` before editing; the refactor target may have drifted.
6. Confirm `templates/templates.go` still uses `//go:embed all:*` — if narrowed, the new skill paths must be added explicitly.
7. Confirm `internal/config/config.go:21` still has the `Agent` field.
