package agent

import (
	"io"

	"github.com/jumppad-labs/spektacular/internal/config"
)

type bobAgent struct{}

func (bobAgent) Name() string { return "bob" }

func (bobAgent) Install(projectPath string, cfg config.Config, out io.Writer) error {
	if err := installWorkflowSkills(projectPath, ".bob/skills", cfg, out); err != nil {
		return err
	}
	return installCommandWrappers(projectPath, ".bob/commands", bobCommandFilename, cfg, out)
}

func bobCommandFilename(skillName string) string {
	return skillName + ".md"
}

func init() {
	register(bobAgent{})
}
