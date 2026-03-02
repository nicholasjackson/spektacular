# Feature: 11 Statistics

<!--
  OVERVIEW
  A concise 2-3 sentence summary of the feature. Answer three questions:
    1. What is being built?
    2. What problem does it solve?
    3. Who benefits and why does it matter?
  Avoid implementation details — this should be readable by any stakeholder.
-->
## Overview

At the end of each command (plan, implement, new), a statistics summary is presented showing the resources consumed and operations performed by the agent. This solves the problem of users having no visibility into what the agent actually did during a run — how many API calls were made, tokens used, files read/written, and time elapsed. Users running spektacular commands benefit by being able to understand agent efficiency, diagnose slow runs, and make informed decisions about cost and usage.


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

- [ ] **Display elapsed time** — Users can see the total elapsed wall-clock time for the command when it completes
- [ ] **Display tokens used** — Users can see the total number of tokens consumed during the command session
- [ ] **Display tools called** — Users can see the total number of tool invocations made by the agent during the command session
- [ ] **Session-wide collection** — The system must collect and accumulate statistics across the entire session, not just individual steps
- [ ] **Statistics shown on completion** — The system must present the statistics summary at the end of each command (plan, implement, new)

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

None.

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

- [ ] **Elapsed time format** — When a command completes, the elapsed time from command start to completion is displayed in hours, minutes, and seconds format
- [ ] **Token count displayed** — When a command completes, the total number of tokens used across the session is displayed
- [ ] **Tool calls shown conditionally** — When a command completes and tool calls were made during the session, the total number of tool invocations is displayed. If no tool calls were made, this line is not displayed
- [ ] **Cross-resume accumulation** — Statistics are collected and accumulated across multiple starts and resumes within a single session, not reset on each resume
- [ ] **Tabular output on completion** — When a session completes, all statistics are presented to the user in a tabular format

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

Different coding agents support different JSON output formats, so statistic collection must be built into the runner interface and implemented by each individual runner. The runner interface should define the contract for reporting statistics, and each concrete runner implementation is responsible for parsing its agent's output to extract the relevant metrics.

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
