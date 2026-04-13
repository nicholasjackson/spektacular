## Step {{step}}: {{title}}

Ask the user: How will you know this feature is working well after delivery?

Be specific:
• Quantitative: 'p99 latency < 200ms', 'error rate < 0.1%'
• Behavioral: 'users complete the flow without support intervention'

Capture their response. If blank, note that no success metrics have been defined.

Once you are satisfied, move to the next step by running the command:

{{config.command}} spec goto --data '{"step":"{{next_step}}"}'

