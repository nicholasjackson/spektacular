## Step {{step}}: {{title}}

plan.md has been written to `{{plan_path}}`.

Now submit the filled context.md the same way — use the `Write` tool to stage it at `.spektacular/tmp/context_template.md`, then:

```
{{config.command}} plan goto --data '{"step":"{{next_step}}"}' --file .spektacular/tmp/context_template.md
```

Spektacular reads the file and writes the final context to `{{context_path}}`.
