## Step {{step}}: {{title}}

context.md has been written to `{{context_path}}`.

Now submit the filled research.md — use the `Write` tool to stage it at `.spektacular/tmp/research_template.md`, then:

```
{{config.command}} plan goto --data '{"step":"{{next_step}}"}' --file .spektacular/tmp/research_template.md
```

Spektacular reads the file and writes the final research to `{{research_path}}`.
