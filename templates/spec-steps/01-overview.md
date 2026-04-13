## Step {{step}}: {{title}}

Ask the user to describe this feature in 2-3 sentences:
• What is being built?
• What problem does it solve?
• Who benefits?

Capture their response. Be specific — avoid generic phrases like 'improve the experience'.
Ask for clarification if the description is vague or incomplete before moving on.

Once you are satisfied with the overview, move to the next step by running the command:

{{config.command}} spec goto --data '{"step":"{{next_step}}"}'
