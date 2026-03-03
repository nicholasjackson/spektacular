# Spec Verify -- Research Notes

## Design Decision: Single Step vs. Per-Section Steps

**Considered**: Adding 7 separate verification steps (one per section), mirroring the collection steps.

**Chosen**: A single verification step that validates all sections in one pass.

**Rationale**:
- A single agent pass has full context of all sections, enabling cross-section validation (e.g., checking that every requirement has a matching acceptance criterion).
- Reduces complexity -- one step, one prompt, one agent invocation.
- The GOTO loop handles iteration naturally when clarification is needed.
- Per-section verification would lose cross-referencing ability and multiply agent invocations.

## Design Decision: Model Selection for Verification

**Spec says**: "A higher level model should be used to validate the specification."

**Options considered**:
1. `claude-opus-4-6` -- Highest capability, used for planning
2. `claude-sonnet-4-6` -- Mid-tier, used for implementation and plan feedback
3. Dynamic: use config `Models.Tiers.Complex`

**Chosen**: `claude-sonnet-4-6` (hardcoded, like all other steps)

**Rationale**:
- Sonnet has sufficient analytical depth for spec validation -- it doesn't need the architectural reasoning depth of Opus.
- Opus is significantly more expensive and slower; validation should be responsive.
- Using Sonnet is "higher level" than Haiku (used for collection), satisfying the spec requirement.
- All existing steps hardcode their model; no dynamic selection infrastructure exists yet.

## Design Decision: Where to Place the Verify Step

**Options considered**:
1. After `OnDone` callback -- Would require changing the `Workflow` interface
2. As a separate command (`spektacular verify <spec>`) -- Spec says "must integrate into existing spec creation steps"
3. As the last step in `SpecCreatorWorkflow` -- Minimal change, follows existing pattern

**Chosen**: Option 3 -- append to `Steps` slice in `SpecCreatorWorkflow`.

**Rationale**: Requires zero changes to the TUI framework or workflow interface. The verification step is just another `WorkflowStep` like the 7 collection steps.

## GOTO Loop Mechanism

The TUI already supports GOTO via `DetectGoto()` in `runner.go:154-160` and `gotoStep()` in `tui.go:177-189`. When the agent outputs `<!-- GOTO: verify -->`, the TUI:

1. Finds the step with `Name == "verify"`
2. Resets questions/answers state
3. Calls `startCurrentStep()` which re-invokes `BuildRunOptions`
4. The session ID is preserved, so the agent has full conversational context

This means the agent can:
- Ask a question, get an answer
- Edit the spec file
- Output `<!-- GOTO: verify -->` to trigger re-validation
- On the next pass, re-read the spec and validate again
- Repeat until satisfied, then output `<!-- FINISHED -->`

## Verification Agent Prompt Design

The prompt needs to balance several concerns:

1. **Section-by-section validation** -- Report on each section individually
2. **Cross-section consistency** -- Check requirements vs. acceptance criteria alignment
3. **Codebase awareness** -- Use tools to explore context
4. **Research with transparency** -- When the agent finds answers in the codebase, it must share those findings with the user for validation
5. **Minimal user disruption** -- Don't ask questions when the spec is already clear
6. **Iterative refinement** -- Loop via GOTO until all issues are resolved

The prompt should explicitly instruct:
- Read the full spec file first
- Explore `.spektacular/knowledge/` and the broader codebase for relevant context
- For each section, evaluate and report: PASS or ISSUE with specific reason
- Group all clarification questions into a single `<!--QUESTION:...-->` block per round
- After receiving answers, edit the spec file, then GOTO verify for re-validation
- If no issues remain, confirm all sections pass and output FINISHED
