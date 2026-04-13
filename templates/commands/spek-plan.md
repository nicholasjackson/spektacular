---
description: Create a new Plan from a Specification
argument-hint: <spec-name>
---

Create a new plan from an existing spec using the {{command}} CLI. The user must provide the name of the spec to plan against as the argument.

Spec name: $ARGUMENTS

If no spec name was provided, ask the user which spec to plan against before proceeding.

Start the plan workflow by running:

```
{{command}} plan new --data '{"name": "<spec_name>"}'
```

The command creates the plan file and state file automatically, then returns JSON with an `instruction` field. Follow that instruction exactly.

Each instruction tells you to call `{{command}} plan goto --data '{"step":"{{next_step}}"}'` when the step is complete.
