package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jumppad-labs/spektacular/internal/spec"
	"github.com/jumppad-labs/spektacular/internal/workflow"
	"github.com/spf13/cobra"
)

// Result is the JSON output from a spec workflow command.
type Result struct {
	Step           string `json:"step"`
	TotalSteps     int    `json:"total_steps"`
	CompletedSteps int    `json:"completed_steps,omitempty"`
	SpecPath       string `json:"spec_path"`
	SpecName       string `json:"spec_name"`
	Instruction    string `json:"instruction"`
}

// StatusResult is the JSON output from --status.
type StatusResult struct {
	SpecName       string            `json:"spec_name"`
	SpecPath       string            `json:"spec_path"`
	CurrentStep    string            `json:"current_step"`
	CompletedSteps []string          `json:"completed_steps"`
	TotalSteps     int               `json:"total_steps"`
	Progress       string            `json:"progress"`
	Steps          []StepStatusEntry `json:"steps"`
}

// StepStatusEntry is one row in the status output.
type StepStatusEntry struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func specStatePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}
	return filepath.Join(cwd, ".spektacular", ".state.json"), nil
}

var specCmd = &cobra.Command{
	Use:   "spec",
	Short: "Manage spec workflow (--new, --next, --step, --status)",
	RunE: func(cmd *cobra.Command, args []string) error {
		isNew, _ := cmd.Flags().GetBool("new")
		isNext, _ := cmd.Flags().GetBool("next")
		stepName, _ := cmd.Flags().GetString("step")
		isStatus, _ := cmd.Flags().GetBool("status")
		dataStr, _ := cmd.Flags().GetString("data")

		set := boolCount(isNew, isNext, stepName != "", isStatus)
		if set != 1 {
			return fmt.Errorf("exactly one of --new, --next, --step, or --status is required")
		}

		var result any
		var err error

		switch {
		case isNew:
			result, err = specNew(dataStr)
		case isNext:
			result, err = specNext()
		case stepName != "":
			result, err = specGoto(stepName)
		case isStatus:
			result, err = specStatus()
		}
		if err != nil {
			return err
		}

		out, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling output: %w", err)
		}

		fmt.Println(string(out))
		return nil
	},
}

func specNew(dataStr string) (*Result, error) {
	if dataStr == "" {
		return nil, fmt.Errorf("--data is required with --new (e.g. --data '{\"name\": \"my-feature\"}')")
	}

	var input struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(dataStr), &input); err != nil {
		return nil, fmt.Errorf("parsing data: %w", err)
	}
	if input.Name == "" {
		return nil, fmt.Errorf("\"name\" is required in --data")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting working directory: %w", err)
	}

	specDir := filepath.Join(cwd, ".spektacular", "specs")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		return nil, fmt.Errorf("creating spec directory: %w", err)
	}

	specPath := filepath.Join(specDir, input.Name+".md")

	if err := spec.RenderScaffold(specPath, input.Name); err != nil {
		return nil, err
	}

	sp, err := specStatePath()
	if err != nil {
		return nil, err
	}

	// Delete any existing state so we start fresh.
	os.Remove(sp)

	steps := spec.Steps()
	wf := workflow.New(steps, sp)
	wf.State().Name = input.Name
	wf.State().ArtifactPath = specPath

	// Advance to the first real step.
	if err := wf.Next(); err != nil {
		return nil, fmt.Errorf("advancing to first step: %w", err)
	}

	cur := wf.Current()
	instruction, err := spec.RenderStep(cur, specPath, wf.NextStepName())
	if err != nil {
		return nil, err
	}

	return &Result{
		Step:        cur,
		TotalSteps:  len(steps),
		SpecPath:    specPath,
		SpecName:    input.Name,
		Instruction: instruction,
	}, nil
}

func specNext() (*Result, error) {
	sp, err := specStatePath()
	if err != nil {
		return nil, err
	}

	steps := spec.Steps()
	wf := workflow.New(steps, sp)

	if err := wf.Next(); err != nil {
		return nil, err
	}

	st := wf.State()

	if wf.IsComplete() {
		return &Result{
			Step:           "done",
			TotalSteps:     len(steps),
			CompletedSteps: len(st.CompletedSteps),
			SpecPath:       st.ArtifactPath,
			SpecName:       st.Name,
			Instruction:    "All steps complete! Review the spec file at " + st.ArtifactPath,
		}, nil
	}

	instruction, err := spec.RenderStep(wf.Current(), st.ArtifactPath, wf.NextStepName())
	if err != nil {
		return nil, err
	}

	return &Result{
		Step:           wf.Current(),
		TotalSteps:     len(steps),
		CompletedSteps: len(st.CompletedSteps),
		SpecPath:       st.ArtifactPath,
		SpecName:       st.Name,
		Instruction:    instruction,
	}, nil
}

func specGoto(stepName string) (*Result, error) {
	sp, err := specStatePath()
	if err != nil {
		return nil, err
	}

	steps := spec.Steps()
	wf := workflow.New(steps, sp)

	if err := wf.Goto(stepName); err != nil {
		return nil, err
	}

	st := wf.State()
	instruction, err := spec.RenderStep(wf.Current(), st.ArtifactPath, wf.NextStepName())
	if err != nil {
		return nil, err
	}

	return &Result{
		Step:           wf.Current(),
		TotalSteps:     len(steps),
		CompletedSteps: len(st.CompletedSteps),
		SpecPath:       st.ArtifactPath,
		SpecName:       st.Name,
		Instruction:    instruction,
	}, nil
}

func specStatus() (*StatusResult, error) {
	sp, err := specStatePath()
	if err != nil {
		return nil, err
	}

	steps := spec.Steps()
	wf := workflow.New(steps, sp)

	st := wf.State()
	infos := wf.StepStatus()

	entries := make([]StepStatusEntry, len(infos))
	for i, info := range infos {
		entries[i] = StepStatusEntry{Name: info.Name, Status: info.Status}
	}

	return &StatusResult{
		SpecName:       st.Name,
		SpecPath:       st.ArtifactPath,
		CurrentStep:    st.CurrentStep,
		CompletedSteps: st.CompletedSteps,
		TotalSteps:     len(steps),
		Progress:       fmt.Sprintf("%d/%d", len(st.CompletedSteps), len(steps)),
		Steps:          entries,
	}, nil
}

func boolCount(flags ...bool) int {
	n := 0
	for _, f := range flags {
		if f {
			n++
		}
	}
	return n
}

func init() {
	specCmd.Flags().Bool("new", false, "Create a new spec workflow")
	specCmd.Flags().Bool("next", false, "Advance to the next step")
	specCmd.Flags().StringP("step", "s", "", "Jump to a specific step (for revision)")
	specCmd.Flags().Bool("status", false, "Show current workflow progress")
	specCmd.Flags().StringP("data", "d", "", "JSON data for the step")
}
