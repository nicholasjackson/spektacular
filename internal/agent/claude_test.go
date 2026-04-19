package agent

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/jumppad-labs/spektacular/internal/config"
	"github.com/stretchr/testify/require"
)

func TestClaudeAgent_Name(t *testing.T) {
	require.Equal(t, "claude", claudeAgent{}.Name())
}

func TestClaudeAgent_Install(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.NewDefault()

	err := claudeAgent{}.Install(tmp, cfg, io.Discard)
	require.NoError(t, err)

	// Exactly three SKILL.md files under .claude/skills/spek-{new,plan,implement}/.
	skillAssertions := map[string]string{
		"spek-new":       "spektacular spec new",
		"spek-plan":      "spektacular plan new",
		"spek-implement": "spektacular implement new",
	}
	for skill, expected := range skillAssertions {
		skillPath := filepath.Join(tmp, ".claude", "skills", skill, "SKILL.md")
		require.FileExists(t, skillPath)
		data, err := os.ReadFile(skillPath)
		require.NoError(t, err)
		require.Contains(t, string(data), expected)
		require.NotContains(t, string(data), "{{command}}")
	}

	// Exactly three command wrappers under .claude/commands/spek/, basenames
	// stripped of the `spek-` prefix.
	commandAssertions := map[string]string{
		"new.md":       "`spek-new` skill",
		"plan.md":      "`spek-plan` skill",
		"implement.md": "`spek-implement` skill",
	}
	for base, expected := range commandAssertions {
		cmdPath := filepath.Join(tmp, ".claude", "commands", "spek", base)
		require.FileExists(t, cmdPath)
		data, err := os.ReadFile(cmdPath)
		require.NoError(t, err)
		require.Contains(t, string(data), expected)
		require.NotContains(t, string(data), "{{command}}")
		require.NotContains(t, string(data), "{{skill}}")
	}

	// Claude command filenames strip the `spek-` prefix — make sure the
	// prefixed variants do NOT exist on disk.
	for _, prefixed := range []string{"spek-new.md", "spek-plan.md", "spek-implement.md"} {
		require.NoFileExists(t, filepath.Join(tmp, ".claude", "commands", "spek", prefixed))
	}

	// Each installed SKILL.md must have a valid frontmatter block that
	// satisfies the agentskills.io naming rules.
	for skill := range skillAssertions {
		validateSkillFrontmatter(t, filepath.Join(tmp, ".claude", "skills", skill, "SKILL.md"))
	}
}
