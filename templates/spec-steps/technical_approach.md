## Step {{step}}: {{title}}

Ask the user: Do you have any technical direction already decided?

Examples:
• Key architectural decisions already made
• Preferred patterns or technologies
• Integration points with existing systems
• Known risks or areas of uncertainty

Capture their response. If blank, note that no technical direction has been decided.

Once you are satisfied, move to the next step by running the command:

{{config.command}} spec goto --data '{"step":"{{next_step}}"}'

