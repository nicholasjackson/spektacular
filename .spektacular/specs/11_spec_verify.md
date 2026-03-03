# Feature: 11 Spec Verify

<!--
  OVERVIEW
  A concise 2-3 sentence summary of the feature. Answer three questions:
    1. What is being built?
    2. What problem does it solve?
    3. Who benefits and why does it matter?
  Avoid implementation details — this should be readable by any stakeholder.
-->
## Overview

Once a spec has been produced it should be verified to ensure the content for each section is clearly defined and well understood. This acts as a sanity check on the user's input, catching vague or incomplete sections before the spec moves into planning and implementation.


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
## Requirements

- [ ] **Validate completed spec** — Once all sections of a spec have been completed, the system must validate that the final spec is correct and contains enough detail for planning and implementation.
- [ ] **Ask user for clarification** — The agent must be able to ask the user clarifying questions when a section is vague, incomplete, or ambiguous.
- [ ] **Perform research for clarification** — The agent must be capable of performing research (e.g. exploring the codebase, reading related files) to clarify any areas of uncertainty before or during validation.

<!--
  CONSTRAINTS
  Hard boundaries the solution must operate within. These are non-negotiable.
  Examples:
    - Must integrate with the existing authentication system
    - Cannot introduce breaking changes to the public API
    - Must support the current minimum supported runtime versions
  Leave blank if there are no constraints.
-->
## Constraints

- Must integrate into the existing spec creation steps.

<!--
  ACCEPTANCE CRITERIA
  The specific, binary conditions that define "done".
  Format: bold title on the checkbox line, verifiable detail indented below.
  Each criterion must be:
    - Independently verifiable (pass/fail, not subjective)
    - Traceable back to a requirement above
    - Testable by someone who didn't write the code
-->
## Acceptance Criteria

- [ ] **Validation outputs issues per section** — When the agent runs validation, it outputs a list of issues found per section, with a specific reason for each issue.
- [ ] **Validation confirms passing sections** — When all sections contain sufficient detail, the agent confirms that all sections pass validation.
- [ ] **Agent asks for clarification when needed** — If the agent has questions regarding a spec section, it asks the user for clarification.
- [ ] **Agent skips clarification when unnecessary** — If the agent has no questions, it does not prompt the user for clarification.
- [ ] **Clarification runs to completion** — The clarification loop continues until all questions the agent has are resolved.
- [ ] **Agent researches the codebase for answers** — The agent explores the codebase and available context to attempt to determine answers for areas of uncertainty before or during validation.
- [ ] **Agent validates research assumptions with user** — The agent always validates any assumptions it made from research with the user before accepting them as resolved.

<!--
  TECHNICAL APPROACH
  High-level technical direction to guide the planning agent. Include:
    - Key architectural decisions already made
    - Preferred patterns or technologies if known
    - Integration points with existing systems
    - Known risks or areas of uncertainty
  Leave blank if you want the planner to propose the approach.
-->
## Technical Approach

A higher level model should be used to validate the specification, for example, if haiku is used to generate the spec, then opus should be used to validate.

<!--
  SUCCESS METRICS
  How you will know the feature is working well after delivery. Be specific:
    - Quantitative: "p99 latency < 200ms", "error rate < 0.1%"
    - Behavioural: "users complete the flow without support intervention"
  Leave blank if not applicable.
-->
## Success Metrics

None.

<!--
  NON-GOALS
  Explicitly state what this spec does NOT cover. This is as important as
  the requirements — it prevents scope creep and sets clear expectations.
  Examples:
    - "Mobile support is out of scope (tracked in #456)"
    - "Internationalisation will be addressed in a follow-up spec"
  Leave blank if there are no explicit exclusions to call out.
-->
## Non-Goals

None.
