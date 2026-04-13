## Step {{step}}: {{title}}

The plan workflow is complete. Three files should now exist next to each other under `{{plan_dir}}`:

- `{{plan_path}}` — the user-scannable plan
- `{{context_path}}` — technical detail for implementation
- `{{research_path}}` — the decision log and rehydration cues

If any of these files is missing, STOP and ask the user how to proceed — do not advance.

Inform the user that the plan workflow is finished and the three documents are ready for review.

When ready to share the plan with the team, use the `share-docs` skill to promote it to the shared namespace.
