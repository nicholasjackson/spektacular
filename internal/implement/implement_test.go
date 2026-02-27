package implement

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadAgentPrompt_ReturnsContent(t *testing.T) {
	content := LoadAgentPrompt()
	require.NotEmpty(t, content)
	require.Contains(t, content, "Execution Agent")
}

func TestLoadPlanContent_RequiresPlanMD(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadPlanContent(dir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "plan.md not found")
}

func TestLoadPlanContent_IncludesAllFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.md"), []byte("# Plan"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "context.md"), []byte("# Context"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "research.md"), []byte("# Research"), 0644))

	content, err := LoadPlanContent(dir)
	require.NoError(t, err)
	require.Contains(t, content, "# Plan")
	require.Contains(t, content, "# Context")
	require.Contains(t, content, "# Research")
}

func TestLoadPlanContent_OptionalFilesSkipped(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.md"), []byte("# Plan Only"), 0644))

	content, err := LoadPlanContent(dir)
	require.NoError(t, err)
	require.Contains(t, content, "# Plan Only")
	require.NotContains(t, content, "context.md")
	require.NotContains(t, content, "research.md")
}

func TestResolvePlanDir_DirectPath(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plan.md"), []byte("plan"), 0644))

	resolved, err := ResolvePlanDir(dir, "/tmp")
	require.NoError(t, err)
	require.Equal(t, dir, resolved)
}

func TestResolvePlanDir_PlanName(t *testing.T) {
	cwd := t.TempDir()
	planDir := filepath.Join(cwd, ".spektacular", "plans", "my-feature")
	require.NoError(t, os.MkdirAll(planDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(planDir, "plan.md"), []byte("plan"), 0644))

	resolved, err := ResolvePlanDir("my-feature", cwd)
	require.NoError(t, err)
	require.Equal(t, planDir, resolved)
}

func TestResolvePlanDir_NotFound(t *testing.T) {
	_, err := ResolvePlanDir("nonexistent", t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "plan.md not found")
}
