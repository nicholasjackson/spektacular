## Step {{step}}: {{title}}

Ask the user to describe this feature in 2-3 sentences:
• What is being built?
• What problem does it solve?
• Who benefits?

Capture their response. Be specific — avoid generic phrases like 'improve the experience'.

**Keep the overview stakeholder-readable.** No file paths, no section names, no step names, no framework or library names, no code identifiers. A non-engineer should be able to read it and understand the value. If you can't explain the feature without naming an implementation artifact, ask the user one more "so that…" question until you find the user-visible value.

If the user volunteers implementation detail ("we add a new FSM state", "append to research.md"), capture it mentally and tell them it will land in Technical Approach — then rephrase the overview at the behavior level.

Ask for clarification if the description is vague, incomplete, or leaks implementation before moving on.

Once you are satisfied with the overview, move to the next step by running the command:

{{config.command}} spec goto --data '{"step":"{{next_step}}"}'
