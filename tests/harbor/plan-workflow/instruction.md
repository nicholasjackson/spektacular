# Create an Implementation Plan using Spektacular

You are testing the `spektacular` CLI tool by driving a complete plan workflow
against a pre-existing specification. The binary is already installed at
`/usr/local/bin/spektacular`.

## Setup

First initialize the project:

```bash
spektacular init claude
```

A specification file already exists at `.spektacular/specs/user-auth.md`
describing a stateless JWT authentication feature. Read it before you start
planning — the plan workflow's first step needs that context.

## Task

Drive the full plan workflow against the `user-auth` specification by using
the `/spek:plan` skill that was installed during init:

```
/spek:plan user-auth
```

The skill will guide you through every plan step from `overview` through
`finished`. Follow each rendered instruction exactly — in particular:

- At the `discovery` step, use your agent-orchestration capability to spawn
  sub-agents in parallel for codebase research, and retrieve the skills the
  step template references (`discover-project-commands`, `discover-test-patterns`,
  `spawn-planning-agents`).
- At the `phases` step, retrieve the `spawn-implementation-agents` skill the
  template references.
- At the `verification` step, retrieve the `gather-project-metadata` and
  `determine-feature-slug` skills the template references. Then follow the
  rendered pipe instructions: spektacular writes each file when you pipe the
  filled content back via `--stdin plan_template` / `--stdin context_template`
  / `--stdin research_template` across the `write_plan`, `write_context`, and
  `write_research` steps. Do **not** use the `Write` tool for these three files
  — the workflow owns the write.

Write meaningful, non-placeholder content for every section of every artefact.
The plan is a plan for *implementing the JWT authentication feature described
in the spec*, so draft content should talk about JWT, tokens, auth middleware,
and related concepts.

## After completion

Copy the `.spektacular` directory to `/logs/artifacts/` so results are
collected:

```bash
cp -r /app/.spektacular /logs/artifacts/spektacular
```

### Success criteria

- The workflow reaches the `finished` state
- All steps appear in `completed_steps` in canonical order
- `plan.md`, `context.md`, `research.md` exist under `.spektacular/plans/user-auth/`
- Each section of each artefact has meaningful, non-placeholder text
- The agent retrieved every template-referenced skill during the step that referenced it
- The agent spawned at least one sub-agent during the `discovery` step
