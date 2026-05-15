# Spec Verify -- Implementation Plan

## Overview
- **Spec**: `11_spec_verify.md`
- **Complexity**: Medium
- **Dependencies**: existing spec creation workflow (`internal/steps/new.go`), TUI workflow framework (`internal/tui/tui.go`), runner infrastructure (`internal/runner/`), embedded agent prompts (`internal/defaults/files/agents/`)

## Current State

The spec creation workflow (`SpecCreatorWorkflow` in `internal/steps/new.go`) runs 7 sequential steps (overview, requirements, acceptance criteria, constraints, technical approach, success metrics, non-goals), each driven by a Haiku-level model. Each step collects user input for one section and writes it to the spec file. After the final non-goals step, the workflow completes and returns the spec path.

**What's missing:**
- There is no verification/validation step after all sections are collected.
- No agent prompt exists for the verification role.
- The `SpecCreatorWorkflow` currently ends immediately after the non-goals step with no post-collection processing.

**What exists that we can reuse:**
- The `tui.Workflow` / `tui.WorkflowStep` framework already supports multi-step workflows, `<!-- FINISHED -->` detection, `<!-- GOTO: step -->` for looping, and the `<!--QUESTION:...-->` protocol for asking the user questions.
- The `GOTO` mechanism (already used in the plan feedback loop in `internal/steps/plan.go:48-82`) is exactly what we need for the clarification loop.
- The runner infrastructure supports per-step model selection via `RunOptions.Model`.
- The embedded defaults system (`internal/defaults/`) supports adding new agent prompt files.

## Implementation Strategy

Add a single new verification step to the end of `SpecCreatorWorkflow`. This step:

1. Uses a higher-tier model (Opus or Sonnet) to read the completed spec file.
2. The agent explores the codebase for context (using Read, Grep, Glob tools).
3. Validates each section for completeness and clarity.
4. Reports issues per section, or confirms all sections pass.
5. If the agent has questions or found vague areas, it asks the user for clarification using `<!--QUESTION:...-->`.
6. If the agent made assumptions from research, it validates those with the user.
7. The agent edits the spec file with any resolved clarifications.
8. Uses `<!-- GOTO: verify -->` to loop back for another validation pass after clarifications.
9. Once satisfied, outputs `<!-- FINISHED -->`.

This is the same pattern used by the plan feedback step (`feedbackStep` in `internal/steps/plan.go`), where GOTO enables a re-validation loop. The session ID carries forward, so the agent retains full context of what was already discussed.

## Phase 1: Create the Verification Agent Prompt

### Changes Required

- **`internal/defaults/files/agents/verify.md`** (new file)
  - Create an embedded agent system prompt for the spec verification agent.
  - The prompt must instruct the agent to:
    - Read the completed spec file
    - Explore the codebase and `.spektacular/knowledge/` for context relevant to the spec
    - Validate each section (Overview, Requirements, Constraints, Acceptance Criteria, Technical Approach, Success Metrics, Non-Goals) for:
      - Sufficient detail for planning/implementation
      - Clarity and lack of ambiguity
      - Testability (for requirements and acceptance criteria)
      - Consistency between sections (e.g., every requirement has matching acceptance criteria)
    - Output a structured validation report listing issues per section with specific reasons
    - Confirm passing sections explicitly
    - If the agent finds areas of uncertainty, first research the codebase to attempt clarification
    - Always validate any assumptions from research with the user before accepting them
    - Ask the user clarifying questions using the `<!--QUESTION:...-->` format when sections are vague/incomplete
    - After clarification, edit the spec file to incorporate resolved answers
    - Use `<!-- GOTO: verify -->` to re-validate after making changes
    - If no issues remain and no questions are needed, confirm all sections pass and output `<!-- FINISHED -->`
    - Use the same question format rules as the spec creator agent (no `AskUserQuestion` tool, use HTML comment format only)
  - Rationale: Separating the verification prompt keeps concerns isolated and follows the existing pattern of one prompt file per agent role (`spec.md`, `planner.md`, `executor.md`).

### Testing Strategy
- Manual: Verify the file is embedded correctly by checking `defaults.ReadFile("agents/verify.md")` doesn't error.
- Unit: Not directly testable in isolation (it's a markdown prompt file), but the embedding is verified by the existing `defaults_test.go` pattern.

### Success Criteria
- [ ] `internal/defaults/files/agents/verify.md` exists and is well-structured
- [ ] `go build ./...` passes (embed directive picks up the new file automatically)

## Phase 2: Add Verification Step to the Spec Creation Workflow

### Changes Required

- **`internal/spec/spec.go`**
  - Add `LoadVerifyAgentSystemPrompt() string` function that reads the embedded `agents/verify.md` file using `defaults.MustReadFile("agents/verify.md")`.
  - Follows the exact pattern of the existing `LoadAgentSystemPrompt()` at line 172.
  - Rationale: Keeps prompt loading centralized in the spec package.

- **`internal/steps/new.go`**
  - Add a `verifyMsg` user prompt template (similar to existing section prompts like `overviewMsg`). This prompt tells the verification agent:
    - The spec file path to read
    - To explore the codebase for context
    - To validate all sections
    - To report issues per section or confirm all pass
    - To ask clarification questions if needed
    - To use `<!-- GOTO: verify -->` after receiving clarification to re-validate
    - To output `<!-- FINISHED -->` when validation is complete
  - Add a `verifyStep(specPath string) tui.WorkflowStep` function that:
    - Uses `spec.LoadVerifyAgentSystemPrompt()` for the system prompt
    - Uses `runner.BuildPromptWithHeader(fmt.Sprintf(verifyMsg, specPath), "Spec Verification")` for the user prompt
    - Sets `Name: "verify"` (critical: this is the GOTO target for the re-validation loop)
    - Sets `StatusLabel: "verifying spec"`
    - Uses a higher-tier model: `Model: "claude-sonnet-4-6"` (Sonnet for the verification step, upgrading from Haiku used in collection steps. The spec says "higher level model" -- Sonnet is the right balance of quality vs cost for validation)
  - Modify `SpecCreatorWorkflow` to append the verify step after the non-goals step:
    ```go
    Steps: []tui.WorkflowStep{
        overviewStep(specPath),
        requirementsStep(specPath),
        acStep(specPath),
        constraintsStep(specPath),
        technicalApproachStep(specPath),
        successMetricsStep(specPath),
        nonGoalsStep(specPath),
        verifyStep(specPath),  // NEW
    },
    ```
  - Update the `Preamble` string to mention "8 sections" instead of "7 sections" and add "Verification" to the section list.
  - Rationale: Minimal change to existing code -- just adds one step and one helper function. The GOTO loop pattern is already proven in the feedback step.

### Testing Strategy
- Unit: `internal/spec/spec_test.go` -- add a test for `LoadVerifyAgentSystemPrompt()` to verify it returns non-empty content.
- Integration: Manual end-to-end test running `spektacular new test-feature` and verifying the verification step executes after all sections are collected.
- Verification: `go build ./...` and `go test ./...`

### Success Criteria
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes
- [ ] Running `spektacular new <name>` in interactive mode shows 8 steps including verification
- [ ] The verification step uses Sonnet (not Haiku)
- [ ] The verification agent reads the full spec and validates each section
- [ ] The verification agent asks clarification questions when sections are vague
- [ ] The verification agent does NOT ask questions when all sections are clear
- [ ] The clarification loop repeats (via GOTO) until resolved
- [ ] The verification agent explores the codebase before/during validation
- [ ] The verification agent validates research-based assumptions with the user
- [ ] After verification completes, the spec file reflects any changes from clarification

## Not In Scope

- **New CLI command or flag** -- Verification is built into the existing `new` command's interactive workflow, not a separate `verify` command.
- **Non-interactive verification** -- The spec says "must integrate into the existing spec creation steps", so this only applies to interactive mode.
- **Modifying the plan or implement workflows** -- Verification is only for spec creation.
- **Config-based model selection for the verify step** -- We hardcode the model like all other steps do. Future work could make this configurable.

## References

- Key files:
  - `internal/steps/new.go` -- Spec creation workflow and section steps
  - `internal/spec/spec.go:172-174` -- `LoadAgentSystemPrompt()` pattern
  - `internal/steps/plan.go:48-82` -- Feedback/GOTO loop pattern
  - `internal/tui/tui.go:42-56` -- `WorkflowStep` and `Workflow` types
  - `internal/runner/runner.go:144-160` -- `DetectFinished`, `DetectGoto` markers
  - `internal/defaults/files/agents/spec.md` -- Existing spec agent prompt (question format rules)
  - `internal/defaults/files/agents/planner.md` -- Planner agent prompt (research workflow reference)
  - `internal/defaults/defaults.go` -- Embedded file system
- Related patterns:
  - The `feedbackStep` in `internal/steps/plan.go` demonstrates the exact GOTO loop pattern needed for iterative clarification
  - All existing steps in `new.go` demonstrate the `WorkflowStep` builder pattern with model selection
