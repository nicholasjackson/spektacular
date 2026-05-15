# Interactive Spec Creation - Research Notes

## Specification Analysis

### Original Requirements
From `.spektacular/specs/9_interactive_spec.md`:

1. **Interactive Mode** - Prompt user to fill out each spec section with guidance
2. **Clarification and Questions** - Ask follow-ups for vague/incomplete input
3. **TUI** - Interactive terminal UI with multiline input and markdown formatting
4. **Save Spec** - Save completed spec to appropriate location

### Implicit Requirements
- Must not break existing `spektacular new` workflow
- Should follow existing patterns for plan/implement commands
- Need to detect when to use interactive vs non-interactive mode
- Multi-turn conversation support for clarification rounds

### Constraints Identified
- Must integrate with existing new command
- Must use existing agent infrastructure (Claude CLI runner)
- Must maintain spec template structure
- Must work within Bubble Tea TUI framework

## Research Process

### Sub-agents Spawned
1. **Codebase Explorer** - Analyzed new command implementation
2. **Agent Interaction Researcher** - Mapped LLM integration patterns
3. **TUI Implementation Explorer** - Found Bubble Tea components and patterns
4. **CLI Architecture Analyst** - Studied Cobra flag handling
5. **Template Structure Researcher** - Examined spec template system

### Files Examined

#### New Command Implementation
- `/home/nicj/code/github.com/jumppad-labs/spektacular/cmd/new.go:1-36` - Current command with title/description flags
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/spec/spec.go:16-75` - Template creation logic with string replacements
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/defaults/files/spec-template.md:1-87` - 7-section template structure with HTML guidance

#### Agent Infrastructure
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/defaults/files/agents/planner.md:1-287` - Example system prompt structure
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/defaults/files/agents/executor.md:1-277` - Another agent example
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/plan/plan.go:53-56` - Agent prompt loading pattern
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/plan/plan.go:117-165` - Multi-turn conversation loop
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/runner/registry.go:10-34` - Runner registration system
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/runner/claude/claude.go:19-113` - Claude CLI subprocess execution

#### Question/Answer System
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/runner/runner.go:99-125` - Question detection via regex `<!--QUESTION:...-->`
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/runner/runner.go:23-90` - Event parsing (session_id, text content, result)
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/runner/runner.go:127-138` - Prompt building with knowledge hint

#### TUI Components
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/tui/tui.go:40-53` - Workflow abstraction (StatusLabel, Start, OnResult)
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/tui/tui.go:59-79` - Model structure with viewport, questions, answers
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/tui/tui.go:247-291` - Key handling delegation
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/tui/tui.go:293-336` - Manual single-line text input (current limitation)
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/tui/tui.go:492-528` - Question panel rendering with numbered options
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/tui/tui.go:374-447` - Agent event handling and question detection
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/tui/tui.go:649-668` - Generic RunAgentTUI entry point
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/tui/tui.go:586-604` - Markdown rendering via Glamour
- `/home/nicj/code/github.com/jumppad-labs/spektacular/internal/tui/theme.go:67-74` - Theme system (5 palettes)

#### CLI Patterns
- `/home/nicj/code/github.com/jumppad-labs/spektacular/cmd/root.go:30` - Command registration
- `/home/nicj/code/github.com/jumppad-labs/spektacular/cmd/init.go:33-36` - Boolean flag example (`--force`)
- `/home/nicj/code/github.com/jumppad-labs/spektacular/cmd/plan.go:40-60` - TTY detection pattern with `term.IsTerminal()`
- `/home/nicj/code/github.com/jumppad-labs/spektacular/cmd/implement.go:44-62` - Similar interactive/non-interactive routing

#### Dependencies
- `/home/nicj/code/github.com/jumppad-labs/spektacular/go.mod` - Bubble Tea v1.3.10, Bubbles v1.0.0, Glamour v0.10.0, Lipgloss v1.1.1, Cobra v1.8.1

### Patterns Discovered

#### Multi-Turn Conversation Pattern
```go
sessionID := ""
currentPrompt := initialPrompt

for {
    events, errc := runner.Run(RunOptions{
        Prompt: currentPrompt,
        SystemPrompt: agentPrompt,
        SessionID: sessionID,  // Maintains context
        CWD: projectPath,
    })

    for event := range events {
        if id := event.SessionID(); id != "" {
            sessionID = id
        }
        if text := event.TextContent(); text != "" {
            questions = append(questions, DetectQuestions(text)...)
        }
        if event.IsResult() {
            return event.ResultText(), nil
        }
    }

    if len(questions) > 0 {
        answer := getUserAnswer(questions)
        currentPrompt = answer
        continue  // Next turn
    }

    break
}
```

#### Workflow Abstraction Pattern
```go
workflow := Workflow{
    StatusLabel: "Creating specification",
    Start: func(cfg config.Config, sessionID string) tea.Cmd {
        return startAgentCmd(...)
    },
    OnResult: func(resultText string) (string, error) {
        return validateOutput()
    },
}
RunAgentTUI(workflow, projectPath, cfg)
```

#### TTY Detection Pattern
```go
if term.IsTerminal(int(os.Stdout.Fd())) && !nonInteractive {
    // Launch TUI
    return tui.RunSomethingTUI(...)
} else {
    // Non-interactive mode
    return nonInteractiveImplementation(...)
}
```

## Key Findings

### Architecture Insights

**Agent System Design:**
- Agents are defined by markdown system prompts in `internal/defaults/files/agents/`
- Loaded via `defaults.MustReadFile("agents/<name>.md")`
- Agent prompts embedded at build time using Go's `embed.FS`
- Same runner infrastructure for all agents (plan, implement, future spec-creator)

**Question Flow:**
- Agent outputs structured questions: `<!--QUESTION:{"questions":[...]}-->`
- TUI detects questions via regex pattern matching
- Questions queued in model, displayed one at a time
- User answers via numbered options (1-9) or free-text input
- Answer sent back to agent via `--resume <session-id>` flag

**Session Management:**
- Claude CLI maintains conversation context via session IDs
- First agent response includes `session_id` in event data
- Subsequent turns pass `--resume <session-id>` to continue conversation
- No server-side state - subprocess per turn

### Existing Implementations

**Template System:**
- Embedded template at `internal/defaults/files/spec-template.md`
- Simple string replacement: `{title}`, `{description}`, `{requirement_1}`, etc.
- No validation of spec structure after creation
- Specs consumed as raw markdown by plan command

**TUI Limitations:**
- Current text input is **manual character-by-character** implementation
- Only supports **single-line** input (Enter submits immediately)
- No clipboard support in current implementation
- No use of bubbles/textarea despite it being available (v1.0.0)

### Reusable Components

**Available from Bubbles v1.0.0:**
- `viewport.Model` - Currently used for scrollable output ✅
- `textarea.Model` - **NOT currently used** but available for multi-line input
- `textinput.Model` - Available for single-line input (could replace manual handling)
- Other: list, spinner, progress, table, paginator, filepicker

**Glamour Markdown Rendering:**
- Already integrated and working
- Supports 5 themes (dracula, github-dark, nord, solarized, monokai)
- Used for rendering agent text output
- Word-wrapping based on terminal width

### Testing Infrastructure

**Existing Tests:**
- Unit tests for spec creation: `internal/spec/spec_test.go`
- Tests use `testify/require` for assertions
- Pattern: `require.NotEmpty()`, `require.NoError()`, `require.Contains()`

**Test Coverage:**
- Template loading tested
- String replacement tested
- File creation tested
- **No agent integration tests** - manual verification only

## Questions & Answers

### Q: Should we use bubbles/textarea or continue manual input handling?
**A**: Use bubbles/textarea
**Impact**: Provides proper multi-line editing, clipboard support, and standard keybindings. Reduces code complexity and maintenance burden.
**Decision**: Phase 2 will integrate textarea.Model

### Q: How should users submit multi-line input?
**A**: Use Ctrl+D (end-of-input) or Ctrl+S (save) keybindings
**Impact**: Enter key can be used for newlines. Standard Unix convention.
**Decision**: Ctrl+D and Ctrl+S both submit, Enter adds newline, Esc cancels

### Q: Should interactive mode be opt-in or opt-out?
**A**: Opt-out (interactive by default, `--noninteractive` to disable)
**Impact**: Spec says "should run in interactive mode by default"
**Decision**: TTY detection + `--noninteractive` flag for opt-out

### Q: What if agent asks too many clarification questions?
**A**: Agent prompt includes guidance: "Maximum 2 rounds of clarification per section unless user provides contradictory info"
**Impact**: Prevents frustrating user with excessive back-and-forth
**Decision**: Let LLM make clarification decisions based on system prompt guidelines

### Q: How to validate agent created the spec correctly?
**A**: Check file exists after completion, rely on agent following system prompt
**Impact**: No structural validation - trust agent to follow template format
**Decision**: Simple file existence check, document structure validation as future enhancement

### Q: Should we support editing previous answers?
**A**: No, not in initial implementation
**Impact**: Would require significant TUI state management complexity
**Decision**: User can Ctrl+C and restart if they make a mistake. Future enhancement: answer history navigation.

### Q: What about mobile or non-TTY environments?
**A**: Falls back to template creation automatically
**Impact**: CI/CD, pipes, SSH without PTY all get template behavior
**Decision**: Same as plan/implement - TTY detection handles this

## Design Decisions

### Decision: Create New Agent vs. Enhance Template System
**Options Considered:**
1. Create spec-creator agent with conversation flow
2. Enhance template system with interactive prompts
3. Build standalone wizard without agent

**Rationale**: Option 1 chosen
- Leverages existing agent infrastructure
- LLM can intelligently detect vague input and ask clarifying questions
- Consistent with plan/implement UX
- Spec says "A new agent system prompt should be created"

**Trade-offs**: Requires Claude CLI to be configured, won't work offline. Acceptable since plan/implement already have this requirement.

---

### Decision: Multi-line Input via Textarea Component
**Options Considered:**
1. Continue manual character handling, add newline support
2. Use bubbles/textarea component
3. Use bubbles/textinput for single-line only

**Rationale**: Option 2 chosen
- Spec requires "multiline input and markdown formatting"
- Textarea provides this out-of-the-box
- Reduces maintenance burden
- Standard editing keybindings (clipboard, navigation)

**Trade-offs**: Adds dependency on bubbles component behavior. Acceptable since bubbles is already a dependency.

---

### Decision: Agent Decides Clarification Logic
**Options Considered:**
1. Hard-coded rules for vague detection (e.g., word count < 10)
2. Let LLM agent decide when to ask clarifying questions
3. Always ask clarifying questions for every section

**Rationale**: Option 2 chosen
- LLM better at detecting vague/incomplete content
- Can adapt to context (e.g., "N/A" is valid for Constraints)
- Spec says "system should ask follow-up questions if the user's input is too vague or incomplete"
- Agent prompt provides guidelines to prevent over-questioning

**Trade-offs**: Less predictable behavior than hard-coded rules. Acceptable since agent prompt provides guard rails (max 2 rounds).

---

### Decision: Preserve Exact Current Behavior in Non-Interactive Mode
**Options Considered:**
1. Modify template creation to be more sophisticated
2. Keep existing spec.Create() completely unchanged
3. Remove template mode entirely (interactive only)

**Rationale**: Option 2 chosen
- Spec says "must integrate with the existing new command"
- No breaking changes for users who script `spektacular new`
- CI/CD environments need deterministic template creation
- Safe rollback path if interactive mode has issues

**Trade-offs**: Maintains two code paths. Acceptable since they're cleanly separated.

---

### Decision: Question Format via HTML Comments
**Options Considered:**
1. JSON-only output from agent (no HTML comments)
2. Structured HTML comment markers (current plan/implement pattern)
3. Custom text parsing (e.g., "QUESTION: ...")

**Rationale**: Option 2 chosen
- Already implemented and tested in TUI
- Regex pattern exists: `<!--QUESTION:([\s\S]*?)-->`
- Invisible to user (HTML comments don't render)
- JSON payload provides structured data for parsing

**Trade-offs**: Couples agent output format to TUI parsing. Acceptable since it's an established pattern.

---

### Decision: Save Spec via Agent's Write Tool
**Options Considered:**
1. Agent outputs spec content, Go code writes file
2. Agent uses Write tool to create file directly
3. Agent outputs JSON, Go code parses and writes

**Rationale**: Option 2 chosen
- Consistent with how plan/implement work
- Agent knows exact format and structure
- Reduces Go code complexity
- Agent prompt instructs exact file path: `.spektacular/specs/{name}.md`

**Trade-offs**: Less control over file writing in Go. Acceptable since validation step checks file exists.

---

## Code Examples & Patterns

### Agent System Prompt Structure
From `internal/defaults/files/agents/planner.md`:

```markdown
# Agent Name

## Role
[Define agent's purpose]

## Core Mission
[What the agent should accomplish]

## Workflow

### Phase 1: [Phase Name]
[Detailed instructions for this phase]

### Phase 2: [Next Phase]
[More instructions]

## Important Guidelines

### [Guideline Category]
[Specific rules and examples]

## Example Flow
[Show example interaction]
```

**Usage Pattern:**
```go
func LoadAgentPrompt() string {
    return string(defaults.MustReadFile("agents/<name>.md"))
}
```

---

### Multi-Turn Conversation Loop
From `internal/plan/plan.go:117-165`:

```go
sessionID := ""
currentPrompt := prompt

for {
    var questionsFound []runner.Question
    var finalResult string

    events, errc := r.Run(runner.RunOptions{
        Prompt:       currentPrompt,
        SystemPrompt: agentPrompt,
        SessionID:    sessionID,
        CWD:          projectPath,
    })

    for event := range events {
        if id := event.SessionID(); id != "" {
            sessionID = id
        }
        if text := event.TextContent(); text != "" {
            if onText != nil {
                onText(text)
            }
            questionsFound = append(questionsFound, runner.DetectQuestions(text)...)
        }
        if event.IsResult() {
            finalResult = event.ResultText()
        }
    }

    if err := <-errc; err != nil {
        return "", fmt.Errorf("agent error: %w", err)
    }

    if len(questionsFound) > 0 && onQuestion != nil {
        answer := onQuestion(questionsFound)
        currentPrompt = answer
        continue
    }

    if finalResult == "" {
        return "", fmt.Errorf("agent completed without result")
    }

    return finalResult, nil
}
```

---

### Textarea Integration Pattern
From bubbles/textarea documentation:

```go
import "github.com/charmbracelet/bubbles/textarea"

type model struct {
    textarea textarea.Model
}

func initialModel() model {
    ti := textarea.New()
    ti.Placeholder = "Enter your response..."
    ti.Focus()
    ti.CharLimit = 10000
    ti.SetWidth(80)
    ti.SetHeight(10)
    return model{textarea: ti}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+d":
            // Submit
            value := m.textarea.Value()
            // Process value...
        case "esc":
            // Cancel
            m.textarea.Reset()
        default:
            // Delegate to textarea
            var cmd tea.Cmd
            m.textarea, cmd = m.textarea.Update(msg)
            return m, cmd
        }
    }
    return m, nil
}

func (m model) View() string {
    return m.textarea.View()
}
```

---

### Question Detection
From `internal/runner/runner.go:99-125`:

```go
var questionPattern = regexp.MustCompile(`<!--QUESTION:([\s\S]*?)-->`)

type Question struct {
    Question string
    Header   string
    Options  []map[string]any
}

func detectQuestions(text string) []Question {
    var questions []Question
    for _, match := range questionPattern.FindAllStringSubmatch(text, -1) {
        var payload struct {
            Questions []struct {
                Question string           `json:"question"`
                Header   string           `json:"header"`
                Options  []map[string]any `json:"options"`
            } `json:"questions"`
        }
        if err := json.Unmarshal([]byte(match[1]), &payload); err != nil {
            continue
        }
        for _, q := range payload.Questions {
            questions = append(questions, Question{
                Question: q.Question,
                Header:   q.Header,
                Options:  q.Options,
            })
        }
    }
    return questions
}
```

**Question Format:**
```html
<!--QUESTION:{"questions":[{"question":"What authentication method?","header":"Auth Method","options":[{"label":"JWT","description":"Stateless tokens"},{"label":"OAuth2","description":"Third-party auth"}]}]}-->
```

---

### TTY Detection and Mode Routing
From `cmd/plan.go:40-60`:

```go
import "golang.org/x/term"

if term.IsTerminal(int(os.Stdout.Fd())) {
    // Interactive TUI mode
    planDir, err := tui.RunPlanTUI(specFile, cwd, cfg)
    if err != nil {
        return err
    }
    fmt.Printf("Plan created: %s\n", planDir)
} else {
    // Non-interactive streaming mode
    planDir, err := plan.RunPlan(
        specFile, cwd, cfg,
        func(text string) { fmt.Print(text) },
        func(questions []runner.Question) string {
            if len(questions) > 0 {
                fmt.Printf("\n[Question] %s\n", questions[0].Question)
            }
            return ""
        },
    )
    if err != nil {
        return err
    }
    fmt.Printf("Plan created: %s\n", planDir)
}
```

---

## Open Questions (All Resolved)

All questions have been resolved through research and design decisions documented above.
