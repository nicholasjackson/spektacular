// Package plan orchestrates the plan-generation workflow.
package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jumppad-labs/spektacular/internal/config"
	"github.com/jumppad-labs/spektacular/internal/defaults"
	"github.com/jumppad-labs/spektacular/internal/runner"
)

// LoadKnowledge returns all markdown files from .spektacular/knowledge/, keyed by
// their path relative to the knowledge directory.
func LoadKnowledge(projectPath string) map[string]string {
	knowledgeDir := filepath.Join(projectPath, ".spektacular", "knowledge")
	result := make(map[string]string)

	entries, err := os.ReadDir(knowledgeDir)
	if err != nil {
		return result // dir missing — no knowledge
	}

	walkDir(knowledgeDir, knowledgeDir, result, entries)
	return result
}

func walkDir(base, dir string, out map[string]string, entries []os.DirEntry) {
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			subEntries, err := os.ReadDir(path)
			if err == nil {
				walkDir(base, path, out, subEntries)
			}
			continue
		}
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			rel = entry.Name()
		}
		data, err := os.ReadFile(path)
		if err == nil {
			out[rel] = string(data)
		}
	}
}

// LoadAgentPrompt returns the embedded planner agent prompt.
func LoadAgentPrompt() string {
	return string(defaults.MustReadFile("agents/planner.md"))
}

// PreparePlanDir creates the plan directory and removes any stale plan.md so
// that WritePlanOutput can detect whether Claude wrote one via the Write tool.
func PreparePlanDir(planDir string) error {
	if err := os.MkdirAll(planDir, 0755); err != nil {
		return fmt.Errorf("creating plan directory: %w", err)
	}
	_ = os.Remove(filepath.Join(planDir, "plan.md"))
	return nil
}

// WritePlanOutput verifies that the agent wrote plan.md to planDir.
// The agent is always responsible for producing the file via its Write tool.
func WritePlanOutput(planDir, _ string) error {
	planFile := filepath.Join(planDir, "plan.md")
	if _, err := os.Stat(planFile); err != nil {
		return fmt.Errorf("agent did not produce plan.md in %s", planDir)
	}
	return nil
}

// RunPlan executes the plan-generation pipeline for specPath.
// onText is called with each text chunk; onQuestion is called when questions are detected.
func RunPlan(
	specPath, projectPath string,
	cfg config.Config,
	onText func(string),
	onQuestion func([]runner.Question) string,
) (string, error) {
	specContent, err := os.ReadFile(specPath)
	if err != nil {
		return "", fmt.Errorf("reading spec file: %w", err)
	}

	specName := stripExt(filepath.Base(specPath))
	planDir := filepath.Join(projectPath, ".spektacular", "plans", specName)

	if err := PreparePlanDir(planDir); err != nil {
		return "", err
	}

	if cfg.Debug.Enabled {
		debugDir := filepath.Join(projectPath, ".spektacular", "debug")
		_ = os.MkdirAll(debugDir, 0755)
		_ = os.WriteFile(filepath.Join(debugDir, "plan-prompt.md"), specContent, 0644)
	}

	r, err := runner.NewRunner(cfg)
	if err != nil {
		return "", fmt.Errorf("creating runner: %w", err)
	}

	logFile := ""
	if cfg.Debug.Enabled && cfg.Debug.LogDir != "" {
		logDir := filepath.Join(projectPath, cfg.Debug.LogDir)
		_ = os.MkdirAll(logDir, 0755)
		logFile = filepath.Join(logDir, time.Now().Format("2006-01-02")+"_plan.log")
	}

	if err := runner.RunSteps(r, []runner.Step{
		{
			Prompts: runner.Prompts{
				User:   runner.BuildPrompt(string(specContent)),
				System: LoadAgentPrompt(),
			},
			LogFile: logFile,
		},
	}, cfg, projectPath, onText, onQuestion); err != nil {
		return "", err
	}

	if err := WritePlanOutput(planDir, ""); err != nil {
		return "", err
	}
	return planDir, nil
}

func stripExt(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return name
	}
	return name[:len(name)-len(ext)]
}
