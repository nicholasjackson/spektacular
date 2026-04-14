---
description: Execute an approved Plan to implement the feature
argument-hint: <plan-name>
---

# What this command does

This command drives a **multi-step interactive workflow** that executes an approved plan in `.spektacular/plans/<name>/plan.md`, producing working code, tests, and a changelog. The workflow is owned by the `{{command}}` CLI, not by you — the CLI is the state machine and you are the executor.

On each turn, the CLI returns JSON containing an `instruction` field. That instruction describes exactly one step (e.g. analyze, implement a phase, verify, update changelog, …). You must:

1. Read the `instruction` carefully.
2. Perform the step — this may mean reading the plan, spawning subagents, editing code, running tests, or writing to the changelog.
3. When the step is complete, run the `goto` command named at the bottom of the instruction to advance the state machine.
4. Read the next `instruction` from the new JSON response and repeat.

**This is a loop. Do not stop after the first step.** Keep looping — step → goto → next instruction → step — until a returned instruction tells you the workflow is *finished*. Only then should you report completion to the user.

# How to start

Plan name: $ARGUMENTS

If no plan name was provided, check `.spektacular/state.json` for an active plan under `data.name`. If one exists, ask the user whether they want to implement that plan, offering the option to name a different one. If no active plan is found, ask the user which plan to implement before proceeding.

The plan file must already exist at `.spektacular/plans/<plan_name>/plan.md`. If it does not, stop and tell the user to run `{{command}} plan` first.

Start the implement workflow by running:

```
{{command}} implement new --data '{"name": "<plan_name>"}'
```

This creates the state file automatically and returns the first `instruction`. From that point on, follow the loop above: do what the instruction says, then call `{{command}} implement goto --data '{"step":"<next_step>"}'` to get the next one. Do not invent step names — every instruction tells you the exact `goto` command to run next.
