## Step {{step}}: {{title}}

Read the spec file in full.

Validate every section (Overview, Requirements, Acceptance Criteria, Constraints, Technical Approach, Success Metrics, Non-Goals) for:
• Completeness — all sections are filled
• Clarity — requirements are specific and testable
• Consistency — sections reference each other appropriately

Report any issues found. If there are gaps or unclear sections, ask the user for clarification.

Once all sections are validated and complete, confirm the spec is ready.

Spec file location: {{spec_path}}

{{#next_step}}Once complete, call: spektacular spec --next{{/next_step}}
{{^next_step}}All steps complete! Review the spec file.{{/next_step}}