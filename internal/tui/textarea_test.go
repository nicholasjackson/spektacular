package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jumppad-labs/spektacular/internal/config"
	"github.com/jumppad-labs/spektacular/internal/runner"
	"github.com/stretchr/testify/require"
)

func TestTextareaInit(t *testing.T) {
	m := &model{width: 80, height: 24}
	m.initTextarea("Test placeholder")

	require.True(t, m.textareaActive)
	require.Equal(t, "Test placeholder", m.textarea.Placeholder)
}

func TestTextareaSubmit(t *testing.T) {
	m := model{
		textareaActive: true,
		questions: []runner.Question{
			{Question: "Test?", Header: "Test"},
			{Question: "More?", Header: "Next"},
		},
		width:    80,
		height:   24,
		workflow: Workflow{StatusLabel: "Test"},
		cfg:      config.NewDefault(),
	}
	m.initTextarea("placeholder")
	m.textarea.SetValue("My multi-line\nanswer")

	// Simulate Ctrl+D
	msg := tea.KeyMsg{Type: tea.KeyCtrlD}
	newModel, _ := m.handleTextareaInput(msg)

	m2 := newModel.(model)
	require.False(t, m2.textareaActive)
	require.Len(t, m2.answers, 1)
	require.Equal(t, "My multi-line\nanswer", m2.answers[0])
	// Should still have one question remaining
	require.Len(t, m2.questions, 1)
}

func TestTextareaCancel(t *testing.T) {
	m := model{textareaActive: true, width: 80, height: 24}
	m.initTextarea("placeholder")
	m.textarea.SetValue("Some text")

	// Simulate Esc
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := m.handleTextareaInput(msg)

	m2 := newModel.(model)
	require.False(t, m2.textareaActive)
	require.Empty(t, m2.answers)
}

func TestJoinAnswers(t *testing.T) {
	answers := []string{"First answer", "Second answer", "Third answer"}
	result := joinAnswers(answers)

	require.Contains(t, result, "First answer")
	require.Contains(t, result, "Second answer")
	require.Contains(t, result, "Third answer")
	require.Contains(t, result, "---")
}
