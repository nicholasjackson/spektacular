# Implement Phase - Implementation Plan

## Overview
- **Specification**: `.spektacular/specs/7_implement_phase.md`
- **Complexity**: Medium
- **Dependencies**: `internal/tui`, `internal/plan`, `internal/runner`, `internal/defaults`, `internal/config`

## Current State Analysis
- **Exists**: Full plan workflow (`spektacular plan <spec-file>`) with interactive TUI, Claude subprocess runner, multi-turn question handling, and agent prompt system. An executor agent definition already exists at `internal/defaults/files/agents/executor.md`.
- **Missing**: No `implement` command, no way to execute plans. The `run` command is a placeholder stub. The TUI is hardcoded for plan generation (loads planner agent, builds plan-specific prompt, validates plan.md output).
- **Key constraint**: The spec requires reusing the TUI and agent execution logic from plan, not duplicating it.

## Implementation Strategy

Refactor the TUI to accept a generic `Workflow` configuration, then create the implement package and command that plugs into the same TUI. This approach:
1. Avoids code duplication between plan and implement
2. Makes the TUI reusable for future workflow types
3. Keeps changes minimal and focused

## Phase 1: Refactor TUI for Workflow Reusability

### Changes Required

- **File**: `internal/tui/tui.go`
  - **Add** `Workflow` struct (new type near top of file, after message types):
    ```go
    // Workflow defines the agent-specific behavior for the TUI.
    // Both plan and implement provide their own Workflow to customize
    // prompt construction and result handling.
    type Workflow struct {
        // StatusLabel is shown in the "thinking" status bar.
        StatusLabel string

        // Start returns a tea.Cmd that builds the prompt and spawns the runner.
        Start func(cfg config.Config, sessionID string) tea.Cmd

        // OnResult is called when the agent produces a terminal result event.
        // Returns the output directory path or an error.
        OnResult func(resultText string) (string, error)
    }
    ```
  - **Modify** `model` struct: replace `specPath string` field with `workflow Workflow`
  - **Modify** `initialModel`: accept `Workflow` instead of `specPath`, set `statusText` using `workflow.StatusLabel`
  - **Modify** `Init()`: call `m.workflow.Start(m.cfg, "")` instead of `startAgentCmd(m.specPath, ...)`
  - **Modify** `handleOtherInput` (line 282) and `handleNumberKey` (line 338): replace `filepath.Base(m.specPath)` with `m.workflow.StatusLabel`
  - **Modify** `handleAgentEvent` result handling (lines 396-411): replace inline plan-dir derivation and `plan.WritePlanOutput()` with `m.workflow.OnResult(event.ResultText())`
  - **Rename** `startAgentCmd` → keep as a private helper, called from `RunPlanTUI`'s workflow closure
  - **Add** `RunAgentTUI(wf Workflow, projectPath string, cfg config.Config) (string, error)` as the generic entry point
  - **Modify** `RunPlanTUI`: create a plan-specific `Workflow` and delegate to `RunAgentTUI`

  **Proposed `RunAgentTUI`**:
  ```go
  func RunAgentTUI(wf Workflow, projectPath string, cfg config.Config) (string, error) {
      m := initialModel(wf, projectPath, cfg)

      p := tea.NewProgram(
          m,
          tea.WithAltScreen(),
          tea.WithMouseCellMotion(),
      )

      finalModel, err := p.Run()
      if err != nil {
          return "", err
      }

      fm := finalModel.(model)
      if fm.errMsg != "" {
          return "", fmt.Errorf("%s", fm.errMsg)
      }
      return fm.resultDir, nil
  }
  ```

  **Proposed `RunPlanTUI` (refactored)**:
  ```go
  func RunPlanTUI(specPath, projectPath string, cfg config.Config) (string, error) {
      wf := Workflow{
          StatusLabel: filepath.Base(specPath),
          Start: func(c config.Config, sessionID string) tea.Cmd {
              return startAgentCmd(specPath, projectPath, c, sessionID)
          },
          OnResult: func(resultText string) (string, error) {
              specName := stripExt(filepath.Base(specPath))
              planDir := filepath.Join(projectPath, ".spektacular", "plans", specName)
              if err := plan.WritePlanOutput(planDir, resultText); err != nil {
                  return "", err
              }
              return planDir, nil
          },
      }
      return RunAgentTUI(wf, projectPath, cfg)
  }
  ```

  **Proposed `handleAgentEvent` result section (replacing lines 384-411)**:
  ```go
  // Result event — terminal
  if event.IsResult() {
      m.toolLine = ""
      if event.IsError() {
          m.errMsg = event.ResultText()
          m.done = true
          m.statusText = "error  press q to exit"
          p := m.currentPalette()
          m = m.withLine(lipgloss.NewStyle().Foreground(p.errColor).Render("• "+m.errMsg) + "\n")
          return m, nil
      }

      resultDir, err := m.workflow.OnResult(event.ResultText())
      if err != nil {
          m.errMsg = err.Error()
          m.done = true
          m.statusText = "error  press q to exit"
          return m, nil
      }
      m.resultDir = resultDir
      m.done = true
      m.statusText = "done  press q to exit"
      p := m.currentPalette()
      m = m.withLine(lipgloss.NewStyle().Foreground(p.success).Render(
          fmt.Sprintf("• completed  output: %s", resultDir),
      ) + "\n")
      return m, nil
  }
  ```

  **Proposed `initialModel` (modified)**:
  ```go
  func initialModel(wf Workflow, projectPath string, cfg config.Config) model {
      return model{
          workflow:    wf,
          projectPath: projectPath,
          cfg:         cfg,
          themeIdx:    0,
          followMode:  true,
          statusText:  "* thinking  " + wf.StatusLabel,
      }
  }
  ```

### Testing Strategy
- Update `TestCurrentPalette_DefaultIsDracula` and other tests that call `initialModel` — pass a `Workflow` instead of `specPath`
- Update `TestWithLine_*` tests similarly
- Add `TestRunAgentTUI_WorkflowCalled` — verify the Workflow's Start function is invoked (can check model state after init)

### Success Criteria
#### Automated Verification
- [ ] `go build ./...` compiles without errors
- [ ] `go test ./internal/tui/...` passes
- [ ] `spektacular plan .spektacular/specs/1_plan_mode.md` still works (regression check)

## Phase 2: Extend Runner for Custom Content Headers

### Changes Required

- **File**: `internal/runner/runner.go`
  - **Add** `BuildPromptWithHeader` function that accepts a custom content section header:
    ```go
    // BuildPromptWithHeader assembles the prompt with a custom content section header.
    func BuildPromptWithHeader(content, agentPrompt string, knowledge map[string]string, header string) string {
        var b strings.Builder
        b.WriteString(agentPrompt)
        b.WriteString("\n\n---\n\n# Knowledge Base\n")
        for filename, content := range knowledge {
            fmt.Fprintf(&b, "\n## %s\n%s\n", filename, content)
        }
        fmt.Fprintf(&b, "\n---\n\n# %s\n\n%s", header, content)
        return b.String()
    }
    ```
  - **Refactor** `BuildPrompt` to delegate to `BuildPromptWithHeader`:
    ```go
    func BuildPrompt(specContent, agentPrompt string, knowledge map[string]string) string {
        return BuildPromptWithHeader(specContent, agentPrompt, knowledge, "Specification to Plan")
    }
    ```

### Testing Strategy
- Existing `TestBuildPrompt_*` tests continue to pass (unchanged behavior)
- Add `TestBuildPromptWithHeader_UsesCustomHeader` — verify custom header appears in output

### Success Criteria
#### Automated Verification
- [ ] `go test ./internal/runner/...` passes
- [ ] Existing `TestBuildPrompt_ContainsAllParts` still passes

## Phase 3: Create Implement Package

### Changes Required

- **File**: `internal/implement/implement.go` (new file)
  ```go
  // Package implement orchestrates the plan-execution workflow.
  package implement

  import (
      "fmt"
      "os"
      "path/filepath"
      "strings"

      "github.com/nicholasjackson/spektacular/internal/config"
      "github.com/nicholasjackson/spektacular/internal/defaults"
      "github.com/nicholasjackson/spektacular/internal/plan"
      "github.com/nicholasjackson/spektacular/internal/runner"
  )

  // LoadAgentPrompt returns the embedded executor agent prompt.
  func LoadAgentPrompt() string {
      return string(defaults.MustReadFile("agents/executor.md"))
  }

  // LoadPlanContent reads plan files from planDir and returns the combined content.
  // plan.md is required; context.md and research.md are optional.
  func LoadPlanContent(planDir string) (string, error) {
      planFile := filepath.Join(planDir, "plan.md")
      planContent, err := os.ReadFile(planFile)
      if err != nil {
          return "", fmt.Errorf("plan.md not found in %s: %w", planDir, err)
      }

      var b strings.Builder

      if ctx, err := os.ReadFile(filepath.Join(planDir, "context.md")); err == nil {
          b.WriteString("## context.md\n")
          b.Write(ctx)
          b.WriteString("\n\n")
      }

      b.WriteString("## plan.md\n")
      b.Write(planContent)
      b.WriteString("\n\n")

      if research, err := os.ReadFile(filepath.Join(planDir, "research.md")); err == nil {
          b.WriteString("## research.md\n")
          b.Write(research)
          b.WriteString("\n\n")
      }

      return b.String(), nil
  }

  // ResolvePlanDir resolves the plan directory from the given argument.
  // It checks: (1) direct path, (2) relative to cwd, (3) plan name in .spektacular/plans/.
  func ResolvePlanDir(arg, cwd string) (string, error) {
      candidates := []string{
          arg,
          filepath.Join(cwd, arg),
          filepath.Join(cwd, ".spektacular", "plans", arg),
      }
      for _, dir := range candidates {
          if _, err := os.Stat(filepath.Join(dir, "plan.md")); err == nil {
              return dir, nil
          }
      }
      return "", fmt.Errorf("plan.md not found: tried %s, %s, and .spektacular/plans/%s",
          arg, filepath.Join(cwd, arg), arg)
  }

  // RunImplement executes the full implementation loop for the given plan directory.
  // onText is called with each text chunk from the agent (may be nil).
  // onQuestion is called when questions are detected; it must return the answer string.
  func RunImplement(
      planDir, projectPath string,
      cfg config.Config,
      onText func(string),
      onQuestion func([]runner.Question) string,
  ) (string, error) {
      planContent, err := LoadPlanContent(planDir)
      if err != nil {
          return "", err
      }

      agentPrompt := LoadAgentPrompt()
      knowledge := plan.LoadKnowledge(projectPath)
      prompt := runner.BuildPromptWithHeader(planContent, agentPrompt, knowledge, "Implementation Plan")

      if cfg.Debug.Enabled {
          logDir := filepath.Join(projectPath, cfg.Debug.LogDir)
          _ = os.MkdirAll(logDir, 0755)
          _ = os.WriteFile(filepath.Join(planDir, "implement-prompt.md"), []byte(prompt), 0644)
      }

      sessionID := ""
      currentPrompt := prompt

      for {
          var questionsFound []runner.Question
          var finalResult string

          events, errc := runner.RunClaude(runner.RunOptions{
              Prompt:    currentPrompt,
              Config:    cfg,
              SessionID: sessionID,
              CWD:       projectPath,
              Command:   "implement",
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
                  if event.IsError() {
                      return "", fmt.Errorf("agent error: %s", event.ResultText())
                  }
                  finalResult = event.ResultText()
              }
          }

          if err := <-errc; err != nil {
              return "", fmt.Errorf("runner error: %w", err)
          }

          if len(questionsFound) > 0 && onQuestion != nil {
              answer := onQuestion(questionsFound)
              currentPrompt = answer
              continue
          }

          if finalResult == "" {
              return "", fmt.Errorf("agent completed without producing a result")
          }
          return planDir, nil
      }
  }
  ```

### Testing Strategy
- `TestLoadAgentPrompt_ReturnsContent` — verify executor prompt is non-empty
- `TestLoadPlanContent_RequiresPlanMD` — error when plan.md missing
- `TestLoadPlanContent_IncludesAllFiles` — includes context.md, plan.md, research.md
- `TestLoadPlanContent_OptionalFilesSkipped` — works with only plan.md
- `TestResolvePlanDir_DirectPath` — resolves when full path given
- `TestResolvePlanDir_PlanName` — resolves when just plan name given
- `TestResolvePlanDir_NotFound` — returns error when plan.md not found

### Success Criteria
#### Automated Verification
- [ ] `go test ./internal/implement/...` passes
- [ ] `go build ./...` compiles

## Phase 4: Create Implement Command

### Changes Required

- **File**: `cmd/implement.go` (new file)
  ```go
  package cmd

  import (
      "fmt"
      "os"
      "path/filepath"

      "github.com/nicholasjackson/spektacular/internal/config"
      "github.com/nicholasjackson/spektacular/internal/implement"
      "github.com/nicholasjackson/spektacular/internal/runner"
      "github.com/nicholasjackson/spektacular/internal/tui"
      "github.com/spf13/cobra"
      "golang.org/x/term"
  )

  var implementCmd = &cobra.Command{
      Use:   "implement <plan-directory>",
      Short: "Execute an implementation plan",
      Args:  cobra.ExactArgs(1),
      RunE: func(cmd *cobra.Command, args []string) error {
          planArg := args[0]

          cwd, err := os.Getwd()
          if err != nil {
              return fmt.Errorf("getting working directory: %w", err)
          }

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

          planDir, err := implement.ResolvePlanDir(planArg, cwd)
          if err != nil {
              return err
          }

          if term.IsTerminal(int(os.Stdout.Fd())) {
              _, err = tui.RunImplementTUI(planDir, cwd, cfg)
              if err != nil {
                  return fmt.Errorf("implementation failed: %w", err)
              }
          } else {
              _, err = implement.RunImplement(planDir, cwd, cfg,
                  func(text string) { fmt.Print(text) },
                  func(questions []runner.Question) string {
                      if len(questions) > 0 {
                          fmt.Printf("\n[Question] %s\n", questions[0].Question)
                      }
                      return ""
                  },
              )
              if err != nil {
                  return fmt.Errorf("implementation failed: %w", err)
              }
          }

          fmt.Println("Implementation complete.")
          return nil
      },
  }
  ```

- **File**: `cmd/root.go` (line 29, add after `planCmd`)
  - **Add**: `rootCmd.AddCommand(implementCmd)` in the `init()` function

### Testing Strategy
- Manual: `spektacular implement --help` shows correct usage
- Manual: `spektacular implement nonexistent` returns clear error about missing plan.md

### Success Criteria
#### Automated Verification
- [ ] `go build ./...` compiles
- [ ] `spektacular --help` shows `implement` command
- [ ] `spektacular implement --help` shows usage with `<plan-directory>` argument

## Phase 5: Add Implement TUI Entry Point

### Changes Required

- **File**: `internal/tui/tui.go`
  - **Add** `RunImplementTUI` function:
    ```go
    // RunImplementTUI launches the interactive TUI for plan implementation.
    func RunImplementTUI(planDir, projectPath string, cfg config.Config) (string, error) {
        wf := Workflow{
            StatusLabel: filepath.Base(planDir),
            Start: func(c config.Config, sessionID string) tea.Cmd {
                return implementStartCmd(planDir, projectPath, c, sessionID)
            },
            OnResult: func(resultText string) (string, error) {
                return planDir, nil
            },
        }
        return RunAgentTUI(wf, projectPath, cfg)
    }
    ```

  - **Add** `implementStartCmd` function:
    ```go
    func implementStartCmd(planDir, projectPath string, cfg config.Config, sessionID string) tea.Cmd {
        return func() tea.Msg {
            planContent, err := implement.LoadPlanContent(planDir)
            if err != nil {
                return agentErrMsg{err: fmt.Errorf("loading plan: %w", err)}
            }
            agentPrompt := implement.LoadAgentPrompt()
            knowledge := plan.LoadKnowledge(projectPath)
            prompt := runner.BuildPromptWithHeader(planContent, agentPrompt, knowledge, "Implementation Plan")

            if cfg.Debug.Enabled {
                _ = os.WriteFile(filepath.Join(planDir, "implement-prompt.md"), []byte(prompt), 0644)
            }

            events, errc := runner.RunClaude(runner.RunOptions{
                Prompt:    prompt,
                Config:    cfg,
                SessionID: sessionID,
                CWD:       projectPath,
                Command:   "implement",
            })
            return readNext(events, errc)
        }
    }
    ```

  - **Add** import for `implement` package in the import block

### Testing Strategy
- The generic `RunAgentTUI` is already tested via `RunPlanTUI` regression
- Add test verifying `implementStartCmd` returns proper error msg for missing plan

### Success Criteria
#### Automated Verification
- [ ] `go build ./...` compiles
- [ ] `go test ./internal/tui/...` passes
- [ ] `go test ./...` all tests pass

## Phase 6: Write Tests

### Changes Required

- **File**: `internal/implement/implement_test.go` (new file)
  ```go
  package implement

  import (
      "os"
      "path/filepath"
      "testing"

      "github.com/stretchr/testify/require"
  )

  func TestLoadAgentPrompt_ReturnsContent(t *testing.T) {
      content := LoadAgentPrompt()
      require.NotEmpty(t, content)
      require.Contains(t, content, "Execution Agent")
  }

  func TestLoadPlanContent_RequiresPlanMD(t *testing.T) {
      dir := t.TempDir()
      _, err := LoadPlanContent(dir)
      require.Error(t, err)
      require.Contains(t, err.Error(), "plan.md not found")
  }

  func TestLoadPlanContent_IncludesAllFiles(t *testing.T) {
      dir := t.TempDir()
      require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.md"), []byte("# Plan"), 0644))
      require.NoError(t, os.WriteFile(filepath.Join(dir, "context.md"), []byte("# Context"), 0644))
      require.NoError(t, os.WriteFile(filepath.Join(dir, "research.md"), []byte("# Research"), 0644))

      content, err := LoadPlanContent(dir)
      require.NoError(t, err)
      require.Contains(t, content, "# Plan")
      require.Contains(t, content, "# Context")
      require.Contains(t, content, "# Research")
  }

  func TestLoadPlanContent_OptionalFilesSkipped(t *testing.T) {
      dir := t.TempDir()
      require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.md"), []byte("# Plan Only"), 0644))

      content, err := LoadPlanContent(dir)
      require.NoError(t, err)
      require.Contains(t, content, "# Plan Only")
      require.NotContains(t, content, "context.md")
      require.NotContains(t, content, "research.md")
  }

  func TestResolvePlanDir_DirectPath(t *testing.T) {
      dir := t.TempDir()
      require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.md"), []byte("plan"), 0644))

      resolved, err := ResolvePlanDir(dir, "/tmp")
      require.NoError(t, err)
      require.Equal(t, dir, resolved)
  }

  func TestResolvePlanDir_PlanName(t *testing.T) {
      cwd := t.TempDir()
      planDir := filepath.Join(cwd, ".spektacular", "plans", "my-feature")
      require.NoError(t, os.MkdirAll(planDir, 0755))
      require.NoError(t, os.WriteFile(filepath.Join(planDir, "plan.md"), []byte("plan"), 0644))

      resolved, err := ResolvePlanDir("my-feature", cwd)
      require.NoError(t, err)
      require.Equal(t, planDir, resolved)
  }

  func TestResolvePlanDir_NotFound(t *testing.T) {
      _, err := ResolvePlanDir("nonexistent", t.TempDir())
      require.Error(t, err)
      require.Contains(t, err.Error(), "plan.md not found")
  }
  ```

- **File**: `internal/runner/runner_test.go` (append new test)
  ```go
  func TestBuildPromptWithHeader_UsesCustomHeader(t *testing.T) {
      prompt := BuildPromptWithHeader("plan content", "agent instructions", nil, "Implementation Plan")
      require.Contains(t, prompt, "# Implementation Plan")
      require.Contains(t, prompt, "plan content")
      require.Contains(t, prompt, "agent instructions")
      require.NotContains(t, prompt, "Specification to Plan")
  }
  ```

- **File**: `internal/tui/tui_test.go` (update existing tests)
  - Update all calls to `initialModel()` to pass a `Workflow` struct instead of `specPath`:
    ```go
    func testWorkflow(label string) Workflow {
        return Workflow{StatusLabel: label}
    }

    // Replace: initialModel("spec.md", "/tmp", config.NewDefault())
    // With:    initialModel(testWorkflow("spec.md"), "/tmp", config.NewDefault())
    ```

### Success Criteria
#### Automated Verification
- [ ] `go test ./...` — all tests pass
- [ ] `go vet ./...` — no issues

## References
- Original specification: `.spektacular/specs/7_implement_phase.md`
- Plan command: `cmd/plan.go:16-66`
- Plan orchestration: `internal/plan/plan.go:78-159`
- TUI implementation: `internal/tui/tui.go:1-637`
- Runner: `internal/runner/runner.go:125-135` (BuildPrompt)
- Executor agent: `internal/defaults/files/agents/executor.md`
- Defaults loader: `internal/defaults/defaults.go:9-29`
