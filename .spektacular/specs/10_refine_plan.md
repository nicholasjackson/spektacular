# Feature: 10 Refine Plan

## Overview
<!--
  OVERVIEW
  A concise 2-3 sentence summary of the feature. Answer three questions:
    1. What is being built?
    2. What problem does it solve?
    3. Who benefits and why does it matter?
  Avoid implementation details — this should be readable by any stakeholder.
-->
Plan mode should allow the refinement and provision of user feedback for plans. This enables the user to iteratively refine a generated plan by providing additional knowledge, guidance, or corrections. Users benefit from higher-quality plans that incorporate their domain expertise and preferences.

## Requirements
<!--
  REQUIREMENTS
  Specific, testable behaviours the feature must deliver.
  Format: bold title on the checkbox line, detail indented below.
  Rules:
    - Use active voice: "Users can...", "The system must..."
    - Each requirement should be independently verifiable
    - Focus on WHAT, not HOW — avoid prescribing implementation
    - Keep each item atomic — one behaviour per line
-->
- [ ] **Prompt for feedback at end of plan mode**
  At the end of plan generation the user is prompted to read the plan and provide feedback.
- [ ] **Feedback text box with clarification loop**
  The system enables a text box for user input. On receiving feedback, the agent enters a clarification loop before generating a new plan.
- [ ] **Multiple feedback turns**
  Multiple rounds of feedback can happen in a single session. After each round, the user is prompted for further feedback.
- [ ] **Feedback via plan file editing**
  The user can provide feedback by directly editing the plan file. The agent checks for changes when told by the user.
- [ ] **New plan produced after each feedback turn**
  A new plan is written after every round of feedback. If no feedback is provided, no new plan is written.
- [ ] **Plan refinement from a new session**
  A user can resume refinement of an existing plan in a new session using the existing `plan` command.

## Constraints
<!--
  CONSTRAINTS
  Hard boundaries the solution must operate within.
-->
None.

## Acceptance Criteria
<!--
  ACCEPTANCE CRITERIA
  The specific, binary conditions that define "done".
  Format: bold title on the checkbox line, verifiable detail indented below.
  Each criterion must be:
    - Independently verifiable (pass/fail, not subjective)
    - Traceable back to a requirement above
    - Testable by someone who didn't write the code
-->
- [ ] **Feedback prompt offers text input, file editing, and quit options**
  The system allows the user to enter feedback in a text box, provide feedback by editing the plan file, or quit without providing any feedback.
- [ ] **Clarification questions asked when feedback is ambiguous**
  When there is ambiguity or the agent does not understand the user's response, it asks "Can you clarify what you mean by ..." or asks a specific question. When everything is clear, the agent progresses to the next step.
- [ ] **"Anything else?" prompt after each feedback round**
  After feedback has been collected and a new plan produced, the agent asks "Anything else?" to prompt a new round of feedback.
- [ ] **Agent detects and acknowledges plan file edits**
  When the user tells the agent they have edited the plan file, the agent checks for changes and says "Let me check the changes you have made to the plan file."
- [ ] **New plan written after every feedback round**
  After every round of feedback the agent writes a new plan. If there is no feedback, no new plan is written.
- [ ] **Existing plan can be resumed in a new session**
  A user can resume an existing plan session using the same `plan` command with the `--resume` flag, where the plan directory becomes the positional argument instead of a spec.

## Technical Approach
<!--
  TECHNICAL APPROACH
  High-level technical direction to guide the planning agent.
-->
For resuming a plan, the existing `plan` command should be used along with a `--resume` flag. When this flag is present, a plan directory becomes the positional argument instead of a spec.

## Success Metrics
<!--
  SUCCESS METRICS
  How you will know the feature is working well after delivery.
-->
None.

## Non-Goals
<!--
  NON-GOALS
  Explicitly state what this spec does NOT cover.
-->
None.
