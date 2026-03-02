// Package implement orchestrates the plan-execution workflow.
package implement

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jumppad-labs/spektacular/internal/config"
	"github.com/jumppad-labs/spektacular/internal/defaults"
	"github.com/jumppad-labs/spektacular/internal/runner"
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

// RunImplement executes the implementation pipeline for the given plan directory.
// onText is called with each text chunk; onQuestion is called when questions are detected.
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

	if cfg.Debug.Enabled {
		debugDir := filepath.Join(projectPath, ".spektacular", "debug")
		_ = os.MkdirAll(debugDir, 0755)
		_ = os.WriteFile(filepath.Join(debugDir, "implement-prompt.md"), []byte(planContent), 0644)
	}

	r, err := runner.NewRunner(cfg)
	if err != nil {
		return "", fmt.Errorf("creating runner: %w", err)
	}

	logFile := ""
	if cfg.Debug.Enabled && cfg.Debug.LogDir != "" {
		logDir := filepath.Join(projectPath, cfg.Debug.LogDir)
		_ = os.MkdirAll(logDir, 0755)
		logFile = filepath.Join(logDir, time.Now().Format("2006-01-02")+"_implement.log")
	}

	if err := runner.RunSteps(r, []runner.Step{
		{
			Prompts: runner.Prompts{
				User:   runner.BuildPromptWithHeader(planContent, "Implementation Plan"),
				System: LoadAgentPrompt(),
			},
			LogFile: logFile,
		},
	}, cfg, projectPath, onText, onQuestion); err != nil {
		return "", err
	}

	return planDir, nil
}
