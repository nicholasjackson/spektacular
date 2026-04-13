## Step {{step}}: {{title}}

Draft the **Component Breakdown** section of `plan.md`. This section lists the components (new or changed) that make up the chosen solution, their responsibilities, and how they interact.

### Rules

- One bullet or short paragraph per component.
- For each component: name it, state what it owns, and describe its relationship to the other components.
- Do NOT list file paths or line numbers in plan.md — those belong in context.md. Component responsibilities, not implementation sites.
- Reuse existing components wherever possible; a new component needs justification.
- Cover both new components and meaningfully-changed existing ones.

### What to produce

A draft Component Breakdown section ready to drop into plan.md at verification time. Present it to the user for review.

Once the user is happy with the component list, advance:

{{config.command}} plan goto --data '{"step":"{{next_step}}"}'
