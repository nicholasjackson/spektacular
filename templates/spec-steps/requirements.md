## Step {{step}}: {{title}}

Ask the user to list the specific, testable behaviours this feature must deliver.

Use active voice:
• 'Users can...'
• 'The system must...'

Each item should be independently verifiable. One behaviour per line.

Format the requirements as a markdown checklist and write them to the Requirements section:
- [ ] **Title** — description

Spec file location: {{spec_path}}

{{#next_step}}Once complete, call: spektacular spec --next{{/next_step}}
{{^next_step}}All steps complete! Review the spec file.{{/next_step}}