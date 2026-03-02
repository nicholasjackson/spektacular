package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jumppad-labs/spektacular/internal/config"
	"github.com/jumppad-labs/spektacular/internal/spec"
	"github.com/jumppad-labs/spektacular/internal/steps"
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

			wf := steps.SpecCreatorWorkflow(name, cwd, cfg)
			specPath, err = tui.RunAgentTUI(wf, cwd, cfg)
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
