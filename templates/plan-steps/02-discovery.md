## Step {{step}}: {{title}}

Research the codebase to understand what's needed to implement the spec you read in the previous step. The output of this step is `research.md` — a **decision log**, not a transcript — that captures your investigation so a future cold session can rehydrate without re-doing the work.

### Step 1: Project Context

Check if reference docs exist and create them if missing:

- Check if `thoughts/notes/commands.md` exists. If not, use the `discover-project-commands` skill. For skill details: `{{config.command}} skill discover-project-commands`
- Check if `thoughts/notes/testing.md` exists. If not, use the `discover-test-patterns` skill. For skill details: `{{config.command}} skill discover-test-patterns`

### Step 2: Codebase Research

Research the codebase in parallel to find:

1. **Files related to the plan** — Organize by category (implementation, tests, config, docs)
2. **Prior research** — Existing plans, research, or tickets in `thoughts/`, `.spektacular/plans/`, `.spektacular/specs/`
3. **Similar implementations** — Code examples to model after, with file:line references
4. **Architecture and integration points** — How the relevant components fit together
5. **Alternatives to consider** — Identify 2-3 viable approaches so the next step can compare them with evidence

Use your agent orchestration capability to parallelize this research. For guidance: `{{config.command}} skill spawn-planning-agents`

### Step 3: Distill findings into research.md — the decision log

You are not writing research.md to disk yet (the verification step will handle that). You are gathering the content that will go into these required sections:

- **Alternatives considered and rejected** — options you considered; for each, what it is, why rejected, with citation (file:line or external reference). This prevents future agents from re-proposing the same dead ends.
- **Chosen approach — evidence** — the file:line or external references that support the option you'll recommend in the next step. Evidence, not the decision itself.
- **Files examined** — terse one-liner per file: `path:line — what was learned`.
- **External references** — papers, RFCs, library docs, blog posts, with a one-line "why this mattered".
- **Prior plans / specs consulted** — links with what was learned from each.
- **Open assumptions** — things assumed but not verified. If any turn out wrong, the implement workflow must STOP and ask.
- **Rehydration cues** — skill invocations and file re-reads that can rebuild context from cold.

Keep this dense. Assume a future agent will read it cold and need to make decisions from it.

### Step 4: Read and Clarify

- Read all findings fully
- Ask only questions the code cannot answer
- Present a summary of key discoveries to the user

Once research is complete, advance:

{{config.command}} plan goto --data '{"step":"{{next_step}}"}'
