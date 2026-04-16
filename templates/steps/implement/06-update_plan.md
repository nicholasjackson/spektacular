## Step {{step}}: {{title}}

Mark the current phase's acceptance criteria as complete in `{{plan_path}}`.

### What to edit

For the current phase (the one you just implemented, tested, and verified):

1. Change the phase heading from `#### - [ ] Phase N.M: <title>` to `#### - [x] Phase N.M: <title>`.
2. Change each `- [ ]` acceptance-criterion checkbox in the phase's `**Acceptance criteria**:` block to `- [x]` **only if** that criterion actually passed verification in the previous step.
3. Leave criteria that did not pass as `- [ ]`. Do not mark them complete just because the phase is "mostly done".

### STOP-on-mismatch

If any acceptance criterion passed verification but describes an outcome that no longer matches what the code actually does (e.g. the criterion says "function X returns Y" but the implementation returns Z and the user authorized the change), STOP. The plan must be updated to reflect the new reality before the checkbox can flip. Ask the user whether to (a) update the criterion text in `{{plan_path}}` first, (b) leave the checkbox unchecked and note the deviation for the changelog, or (c) flip it anyway because the user accepts the divergence.

### Advance

Once checkboxes are updated:

```
{{config.command}} implement goto --data '{"step":"{{next_step}}"}'
```
