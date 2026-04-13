## Step {{step}}: {{title}}

Complete the following template with all of the information you have gathered so far

```markdown
{{spec_template}}
```

Review all the information gathered across every step and validate the complete spec for:
• Completeness — all sections are filled
• Clarity — requirements are specific and testable
• Consistency — sections reference each other appropriately

Report any issues to the user and ask for clarification until you are confident the spec is correct and complete.

Once the user is happy, produce the final complete spec and pipe it to the following command:

cat <<'EOF' | {{config.command}} spec goto --data '{"step":"{{next_step}}"}' --stdin spec_template
<complete filled spec here>
EOF

