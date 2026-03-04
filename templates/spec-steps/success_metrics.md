## Step {{step}}: {{title}}

Ask the user: How will you know this feature is working well after delivery?

Be specific:
• Quantitative: 'p99 latency < 200ms', 'error rate < 0.1%'
• Behavioral: 'users complete the flow without support intervention'

Write their response to the Success Metrics section. If blank, write 'None.'

Spec file location: {{spec_path}}

{{#next_step}}Once complete, call: spektacular spec --next{{/next_step}}
{{^next_step}}All steps complete! Review the spec file.{{/next_step}}