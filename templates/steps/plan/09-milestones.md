## Step {{step}}: {{title}}

Define 2-4 milestones that deliver the spec you have in context.

Each milestone must have:

- **Title**: A user-facing description of what changes (not an engineering label like "Refactor X").
- **What changes**: A one-paragraph summary describing the user-visible difference when the milestone lands. Written in plain language, no file paths, no shell commands. This paragraph is what a reader of plan.md uses to decide whether the milestone is worth doing — make it outcome-focused, not implementation-focused.
- **Validation point**: How to confirm the milestone is done before moving on.

Rules:

- Each milestone should be independently deliverable.
- Milestones should build on each other in order.
- NO open questions — resolve any uncertainties now by asking the user.
- Purely internal cleanups (e.g. refactors with no user-visible effect) are allowed, but the "What changes" paragraph must say so explicitly and explain why the cleanup is worth its own milestone.

Present the milestones to the user for review. Once agreed, advance:

{{config.command}} plan goto --data '{"step":"{{next_step}}"}'
