# Feature: 12 Bob Runner

<!--
  OVERVIEW
  A concise 2-3 sentence summary of the feature. Answer three questions:
    1. What is being built?
    2. What problem does it solve?
    3. Who benefits and why does it matter?
  Avoid implementation details — this should be readable by any stakeholder.
-->
## Overview

Spektacular should support IBM Bob by wrapping its CLI, enabling users to run Bob commands through Spektacular's interface. This solves the problem of managing Bob workflows separately from other build tools that Spektacular already orchestrates. Developers using IBM Bob alongside other Spektacular-supported tools benefit from a unified experience.


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

- [ ] **Parse Bob CLI output** — The system must parse the JSON-stream output format from the Bob CLI and render it to the terminal
- [ ] **Resume sessions** — Users can resume existing Bob sessions and handle interactive questions from the Bob CLI
- [ ] **Multiple model support** — Users can select from multiple models when running Bob
- [ ] **Runner selection** — Users can select the runner (e.g. Claude, Bob) via the config file

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

- Must integrate with the existing events and runner architecture

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

- [ ] **JSON-stream parsing** — JSON-stream output from the Bob CLI is parsed and can be handled by the TUI
- [ ] **Question detection** — Questions in the Bob CLI output are detected and handled in the TUI
- [ ] **Output completion detection** — The system can determine when Bob CLI output has finished to move to the next steps
- [ ] **GOTO flow control** — The system can handle `<!-- GOTO -->` flow control directives from the Bob CLI output
- [ ] **Message and tool call rendering** — Messages and tool calls from Bob are rendered to the TUI
- [ ] **Session resumption after interactions** — After a `<!-- QUESTION -->` or `<!-- GOTO -->` has been handled, the previous session is resumed
- [ ] **Session ID capture** — Spektacular captures the session ID from Bob if one does not already exist
- [ ] **Session ID reuse** — If a session ID is already present, it is reused when resuming Bob
- [ ] **Model selection** — Spektacular can select the model when interacting with Bob
- [ ] **Model list** — Spektacular maintains a list of models that can be used with Bob
- [ ] **Default model fallback** — When no model is defined, Spektacular uses a default model
- [ ] **Runner config selection** — The config file specifies which runner to use and Spektacular loads the correct runner accordingly

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

- A new runner should be created for the Bob CLI that implements the existing runner interface
- Bob's stream-json event format differs significantly from Claude's. The Bob runner must translate Bob events (`init`, `message` with `delta`, `tool_use`, `tool_result`, `result`) into Spektacular's internal Event structure so the existing `RunSteps` machinery and Event helper methods (`TextContent`, `ToolUses`, `SessionID`) work correctly. Where possible, shared abstractions should be used to satisfy both runners
- The event mapping strategy is documented in `.spektacular/knowledge/architecture/bob-output-spec.md`
- Runner selection should be driven by the config file, using the existing runner registry

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
