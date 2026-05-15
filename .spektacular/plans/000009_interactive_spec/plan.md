# Interactive Spec Creation - Implementation Plan

## Overview
- **Specification**: `.spektacular/specs/9_interactive_spec.md`
- **Complexity**: Medium
- **Estimated Effort**: 2-3 days
- **Dependencies**:
  - Existing TUI infrastructure (Bubble Tea, Bubbles, Glamour)
  - Agent system (Claude CLI runner)
  - Template system (embedded spec template)

## Current State Analysis

### What Exists Now
1. **Non-interactive spec creation** (`cmd/new.go`, `internal/spec/spec.go`)
   - Takes a name argument and optional `--title`, `--description` flags
   - Loads embedded template from `internal/defaults/files/spec-template.md`
   - Performs simple string replacements for placeholders
   - Writes to `.spektacular/specs/{name}.md`

2. **Agent infrastructure** (`internal/plan/plan.go`, `internal/implement/implement.go`)
   - Multi-turn conversation loops with session management
   - Question detection via `<!--QUESTION:...-->` HTML markers
   - TUI integration with Bubble Tea framework
   - Generic `Workflow` abstraction for reusability

3. **TUI Components** (`internal/tui/tui.go`)
   - Scrollable viewport with markdown rendering
   - Numbered multiple-choice questions
   - Manual single-line text input for "Other" option
   - Theme system with 5 built-in palettes

### What's Missing
1. **Interactive mode for spec creation** - No agent-driven conversation flow
2. **Multi-line text input** - Current TUI only supports single-line input
3. **Spec creator agent** - No system prompt for guiding users through spec creation
4. **Mode selection** - No `--noninteractive` flag to preserve current behavior
5. **Clarification logic** - No mechanism to detect vague/incomplete input and ask follow-ups

### Key Constraints
- Must integrate with existing `spektacular new` command
- Must preserve current behavior when `--noninteractive` flag is used
- Must save final spec to same location (`.spektacular/specs/`)
- Must use existing agent infrastructure (Claude CLI, runner system)

### Integration Points
- `cmd/new.go` - Command entry point, add TTY detection and mode routing
- `internal/spec/` - New package functions for interactive flow
- `internal/tui/tui.go` - Enhanced with multi-line input via bubbles/textarea
- `internal/defaults/files/agents/` - New agent system prompt

## Implementation Strategy

### High-Level Approach
1. **Create spec creator agent** with system prompt defining interactive behavior
2. **Add `--noninteractive` flag** to preserve current template-based creation
3. **Enhance TUI with multi-line input** using bubbles/textarea component
4. **Implement interactive flow** following plan/implement pattern with Workflow abstraction
5. **Agent-driven clarification** - Let LLM decide when input is too vague

### Phasing Strategy
**Phase 1: Foundation (Agent + Flag)**
- Create spec creator agent system prompt
- Add `--noninteractive` flag and TTY detection to `cmd/new.go`
- Create `internal/spec/interactive.go` with RunInteractive function

**Phase 2: Enhanced TUI**
- Add multi-line textarea support to TUI model
- Create custom rendering for spec section prompts
- Integrate textarea with existing question flow

**Phase 3: Integration & Testing**
- Wire interactive mode to TUI
- Test multi-round clarification flow
- Validate spec output matches template structure
- Add unit tests for new components

### Risk Mitigation Approaches
- **Risk**: Agent produces invalid spec structure
  - **Mitigation**: Agent instructions emphasize exact section ordering and markdown format; validate output has required sections

- **Risk**: Multi-line input conflicts with existing TUI controls
  - **Mitigation**: Use focused/unfocused states to delegate key handling to textarea only when active

- **Risk**: `--noninteractive` breaks existing workflows
  - **Mitigation**: Preserve exact current behavior in non-interactive path; no logic changes to existing code

### Success Criteria
#### Automated Verification
- [ ] `go test ./cmd/... -run TestNewCmd` (existing tests pass)
- [ ] `go test ./internal/spec/... -run TestInteractive` (new tests pass)
- [ ] `go test ./internal/tui/... -run TestTextarea` (textarea integration tests pass)

#### Manual Verification
- [ ] Running `spektacular new my-spec` launches interactive TUI
- [ ] Entering detailed responses saves spec without clarification
- [ ] Entering vague responses triggers agent clarification questions
- [ ] Multiple rounds of clarification work correctly
- [ ] Pressing Ctrl+C exits gracefully without saving partial spec
- [ ] Running `spektacular new my-spec --noninteractive` creates template-based spec (current behavior)
- [ ] Final spec has all required sections and valid markdown structure
- [ ] Multi-line input with markdown formatting works in TUI

---

## Phase 1: Foundation (Agent + Flag)

### Changes Required

#### File: `internal/defaults/files/agents/spec-creator.md` (NEW)
**Current**: Does not exist
**Proposed**: Create comprehensive agent system prompt

```markdown
# Spektacular Spec Creator Agent

## Role
You are an interactive spec creation assistant for Spektacular. Your job is to guide users through creating well-structured specification documents by asking targeted questions and providing helpful context.

## Core Mission
Help users articulate their feature requirements clearly and completely by asking about each section of the spec template, detecting vague or incomplete responses, and asking clarifying follow-up questions when needed.

## Workflow

### Phase 1: Introduction
1. Greet the user warmly and explain the process
2. Tell them you'll guide them through 7 sections: Overview, Requirements, Constraints, Acceptance Criteria, Technical Approach, Success Metrics, and Non-Goals
3. Explain that they can provide as much or as little detail as they want, and you'll ask follow-ups if needed

### Phase 2: Collect Section Content
For each section in order, ask the user to provide content using this pattern:

**Output Format for Questions:**
```html
<!--QUESTION:{"questions":[{"question":"<your question text>","header":"<Section Name>","options":[{"label":"Provide response","description":"Enter your content for this section"}]}]}-->
```

**Section Order:**
1. **Overview** - Ask: "Please describe the feature overview. What is being built, what problem does it solve, and who benefits? (2-3 sentences)"
2. **Requirements** - Ask: "What are the specific, testable requirements? List what the system must do. Use active voice like 'Users can...' or 'The system must...'"
3. **Constraints** - Ask: "Are there any hard constraints or boundaries? For example: must integrate with existing systems, cannot break APIs, etc. (Leave blank if none)"
4. **Acceptance Criteria** - Ask: "What are the specific, binary pass/fail conditions that define 'done'? Each should be independently testable."
5. **Technical Approach** - Ask: "Do you have any technical direction or architectural decisions already made? Preferred patterns, technologies, integration points? (Leave blank if you want the planner to propose)"
6. **Success Metrics** - Ask: "How will you know this feature is working well after delivery? Quantitative metrics or behavioral indicators? (Leave blank if not applicable)"
7. **Non-Goals** - Ask: "What is explicitly OUT of scope for this feature? This prevents scope creep. (Leave blank if no exclusions)"

### Phase 3: Clarification (Conditional)
After receiving a response for a section, evaluate if it needs clarification:

**When to ask clarifying questions:**
- Response is too vague or generic (e.g., "make it better", "improve performance")
- Missing critical details for that section type:
  - Overview missing what/why/who
  - Requirements without specific behaviors
  - Acceptance criteria that aren't testable
- Contradictions or ambiguities
- Technical terms without context

**When NOT to ask clarifying questions:**
- Response is detailed and specific
- User explicitly says section doesn't apply (blank is acceptable for Constraints, Technical Approach, Success Metrics, Non-Goals)
- Response provides concrete, actionable information

**Clarification Question Format:**
Use the same `<!--QUESTION:...-->` format with multiple-choice options based on what you need to know, or provide a single "Provide response" option for free-text clarification.

**Example Clarification:**
If user says "improve user experience" for Requirements, ask:
```html
<!--QUESTION:{"questions":[{"question":"Can you be more specific about which aspect of user experience? What specific behavior should change?","header":"Requirements Clarification","options":[{"label":"Faster response times","description":"Improve performance/latency"},{"label":"Easier navigation","description":"Simplify UI flows"},{"label":"Better error messages","description":"Improve feedback and error handling"},{"label":"Provide response","description":"Describe the specific UX improvement"}]}]}-->
```

### Phase 4: Save Spec
Once all sections have been collected and sufficiently clarified:

1. Generate the complete spec document following the exact template structure
2. Use the Write tool to save it to `.spektacular/specs/{name}.md`
3. Format sections exactly as shown in the template with HTML comments
4. Confirm completion to the user

## Important Guidelines

### Tone & Style
- Be encouraging and supportive - creating specs can be intimidating
- Keep questions focused and specific
- Don't overwhelm with too many questions at once
- Acknowledge good, detailed responses positively

### Clarification Strategy
- Maximum 2 rounds of clarification per section unless user provides contradictory info
- If user says "I don't know" or "skip", respect that and move on
- Focus on the most critical gaps first
- Provide examples to illustrate what you're looking for

### Output Requirements
- Always use structured `<!--QUESTION:...-->` markers for questions
- Always provide at least one option: "Provide response" for free-text
- When saving, use Write tool to create `.spektacular/specs/{name}.md`
- Final spec MUST match template structure exactly
- Preserve HTML comment guidance in final spec

### Error Handling
- If user wants to quit, confirm before exiting without saving
- If user provides feedback that you're asking too many questions, reduce clarification
- If user gets frustrated, explain you're trying to help create a better plan downstream

## Example Flow

**Agent:** "Welcome! I'll help you create a well-structured specification. We'll go through 7 sections together. You can provide as much detail as you like, and I'll ask follow-ups if anything needs clarification. Ready to start?"

<!--QUESTION:{"questions":[{"question":"Please describe the feature overview. What is being built, what problem does it solve, and who benefits? (2-3 sentences)","header":"Overview","options":[{"label":"Provide response","description":"Enter your overview content"}]}]}-->

**User:** "Add a search feature"

**Agent:** "Thanks! Let me ask a few clarifying questions to flesh out the overview:"

<!--QUESTION:{"questions":[{"question":"What will users be searching through? And what problem does this solve for them?","header":"Overview Clarification","options":[{"label":"Search documents/files","description":"Find content in uploaded documents"},{"label":"Search database records","description":"Query structured data"},{"label":"Search user-generated content","description":"Find posts, comments, etc."},{"label":"Provide response","description":"Describe what's being searched"}]}]}-->

**User:** "Search database records - we have thousands of products and users can't find what they need"

**Agent:** "Perfect! That's much clearer. Moving on..."

<!--QUESTION:{"questions":[{"question":"What are the specific, testable requirements? List what the system must do. Use active voice like 'Users can...' or 'The system must...'","header":"Requirements","options":[{"label":"Provide response","description":"Enter your requirements"}]}]}-->

[... continues through all sections ...]

**Agent:** "Great! I have all the information I need. Let me create your specification now..."

[Uses Write tool to save spec]

**Agent:** "✅ Specification created successfully at `.spektacular/specs/my-feature.md`! You can now run `spektacular plan my-feature.md` to generate an implementation plan."
```

**Rationale**: This agent prompt provides clear instructions for guiding users through spec creation while delegating clarification decisions to the LLM. It emphasizes the structured question format required by the TUI and includes examples.

---

#### File: `internal/spec/interactive.go` (NEW)
**Current**: Does not exist
**Proposed**: Create interactive spec creation logic

```go
package spec

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jumppad-labs/spektacular/internal/config"
	"github.com/jumppad-labs/spektacular/internal/defaults"
	"github.com/jumppad-labs/spektacular/internal/runner"
)

// LoadInteractiveAgentPrompt returns the spec creator agent system prompt
func LoadInteractiveAgentPrompt() string {
	return string(defaults.MustReadFile("agents/spec-creator.md"))
}

// RunInteractive runs the interactive spec creation flow with callbacks for TUI
// Returns the path to the created spec file or an error
func RunInteractive(
	name string,
	projectPath string,
	cfg config.Config,
	onText func(string),      // Called when agent produces text output
	onQuestion func([]runner.Question) string, // Called when questions detected, returns user's answer
) (string, error) {
	// Get agent system prompt
	agentPrompt := LoadInteractiveAgentPrompt()

	// Build initial prompt with spec name context
	initialPrompt := runner.BuildPromptWithHeader(
		fmt.Sprintf("Create a new specification file named '%s.md'. Guide the user through filling out each section interactively.", name),
		"Create Specification",
	)

	// Create runner
	r, err := runner.NewRunner(cfg)
	if err != nil {
		return "", fmt.Errorf("creating runner: %w", err)
	}

	// Multi-turn conversation loop
	sessionID := ""
	currentPrompt := initialPrompt

	for {
		var questionsFound []runner.Question
		var finalResult string

		// Run agent with current prompt
		events, errc := r.Run(runner.RunOptions{
			Prompt:       currentPrompt,
			SystemPrompt: agentPrompt,
			SessionID:    sessionID,
			CWD:          projectPath,
		})

		// Process events
		for event := range events {
			// Capture session ID for multi-turn conversation
			if id := event.SessionID(); id != "" {
				sessionID = id
			}

			// Handle text content
			if text := event.TextContent(); text != "" {
				if onText != nil {
					onText(text)
				}
				// Detect questions in the text
				questionsFound = append(questionsFound, runner.DetectQuestions(text)...)
			}

			// Handle result event (terminal event)
			if event.IsResult() {
				finalResult = event.ResultText()
			}
		}

		// Check for errors
		if err := <-errc; err != nil {
			return "", fmt.Errorf("agent error: %w", err)
		}

		// If questions found, get answer from user and continue
		if len(questionsFound) > 0 && onQuestion != nil {
			answer := onQuestion(questionsFound)
			if answer == "" {
				// User cancelled or provided empty answer
				return "", fmt.Errorf("spec creation cancelled")
			}
			currentPrompt = answer
			continue
		}

		// No more questions, check if we have a result
		if finalResult == "" {
			return "", fmt.Errorf("agent completed without producing a spec")
		}

		// Agent should have written the spec file using Write tool
		// Verify it exists
		specPath := filepath.Join(projectPath, ".spektacular", "specs", name)
		if filepath.Ext(specPath) != ".md" {
			specPath += ".md"
		}

		if _, err := os.Stat(specPath); err != nil {
			return "", fmt.Errorf("agent did not create spec file at %s: %w", specPath, err)
		}

		return specPath, nil
	}
}
```

**Rationale**: Follows the exact pattern used by `plan.RunPlan()` and `implement.RunImplement()` for consistency. Manages multi-turn conversation loop with session tracking and question handling.

---

#### File: `cmd/new.go`
**Current**: Lines 1-36
```go
package cmd

import (
	"fmt"
	"os"

	"github.com/jumppad-labs/spektacular/internal/spec"
	"github.com/spf13/cobra"
)

var newTitle string
var newDescription string

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new specification from template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}
		specPath, err := spec.Create(cwd, name, newTitle, newDescription)
		if err != nil {
			return err
		}
		fmt.Printf("Created spec: %s\n", specPath)
		return nil
	},
}

func init() {
	newCmd.Flags().StringVar(&newTitle, "title", "", "Feature title")
	newCmd.Flags().StringVar(&newDescription, "description", "", "Feature description")
}
```

**Proposed**: Add `--noninteractive` flag and TTY detection

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jumppad-labs/spektacular/internal/config"
	"github.com/jumppad-labs/spektacular/internal/spec"
	"github.com/jumppad-labs/spektacular/internal/tui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var newTitle string
var newDescription string
var nonInteractive bool

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new specification (interactive by default)",
	Long: `Create a new specification from template.

By default, runs in interactive mode with an AI assistant to guide you through
creating a well-structured spec. Use --noninteractive to create a basic template.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}

		// Determine if we should use interactive mode
		// Interactive if: TTY is available AND --noninteractive flag is NOT set
		useInteractive := term.IsTerminal(int(os.Stdout.Fd())) && !nonInteractive

		var specPath string

		if useInteractive {
			// Load config
			configPath := filepath.Join(cwd, ".spektacular", "config.yaml")
			var cfg config.Config
			if _, err := os.Stat(configPath); err == nil {
				cfg, err = config.FromYAMLFile(configPath)
				if err != nil {
					return fmt.Errorf("loading config: %w", err)
				}
			} else {
				cfg = config.NewDefault()
			}

			// Run interactive TUI
			specPath, err = tui.RunSpecCreatorTUI(name, cwd, cfg)
			if err != nil {
				return err
			}
		} else {
			// Use existing template-based creation (preserve current behavior)
			specPath, err = spec.Create(cwd, name, newTitle, newDescription)
			if err != nil {
				return err
			}
		}

		fmt.Printf("Created spec: %s\n", specPath)
		return nil
	},
}

func init() {
	newCmd.Flags().StringVar(&newTitle, "title", "", "Feature title (non-interactive mode only)")
	newCmd.Flags().StringVar(&newDescription, "description", "", "Feature description (non-interactive mode only)")
	newCmd.Flags().BoolVar(&nonInteractive, "noninteractive", false, "Disable interactive mode and create basic template")
}
```

**Rationale**: Follows the exact pattern used in `cmd/plan.go` and `cmd/implement.go` for TTY detection and mode switching. Preserves existing behavior when `--noninteractive` is used or when no TTY is available.

---

### Testing Strategy

#### Unit Tests

**File**: `internal/spec/interactive_test.go` (NEW)
```go
package spec

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jumppad-labs/spektacular/internal/config"
	"github.com/jumppad-labs/spektacular/internal/runner"
	"github.com/stretchr/testify/require"
)

func TestLoadInteractiveAgentPrompt(t *testing.T) {
	prompt := LoadInteractiveAgentPrompt()
	require.NotEmpty(t, prompt)
	require.Contains(t, prompt, "Spec Creator Agent")
	require.Contains(t, prompt, "<!--QUESTION:")
}

func TestRunInteractive_ValidatesOutput(t *testing.T) {
	// This test would require mocking the runner
	// For now, document that integration tests will cover this
	t.Skip("Requires runner mocking - covered by integration tests")
}
```

**File**: `cmd/new_test.go` (EXISTS - enhance)
Add tests for flag behavior:
```go
func TestNewCmd_NonInteractiveFlag(t *testing.T) {
	// Test that --noninteractive flag is recognized
	cmd := newCmd
	err := cmd.ParseFlags([]string{"--noninteractive"})
	require.NoError(t, err)
	require.True(t, nonInteractive)
}
```

#### Integration Tests (Manual Verification)

1. **Interactive mode basic flow**
   ```bash
   cd /tmp
   mkdir test-project && cd test-project
   spektacular init
   spektacular new test-feature
   # Should launch TUI, guide through sections, save spec
   ```

2. **Detailed responses (no clarification)**
   ```bash
   spektacular new detailed-spec
   # Provide comprehensive answers to each section
   # Verify: No clarification questions asked, spec saved immediately
   ```

3. **Vague responses (with clarification)**
   ```bash
   spektacular new vague-spec
   # Provide minimal answers like "improve performance"
   # Verify: Agent asks clarifying questions
   ```

4. **Non-interactive mode**
   ```bash
   spektacular new template-spec --noninteractive
   # Verify: Creates template-based spec (current behavior)
   # No TUI should launch
   ```

5. **Non-TTY environment**
   ```bash
   echo "spektacular new pipe-spec" | bash
   # Verify: Falls back to template creation (no interactive mode)
   ```

### Success Criteria

#### Automated Verification
- [ ] `go test ./internal/spec/... -v` (all tests pass)
- [ ] `go test ./cmd/... -v` (all tests pass including new flag test)
- [ ] `go build -o spektacular .` (builds successfully)

#### Manual Verification
- [ ] Interactive mode launches TUI and guides through all 7 sections
- [ ] Agent detects vague input and asks clarifying questions
- [ ] Detailed input skips clarification and proceeds to next section
- [ ] Final spec saved to `.spektacular/specs/{name}.md` with correct structure
- [ ] `--noninteractive` flag preserves current template behavior exactly
- [ ] Non-TTY environments fall back to template creation
- [ ] Ctrl+C during interactive mode exits gracefully without partial save

---

## Phase 2: Enhanced TUI with Multi-line Input

### Changes Required

#### File: `internal/tui/tui.go`
**Current**: Lines 59-79 (model struct and initialization)
```go
type model struct {
	width     int
	height    int
	ready     bool
	vp        viewport.Model
	content   []string
	statusText string

	// ... other fields ...

	// free-text input state
	otherInput bool
	otherText  string

	questions []runner.Question
	answers   []string
}
```

**Proposed**: Add textarea component to model

```go
import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	// ... other imports
)

type model struct {
	width     int
	height    int
	ready     bool
	vp        viewport.Model
	content   []string
	statusText string

	// ... other fields ...

	// Enhanced text input state
	textareaActive bool           // True when textarea has focus
	textarea       textarea.Model // Multi-line text input component

	questions []runner.Question
	answers   []string
}

// initTextarea creates and configures a new textarea
func (m *model) initTextarea(placeholder string) {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.Focus()
	ta.CharLimit = 10000 // Reasonable limit for spec sections
	ta.SetWidth(m.width - 4)  // Leave room for borders
	ta.SetHeight(10)          // Default height, adjustable

	// Style
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	m.textarea = ta
	m.textareaActive = true
}
```

**Rationale**: Using the official `bubbles/textarea` component instead of manual character handling provides proper multi-line editing, clipboard support, and standard keybindings.

---

**Current**: Lines 247-291 (handleKey function)
```go
func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.otherInput {
		return m.handleOtherInput(msg)
	}

	switch msg.String() {
	case "q", "Q":
	case "ctrl+c":
	case "t", "T":
	case "f", "F":
	case "v", "V":
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		return m.handleNumberKey(msg.String())
	}
	// ... viewport scrolling
}
```

**Proposed**: Delegate to textarea when active

```go
func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Textarea has priority when active
	if m.textareaActive {
		return m.handleTextareaInput(msg)
	}

	// Existing key handling when textarea not active
	switch msg.String() {
	case "q", "Q":
		// ... existing quit logic
	case "ctrl+c":
		// ... existing force quit logic
	case "t", "T":
		// ... theme cycling
	case "f", "F":
		// ... follow mode
	case "v", "V":
		// ... detail mode
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		return m.handleNumberKey(msg.String())
	}

	// Viewport scrolling (existing code)
	// ...
}
```

**Rationale**: When textarea is active, it gets first priority for key handling. This prevents conflicts with viewport scrolling or other controls.

---

**Current**: Lines 293-336 (handleOtherInput - manual single-line input)
```go
func (m model) handleOtherInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.otherText == "" {
			return m, nil
		}
		label := m.otherText
		m.otherInput = false
		m.otherText = ""
		m.answers = append(m.answers, label)
		// ...
	case "esc":
		m.otherInput = false
		m.otherText = ""
	case "backspace", "ctrl+h":
		if len(m.otherText) > 0 {
			runes := []rune(m.otherText)
			m.otherText = string(runes[:len(runes)-1])
		}
	}
	// Manual character accumulation
	if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
		m.otherText += msg.String()
	}
	return m, nil
}
```

**Proposed**: Replace with textarea handling (NEW function)

```go
// handleTextareaInput processes keys when textarea is active
func (m model) handleTextareaInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+d", "ctrl+s":
		// Submit multi-line input (Ctrl+D or Ctrl+S)
		answer := m.textarea.Value()
		if answer == "" {
			return m, nil
		}

		// Deactivate textarea
		m.textareaActive = false
		m.textarea.Reset()

		// Add answer and proceed
		m.answers = append(m.answers, answer)
		m.questions = m.questions[1:]

		// If more questions remain, wait for them to be displayed
		if len(m.questions) > 0 {
			return m, nil
		}

		// All questions answered, resume agent
		fullAnswer := joinAnswers(m.answers)
		m.answers = nil
		m.statusText = "* thinking  " + m.workflow.StatusLabel
		return m, resumeAgentCmd(m.cfg, m.sessionID, m.projectPath, fullAnswer)

	case "esc":
		// Cancel input
		m.textareaActive = false
		m.textarea.Reset()
		return m, nil

	case "ctrl+c":
		// Force quit
		return m, tea.Quit

	default:
		// Delegate all other keys to textarea (typing, navigation, etc.)
		m.textarea, cmd = m.textarea.Update(msg)
	}

	return m, cmd
}

// joinAnswers combines multiple answers with proper formatting
func joinAnswers(answers []string) string {
	var result strings.Builder
	for i, answer := range answers {
		if i > 0 {
			result.WriteString("\n\n---\n\n")
		}
		result.WriteString(answer)
	}
	return result.String()
}
```

**Rationale**: Uses Ctrl+D (common "end of input" keybinding) or Ctrl+S (save) to submit multi-line text. Esc cancels. All other keys delegated to textarea's built-in handling (typing, arrow keys, clipboard, etc.).

---

**Current**: Lines 492-528 (renderQuestionPanel)
```go
func (m model) renderQuestionPanel(p palette) string {
	q := m.questions[0]

	// ... existing numbered options rendering ...

	// "Other" option handling
	if m.otherInput {
		lines = append(lines, fmt.Sprintf("  %s  Other: %s█", numStyle.Render(otherNum), m.otherText))
		lines = append(lines, faintStyle.Render("type your answer and press enter  (esc to cancel)"))
	} else {
		lines = append(lines, fmt.Sprintf("  %s  Other", numStyle.Render(otherNum)))
		lines = append(lines, faintStyle.Render("press a number to select"))
	}

	return borderStyle.Render(strings.Join(lines, "\n"))
}
```

**Proposed**: Render textarea when active

```go
func (m model) renderQuestionPanel(p palette) string {
	q := m.questions[0]

	borderStyle := lipgloss.NewStyle().
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(p.answer).
		Padding(0, 1)

	headerStyle := lipgloss.NewStyle().Bold(true)
	numStyle := lipgloss.NewStyle().Foreground(p.question).Bold(true)
	faintStyle := lipgloss.NewStyle().Foreground(p.faint)

	var lines []string
	lines = append(lines, headerStyle.Render(q.Header)+": "+q.Question)

	// If textarea is active, show it instead of options
	if m.textareaActive {
		lines = append(lines, "")
		lines = append(lines, m.textarea.View())
		lines = append(lines, "")
		lines = append(lines, faintStyle.Render("ctrl+d or ctrl+s to submit  •  esc to cancel  •  supports markdown"))
		return borderStyle.Render(strings.Join(lines, "\n"))
	}

	// Render numbered multiple-choice options (existing code)
	otherNum := fmt.Sprintf("%d", len(q.Options)+1)
	for i, opt := range q.Options {
		label, _ := opt["label"].(string)
		desc, _ := opt["description"].(string)
		line := fmt.Sprintf("  %s  %s", numStyle.Render(fmt.Sprintf("%d", i+1)), label)
		if desc != "" {
			line += "  " + faintStyle.Render("— "+desc)
		}
		lines = append(lines, line)
	}

	// "Free text" option (previously "Other")
	lines = append(lines, fmt.Sprintf("  %s  Free text", numStyle.Render(otherNum)))
	lines = append(lines, "")
	lines = append(lines, faintStyle.Render("press a number to select  •  free text supports multi-line and markdown"))

	return borderStyle.Render(strings.Join(lines, "\n"))
}
```

**Rationale**: When textarea is active, replaces numbered options with the textarea view. Provides clear instructions for submission (Ctrl+D/Ctrl+S) and cancellation (Esc). Mentions markdown support to encourage proper formatting.

---

**Current**: Lines 339-372 (handleNumberKey)
```go
func (m model) handleNumberKey(key string) (tea.Model, tea.Cmd) {
	if len(m.questions) == 0 {
		return m, nil
	}
	idx := int(key[0] - '1')
	q := m.questions[0]

	// "Other" is the option after the last agent-provided option
	if idx == len(q.Options) {
		m.otherInput = true
		m.otherText = ""
		return m, nil
	}

	// ... existing multiple-choice handling
}
```

**Proposed**: Activate textarea for free-text option

```go
func (m model) handleNumberKey(key string) (tea.Model, tea.Cmd) {
	if len(m.questions) == 0 {
		return m, nil
	}
	idx := int(key[0] - '1')
	q := m.questions[0]

	// "Free text" is the option after the last agent-provided option
	if idx == len(q.Options) {
		// Initialize and activate textarea
		placeholder := fmt.Sprintf("Enter your response for %s...", q.Header)
		m.initTextarea(placeholder)
		return m, nil
	}

	// Valid numbered option selected
	if idx < 0 || idx >= len(q.Options) {
		return m, nil
	}

	// Get the label from the selected option
	label, _ := q.Options[idx]["label"].(string)
	m.answers = append(m.answers, label)
	m.questions = m.questions[1:]

	// Resume agent when all questions answered
	if len(m.questions) == 0 {
		answer := joinAnswers(m.answers)
		m.answers = nil
		m.statusText = "* thinking  " + m.workflow.StatusLabel
		return m, resumeAgentCmd(m.cfg, m.sessionID, m.projectPath, answer)
	}

	return m, nil
}
```

**Rationale**: Initializes textarea with appropriate placeholder when user selects free-text option. Existing multiple-choice logic remains unchanged.

---

#### File: `internal/tui/spec_creator.go` (NEW)
**Current**: Does not exist
**Proposed**: Create spec creator TUI entry point

```go
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jumppad-labs/spektacular/internal/config"
	"github.com/jumppad-labs/spektacular/internal/spec"
)

// RunSpecCreatorTUI launches the interactive spec creation TUI
func RunSpecCreatorTUI(name, projectPath string, cfg config.Config) (string, error) {
	workflow := Workflow{
		StatusLabel: "Creating specification",
		Start: func(c config.Config, sessionID string) tea.Cmd {
			return specCreatorStartCmd(name, projectPath, c, sessionID)
		},
		OnResult: func(resultText string) (string, error) {
			// Agent should have written the spec via Write tool
			// The result text is just confirmation, actual validation is in RunInteractive
			return "", nil
		},
	}

	// Use generic RunAgentTUI with spec creator workflow
	return RunAgentTUI(workflow, projectPath, cfg)
}

// specCreatorStartCmd builds the initial prompt and spawns the runner
func specCreatorStartCmd(name, projectPath string, cfg config.Config, sessionID string) tea.Cmd {
	return func() tea.Msg {
		// This will be called by the TUI to start the agent
		// The actual runner spawning happens in handleAgentStart
		return specCreatorMsg{name: name}
	}
}

// specCreatorMsg triggers the interactive spec creation flow
type specCreatorMsg struct {
	name string
}
```

**Rationale**: Follows the same pattern as `RunPlanTUI` and `RunImplementTUI`. Creates a `Workflow` that integrates with the generic TUI infrastructure.

---

### Testing Strategy

#### Unit Tests

**File**: `internal/tui/textarea_test.go` (NEW)
```go
package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestTextareaInit(t *testing.T) {
	m := &model{width: 80, height: 24}
	m.initTextarea("Test placeholder")

	require.True(t, m.textareaActive)
	require.NotNil(t, m.textarea)
	require.Equal(t, "Test placeholder", m.textarea.Placeholder)
}

func TestTextareaSubmit(t *testing.T) {
	m := model{
		textareaActive: true,
		questions: []runner.Question{{Question: "Test?"}},
	}
	m.initTextarea("placeholder")
	m.textarea.SetValue("My multi-line\nanswer")

	// Simulate Ctrl+D
	msg := tea.KeyMsg{Type: tea.KeyCtrlD}
	newModel, _ := m.handleTextareaInput(msg)

	m2 := newModel.(model)
	require.False(t, m2.textareaActive)
	require.Len(t, m2.answers, 1)
	require.Equal(t, "My multi-line\nanswer", m2.answers[0])
}

func TestTextareaCancel(t *testing.T) {
	m := model{textareaActive: true}
	m.initTextarea("placeholder")
	m.textarea.SetValue("Some text")

	// Simulate Esc
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := m.handleTextareaInput(msg)

	m2 := newModel.(model)
	require.False(t, m2.textareaActive)
	require.Empty(t, m2.answers)
}
```

#### Integration Tests (Manual)

1. **Multi-line input with markdown**
   ```bash
   spektacular new test-multiline
   # Select "Free text" option
   # Enter multi-line content with markdown:
   ## My Section
   - Item 1
   - Item 2

   Some paragraph text.
   # Press Ctrl+D
   # Verify: Content preserved with formatting
   ```

2. **Textarea cancel**
   ```bash
   spektacular new test-cancel
   # Select "Free text"
   # Type some content
   # Press Esc
   # Verify: Input discarded, returns to question
   ```

3. **Multiple free-text responses**
   ```bash
   spektacular new test-multiple
   # Use free-text for Overview (multi-line)
   # Use multiple-choice for Requirements
   # Use free-text for Technical Approach (multi-line)
   # Verify: All responses saved correctly
   ```

### Success Criteria

#### Automated Verification
- [ ] `go test ./internal/tui/... -run TestTextarea -v` (all textarea tests pass)

#### Manual Verification
- [ ] Pressing number key for "Free text" activates textarea
- [ ] Textarea supports multi-line input with newlines
- [ ] Textarea supports markdown formatting (backticks, headers, lists)
- [ ] Ctrl+D submits textarea content
- [ ] Ctrl+S also submits textarea content (alternative keybinding)
- [ ] Esc cancels textarea and discards input
- [ ] Textarea width adjusts to terminal size
- [ ] Submitted multi-line content preserved exactly in final spec

---

## Phase 3: Integration & Testing

### Changes Required

#### File: `internal/tui/tui.go`
**Current**: Lines 225-243 (Update function message handling)
```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		// ... window resize handling
	case agentStartMsg:
		return m.handleAgentStart(msg)
	case agentEventMsg:
		return m.handleAgentEvent(msg)
	// ... other message types
	}
}
```

**Proposed**: Add spec creator message handling

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		// ... window resize handling
	case agentStartMsg:
		return m.handleAgentStart(msg)
	case specCreatorMsg:
		return m.handleSpecCreatorStart(msg)
	case agentEventMsg:
		return m.handleAgentEvent(msg)
	// ... other message types
	}
}

// handleSpecCreatorStart initiates the spec creator workflow
func (m model) handleSpecCreatorStart(msg specCreatorMsg) (tea.Model, tea.Cmd) {
	m.statusText = "* thinking  Creating specification"

	// Build initial prompt
	agentPrompt := spec.LoadInteractiveAgentPrompt()
	initialPrompt := runner.BuildPromptWithHeader(
		fmt.Sprintf("Create a new specification file named '%s.md'. Guide the user through filling out each section interactively.", msg.name),
		"Create Specification",
	)

	// Create runner
	r, err := runner.NewRunner(m.cfg)
	if err != nil {
		m.errMsg = fmt.Sprintf("Error creating runner: %v", err)
		m.done = true
		return m, tea.Quit
	}

	// Run agent
	events, errc := r.Run(runner.RunOptions{
		Prompt:       initialPrompt,
		SystemPrompt: agentPrompt,
		SessionID:    m.sessionID,
		CWD:          m.projectPath,
	})

	return m, waitForEvent(events, errc)
}
```

**Rationale**: Adds message handling for spec creator workflow, following the same pattern as plan and implement commands.

---

### Final Validation

Before marking this phase complete, run comprehensive validation:

#### 1. Build & Unit Tests
```bash
go build -o spektacular .
go test ./... -v
```

#### 2. End-to-End Scenarios

**Scenario A: Happy Path (Detailed Responses)**
```bash
cd /tmp/test-spektacular
spektacular init
spektacular new detailed-feature
# For each section, provide detailed, well-structured responses
# Expected: No clarification questions, smooth flow through all sections
# Expected: Spec saved to .spektacular/specs/detailed-feature.md
cat .spektacular/specs/detailed-feature.md
# Verify: All sections present, content matches input
```

**Scenario B: Vague Responses (Clarification Flow)**
```bash
spektacular new vague-feature
# Overview: "Add a dashboard"
# Expected: Agent asks "What will the dashboard show? Who will use it?"
# Provide clarification
# Requirements: "Make it fast"
# Expected: Agent asks specific performance questions
# Continue until all sections complete
# Verify: Final spec has clarified, detailed content
```

**Scenario C: Multi-line Markdown Input**
```bash
spektacular new markdown-feature
# Select "Free text" for Overview
# Enter:
## Background
This feature adds **real-time** notifications.

### Key Points
- Users get instant updates
- Reduces email spam
- Improves engagement

# Press Ctrl+D
# Verify: Markdown formatting preserved in spec
```

**Scenario D: Multiple Clarification Rounds**
```bash
spektacular new stubborn-feature
# Overview: "improve things"
# Clarification 1: "improve UX"
# Clarification 2: "make UI better"
# Expected: Agent continues asking for specifics (max 2 rounds per spec instructions)
# Eventually accepts or moves on
```

**Scenario E: Non-interactive Fallback**
```bash
spektacular new template-feature --noninteractive
# Expected: Creates template with placeholders (current behavior)
# No TUI should appear
diff .spektacular/specs/template-feature.md internal/defaults/files/spec-template.md
# Should be nearly identical (just title substitution)
```

**Scenario F: Graceful Exit**
```bash
spektacular new cancelled-feature
# Start answering questions
# Press Ctrl+C during Overview
# Expected: TUI exits immediately
# No partial spec file should exist
ls .spektacular/specs/cancelled-feature.md
# Should not exist
```

#### 3. Spec Structure Validation

After creating a spec via interactive mode, validate structure:

```bash
SPEC=".spektacular/specs/test-feature.md"

# Check all required sections exist
grep -q "## Overview" $SPEC || echo "FAIL: Missing Overview"
grep -q "## Requirements" $SPEC || echo "FAIL: Missing Requirements"
grep -q "## Constraints" $SPEC || echo "FAIL: Missing Constraints"
grep -q "## Acceptance Criteria" $SPEC || echo "FAIL: Missing Acceptance Criteria"
grep -q "## Technical Approach" $SPEC || echo "FAIL: Missing Technical Approach"
grep -q "## Success Metrics" $SPEC || echo "FAIL: Missing Success Metrics"
grep -q "## Non-Goals" $SPEC || echo "FAIL: Missing Non-Goals"

# Check HTML comments preserved
grep -q "<!--" $SPEC || echo "FAIL: Missing HTML comment guidance"

# Check proper markdown structure
grep -q "^# Feature:" $SPEC || echo "FAIL: Missing main title"
```

#### 4. Agent Integration Test

Verify the created spec works with plan command:

```bash
spektacular new integration-test
# Fill out complete spec interactively
spektacular plan integration-test.md
# Expected: Plan command reads spec successfully and creates plan
# No errors about spec format
```

### Success Criteria

#### Automated Verification
- [ ] `go build -o spektacular .` (builds successfully)
- [ ] `go test ./... -v` (all tests pass)
- [ ] `go test ./cmd/... -run TestNewCmd` (new command tests pass)
- [ ] `go test ./internal/spec/... -run TestInteractive` (interactive logic tests pass)
- [ ] `go test ./internal/tui/... -run TestTextarea` (TUI enhancements pass)

#### Manual Verification
- [ ] Scenario A: Happy path with detailed responses works without clarification
- [ ] Scenario B: Vague responses trigger clarification questions
- [ ] Scenario C: Multi-line markdown input preserves formatting
- [ ] Scenario D: Multiple clarification rounds work (up to agent's limit)
- [ ] Scenario E: `--noninteractive` creates template (exact current behavior)
- [ ] Scenario F: Ctrl+C exits gracefully without partial files
- [ ] All 7 spec sections are collected in correct order
- [ ] HTML comment guidance preserved in final spec
- [ ] Created spec passes `spektacular plan` command
- [ ] Textarea supports clipboard paste (Ctrl+V)
- [ ] Terminal resize doesn't break textarea layout

---

## Migration & Rollout

### Breaking Changes
**None** - The `--noninteractive` flag preserves exact current behavior. Interactive mode is additive.

### Feature Flag Strategy
Not applicable - Mode selection based on TTY detection and `--noninteractive` flag.

### Rollback Plan
If interactive mode has issues:
1. Add temporary flag `--force-template` as alias for `--noninteractive`
2. Update documentation to recommend `--noninteractive` until issues resolved
3. If critical: Change default to non-interactive, add `--interactive` opt-in flag
4. Hotfix specific agent prompt issues without requiring code changes (agent prompt is embedded but can be overridden via config in future enhancement)

### Monitoring & Alerting
Not applicable for CLI tool - no telemetry.

User feedback channels:
- GitHub issues for bug reports
- Monitor for reports of:
  - Agent asking too many clarification questions
  - Agent not detecting vague input
  - Textarea input bugs (missing characters, clipboard issues)
  - Spec structure validation failures

### Deployment Steps
1. Merge PR with all three phases
2. Build new binary: `go build -o spektacular .`
3. Install locally: `cp spektacular /usr/local/bin/` (or platform equivalent)
4. Update documentation:
   - README.md: Add section on interactive mode
   - Mention `--noninteractive` for CI/CD or scripting scenarios
5. Create example video/GIF demonstrating interactive flow
6. Announce in release notes:
   - New interactive mode for guided spec creation
   - `--noninteractive` preserves old behavior
   - Multi-line markdown input support

---

## References

### Original Specification
`.spektacular/specs/9_interactive_spec.md`

### Key Files Examined

**Agent Infrastructure:**
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/defaults/files/agents/planner.md:1-287` - Example agent prompt structure
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/plan/plan.go:117-165` - Multi-turn conversation loop pattern
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/runner/runner.go:99-125` - Question detection system

**TUI Components:**
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/tui/tui.go:40-53` - Workflow abstraction
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/tui/tui.go:293-336` - Current manual text input
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/tui/tui.go:492-528` - Question panel rendering

**Current Implementation:**
- `/home/nicj/code/github.com/jumppad-labs/spektacular/cmd/new.go:14-36` - Command definition
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/spec/spec.go:16-63` - Template creation logic
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/defaults/files/spec-template.md:1-87` - Spec template structure

**Reference Patterns:**
- `/home/nicj/code/github.com/jumppad-labs/spektacular/cmd/plan.go:40-60` - TTY detection pattern
- `/home/nicj/code/github.com/jumppad-labs/spektacular/cmd/init.go:33-36` - Boolean flag pattern

### Related Patterns
- Bubble Tea documentation: https://github.com/charmbracelet/bubbletea
- Bubbles textarea component: https://github.com/charmbracelet/bubbles/tree/master/textarea
- Cobra CLI framework: https://github.com/spf13/cobra
