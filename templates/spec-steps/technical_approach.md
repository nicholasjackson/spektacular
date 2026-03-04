## Step {{step}}: {{title}}

Ask the user: Do you have any technical direction already decided?

Examples:
• Key architectural decisions already made
• Preferred patterns or technologies
• Integration points with existing systems
• Known risks or areas of uncertainty

Write their response to the Technical Approach section. If blank, write 'None.'

Spec file location: {{spec_path}}

{{#next_step}}Once complete, call: spektacular spec --next{{/next_step}}
{{^next_step}}All steps complete! Review the spec file.{{/next_step}}