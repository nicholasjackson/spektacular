package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit_Claude(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd.SetArgs([]string{"init", "claude"})
	err := rootCmd.Execute()
	require.NoError(t, err)

	// .spektacular directory created
	_, err = os.Stat(filepath.Join(dir, ".spektacular"))
	require.NoError(t, err)

	// command template written with default command name rendered
	destPath := filepath.Join(dir, ".claude", "commands", "spek", "new.md")
	got, err := os.ReadFile(destPath)
	require.NoError(t, err)
	require.Contains(t, string(got), "spektacular spec new")
	require.NotContains(t, string(got), "{{command}}")
}

func TestInit_Bob(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd.SetArgs([]string{"init", "bob"})
	err := rootCmd.Execute()
	require.NoError(t, err)

	// .spektacular directory created
	_, err = os.Stat(filepath.Join(dir, ".spektacular"))
	require.NoError(t, err)

	// command template written with default command name rendered
	destPath := filepath.Join(dir, ".bob", "commands", "spek-new.md")
	got, err := os.ReadFile(destPath)
	require.NoError(t, err)
	require.Contains(t, string(got), "spektacular spec new")
	require.NotContains(t, string(got), "{{command}}")
}

func TestInit_InvalidAgent(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd.SetArgs([]string{"init", "unknown"})
	err := rootCmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown agent")
}

func TestInit_CustomCommand(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	// First init to create .spektacular with default config
	rootCmd.SetArgs([]string{"init", "claude"})
	require.NoError(t, rootCmd.Execute())

	// Override the command in config
	configPath := filepath.Join(dir, ".spektacular", "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("command: \"go run .\"\n"), 0644))

	// Re-init — should use custom command
	rootCmd.SetArgs([]string{"init", "claude"})
	require.NoError(t, rootCmd.Execute())

	destPath := filepath.Join(dir, ".claude", "commands", "spek", "new.md")
	got, err := os.ReadFile(destPath)
	require.NoError(t, err)
	require.Contains(t, string(got), "go run . spec new")
	require.NotContains(t, string(got), "{{command}}")
}

func TestInit_Idempotent(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	// Create a sibling file that should survive re-init
	claudeDir := filepath.Join(dir, ".claude", "commands")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "other.md"), []byte("keep"), 0644))

	rootCmd.SetArgs([]string{"init", "claude"})
	require.NoError(t, rootCmd.Execute())

	rootCmd.SetArgs([]string{"init", "claude"})
	require.NoError(t, rootCmd.Execute())

	// sibling file still exists
	data, err := os.ReadFile(filepath.Join(claudeDir, "other.md"))
	require.NoError(t, err)
	require.Equal(t, "keep", string(data))
}
