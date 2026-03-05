## Step {{step}}: {{title}}

Ask the user: Are there any hard constraints or boundaries the solution must operate within?

Examples:
• Must integrate with the existing authentication system
• Cannot introduce breaking changes to the public API
• Must support the current minimum supported runtime versions

Capture their response. If blank, note that there are no constraints.

Once you are satisfied, move to the next step by running the command:

{{config.command}} spec goto --data '{"step":"{{next_step}}"}'

