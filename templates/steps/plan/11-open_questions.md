## Step {{step}}: {{title}}

Draft the **Open Questions** section of `plan.md`.

### Strict scope

This section is **strictly for questions that genuinely cannot be resolved until implementation begins**. Anything that can be answered by asking the user, reading the code, or running a quick experiment must be resolved now — not parked here.

Examples of what belongs here:

- "Whether the downstream API returns X or Y under condition Z — only discoverable by exercising it"
- "Whether the refactor exposes a hidden assumption in an untested code path — will surface during implementation"

Examples of what does NOT belong here:

- "Which library should we use?" → ask the user now
- "What is the current shape of the API?" → read the code now
- "How should we name the new field?" → decide now
- "Is there a test already covering this?" → check now

### What to produce

A draft Open Questions section. If, after a genuine pass, there are no impl-time-only uncertainties, state that explicitly: an empty Open Questions section is the expected healthy outcome.

For every item you keep: state the question, state what it depends on, and state what the implementer should do when they hit it (usually: STOP and ask the user).

Once the user agrees the list is correctly scoped, advance:

{{config.command}} plan goto --data '{"step":"{{next_step}}"}'
