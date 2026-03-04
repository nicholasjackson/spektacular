## Step {{step}}: {{title}}

Read all requirements from the spec file Requirements section.

For each requirement, ask the user: "What is the pass/fail condition that proves this is done?"

A good criterion:
• Describes an observable outcome
• Passes or fails — no subjective judgment
• Is traceable to this requirement

Example: "When X happens, Y is visible / saved / returned."

Write all criteria to the Acceptance Criteria section as a checklist.

Spec file location: {{spec_path}}

{{#next_step}}Once complete, call: spektacular spec --next{{/next_step}}
{{^next_step}}All steps complete! Review the spec file.{{/next_step}}