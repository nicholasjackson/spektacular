package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cbroglie/mustache"
	"github.com/jumppad-labs/spektacular/internal/project"
	"github.com/jumppad-labs/spektacular/templates"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init <claude|bob>",
	Short: "Initialise a Spektacular project for the specified agent",
	Args:  cobra.ExactArgs(1),
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	agent := args[0]
	if agent != "claude" && agent != "bob" {
		return fmt.Errorf("unknown agent %q: must be \"claude\" or \"bob\"", agent)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	if err := project.Init(cwd, true); err != nil {
		return fmt.Errorf("initialising project: %w", err)
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Store agent type in config.
	cfg.Agent = agent
	cfgPath := filepath.Join(cwd, ".spektacular", "config.yaml")
	if err := cfg.ToYAMLFile(cfgPath); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	commands := []struct {
		srcName    string
		claudeDest string
		bobDest    string
	}{
		{"spek-new.md", filepath.Join(".claude", "commands", "spek", "new.md"), filepath.Join(".bob", "commands", "spek-new.md")},
		{"spek-plan.md", filepath.Join(".claude", "commands", "spek", "plan.md"), filepath.Join(".bob", "commands", "spek-plan.md")},
		{"spek-implement.md", filepath.Join(".claude", "commands", "spek", "implement.md"), filepath.Join(".bob", "commands", "spek-implement.md")},
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Spektacular initialised for %s.\n", agent)
	fmt.Fprintf(cmd.OutOrStdout(), "  Project:  %s\n", filepath.Join(cwd, ".spektacular"))

	for _, c := range commands {
		tmplBytes, err := templates.FS.ReadFile("commands/" + c.srcName)
		if err != nil {
			return fmt.Errorf("reading embedded command template %s: %w", c.srcName, err)
		}

		rendered, err := mustache.Render(string(tmplBytes), map[string]string{"command": cfg.Command})
		if err != nil {
			return fmt.Errorf("rendering command template %s: %w", c.srcName, err)
		}

		var destPath string
		switch agent {
		case "claude":
			destPath = filepath.Join(cwd, c.claudeDest)
		case "bob":
			destPath = filepath.Join(cwd, c.bobDest)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("creating command directory: %w", err)
		}

		if err := os.WriteFile(destPath, []byte(rendered), 0644); err != nil {
			return fmt.Errorf("writing command template: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "  Command:  %s\n", destPath)
	}

	return nil
}

