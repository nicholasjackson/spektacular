## Step {{step}}: {{title}}

Ask the user: What is explicitly OUT of scope for this feature?

Examples:
• 'Mobile support is out of scope (tracked in #456)'
• 'Internationalisation will be addressed in a follow-up spec'

Write their response to the Non-Goals section. If blank, write 'None.'

Spec file location: {{spec_path}}

{{#next_step}}Once complete, call: spektacular spec --next{{/next_step}}
{{^next_step}}All steps complete! Review the spec file.{{/next_step}}