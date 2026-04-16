## Step {{step}}: {{title}}

Review the requirements captured in the previous step.

For each requirement, ask the user: "What is the pass/fail condition that proves this is done?"

A good criterion:
• Describes an observable outcome
• Passes or fails — no subjective judgment
• Is traceable to this requirement

Example: "When X happens, Y is visible / saved / returned."

**Stay at the blackbox level.** A criterion should be something a tester who has never read the source could verify by observing inputs and outputs — files on disk, API responses, UI state, log lines. Avoid criteria that inspect internal plumbing:

• Good: *"After the user accepts a candidate, a file exists under the knowledge directory containing the candidate's title and content."*
• Bad: *"Running the FSM goto from source state X is accepted and from source state Y is rejected."* — that's a unit test of an internal state machine, not an acceptance criterion.

If the user gives you a criterion that names internal states, private functions, or specific code paths, rephrase it as an observable outcome and tell them the internal check belongs in tests, not the spec.

Capture all criteria. Ask for clarification on any that are subjective, not traceable to a requirement, or that test internal plumbing before moving on.

Once you are satisfied with the acceptance criteria, move to the next step by running the command:

{{config.command}} spec goto --data '{"step":"{{next_step}}"}'

