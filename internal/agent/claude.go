package agent

import (
	"io"
	"strings"

	"github.com/jumppad-labs/spektacular/internal/config"
)

type claudeAgent struct{}

func (claudeAgent) Name() string { return "claude" }

func (claudeAgent) Install(projectPath string, cfg config.Config, out io.Writer) error {
	if err := installWorkflowSkills(projectPath, ".claude/skills", cfg, out); err != nil {
		return err
	}
	return installCommandWrappers(projectPath, ".claude/commands/spek", claudeCommandFilename, cfg, out)
}

func claudeCommandFilename(skillName string) string {
	return strings.TrimPrefix(skillName, "spek-") + ".md"
}

func init() {
	register(claudeAgent{})
}
