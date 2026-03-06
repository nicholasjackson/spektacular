# Plan Command — Implementation Plan

## Context
Spektacular already has a `spec` command for feature specifications. We want a parallel `plan` command implementing the eveld/claude plan workflow in spektacular's markdown-template + step-based FSM format.

Key insight: Rather than installing skills to Claude's `.claude/skills/` or writing Bob rules to `.bob/rules/`, implement a unified **Spektacular Skill Library** that both agents can access via a CLI command. This keeps skills version-controlled, agent-agnostic, and dynamically fetchable.

Changes:
1. New `skill` command: `spektacular skill <name>` returns skill definition from `internal/skills/`
2. New `plan` command with steps/templates that branch per agent type (claude/bob)
3. Templates/agents fetch skills via `{{config.command}} skill <name>` during workflow execution
4. Updated `init` command to store agent type in config

---

## What We're Building

```
spektacular plan new --data '{"name":"my-feature"}'
spektacular plan goto --data '{"step":"discovery"}'
spektacular plan status
spektacular plan steps
```

Output: `JSON { step, plan_path, plan_name, instruction }`

---

## Step Sequence

| # | Name | Agent Does |
|---|------|------------|
| 0 | `new` | Initializes state (`.spektacular/plan-<name>/state.json`), stores overview data; does NOT create plan document yet |
| 1 | `overview` | Displays provided overview, asks for clarification/refinement (not first-time input) |
| 2 | `discovery` | Installs project context docs + runs `spawn-planning-agents` skill + asks clarifying questions |
| 3 | `approach` | Presents design options with trade-offs; gets user agreement + out-of-scope items |
| 4 | `milestones` | Defines 2-4 milestones with user-facing goals and testable outcomes |
| 5 | `phases` | Breaks milestones into phases with file:line changes and split success criteria |
| 6 | `verification` | Gathers metadata, fills in plan document template, pipes completed doc back via stdin |
| 7 | `finished` | Confirms plan saved, mentions `share-docs` for promotion |

FSM: `start → new → overview → discovery → approach → milestones → phases → verification → finished`

---

## Files to Create / Modify

### NEW: `cmd/skill.go`
New command exposing the Spektacular Skill Library:
- `skillCmd` — parent command
- `skillCmd get <name>` — returns skill definition (JSON: `{ name, title, description, instructions }`)
- Example: `spektacular skill discover-project-commands` returns the skill definition
- Used by agents via `{{config.command}} skill <skill_name>` to fetch skill details during workflow
- Skills stored in `internal/skills/skill_*.md` (embedded)

---

## Files to Create

### `cmd/plan.go`
Exact mirror of `cmd/spec.go` (`cmd/spec.go` is the primary reference):
- `planCmd` parent + `planNewCmd`, `planGotoCmd`, `planStatusCmd`, `planStepsCmd`
- `plan new` accepts `--data '{"name":"...", "overview":"..."}' or `--stdin overview` to provide context upfront
- Output type: `Result{ Step, PlanPath, PlanName, Instruction }`
- Calls `plan.Steps()` from `internal/plan/steps.go`
- State dir: `.spektacular/plan-<name>/state.json`
- Plan doc: `.spektacular/plans/<name>.md`
- Workflow data includes `overview` (from init or refined during overview step)

### `internal/plan/steps.go`
Mirror of `internal/spec/steps.go`:
- 8 steps (new + 7)
- `writeStepResult` vars: `step`, `title`, `plan_path`, `next_step`, `config`
- Special `new`: initializes state only (no document created yet); stores overview data
- Special `verification`: injects `plan_template` (blank or partial plan file content) so template displays it
- Special `finished`: if `plan_template` in data (piped via `--stdin`), writes plan file to `.spektacular/plans/<name>.md`

### `templates/plan-scaffold.md`
Plan document template (mirrors eveld/claude `plan-document.md`). The scaffold is displayed in the `verification` step for the agent to fill in. Key sections with HTML comment instructions:
- Overview (pre-filled from earlier steps)
- Current State Analysis
- Desired End State + Key Discoveries
- What We're NOT Doing
- Implementation Approach
- Project References (`thoughts/notes/commands.md`, `thoughts/notes/testing.md`)
- Token Management Strategy (Low/Medium/High complexity tiers)
- Milestones (with nested Phases → Changes Required → Automated+Manual Success Criteria)
- Testing Strategy
- Performance Considerations
- Migration Notes
- References
- Changelog (left empty for implement workflow to fill)

### `templates/plan-steps/overview.md`
Overview is provided upfront via `--data '{"overview":"..."}' or `--stdin overview`. Template displays the overview and asks: does this accurately describe what we're planning? Any clarifications or refinements needed? Agent updates overview if needed, then advances:
```
{{config.command}} plan goto --data '{"step":"{{next_step}}", "overview":"{{overview}}"}'
```

### `templates/plan-steps/discovery.md`
Uses the provided `{{overview}}` to guide research. Full orchestration with agent-agnostic instructions.

**Step 1: Project context (check/create reference docs):**
- Check if `thoughts/notes/commands.md` exists → if not, use the `discover-project-commands` skill. If you need skill details: `{{config.command}} skill discover-project-commands`
- Check if `thoughts/notes/testing.md` exists → if not, use the `discover-test-patterns` skill. If you need skill details: `{{config.command}} skill discover-test-patterns`

**Step 2 — Codebase research (guided by overview):**

Research the codebase in parallel to find:
1. **Files related to {{overview}}** — Organize by category (implementation, tests, config, docs)
2. **Prior research about {{overview}}** — Find existing plans, research, tickets in thoughts/
3. **Similar implementations to {{overview}}** — Find code examples with file:line references
4. **Architecture and integration points** — How do the relevant components fit together?

Use your agent orchestration capability to parallelize this research. If you need guidance on how to structure agent orchestration: `{{config.command}} skill spawn-planning-agents`

**Step 3 — Read + clarify (all agents):**
- Read all findings fully
- Ask only questions code can't answer via `AskUserQuestion`
- Advance: `{{config.command}} plan goto --data '{"step":"{{next_step}}"}'`

### `templates/plan-steps/approach.md`
Tell agent to: present 2-3 design options with pros/cons (with file:line refs from research), get user agreement on chosen direction and explicit out-of-scope items. Advance.

### `templates/plan-steps/milestones.md`
Tell agent to: define 2-4 milestones with user-facing goals + testable outcomes + validation points. NO open questions. Advance.

### `templates/plan-steps/phases.md`
For each milestone, define implementation phases with:
- **Complexity**: Low/Medium/High
- **Token estimate**: ~{N}k tokens
- **Agent strategy**: How to break the phase down (Parallel Analysis / Sequential / Minimal)
- **File changes**: Specific file:line changes from research findings
- **Success criteria**:
  - Automated (commands from `thoughts/notes/commands.md`)
  - Manual (concrete verification steps)

NO open questions allowed.

Use your agent orchestration capability to parallelize implementation work. If you need guidance on orchestration strategy: `{{config.command}} skill spawn-implementation-agents`

Advance: `{{config.command}} plan goto --data '{"step":"{{next_step}}"}'`

### `templates/plan-steps/verification.md`
Tell agent to:
1. Use the `gather-project-metadata` skill (ISO timestamp, git commit, branch, repo). If needed: `{{config.command}} skill gather-project-metadata`
2. Use the `determine-feature-slug` skill (namespace, NNNN number, confirm via AskUserQuestion). If needed: `{{config.command}} skill determine-feature-slug`
3. Display the plan scaffold (empty at first, or pre-filled with Overview + metadata)
4. Fill in ALL sections — no placeholders, no open questions
5. Review: completeness, specificity (file:line), automated+manual split
6. Pipe completed plan back via heredoc (will be written to disk in `finished` step):
```
cat <<'EOF' | {{config.command}} plan goto --data '{"step":"{{next_step}}"}' --stdin plan_template
<complete filled plan here>
EOF
```

### `templates/plan-steps/finished.md`
Tell agent: plan is complete at `{{plan_path}}`. When ready for team sharing, use `share-docs` skill.

---

## Config Changes: `internal/config/config.go`

Add `Agent` field to `Config` struct:
```go
type Config struct {
    Command string      `yaml:"command"`
    Agent   string      `yaml:"agent"`   // "claude" or "bob"
    Debug   DebugConfig `yaml:"debug"`
}
```

`init` sets `cfg.Agent = agent` ("claude" or "bob") and writes config via `cfg.ToYAMLFile(...)`.

`steps.go` injects agent flags into mustache vars:
```go
vars["is_claude"] = cfg.Agent == "claude"
vars["is_bob"]    = cfg.Agent == "bob"
```

---

## Updated `cmd/init.go`

Extend `runInit` to:
1. Set `cfg.Agent` and persist config
2. That's it — no skill installation needed

**For both agents**:

Skills are built into spektacular's `internal/skills/` and accessed uniformly via `spektacular skill <name>` CLI command.

No `.claude/skills/`, no `.bob/rules/` installation by init — agents fetch skills on-demand during workflow.

**Skills are stored in spektacular binary**:
- Embedded in `internal/skills/` as markdown files
- Accessed via `cmd/skill.go` command
- Both Claude and Bob reference via: `{{config.command}} skill <name>` in templates

### `cmd/root.go` (modify)
Add `skillCmd` and `planCmd` to `rootCmd.AddCommand(...)` in `init()`.

---

## Critical Reference Files

| File | Purpose |
|------|---------|
| `cmd/spec.go` | Copy/adapt for `cmd/plan.go` |
| `cmd/spec.go` | Also reference for `cmd/skill.go` pattern (query-based command) |
| `internal/spec/steps.go` | Copy/adapt for `internal/plan/steps.go` |
| `templates/spec-steps/overview.md` | Template advance pattern |
| `templates/spec-steps/verification.md` | Stdin pipe heredoc pattern |
| `templates/spec-steps/finished.md` | Finish pattern |
| `templates/spec-scaffold.md` | Scaffold pattern |
| `templates/commands/spek-new.md` | Command template pattern |
| `internal/config/config.go` | Add Agent field |
| `cmd/init.go` | Update to set cfg.Agent |
| `cmd/root.go` | Add `skillCmd` and `planCmd` |
| `internal/skills/` | New directory for embedded skill definitions |

---

## Verification

1. `go build ./...` — compiles cleanly
2. `spektacular skill discover-project-commands` → returns skill definition (JSON or formatted output)
3. `spektacular skill spawn-planning-agents` → returns skill definition
4. `spektacular init claude` — stores agent type in config
5. `spektacular plan new --data '{"name":"test-plan", "overview":"..."}'` → initializes state (no document yet), outputs `overview` JSON
6. `spektacular plan goto --data '{"step":"discovery"}'` → JSON with agent-branched instructions (references skills via `spektacular skill` calls)
7. `spektacular plan status` → step progress JSON
8. `spektacular plan steps` → 8 step names
9. Pipe test: `cat <<'EOF' | spektacular plan goto --data '{"step":"finished"}' --stdin plan_template` → writes completed plan to `.spektacular/plans/test-plan.md`
