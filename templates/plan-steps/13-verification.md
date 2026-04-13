## Step {{step}}: {{title}}

Time to compile the three plan documents and write them to disk.

### Step 1: Gather Metadata

Use the `gather-project-metadata` skill to collect: ISO timestamp, git commit, branch, and repository info.
For skill details: `{{config.command}} skill gather-project-metadata`

### Step 2: Determine Feature Slug

Use the `determine-feature-slug` skill to determine the plan file namespace and number.
For skill details: `{{config.command}} skill determine-feature-slug`

### Step 3: Fill in the Three Scaffolds

Fill in ALL sections of all three scaffolds — no placeholders, no open questions.

#### plan.md scaffold

```markdown
{{plan_template}}
```

#### context.md scaffold

```markdown
{{context_template}}
```

#### research.md scaffold

```markdown
{{research_template}}
```

### Step 4: Review

Verify the three filled documents together before writing:

- **plan.md** — readable in under a minute; every phase has a summary paragraph, a `*Technical detail:*` link, and outcome-based acceptance criteria; no shell commands anywhere.
- **context.md** — per-phase technical notes under headings matching plan.md's `*Technical detail:*` anchors; current state analysis; testing strategy; migration notes.
- **research.md** — alternatives considered and rejected with citations; files examined; open assumptions; rehydration cues. Dense enough to rehydrate a cold session.

### Step 5: Write the Files Directly

Use the `Write` tool to write all three files to disk:

- `{{plan_path}}` — the filled plan.md
- `{{context_path}}` — the filled context.md
- `{{research_path}}` — the filled research.md

Do **not** pipe content back via stdin. Write each file directly using the `Write` tool.

### Step 6: Advance

Once all three files are written, advance to the finished step:

{{config.command}} plan goto --data '{"step":"{{next_step}}"}'
