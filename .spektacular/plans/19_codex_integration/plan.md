# Plan: 19_codex_integration

<!-- Metadata -->
<!-- Created: 2026-04-17T16:18:45Z -->
<!-- Commit: fc21712ff99f1d6b4af4c826e6003f4471bdffce -->
<!-- Branch: f-codex -->
<!-- Repository: git@github.com:jumppad-labs/spektacular.git -->

## Overview

Spektacular separates its agent-facing interface into two artefact types — skills (the canonical workflow entry points, installable by any agent) and commands (thin slash-command wrappers, only for agents that support them). A new `Agent` interface drives per-agent install behaviour, and a Codex integration ships alongside the existing Claude and Bob ones. Development teams get the same spec-driven workflow regardless of which AI coding agent they use.

## Architecture & Design Decisions

The shape of the solution is a new `internal/agent` package that defines an `Agent` interface, a registry map keyed by agent name, and one small implementation per supported agent. The existing `cmd/init` dispatcher becomes a thin shell that validates the agent name through the registry, runs the standard `.spektacular/` scaffolding, and delegates artefact installation to the selected agent. This cleanly satisfies the spec's "isolated extensibility" criterion: adding a future agent is a new file that implements the interface plus a single registry entry, with no changes to workflow logic.

Skills become the canonical workflow entry points. The three top-level workflow artefacts (`spek-new`, `spek-plan`, `spek-implement`) move from `templates/commands/` into directory-per-skill `SKILL.md` files under `templates/skills/workflows/`, following the [agentskills.io specification](https://agentskills.io/specification) adopted by all three supported agents. Because Codex, Bob, and Claude Code all consume byte-identical SKILL.md content, the three files live once in the embedded FS and each agent implementation copies them to its own install directory (`.claude/skills/`, `.bob/skills/`, `.agents/skills/`). The remaining seven helper skills (`spawn-planning-agents`, `verify-implementation`, etc.) continue to be retrieved on-demand via `{{command}} skill <name>` — they are not installed to disk.

Commands become thin wrappers. For Claude and Bob — the two agents that support slash commands — a single shared wrapper template renders a one-screen file whose sole job is to tell the agent to invoke the named skill. Codex does not support per-repo slash commands (its custom-prompts mechanism is user-scoped only), so Codex installs skills and nothing else. This split is a breaking change for existing Claude/Bob users, who re-run `init` to pick up the new layout; the spec explicitly accepts breaking changes.

Alternatives considered included a declarative struct table of agent descriptors and a YAML-configured registry. Both were rejected: the struct-table approach conflates config with behaviour and makes changes bleed across agents, and the YAML approach contradicts the spec's explicit call for a Go interface. See [research.md § Alternatives considered and rejected](./research.md#alternatives-considered-and-rejected) for the evidence.

## Component Breakdown

- **`internal/agent` (new package)** — Defines the `Agent` interface, the registry map, and per-agent implementations (`ClaudeAgent`, `BobAgent`, `CodexAgent`). Exposes `Lookup(name)` for the CLI dispatcher. Each implementation knows its skill install directory, command install directory (if any), and command filename convention.

- **Skill-install helper (inside `internal/agent`)** — Shared routine that reads the embedded `SKILL.md` templates, renders the `{{command}}` placeholder against the loaded config, and writes the three skills into the agent-chosen target directory. Reused byte-for-byte across all three agent implementations.

- **Command-wrapper helper (inside `internal/agent`)** — Shared routine that renders the thin command-wrapper template ("invoke the `<skill>` skill") into an agent's command directory. Used by Claude and Bob; Codex skips it entirely.

- **`cmd/init` (changed)** — Loses its hardcoded agent check and inline commands table. Becomes a roughly-12-line orchestrator that validates through `agent.Lookup`, runs `project.Init`, persists the agent name to config, and calls `Install`. The unknown-agent error is driven by the registry.

- **`templates/skills/workflows/` (new)** — Three directory-per-skill packages: `spek-new/SKILL.md`, `spek-plan/SKILL.md`, `spek-implement/SKILL.md`. Frontmatter follows agentskills.io; body is the workflow instruction text currently living in the old command templates with the `{{command}}` placeholder preserved.

- **`templates/commands/wrapper.md` (new)** — Single shared thin-command template used to generate the Claude and Bob slash-command files. Replaces three full-content command templates.

- **Old `templates/commands/spek-*.md` (deleted)** — Their content has moved to the three SKILL.md files; the wrapper replaces their role.

- **`cmd/init_test.go` (changed)** — Extended to cover all three agents: both skills and commands for Claude and Bob, skills-only for Codex, unknown-agent error naming all three supported agents, preserved idempotency.

- **`internal/agent/*_test.go` (new)** — Per-agent unit tests that call `Install` against a temp dir and assert the exact set of files written. Plus a registry `Lookup` test and a shared frontmatter-validity helper.

## Data Structures & Interfaces

**`Agent` interface** (package `internal/agent`):

```go
type Agent interface {
    Name() string
    Install(projectPath string, cfg config.Config, out io.Writer) error
}
```

`Name()` returns the canonical identifier used in `spektacular init <name>` and stored in config. `Install` is idempotent: it writes all skills (and commands, if supported) into `projectPath`, emitting one human-readable line per artefact to `out`.

**Registry** (package `internal/agent`):

```go
var registry = map[string]Agent{
    "claude": &claudeAgent{},
    "bob":    &bobAgent{},
    "codex":  &codexAgent{},
}

func Lookup(name string) (Agent, error)
func Supported() []string
```

`Lookup` returns an `ErrUnknownAgent` whose message lists `Supported()` when the name is missing. The error drives the CLI's unknown-agent response.

**Internal workflow-skill descriptor** (inside `internal/agent`):

```go
type workflowSkill struct {
    Name         string // "spek-new", "spek-plan", "spek-implement"
    TemplatePath string // path inside templates.FS
}
var workflowSkills = []workflowSkill{ /* three entries */ }
```

Single source of truth for which skills every agent installs.

**SKILL.md frontmatter** (on-disk wire format, per agentskills.io):

```yaml
---
name: spek-plan
description: Create a new Plan from a Specification.
---
```

Exactly the two required fields. No Go struct represents this at runtime — it is rendered from the embedded template.

**Config** — `config.Config.Agent` already exists; no shape change required.

## Implementation Detail

The plan introduces a new module boundary at `internal/agent`. This package owns every agent-specific install decision: directory paths, filename conventions, and whether the agent supports commands. The CLI dispatcher shrinks to a thin orchestrator, and all per-agent knowledge moves behind the interface. A reviewer reading `cmd/init.go` after the change sees no hardcoded agent names.

Artefacts become data rather than control flow. The three workflow skills are described once as an embedded-template slice shared across every agent. Per-agent code contains only its install-path strategy and its answer to "do I install commands?" If a future agent adopts the same SKILL.md format at a different path, adding it is a ~30-line file plus a registry entry.

The skills-first shift changes which file a developer thinks of as "the workflow". Today, the three `templates/commands/spek-*.md` files carry the workflow instructions and are also the user-facing entry points. After this plan lands, the workflow instructions live in `templates/skills/workflows/*/SKILL.md`, and the command files shrink to a one-screen wrapper that just points at the named skill. This matches the direction every supported agent is moving — commands are convenience, skills are the contract.

Embedded-template layout follows the artefact, not the agent. Because all three agents consume byte-identical skill files, there is no per-agent skill copy under `templates/`. A developer adding a new workflow skill edits exactly one `SKILL.md` file; all three integrations pick it up automatically via the shared `workflowSkills` slice. The only agent-specific template is the shared command wrapper, and even it is generic — it just names the skill it wraps.

Existing patterns are preserved. `cbroglie/mustache` remains the single templating engine for `{{command}}` rendering. `project.Init` still does the `.spektacular/` skeleton work. Config persistence stays in `cmd/init`. The workflow state machine, step templates, and CLI framing are untouched.

## Dependencies

- **`internal/config`** — provides `Config` with the existing `Agent` field; the agent name continues to persist here at init time. No change required.
- **`internal/project`** — provides `project.Init` for the `.spektacular/` skeleton; called unchanged.
- **`templates` (embed.FS)** — `go:embed all:*` already exposes everything under `templates/`; the new `templates/skills/workflows/` tree and `templates/commands/wrapper.md` are picked up automatically.
- **`github.com/cbroglie/mustache`** — already used for `{{command}}` rendering in `cmd/init.go`; reused unchanged for both skill bodies and the command wrapper.
- **`github.com/spf13/cobra`** — already used for the CLI; unchanged.
- **No new external libraries.**
- **[agentskills.io/specification](https://agentskills.io/specification)** — the shared SKILL.md format adopted by all three supported agents. Governs frontmatter fields and `name` validation. Reference only, not a runtime dependency.
- **[developers.openai.com/codex/skills](https://developers.openai.com/codex/skills)** — confirms Codex scans `.agents/skills/` at repo scope.
- **[bob.ibm.com/docs/ide/features/skills](https://bob.ibm.com/docs/ide/features/skills)** — confirms Bob scans `.bob/skills/` at workspace scope.
- **[code.claude.com/docs/en/skills](https://code.claude.com/docs/en/skills)** — confirms Claude Code scans `.claude/skills/` at project scope.
- **Prior plans** — `.spektacular/plans/15_implementation/` established the on-demand `{{command}} skill <name>` retrieval pattern we keep in place for the runtime helper skills.
- **Planning / ordering** — no plan must land first; this work is self-contained.

## Testing Approach

The strategy is almost entirely **unit tests at the CLI integration boundary**, following the existing `cmd/init_test.go` shape: temp directory, run the Cobra command, assert files on disk with `testify/require`. No new test frameworks are introduced; this is the established pattern across the repo.

Coverage concentrates on three areas in priority order:

1. **`cmd/init` dispatcher** — highest-value assertions. One test per supported agent confirms the expected artefact tree. Dedicated tests cover the unknown-agent error path, custom-command rendering (carried over), and idempotency against sibling files (carried over).
2. **`internal/agent` per-agent implementations** — direct unit tests that call `Install` against a `t.TempDir()` and assert the exact file set. Isolates per-agent behaviour from CLI plumbing. Registry `Lookup` gets a tiny test for the not-found error.
3. **SKILL.md validity** — a small shared helper parses every installed SKILL.md's frontmatter and asserts `name` and `description` are present, the `name` matches its parent directory, and the `name` passes agentskills.io naming rules.

**Load-bearing assertions in plain language:** after `init claude`, three SKILL.md files under `.claude/skills/spek-{new,plan,implement}/` **and** three command files under `.claude/commands/spek/`. After `init bob`, three SKILL.md files under `.bob/skills/spek-{new,plan,implement}/` **and** three command files under `.bob/commands/`. After `init codex`, three SKILL.md files under `.agents/skills/spek-{new,plan,implement}/` **and zero** command files anywhere. For every agent, no `{{command}}` placeholder survives rendering. `init unknown` returns an error naming all three supported agents. Re-running `init` leaves unrelated sibling files intact.

Tests follow existing conventions — per-package test files, `t.TempDir()` + `t.Chdir()` setup, `require` assertions. No new helpers unless shared frontmatter validation is genuinely reused across all three agent tests.

**Deliberate gaps.** No end-to-end test that spawns Claude/Bob/Codex and invokes the installed skills — those tools aren't available in CI and the spec's success metric is manual developer verification. No tests for the on-demand `{{command}} skill <name>` retrieval path (unchanged by this plan). No tests for workflow step templates, the stepkit runtime, or the spec/plan/implement commands themselves (also unchanged).

## Milestones & Phases

### Milestone 1: Skills become the canonical workflow artefact

**What changes**: After re-running `spektacular init claude` or `init bob`, the three workflow entry points (`spek-new`, `spek-plan`, `spek-implement`) live on disk as proper `SKILL.md` files under the agent's skills directory, and the matching slash-command files are trimmed down to thin wrappers that just invoke the skill. A developer reading either artefact sees the same workflow description in the skill; the command file becomes a one-line shortcut. Internally, the init dispatcher no longer special-cases agent names — it looks them up through a new agent interface.

#### - [x] Phase 1.1: Introduce the `internal/agent` package

The new package defines the `Agent` interface, a registry map, a `Lookup` function that returns a clear error for unknown names, a shared helper that installs the three workflow skills into a caller-specified directory, and a shared helper that writes a thin command-wrapper file. It is self-contained — no other package imports it yet. This phase isolates the refactor from the wiring change so the first commit is reviewable on its own.

*Technical detail:* [context.md#phase-11](./context.md#phase-11-introduce-internal-agent-package)

**Acceptance criteria**:

- [x] The package compiles and its unit tests pass in isolation.
- [x] `Lookup("unknown")` returns an error that names the supported agents; `Lookup` returns a non-nil agent for every registered name (initially none; registrations land in later phases).
- [x] The skill-install helper, when pointed at a target directory, writes exactly three `SKILL.md` files with fully rendered `{{command}}` and nothing else.
- [x] The command-wrapper helper writes one file per workflow entry when called.

#### - [x] Phase 1.2: Convert workflow templates to SKILL.md format

The three workflow bodies move from `templates/commands/spek-*.md` into directory-per-skill `SKILL.md` files under `templates/skills/workflows/spek-{new,plan,implement}/`, each with agentskills.io-compliant frontmatter. The old command files are deleted and replaced by a single shared wrapper template that any agent supporting commands can render. The embedded FS picks up the new layout automatically via the existing `go:embed all:*` directive. No Go code changes in this phase — it's purely a content reshape.

*Technical detail:* [context.md#phase-12](./context.md#phase-12-convert-workflow-templates-to-skillmd)

**Acceptance criteria**:

- [x] Three `SKILL.md` files exist, each with a valid `name` matching its parent directory and a non-empty `description`.
- [x] The shared command wrapper template exists and, when rendered, produces a one-screen file that directs the reader to the named skill.
- [x] The original three `templates/commands/spek-*.md` files no longer exist.
- [x] `go build ./...` succeeds — no build-time references to the old paths remain.

#### - [x] Phase 1.3: Register Claude and Bob agents and wire up `cmd/init`

Claude and Bob agent implementations arrive: each knows its skill install directory (`.claude/skills/`, `.bob/skills/`), its command install directory and filename convention, and delegates to the shared helpers. Both register themselves in the package-level registry. `cmd/init.go` is rewritten around `agent.Lookup` — the inline commands table and hardcoded agent names are gone. `cmd/init_test.go` is expanded to assert both skill and command artefacts for Claude and Bob, stronger unknown-agent error text, and preserved idempotency.

*Technical detail:* [context.md#phase-13](./context.md#phase-13-register-claude-and-bob-and-wire-up-init)

**Acceptance criteria**:

- [x] `spektacular init claude` produces three SKILL.md files under `.claude/skills/spek-{new,plan,implement}/` and three command files under `.claude/commands/spek/`.
- [x] `spektacular init bob` produces three SKILL.md files under `.bob/skills/spek-{new,plan,implement}/` and three command files under `.bob/commands/`.
- [x] `spektacular init unknown` returns an error that lists the registered agents.
- [x] No installed file contains a remaining `{{command}}` placeholder.
- [x] Re-running `init` leaves unrelated sibling files in `.claude/`/`.bob/` untouched.
- [x] `go test ./...` passes.

### Milestone 2: Codex works out of the box

**What changes**: `spektacular init codex` becomes a supported command. It installs the same three workflow skills into `.agents/skills/spek-{new,plan,implement}/SKILL.md`, the directory Codex actually scans, and writes no command files. A Codex user, after running this, can invoke `$spek-plan`, `$spek-new`, or `$spek-implement` and get the same workflow behaviour Claude and Bob users already have.

#### - [x] Phase 2.1: Add the Codex agent

A `codexAgent` implementation is added to the `internal/agent` package, registered alongside Claude and Bob. It installs the three workflow skills into `.agents/skills/spek-{new,plan,implement}/SKILL.md` and writes no command files. The init dispatcher and unknown-agent error now recognise `codex` without any further code change. Tests cover the Codex install tree, the absence of any command files, and the updated unknown-agent error naming all three agents.

*Technical detail:* [context.md#phase-21](./context.md#phase-21-add-the-codex-agent)

**Acceptance criteria**:

- [x] `spektacular init codex` produces three SKILL.md files under `.agents/skills/spek-{new,plan,implement}/` and zero files under `.claude/`, `.bob/`, or any `commands/` directory.
- [x] A SKILL.md frontmatter check confirms each installed skill passes agentskills.io naming and required-field rules.
- [x] `spektacular init unknown` error message names `claude`, `bob`, and `codex`.
- [x] `go test ./...` passes.
- [ ] Manual smoke test against a live Codex session: typing `$spek-plan` launches the plan workflow without error.

## Open Questions

None. Install paths, SKILL.md format, artefact scope, and command-wrapper design were all resolved during discovery and architecture; see [research.md § Chosen approach — evidence](./research.md#chosen-approach--evidence). An empty Open Questions section is the expected healthy outcome here.

## Out of Scope

- **Installing the seven runtime helper skills to disk** (`spawn-planning-agents`, `spawn-implementation-agents`, `follow-test-patterns`, `gather-project-metadata`, `update-changelog`, `verify-implementation`, `determine-feature-slug`). They remain on-demand via `spektacular skill <name>`. Promoting them to installed SKILL.md files is a potential follow-up.
- **A migration path for existing Claude/Bob projects** that have old-style command files from a prior `init`. Users re-run `init` to pick up the new layout; we do not detect or clean up artefacts from older runs. Breaking-change acceptance was explicit in the spec.
- **Global / user-scope installs** (`~/.claude/skills/`, `~/.bob/skills/`, `~/.codex/prompts/`). `init` continues to target the project directory only.
- **Codex slash commands via `~/.codex/prompts/`**. Codex's slash-command mechanism is user-scoped and not supported per-repo, so it is deliberately not installed. Revisiting this would require its own spec.
- **Automated tests that actually run Claude Code, Bob, or Codex against the installed artefacts.** Success metrics rely on manual developer testing per the spec; live-agent end-to-end coverage is a separate concern.
- **Any change to the workflow step templates, the stepkit runtime, or the `spec` / `plan` / `implement` CLI commands themselves.** This plan only touches the `init` pipeline and the artefact templates it emits.

## Changelog

### 2026-04-19 — Phase 1.1: Introduce the `internal/agent` package

**What was done**: Added a new `internal/agent` package defining the `Agent` interface, a package-level registry map with private `register`, a `Lookup` helper that returns an error wrapping `ErrUnknownAgent` (with the list of supported agents in the message), a `Supported` helper returning a sorted name slice, a `workflowSkills` descriptor slice, and two unexported shared helpers (`installWorkflowSkills`, `installCommandWrappers`) that render mustache `{{command}}` (plus `{{skill}}` / `{{description}}` for the wrapper) from a swappable `sourceFS fs.FS = templates.FS`. Package is isolated — no other package imports it yet. Unit tests cover the Lookup/Supported semantics and the two install helpers using an `fstest.MapFS` override.

**Deviations**: Added a package-level `sourceFS fs.FS = templates.FS` indirection that the plan did not explicitly describe. This was necessary so Phase 1.1's tests could pass in isolation — the real `templates/skills/workflows/` and `templates/commands/wrapper.md` files don't land until Phase 1.2. The runtime behaviour is identical; only tests reassign `sourceFS`.

**Files changed**:
- `internal/agent/agent.go` (new)
- `internal/agent/skills.go` (new)
- `internal/agent/commands.go` (new)
- `internal/agent/agent_test.go` (new)

**Discoveries**:
- The three workflow entries share rendering context: skills render only `{{command}}`, but wrappers additionally render `{{skill}}` and `{{description}}`. Phase 1.2's `templates/commands/wrapper.md` must use those placeholders, and Phase 1.3's per-agent descriptions already live in `internal/agent/commands.go`'s `workflowDescriptions` map.
- `installCommandWrappers` takes a `filename func(skillName string) string` so each agent can pick its own basename convention (Claude strips the `spek-` prefix; Bob keeps it). This contract is what Phase 1.3's `claude.go` / `bob.go` must satisfy.
- The unused-function info diagnostics on `register`, `installWorkflowSkills`, and `installCommandWrappers` are expected and will clear in Phase 1.3 when Claude/Bob wire themselves into the registry.

### 2026-04-19 — Phase 1.2: Convert workflow templates to SKILL.md format

**What was done**: Moved the three workflow bodies out of `templates/commands/spek-{new,plan,implement}.md` into directory-per-skill `templates/skills/workflows/spek-{new,plan,implement}/SKILL.md` files with agentskills.io-compliant frontmatter (`name`, `description`). Deleted the old command files. Added `templates/commands/wrapper.md` as the single shared thin-command template with `{{command}}`, `{{skill}}`, and `{{description}}` placeholders. No Go code changes.

**Deviations**: None.

**Files changed**:
- `templates/skills/workflows/spek-new/SKILL.md` (new)
- `templates/skills/workflows/spek-plan/SKILL.md` (new)
- `templates/skills/workflows/spek-implement/SKILL.md` (new)
- `templates/commands/wrapper.md` (new)
- `templates/commands/spek-new.md` (deleted)
- `templates/commands/spek-plan.md` (deleted)
- `templates/commands/spek-implement.md` (deleted)

**Discoveries**:
- **Intermediate-state breakage**: between this phase and Phase 1.3, `go test ./cmd/...` fails at runtime because the current `cmd/init.go:62` still reads `commands/spek-*.md` via `templates.FS.ReadFile`. `go build ./...` stays green. Phase 1.3's rewrite of `cmd/init.go` removes that read and restores test-green.
- The SKILL.md bodies preserve `$ARGUMENTS` verbatim. For Claude/Bob, the wrapper template renders `Pass ${ARGUMENTS} as input to the skill.` so the agent forwards the user's positional arg. For Codex, `$spek-plan <name>` passes the tail to the skill body directly and the `$ARGUMENTS` token is read by the agent as a placeholder for the invocation argument.
- `templates.FS` (`//go:embed all:*`) requires no changes to the embed directive — the new `templates/skills/workflows/...` tree is picked up automatically.

### 2026-04-19 — Phase 1.3: Register Claude and Bob agents and wire up `cmd/init`

**What was done**: Added `internal/agent/claude.go` and `internal/agent/bob.go` — each a ~25-line `Agent` implementation that delegates to the shared `installWorkflowSkills` and `installCommandWrappers` helpers with its own target directory and filename convention, and registers itself in `init()`. Rewrote `cmd/init.go` as a 50-line dispatcher that validates via `agent.Lookup`, runs `project.Init`, persists the agent name to config, and calls `a.Install`. The inline commands table and `claude`/`bob` string switches are gone. Replaced all five tests in `cmd/init_test.go` for the new artefact layout and added `internal/agent/claude_test.go` and `internal/agent/bob_test.go` that exercise `Install` directly against the real `templates.FS`.

**Deviations**: None.

**Files changed**:
- `internal/agent/claude.go` (new)
- `internal/agent/bob.go` (new)
- `internal/agent/claude_test.go` (new)
- `internal/agent/bob_test.go` (new)
- `cmd/init.go` (rewritten)
- `cmd/init_test.go` (rewritten)

**Discoveries**:
- `initCmd`'s `Short` string calls `strings.Join(agent.Supported(), ", ")` at package-variable initialization time. Go guarantees imported packages' `init()` functions (where `claudeAgent`/`bobAgent` register themselves) run before the importing package's variable initializers, so the sorted name list is populated correctly. This same pattern carries straight into Phase 2.1 when `codex` joins the list.
- `claudeCommandFilename` strips the `spek-` prefix (so `.claude/commands/spek/new.md`) while `bobCommandFilename` keeps it (so `.bob/commands/spek-new.md`). This is the load-bearing difference between the two agents' command layouts; the skills trees are byte-identical.
- `TestInit_InvalidAgent` now uses `errors.Is(err, agent.ErrUnknownAgent)` — the wrap-with-sentinel pattern from Phase 1.1 works as designed across the package boundary.
- Running `init` inside the spektacular repo's own root during smoke tests will install SKILL.md files into the *developer's* `.claude/skills/` — confirmed the smoke was run inside `mktemp -d` to avoid polluting the working tree.

### 2026-04-19 — Phase 1.3 follow-up: SKILL.md wording

**What was done**: Changed the first heading and first-paragraph noun in all three workflow SKILL.md files from "command" to "skill" (e.g. "This skill drives a multi-step interactive workflow…"). The `{{command}}` placeholder for the CLI binary name is unchanged — that's the correct noun for the tool that runs the state machine. Caught during review by the user before Phase 2.1.

**Deviations**: None — this is a copy fix, not a behaviour change.

**Files changed**:
- `templates/skills/workflows/spek-new/SKILL.md`
- `templates/skills/workflows/spek-plan/SKILL.md`
- `templates/skills/workflows/spek-implement/SKILL.md`

**Discoveries**: None.

### 2026-04-19 — Phase 2.1: Add the Codex agent

**What was done**: Added `internal/agent/codex.go` (~18 lines) — a `codexAgent` implementing `Agent` whose `Install` calls only `installWorkflowSkills(projectPath, ".agents/skills", ...)`; no command wrappers because Codex has no per-repo slash-command mechanism. Registered via `init()` alongside Claude and Bob. The init dispatcher and unknown-agent error now recognise `codex` automatically — no changes to `cmd/init.go`. Added `TestInit_Codex` in `cmd/init_test.go`, updated `TestInit_InvalidAgent` to require the error names all three agents, added `internal/agent/codex_test.go` that exercises `Install` directly and asserts no `.claude/`, `.bob/`, or `.agents/commands/` sibling directories are created. Added `validateSkillFrontmatter(t, path)` in `internal/agent/agent_test.go` — it parses the YAML frontmatter, asserts `name` matches the parent directory and the agentskills.io regex (`^[a-z0-9]+(-[a-z0-9]+)*$`, 1-64 chars), and that `description` is non-empty. Retrofitted the existing Claude/Bob/Codex per-agent install tests to call this helper on every installed SKILL.md.

**Deviations**: The "Manual smoke test against a live Codex session" acceptance criterion is left unchecked — it requires an interactive Codex IDE, which is out of scope for automated CI. The user should run `go run . init codex` in a temp directory and type `$spek-plan` in a live Codex session to close this criterion.

**Files changed**:
- `internal/agent/codex.go` (new)
- `internal/agent/codex_test.go` (new)
- `internal/agent/agent_test.go` (added `validateSkillFrontmatter`)
- `internal/agent/claude_test.go` (retrofitted to call `validateSkillFrontmatter`)
- `internal/agent/bob_test.go` (retrofitted to call `validateSkillFrontmatter`)
- `cmd/init_test.go` (added `TestInit_Codex`; extended `TestInit_InvalidAgent`)

**Discoveries**:
- The registry pattern established in Phase 1.1 paid off: adding Codex required zero changes to the dispatcher. `agent.Supported()` automatically picks up the new registration because the `init()` function runs at package load. The `Short` string on `initCmd` still just reads `strings.Join(agent.Supported(), ", ")` and now says `(bob, claude, codex)`.
- `validateSkillFrontmatter` uses `strings.SplitN(raw, "---", 3)` to peel off the frontmatter block — simpler than pulling in a proper markdown parser and safe for our controlled templates where `---` only appears as the frontmatter delimiters.
- Final test count: 123 tests across 11 packages. Phase 1.1 added 5, Phase 1.3 added 7 + rewrote 5, Phase 2.1 added 3 + updated 1 = 16 new/changed tests for this feature.
