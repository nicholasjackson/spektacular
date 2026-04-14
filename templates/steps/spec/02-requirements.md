## Step {{step}}: {{title}}

Ask the user to list the specific, testable behaviours this feature must deliver.

Use active voice:
• 'Users can...'
• 'The system must...'

Each item should be independently verifiable. One behaviour per line.

**Requirements describe WHAT, not HOW.** Apply this test to every requirement before you write it down:

> *If this were implemented with a completely different mechanism, would the requirement still be true?*

If the answer is no, the requirement is prescribing implementation — rephrase it at the behavior level and park the mechanism for Technical Approach.

Example: the user says *"candidates are appended to `research.md` under a `## Knowledge Candidates` section."* That names a specific file and section — a different implementation (a sidecar file, an in-memory queue) would falsify the requirement. Capture it as *"the workflow surfaces candidate knowledge for review before anything becomes persistent"* and tell the user the `research.md` detail will land in Technical Approach.

Warning signs that a requirement is leaking HOW: specific file paths, section or heading names, step or state names, framework/library names, code identifiers, numeric step positions ("after step 13"). If you see any, rephrase.

Capture the requirements. Ask for clarification on any that are vague, ambiguous, not independently verifiable, or that leak implementation before moving on.

Once you are satisfied with the requirements, move to the next step by running the command:

{{config.command}} spec goto --data '{"step":"{{next_step}}"}'

