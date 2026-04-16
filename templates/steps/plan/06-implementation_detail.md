## Step {{step}}: {{title}}

Draft the **Implementation Detail** section of `plan.md`.

### This section is high-level only

Sketch new patterns being introduced, major code-shape changes, and code-structure UX — enough for a reviewer to spot missing patterns or design gaps. This section is **high-level only**. Per-phase file:line work stays in `context.md`.

If you find yourself writing "in file X at line Y", stop and move that content to context.md. The test for "too low-level" is: could this be written before the phases are defined? If no, it belongs in context.md.

### What to include

- New patterns the plan introduces (e.g. a new state machine, a new agent orchestration shape, a new module boundary)
- Major code-shape changes (e.g. splitting one package into three, or introducing a new interface that replaces a concrete type across the codebase)
- Code-structure UX: what a developer reading the changed code will experience
- Where existing patterns in the codebase are being followed vs. where new ones are being introduced

### What NOT to include

- Specific file paths or line numbers
- Per-phase file changes
- Function signatures that belong in § Data Structures & Interfaces
- Shell commands

### What to produce

A draft Implementation Detail section ready to drop into plan.md at verification time. Present it to the user for review.

Once the user is happy with the sketch, advance:

{{config.command}} plan goto --data '{"step":"{{next_step}}"}'
