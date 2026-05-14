# Interactive Spec Creation - Context

## Quick Summary
Add interactive mode to `spektacular new` command that uses an AI agent to guide users through creating well-structured specifications by asking about each section and requesting clarification for vague input. Uses existing TUI framework enhanced with multi-line textarea support.

## Key Files & Locations

### New Files to Create
- **Agent System Prompt**: `internal/defaults/files/agents/spec-creator.md`
- **Interactive Logic**: `internal/spec/interactive.go`
- **TUI Entry Point**: `internal/tui/spec_creator.go`
- **Unit Tests**: `internal/spec/interactive_test.go`, `internal/tui/textarea_test.go`

### Files to Modify
- **Command Entry**: `cmd/new.go` (add TTY detection, `--noninteractive` flag)
- **TUI Core**: `internal/tui/tui.go` (add textarea support, enhance model, update handlers)

### Integration Files
- **Template**: `internal/defaults/files/spec-template.md` (reference for structure)
- **Runner**: `internal/runner/runner.go` (question detection - no changes needed)
- **Config**: `internal/config/config.go` (agent settings - no changes needed)

## Dependencies

### Code Dependencies
- **Internal Modules**:
  - `internal/config` - Configuration loading
  - `internal/defaults` - Embedded file access
  - `internal/runner` - Agent execution and event handling
  - `internal/spec` - Existing spec creation (preserved)
  - `internal/tui` - Bubble Tea TUI framework

- **External Packages**:
  - `github.com/charmbracelet/bubbletea` v1.3.10 - TUI framework
  - `github.com/charmbracelet/bubbles/textarea` v1.0.0 - Multi-line input (NEW usage)
  - `github.com/charmbracelet/bubbles/viewport` v1.0.0 - Scrollable output (existing)
  - `github.com/charmbracelet/glamour` v0.10.0 - Markdown rendering
  - `github.com/charmbracelet/lipgloss` v1.1.1 - Styling
  - `github.com/spf13/cobra` v1.8.1 - CLI framework
  - `golang.org/x/term` - TTY detection

### External Dependencies
- **Claude CLI** - Required for agent execution (same as plan/implement)
- **Configuration** - `.spektacular/config.yaml` with agent settings

### Database Changes
None - File-based system only

## Environment Requirements

### Configuration Variables
No new environment variables required. Uses existing config:
- `agent.command` - Path to Claude CLI (default: "claude")
- `agent.args` - Additional CLI arguments
- `agent.allowed_tools` - Tools available to agent (default: includes Write for spec creation)

### Migration Scripts
None required - Additive feature

### Feature Flags
Mode determined by:
- **TTY Detection**: `term.IsTerminal(os.Stdout.Fd())`
- **`--noninteractive` Flag**: Forces template mode
- **Logic**: Interactive if (TTY detected AND NOT --noninteractive)

## Integration Points

### CLI Commands
- **`spektacular new <name>`** - Interactive mode (NEW)
- **`spektacular new <name> --noninteractive`** - Template mode (current behavior)
- **`spektacular new <name> --title "..." --description "..."`** - Template with custom values (non-interactive only)

### Spec Template Structure
Sections in order (from template):
1. **Title** - Feature: {name}
2. **Overview** - 2-3 sentence summary
3. **Requirements** - Checkbox list with bold titles
4. **Constraints** - Hard boundaries
5. **Acceptance Criteria** - Binary pass/fail conditions
6. **Technical Approach** - Architectural guidance
7. **Success Metrics** - Quantitative/behavioral measures
8. **Non-Goals** - Explicit exclusions

All sections include HTML comment guidance that must be preserved.

### Agent Communication
- **Input**: Initial prompt with spec name
- **Output**: Structured questions via `<!--QUESTION:...-->` markers
- **Session**: Multi-turn conversation using `--resume <session-id>`
- **Completion**: Agent writes spec using Write tool to `.spektacular/specs/{name}.md`

### TUI Workflow
1. User runs `spektacular new my-feature`
2. TUI launches with "Creating specification" status
3. Agent asks about Overview (question 1 of 7)
4. User provides answer via:
   - Numbered option (1-9) for multiple choice
   - Free text option for multi-line textarea input
5. Agent evaluates answer:
   - If detailed → moves to next section
   - If vague → asks clarifying question
6. Repeat for all 7 sections
7. Agent writes final spec file
8. TUI shows success message with file path

### File Outputs
- **Location**: `.spektacular/specs/{name}.md`
- **Format**: Markdown with HTML comment guidance
- **Structure**: Matches `spec-template.md` exactly
- **Validation**: File existence check only (no schema validation)

## Component Architecture

### Agent Layer
```
spec-creator.md (system prompt)
    ↓
LoadInteractiveAgentPrompt()
    ↓
RunInteractive() - multi-turn loop
    ↓
Claude CLI runner (subprocess)
    ↓
Event stream (session_id, text, questions, result)
```

### TUI Layer
```
RunSpecCreatorTUI() - entry point
    ↓
Workflow{Start, OnResult}
    ↓
RunAgentTUI() - generic agent handler
    ↓
model.Update() - Bubble Tea event loop
    ↓
handleTextareaInput() - multi-line input
handleNumberKey() - multiple choice
handleAgentEvent() - process agent output
```

### Command Layer
```
spektacular new <name>
    ↓
TTY detection + flag check
    ↓
Interactive          Non-Interactive
    ↓                    ↓
RunSpecCreatorTUI()  spec.Create() (template)
    ↓                    ↓
.spektacular/specs/{name}.md
```

## Testing Strategy

### Unit Tests
- `TestLoadInteractiveAgentPrompt()` - Agent prompt loads correctly
- `TestTextareaInit()` - Textarea initialization
- `TestTextareaSubmit()` - Ctrl+D submits content
- `TestTextareaCancel()` - Esc discards input
- `TestNewCmd_NonInteractiveFlag()` - Flag parsing

### Integration Tests (Manual)
- **Scenario A**: Detailed responses (no clarification)
- **Scenario B**: Vague responses (with clarification)
- **Scenario C**: Multi-line markdown input
- **Scenario D**: Multiple clarification rounds
- **Scenario E**: `--noninteractive` flag behavior
- **Scenario F**: Ctrl+C graceful exit

### Success Validation
```bash
# Build and test
go build -o spektacular .
go test ./... -v

# Create spec interactively
spektacular new test-feature

# Validate structure
grep -q "## Overview" .spektacular/specs/test-feature.md
grep -q "## Requirements" .spektacular/specs/test-feature.md
# ... all 7 sections

# Verify works with plan command
spektacular plan test-feature.md
```

## Implementation Phases

### Phase 1: Foundation (Agent + Flag)
- Create agent system prompt
- Add `--noninteractive` flag to cmd/new.go
- Implement `internal/spec/interactive.go` with multi-turn loop
- Add prompt loading function
- **Test**: Flag recognized, agent prompt loads

### Phase 2: Enhanced TUI
- Add textarea.Model to TUI model struct
- Implement `handleTextareaInput()` for multi-line submission
- Update `renderQuestionPanel()` to show textarea when active
- Modify `handleNumberKey()` to activate textarea for free-text
- **Test**: Textarea accepts multi-line input, submits on Ctrl+D

### Phase 3: Integration
- Wire interactive mode to TUI in cmd/new.go
- Create `RunSpecCreatorTUI()` entry point
- Add agent event handling for spec creator
- Validate spec file creation
- **Test**: End-to-end flow from command to saved spec

## Performance Considerations

### Agent Response Time
- First response: ~2-3 seconds (agent initialization)
- Follow-up questions: ~1-2 seconds (session resume)
- Total time: ~15-30 seconds for complete spec (7 sections + potential clarifications)

### TUI Responsiveness
- Textarea rendering: <50ms (handled by bubbles component)
- Viewport scrolling: <16ms (60fps target)
- Theme switching: <100ms
- No lag expected with reasonable spec sizes (<10KB)

## Known Limitations

### Current Implementation
1. **No answer editing** - Can't go back to previous sections (Ctrl+C and restart)
2. **No answer preview** - Can't review all answers before submission
3. **No spec validation** - Trusts agent to follow template structure
4. **Requires Claude CLI** - Won't work offline
5. **Single conversation thread** - No branching or parallel section filling

### Future Enhancements
- Answer history navigation (up/down arrows to edit previous sections)
- Spec preview panel showing accumulated content
- Structural validation (check all sections present)
- Offline mode with basic prompts (no clarification)
- Template selection (different spec types)
- Spec editing mode (modify existing specs interactively)

## Rollback Plan

If issues arise after deployment:

1. **Immediate**: Document `--noninteractive` flag workaround
2. **Short-term**: Add `--force-template` as alias (clearer naming)
3. **Medium-term**: Add environment variable `SPEKTACULAR_INTERACTIVE=false` to disable globally
4. **Long-term**: If critical, swap default behavior (add `--interactive` opt-in flag)

**Rollback Command**:
```bash
# For users experiencing issues
export SPEKTACULAR_NEW_TEMPLATE=1  # Future env var
spektacular new my-spec --noninteractive  # Current workaround
```

## References

### Specification
- Original spec: `.spektacular/specs/9_interactive_spec.md`

### Key Patterns
- Agent conversation: `internal/plan/plan.go:117-165`
- TTY detection: `cmd/plan.go:40-60`
- Workflow abstraction: `internal/tui/tui.go:40-53`
- Question rendering: `internal/tui/tui.go:492-528`

### External Documentation
- Bubble Tea: https://github.com/charmbracelet/bubbletea
- Bubbles textarea: https://github.com/charmbracelet/bubbles/tree/master/textarea
- Cobra: https://github.com/spf13/cobra
- Glamour: https://github.com/charmbracelet/glamour
