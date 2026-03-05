# Workflow Step Architecture

## Overview

A workflow is a linear FSM (finite state machine) driven by `looplab/fsm`. Each step is a
`workflow.StepConfig` that wires together an FSM event name, source/destination states, and an
optional callback. Steps are defined in domain packages (e.g. `internal/spec`) and passed into
`workflow.New()`.

## Packages

| Package | Responsibility |
|---|---|
| `internal/workflow` | FSM, state persistence, `Data` interface, `Config` |
| `internal/spec` | Spec-specific steps, templates, result types |
| `cmd` | Thin command handlers — parse input, create workflow, call `Next`/`Goto`, write output |

## Adding a Step

### 1. Add the step to `spec.Steps()`

```go
// internal/spec/steps.go
func Steps() []workflow.StepConfig {
    return []workflow.StepConfig{
        // ...existing steps...
        {Name: "my_step", Src: []string{"previous_step"}, Dst: "my_step", Callback: myStep()},
    }
}
```

The FSM uses `Src` to guard transitions — calling `goto my_step` from any state not in `Src`
returns an error. `Dst` is the state the FSM moves into after the step fires.

### 2. Write the callback

```go
func myStep() workflow.StepCallback {
    return func(data workflow.Data, out workflow.ResultWriter, cfg workflow.Config) error {
        return writeStepResult("my_step", "spec-steps/my_step.md", data, out, cfg)
    }
}
```

The callback receives:
- `data workflow.Data` — read/write key-value store persisted in state. Read spec context here
  (`spec_path`, `name`, etc.). Write anything the step needs to record for later steps.
- `out workflow.ResultWriter` — write the result JSON the agent receives.
- `cfg workflow.Config` — runtime config (`Command`, `DryRun`). Not persisted.

**Rules:**
- The callback must not access workflow internals (current step, completed steps, totals).
- The callback must not write to any file directly — use the `Data` store.
- If `cfg.DryRun` is true, skip all side effects.

### 3. Add the template

Create `templates/spec-steps/my_step.md`. The template receives the bundle constructed in
`writeStepResult` plus `config`:

```
{{step}}          — step name (snake_case)
{{title}}         — step name formatted as title case
{{spec_path}}     — path to the spec file
{{config.command}} — the CLI binary name (e.g. "spektacular")
```

Templates should instruct the agent what to do, where to read/write, and what to call next.

### 4. Add a template for spec-scaffold if needed

The spec scaffold (`templates/spec-scaffold.md`) defines the initial Markdown structure of a new
spec file. Add a new section heading if the step writes to a new section.

## Data Store

State is split into two concerns:

**Core** (managed by workflow, not accessible to callbacks):
- `current_step`, `completed_steps`, timestamps

**Data** (`map[string]any`, persisted alongside core, accessible to callbacks):
- Callbacks read and write arbitrary keys here.
- Seeded by the command handler before the workflow starts (e.g. `name`, `spec_path`).
- Mutations from callbacks are persisted on the next state transition.

## Callback Signature

```go
type StepCallback func(data Data, out ResultWriter, cfg Config) error
```

This is the only interface between the workflow engine and domain logic. Keep it narrow.

## Config

`workflow.Config` carries runtime-only values that are not persisted:

```go
type Config struct {
    Command string // CLI binary name, injected into templates as {{config.command}}
    DryRun  bool   // skip all side effects when true
}
```

Set at workflow construction time by the command handler. Never stored in state.

## Flow: `spec new`

```
runSpecNew
  → workflow.New(spec.Steps(), statePath, wfCfg)
  → wf.SetData("name", ...)
  → wf.SetData("spec_path", ...)
  → wf.Next()        // fires "new" step: creates spec scaffold file, no output
  → wf.Next(out)     // fires "overview" step: renders template, writes Result to agent
```

## Flow: `spec goto`

```
runSpecGoto
  → workflow.New(spec.Steps(), statePath, wfCfg)  // loads persisted state
  → wf.Goto(stepName, out)                         // FSM errors if transition invalid
```

The FSM guards against calling a step from the wrong state. The agent must always call steps in
order unless the workflow explicitly allows branching via multiple `Src` entries.
