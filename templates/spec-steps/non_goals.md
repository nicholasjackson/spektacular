## Step {{step}}: {{title}}

Ask the user: What is explicitly OUT of scope for this feature?

Examples:
• 'Mobile support is out of scope (tracked in #456)'
• 'Internationalisation will be addressed in a follow-up spec'

Capture their response. If blank, note that no non-goals have been defined.

Once you have captured the information from the user move to the next step by running the command:

{{config.command}} spec goto --data '{"step":"{{next_step}}"}'

