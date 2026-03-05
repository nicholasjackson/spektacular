## Step {{step}}: {{title}}

Ask the user to list the specific, testable behaviours this feature must deliver.

Use active voice:
• 'Users can...'
• 'The system must...'

Each item should be independently verifiable. One behaviour per line.

Capture the requirements. Ask for clarification on any that are vague, ambiguous, or not independently verifiable before moving on.

Once you are satisfied with the requirements, move to the next step by running the command:

{{config.command}} spec goto --data '{"step":"{{next_step}}"}'

