# Feature: 9 Interactive Spec

<!--
  OVERVIEW
  A concise 2-3 sentence summary of the feature. Answer three questions:
    1. What is being built?
    2. What problem does it solve?
    3. Who benefits and why does it matter?
  Avoid implementation details — this should be readable by any stakeholder.
-->
## Overview
The new command should should have an interactive mode that guides the user 
through the process of creating a new spec. The intention is to make it easier for
a user to create a well-structured spec which in turn leads to a better plan.

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
- [ ] **Interactive Mode**
  The user should be prompted to fill out each section of the spec (Overview, Requirements, Constraints, etc.) 
  with guidance on what to include.
- [ ] **Clarification and Questions**
  The system should ask follow-up questions if the user's input is too vague or incomplete, to help them flesh out the spec.
- [ ] **TUI**
  The user should be able to answer questions uing the interactive TUI, with support for multiline input and markdown formatting.
- [ ] **Save Spec**
  Once all sections are completed, the spec should be saved to the appropriate location.

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
- Must integrate with the existing new command

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
- [ ] **Running in interactive mode creates a new spec**
  When running in interactive mode the user should be prompted to fill out
  the spec.
- [ ] **Running in normal mode creates a new spec using the template**
  When running in normal mode the user should get the template as they do now.
- [ ] **Interactive mode should prompt for clarification**
  If the user provides vague or incomplete input, the system should ask
  follow-up questions to help them flesh out the spec.
- [ ] **Interactive mode should save the users responses**
  If there is no need for clairification the system should just save the users
  responses to the spec file.
- [ ] **Clarification should support multiple rounds**
  If the user continues to provide vague or incomplete input, the system should
  continue to ask follow-up questions until the spec is sufficiently fleshed out.
- [ ] **Detailed responses should not trigger clarification**
  If the user provides detailed responses, the system should not ask for clarification and should save the spec.
- [ ] **TUI should support multiline input and markdown formatting**
  The user should be able to enter multiline responses and use markdown formatting in the TUI. 
- [ ] **The spec should be saved**
  After the user has completed all sections of the spec, it should be saved to the appropriate location.


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
A new agent system prompt should be created for the interactive mode that details it's
behaviour.
The new command should run in interactive mode by default, the flat --noninteractive flag 
can be used to disable it and get the current behaviour.
The LLM should be called in the same way that plan and implement are.

<!--
  SUCCESS METRICS
  How you will know the feature is working well after delivery. Be specific:
    - Quantitative: "p99 latency < 200ms", "error rate < 0.1%"
    - Behavioural: "users complete the flow without support intervention"
  Leave blank if not applicable.
-->
## Success Metrics


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
