## Step {{step}}: {{title}}

Draft the **Dependencies** section of `plan.md`. List the internal packages, external libraries, upstream specs, and prior plans this work depends on.

### Rules

- One bullet per dependency.
- For each: a one-line note on what it provides and whether it needs any changes.
- Cover both runtime dependencies (imported packages, external services) and planning dependencies (prior specs or plans that must land first).
- If a dependency must land or change before this plan starts, flag that explicitly.
- If there are no meaningful dependencies, say so explicitly — an empty section is not acceptable.

### What to produce

A draft Dependencies section ready to drop into plan.md at verification time. Present it to the user for review.

Once the user is happy with the dependency list, advance:

{{config.command}} plan goto --data '{"step":"{{next_step}}"}'
