package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jumppad-labs/spektacular/internal/config"
	"github.com/jumppad-labs/spektacular/internal/implement"
	"github.com/jumppad-labs/spektacular/internal/runner"
	"github.com/jumppad-labs/spektacular/internal/tui"
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
