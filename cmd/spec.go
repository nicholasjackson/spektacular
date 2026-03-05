package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/jumppad-labs/spektacular/internal/output"
	"github.com/jumppad-labs/spektacular/internal/spec"
	"github.com/jumppad-labs/spektacular/internal/store"
	"github.com/jumppad-labs/spektacular/internal/workflow"
	"github.com/spf13/cobra"
)

var nameRegexp = regexp.MustCompile(`^[a-z0-9_-]+$`)

// Schema types for --schema output.
type schemaProp struct {
	Type    string      `json:"type"`
	Enum    []string    `json:"enum,omitempty"`
	Pattern string      `json:"pattern,omitempty"`
	MaxLen  int         `json:"maxLength,omitempty"`
	Items   *schemaProp `json:"items,omitempty"`
}

type schemaObj struct {
	Type       string                 `json:"type"`
	Properties map[string]*schemaProp `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

type commandSchema struct {
	Input  *schemaObj `json:"input"`
	Output *schemaObj `json:"output"`
}

var resultOutputSchema = &schemaObj{
	Type: "object",
	Properties: map[string]*schemaProp{
		"step":        {Type: "string"},
		"spec_path":   {Type: "string"},
		"spec_name":   {Type: "string"},
		"instruction": {Type: "string"},
	},
}

var statusOutputSchema = &schemaObj{
	Type: "object",
	Properties: map[string]*schemaProp{
		"spec_name":       {Type: "string"},
		"spec_path":       {Type: "string"},
		"current_step":    {Type: "string"},
		"completed_steps": {Type: "array", Items: &schemaProp{Type: "string"}},
		"total_steps":     {Type: "integer"},
		"progress":        {Type: "string"},
		"steps":           {Type: "array"},
	},
}

var specCmd = &cobra.Command{
	Use:   "spec",
	Short: "Manage spec workflow",
}

var specNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new spec workflow",
	RunE:  runSpecNew,
}

var specGotoCmd = &cobra.Command{
	Use:   "goto",
	Short: "Jump to a named step",
	RunE:  runSpecGoto,
}

var specStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current workflow progress",
	RunE:  runSpecStatus,
}

var specStepsCmd = &cobra.Command{
	Use:   "steps",
	Short: "List available workflow step names",
	RunE:  runSpecSteps,
}

func stateFilePath(dataDir string) string {
	return filepath.Join(dataDir, "state.json")
}

// readStdinIntoWorkflow reads all of stdin and stores it in the workflow data
// under stdinKey, if stdinKey is non-empty.
func readStdinIntoWorkflow(cmd *cobra.Command, wf interface{ SetData(string, any) }, stdinKey string) error {
	if stdinKey == "" {
		return nil
	}
	content, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}
	wf.SetData(stdinKey, string(content))
	return nil
}

func runSpecNew(cmd *cobra.Command, _ []string) error {
	if schema, _ := cmd.Flags().GetBool("schema"); schema {
		s := commandSchema{
			Input: &schemaObj{
				Type: "object",
				Properties: map[string]*schemaProp{
					"name": {Type: "string", Pattern: "^[a-z0-9_-]+$", MaxLen: 64},
				},
				Required: []string{"name"},
			},
			Output: resultOutputSchema,
		}
		return output.Write(cmd.OutOrStdout(), s, "")
	}

	dataStr, _ := cmd.Flags().GetString("data")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if dataStr == "" {
		return fmt.Errorf("--data is required (e.g. --data '{\"name\":\"my-feature\"}')")
	}
	var input struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(dataStr), &input); err != nil {
		return fmt.Errorf("parsing --data: %w", err)
	}
	if input.Name == "" || !nameRegexp.MatchString(input.Name) || len(input.Name) > 64 {
		return fmt.Errorf("name must match ^[a-z0-9_-]+$ and be at most 64 characters")
	}

	dataDir, err := specDataDir()
	if err != nil {
		return err
	}
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	statePath := stateFilePath(dataDir)
	if dryRun {
		statePath += ".dryrun-tmp"
	} else {
		_ = os.Remove(statePath)
	}

	wfCfg := workflow.Config{Command: cfg.Command, DryRun: dryRun}
	steps := spec.Steps()
	wf := workflow.New(steps, statePath, wfCfg, store.NewFileStore(dataDir))
	wf.SetData("name", input.Name)

	stdinKey, _ := cmd.Flags().GetString("stdin")
	if err := readStdinIntoWorkflow(cmd, wf, stdinKey); err != nil {
		return err
	}

	// Advance through the internal "new" step (creates spec file, no output).
	if err := wf.Next(); err != nil {
		return output.WriteError(cmd.ErrOrStderr(), err)
	}

	// Advance to "overview" and write the instruction.
	out := output.New(cmd.OutOrStdout(), globalFields)
	if err := wf.Next(out); err != nil {
		return output.WriteError(cmd.ErrOrStderr(), err)
	}
	return nil
}

func runSpecGoto(cmd *cobra.Command, _ []string) error {
	if schema, _ := cmd.Flags().GetBool("schema"); schema {
		s := commandSchema{
			Input: &schemaObj{
				Type: "object",
				Properties: map[string]*schemaProp{
					"step": {Type: "string", Enum: workflow.New(spec.Steps(), "", workflow.Config{}, nil).StepNames()},
				},
				Required: []string{"step"},
			},
			Output: resultOutputSchema,
		}
		return output.Write(cmd.OutOrStdout(), s, "")
	}

	dataStr, _ := cmd.Flags().GetString("data")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if dataStr == "" {
		return fmt.Errorf("--data is required (e.g. --data '{\"step\":\"requirements\"}')")
	}
	var input map[string]any
	if err := json.Unmarshal([]byte(dataStr), &input); err != nil {
		return fmt.Errorf("parsing --data: %w", err)
	}
	stepVal, _ := input["step"].(string)
	if stepVal == "" {
		return fmt.Errorf("\"step\" is required in --data")
	}

	dataDir, err := specDataDir()
	if err != nil {
		return err
	}
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	wfCfg := workflow.Config{Command: cfg.Command, DryRun: dryRun}
	steps := spec.Steps()
	wf := workflow.New(steps, stateFilePath(dataDir), wfCfg, store.NewFileStore(dataDir))

	for k, v := range input {
		if k != "step" {
			wf.SetData(k, v)
		}
	}

	stdinKey, _ := cmd.Flags().GetString("stdin")
	if err := readStdinIntoWorkflow(cmd, wf, stdinKey); err != nil {
		return err
	}

	out := output.New(cmd.OutOrStdout(), globalFields)
	if err := wf.Goto(stepVal, out); err != nil {
		return output.WriteError(cmd.ErrOrStderr(), err)
	}
	return nil
}

func runSpecStatus(cmd *cobra.Command, _ []string) error {
	if schema, _ := cmd.Flags().GetBool("schema"); schema {
		s := commandSchema{Input: nil, Output: statusOutputSchema}
		return output.Write(cmd.OutOrStdout(), s, "")
	}

	dataDir, err := specDataDir()
	if err != nil {
		return err
	}

	steps := spec.Steps()
	wf := workflow.New(steps, stateFilePath(dataDir), workflow.Config{}, nil)
	st := wf.State()

	specName, _ := wf.GetData("name")
	specPath := filepath.Join(dataDir, spec.SpecFilePath(fmt.Sprintf("%v", specName)))

	stepInfos := wf.StepStatus()
	entries := make([]spec.StepEntry, len(stepInfos))
	for i, info := range stepInfos {
		entries[i] = spec.StepEntry{Name: info.Name, Status: info.Status}
	}

	out := output.New(cmd.OutOrStdout(), globalFields)
	return out.WriteResult(spec.StatusResult{
		SpecName:       fmt.Sprintf("%v", specName),
		SpecPath:       specPath,
		CurrentStep:    wf.Current(),
		CompletedSteps: st.CompletedSteps,
		TotalSteps:     len(steps),
		Progress:       fmt.Sprintf("%d/%d", len(st.CompletedSteps), len(steps)),
		Steps:          entries,
	})
}

func runSpecSteps(cmd *cobra.Command, _ []string) error {
	if schema, _ := cmd.Flags().GetBool("schema"); schema {
		s := commandSchema{
			Input: nil,
			Output: &schemaObj{
				Type: "object",
				Properties: map[string]*schemaProp{
					"steps": {Type: "array", Items: &schemaProp{Type: "string"}},
				},
			},
		}
		return output.Write(cmd.OutOrStdout(), s, "")
	}

	wf := workflow.New(spec.Steps(), "", workflow.Config{}, nil)
	out := output.New(cmd.OutOrStdout(), globalFields)
	return out.WriteResult(spec.StepsResult{Steps: wf.StepNames()})
}

func init() {
	specCmd.PersistentFlags().Bool("schema", false, "Print the input/output schema for this subcommand and exit")
	specCmd.PersistentFlags().BoolP("dry-run", "n", false, "Validate and preview without writing any files or persisting state")

	specNewCmd.Flags().StringP("data", "d", "", `JSON input (e.g. '{"name":"my-feature"}')`)
	specNewCmd.Flags().String("stdin", "", "Read stdin and store it in workflow data under this key")
	specGotoCmd.Flags().StringP("data", "d", "", `JSON input (e.g. '{"step":"requirements"}')`)
	specGotoCmd.Flags().String("stdin", "", "Read stdin and store it in workflow data under this key")

	specCmd.AddCommand(specNewCmd, specGotoCmd, specStatusCmd, specStepsCmd)
}
