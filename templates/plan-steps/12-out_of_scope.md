## Step {{step}}: {{title}}

Draft the **Out of Scope** section of `plan.md`. This section lists the explicit exclusions agreed during planning — what this plan is NOT doing, and where those things are tracked instead.

### Rules

- One bullet per exclusion.
- State what is not being done, in plain language.
- Where a follow-up exists (another spec, another plan, a ticket), name it so a reader can find it.
- Pull out-of-scope items from three places:
  1. The spec's § Non-Goals section
  2. Anything the user said "not now" to during the architecture step
  3. Anything the chosen design deliberately leaves to a later plan
- If the plan truly has no exclusions worth naming, say so explicitly — an empty section is usually a sign something was missed.

### What to produce

A draft Out of Scope section ready to drop into plan.md at verification time. Present it to the user for review and confirm each exclusion.

Once the user agrees on the exclusions, advance:

{{config.command}} plan goto --data '{"step":"{{next_step}}"}'
