# Plan: Agent-Optimised CLI Improvements

Reference: [CLI Design for AI Agents](../knowledge/architecture/cli-design-for-ai-agents.md)

## Gap Analysis

Comparing the current CLI against the guiding principles from the article:

| Principle | Status | Notes |
|---|---|---|
| Raw JSON Over Bespoke Flags | PARTIAL | `--data` is good; four boolean mode flags (`--new`, `--next`, `--step`, `--status`) are fragile for agents; no consistent `--data`-only input convention |
| Runtime Schema Introspection | MISSING | No `spektacular spec steps` or `spektacular spec schema` subcommand |
| Context Window Efficiency | PARTIAL | Instructions are full markdown; no field masks — addressed by Phase 5 `--fields` flag |
| Input Validation | WEAK | `name` used directly as a filename with no sanitization (path traversal risk) |
| Agent Skills Documentation | PARTIAL | Slash command files via `init` are good; no standalone/generic agent docs |
| Multi-Surface Architecture | MISSING | CLI-only; no MCP tool surface |
| Safety Rails | MISSING | No `--dry-run`; crafted spec names could inject content into rendered instructions |

## Phases

### Phase 1 — Input Validation (HIGH priority, security)

**Problem:** The `name` value in `--data '{"name":"..."}'` is used directly as a filename (`specs/<name>.md`) with no sanitization. A hallucinated or adversarial value like `../../etc/passwd` or `name; rm -rf .` would be accepted.

**Changes:**
- Validate `name` against `[a-z0-9_-]+` (reject anything else with a clear error)
- Enforce a max length (e.g. 64 chars)
- Return an error (not silent default) when `.spektacular/config.yaml` is present but invalid

**Files:** `cmd/spec.go`, `internal/config/config.go`

---

### Phase 2 — Schema Introspection (MEDIUM priority)

**Problem:** An agent has no way to discover valid step names, the `--data` JSON schema, or the output field schema at runtime without prior documentation.

**Changes:**
- Add `--schema` flag to every subcommand — returns that subcommand's `--data` input schema and output schema as JSON, then exits
- Add `spektacular spec steps` subcommand — returns JSON array of valid step names (useful for `spec goto`)

**Example: `spektacular spec goto --schema`**
```json
{
  "input": {
    "type": "object",
    "properties": {
      "step": { "type": "string", "enum": ["overview", "requirements", "acceptance_criteria", "constraints", "technical_approach", "success_metrics", "non_goals", "verification"] }
    },
    "required": ["step"]
  },
  "output": {
    "type": "object",
    "properties": {
      "step": { "type": "string" },
      "total_steps": { "type": "integer" },
      "completed_steps": { "type": "integer" },
      "spec_path": { "type": "string" },
      "spec_name": { "type": "string" },
      "instruction": { "type": "string" }
    }
  }
}
```

**Files:** `cmd/spec.go` (add `--schema` flag to each subcommand), `internal/spec/steps.go`

---

### Phase 3 — Subcommand Restructure (MEDIUM priority)

**Problem:** Two problems in one:

1. The `--action` flag approach fails for `goto` because the step name has to go somewhere — either as a flag (requires knowing a per-action schema) or in `--data` (opaque to the agent). Bool flags have the same ambiguity problem.

2. The orchestration logic (`specNew`, `specNext`, `specGoto`, `specStatus`) lives directly in `cmd/spec.go`. The `cmd/` layer should only parse flags and write output; workflow logic (loading state, advancing FSM, rendering steps, building results) belongs in `internal/spec/`.

**Changes:**

First, give each step an explicit callback — no string-based template dispatch:

`StepConfig` gains a `Render` field. Each step in `Steps()` registers its own function explicitly:

```go
type StepConfig struct {
    Name   string
    Src    []string
    Dst    string
    Render func(specPath, nextStep, command string) (string, error)
}

func Steps() []StepConfig {
    return []StepConfig{
        {Name: "overview",    ..., Render: renderOverview},
        {Name: "requirements",..., Render: renderRequirements},
        // ...
    }
}

func renderOverview(specPath, nextStep, command string) (string, error) {
    return renderTemplate("spec-steps/overview.md", specPath, nextStep, command)
}
```

The generic `RenderStep(stepName, ...)` function is removed. Callers look up the step by name, get its `Render` callback, and call it directly. No magic — if a step has no callback registered, it's a compile-time/init error, not a silent runtime miss.

Then, move orchestration out of `cmd/` into `internal/spec/`:
- Add `internal/spec/spec.go` exposing `New(name, dataDir) Result`, `Next(dataDir) Result`, `Goto(step, dataDir) Result`, `Status(dataDir) StatusResult`
- `cmd/spec.go` subcommands become thin: parse `--data`, call the internal function, marshal and print result

Then, replace bool flags with subcommands:

```
spektacular spec new --data '{"name":"foo"}'            # creates new spec
spektacular spec next                                   # advance to next step
spektacular spec goto --data '{"step":"requirements"}'  # --data only, no positional args
spektacular spec status                                 # read-only
spektacular spec steps                                  # introspection (from Phase 2)
```

All input goes through `--data` as a JSON payload. No positional arguments on any subcommand.

**Note:** The old bool flag surface can be kept temporarily for backwards compatibility but should be marked deprecated.

**Files:** `internal/spec/steps.go` (add `Render` field to `StepConfig`, explicit per-step functions, remove `RenderStep`), new `internal/spec/spec.go` (orchestration), refactor `cmd/spec.go` into subcommands

---

### Phase 4 — Safety Rails (LOW priority)

**Problem:** All mutating subcommands (`spec new`, `spec next`, `spec goto`) write to disk with no preview. A `--dry-run` flag lets an agent validate or peek before committing.

**Behaviour per subcommand:**

| Subcommand | Mutation | `--dry-run` returns |
|---|---|---|
| `spec new` | Creates `specs/<name>.md` + `.state.json` | Same JSON response; no files written |
| `spec next` | Advances `.state.json` to next step | Same JSON response (next step's instruction); state not written — agent can peek without committing |
| `spec goto` | Jumps `.state.json` to named step | Same JSON response; state not written |
| `spec status` | None (read-only) | N/A — `--dry-run` is a no-op |

For `spec next` and `spec goto` the output is identical to a real run — the value is that state is not persisted, allowing an agent to preview the next instruction without advancing.

**Additional change:**
- Sanitize `spec_name` and `spec_path` before interpolating into instruction output (prevent injected content looping back to the agent)

**Files:** `cmd/spec.go`, `internal/spec/steps.go`

---

### Phase 5 — Field Selection (LOW priority)

**Problem:** Commands return full JSON responses even when an agent only needs one or two fields. Full responses waste context window tokens — particularly `instruction`, which contains an entire rendered markdown template.

**Change:** `--fields` is a **persistent root flag** — declared once on the root command and available to every subcommand. All commands that return data funnel their result through a shared output writer that applies the filter before marshaling to JSON.

This is a deliberately minimal field mask — flat JSON array of field name strings, no nesting, no aliases. Consistent with `--data`; an agent can construct `--fields` directly from the field names in `--schema` output.

**Architecture:**

```go
// internal/output/writer.go
func Write(v any, fields []string) error {
    // if fields is empty, marshal v directly
    // otherwise reflect over v, keep only named fields, error on unknown
}
```

Every command returns its result to a single call site in `cmd/root.go` (or via a shared `RunCommand` wrapper) that calls `output.Write(result, fields)`. No command marshals JSON directly.

**Examples:**

```
# Agent only needs the instruction, not metadata
spektacular --fields '["instruction"]' spec next
# {"instruction": "..."}

# Agent checking progress without the instruction blob
spektacular --fields '["current_step","completed_steps","progress"]' spec status
# {"current_step": "requirements", "completed_steps": ["overview"], "progress": "2/8"}

# Agent resuming — just needs spec_path
spektacular --fields '["spec_path","step"]' spec new --data '{"name":"foo"}'
# {"spec_path": "/abs/path/foo.md", "step": "overview"}
```

**Rules:**
- `--fields` is optional; omitting it returns the full response (no breaking change)
- Unknown field names return an error immediately (helps agents detect typos/hallucinations)
- `--schema` output documents all valid field names per subcommand
- `--schema` output itself is never filtered (always returns the full schema)

**Files:** new `internal/output/writer.go`, `cmd/root.go` (register persistent `--fields` flag, wire output writer)

## Success Criteria

- [ ] Phase 1: `spektacular spec --new --data '{"name":"../../bad"}'` returns a clear validation error
- [ ] Phase 1: Invalid config.yaml surfaces an error instead of silently defaulting
- [ ] Phase 2: `spektacular spec steps` returns valid JSON with all 8 step names
- [ ] Phase 2: `spektacular spec new --schema` returns JSON schema for `--data` input and output
- [ ] Phase 2: `spektacular spec goto --schema` returns JSON schema including valid step enum values
- [ ] Phase 3: `spektacular spec new --data '{"name":"foo"}'` works equivalently to current `--new`
- [ ] Phase 3: `spektacular spec goto --data '{"step":"requirements"}'` works equivalently to current `--step requirements`
- [ ] Phase 4: `spektacular spec new --dry-run --data '{"name":"foo"}'` returns JSON without creating files
- [ ] Phase 4: `spektacular spec next --dry-run` returns next step's instruction without advancing state
- [ ] Phase 5: `spektacular --fields '["instruction"]' spec next` returns only `{"instruction":"..."}`
- [ ] Phase 5: `spektacular --fields '["unknown_field"]' spec next` returns an error
- [ ] Phase 5: omitting `--fields` returns the full response unchanged
- [ ] Phase 5: `--fields` works identically across all data-returning subcommands
