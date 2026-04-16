## Step {{step}}: {{title}}

Draft the **Data Structures & Interfaces** section of `plan.md`. This section captures the types, interface signatures, and serialization boundaries introduced or changed by the plan.

### Rules

- Focus on contracts between components, not internal representation detail.
- Show type shapes in pseudocode or a short code block where it helps — but keep it concise; this is a plan, not source.
- Name each type or interface and describe its role in one or two sentences.
- If no new data structures or interfaces are introduced, say so explicitly — an empty section is not acceptable.
- Per-field implementation detail (defaults, validation, wire-format nuance) belongs in context.md, not here.

### What to produce

A draft Data Structures & Interfaces section ready to drop into plan.md at verification time. Present it to the user for review.

Once the user is happy with the contracts, advance:

{{config.command}} plan goto --data '{"step":"{{next_step}}"}'
