## Step {{step}}: {{title}}

Decide the shape of the solution and lock in the chosen direction. This step produces the **Architecture & Design Decisions** section of `plan.md` — the load-bearing section of the whole plan. A reviewer should be able to spot missing patterns or design gaps from this section alone.

### Step 1: Present Options

Present 2-3 design options to the user. For each:

- **Option name** and brief description
- **Pros**: Advantages, with file:line references from research.md where they apply
- **Cons**: Disadvantages, risks, complexity
- **Effort estimate**: Relative complexity (Low / Medium / High)

Ground each option in the research findings gathered during the discovery step.

### Step 2: Get Agreement

Get the user's explicit agreement on:

1. **Chosen direction** — which option to pursue
2. **Key design decisions** — the non-obvious trade-offs the chosen direction makes

Rejected options go into `research.md § Alternatives considered and rejected` with citations when the verification step writes the files.

### Step 3: Draft the Architecture & Design Decisions section

Draft 2-4 short paragraphs describing:

- The shape of the chosen solution
- The key design decisions and their trade-offs
- Why this direction beats the alternatives
- A cross-reference to `research.md#alternatives-considered-and-rejected` so plan.md readers can drill into the evidence

Keep this section self-contained. Do NOT write `see context.md for …` — plan.md must stand on its own for readers outside the Milestones & Phases block.

Once the user has agreed on the chosen direction and the draft is ready, advance:

{{config.command}} plan goto --data '{"step":"{{next_step}}"}'
