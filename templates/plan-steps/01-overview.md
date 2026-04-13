## Step {{step}}: {{title}}

Read the specification file at `{{spec_path}}` to understand what is being planned.

This spec is the source of truth for the plan's scope, requirements, and constraints. Keep it in mind throughout the remaining planning steps — subsequent prompts will not repeat its contents.

Once you have read and understood the spec, advance to the next step:

{{config.command}} plan goto --data '{"step":"{{next_step}}"}'
