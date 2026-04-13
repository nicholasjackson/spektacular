## Step {{step}}: {{title}}

Review the requirements captured in the previous step.

For each requirement, ask the user: "What is the pass/fail condition that proves this is done?"

A good criterion:
• Describes an observable outcome
• Passes or fails — no subjective judgment
• Is traceable to this requirement

Example: "When X happens, Y is visible / saved / returned."

Capture all criteria. Ask for clarification on any that are subjective or not traceable to a requirement before moving on.

Once you are satisfied with the acceptance criteria, move to the next step by running the command:

{{config.command}} spec goto --data '{"step":"{{next_step}}"}'

