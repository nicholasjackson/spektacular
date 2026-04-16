## Step {{step}}: {{title}}

Draft the **Testing Approach** section of `plan.md`.

### This section is high-level only

Describe the overall testing strategy and test types. This section is **high-level only**. Per-phase testing detail — which specific tests live in which specific files — stays in `context.md`.

If you find yourself writing "a test in file X asserts Y on line Z", stop and move that content to context.md.

### What to include

- The kinds of tests being added (unit, integration, contract, regression, end-to-end)
- Which components get the most coverage and why
- The load-bearing assertions — what, in plain language, the tests guarantee
- Where tests slot into existing test conventions in the project
- Any deliberate gaps (e.g. "not adding integration tests because the contract is exercised by unit tests")

### What NOT to include

- Specific test file paths
- Per-phase test lists
- Shell commands to run the tests

### What to produce

A draft Testing Approach section ready to drop into plan.md at verification time. Present it to the user for review.

Once the user is happy with the testing strategy, advance:

{{config.command}} plan goto --data '{"step":"{{next_step}}"}'
