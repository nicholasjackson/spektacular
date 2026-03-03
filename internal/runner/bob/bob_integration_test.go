package bob

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jumppad-labs/spektacular/internal/config"
	"github.com/jumppad-labs/spektacular/internal/runner"
	"github.com/stretchr/testify/require"
)

// writeMockBob creates a shell script that outputs the given JSONL lines to stdout.
// It returns the path to the executable script.
func writeMockBob(t *testing.T, jsonLines string) string {
	t.Helper()
	dir := t.TempDir()

	// Write JSONL data to a separate file to avoid shell escaping issues.
	dataFile := filepath.Join(dir, "output.jsonl")
	err := os.WriteFile(dataFile, []byte(jsonLines), 0644)
	require.NoError(t, err)

	// Write a shell script that cats the data file.
	scriptContent := fmt.Sprintf("#!/bin/sh\ncat '%s'\n", dataFile)
	scriptFile := filepath.Join(dir, "mockbob")
	err = os.WriteFile(scriptFile, []byte(scriptContent), 0755)
	require.NoError(t, err)

	return scriptFile
}

// runBob runs the Bob runner with the given mock script and collects all events.
func runBob(t *testing.T, scriptPath string, opts runner.RunOptions) ([]runner.Event, error) {
	t.Helper()
	opts.Config.Agent.Command = scriptPath
	b := New()
	events, errc := b.Run(opts)
	var got []runner.Event
	for e := range events {
		got = append(got, e)
	}
	return got, <-errc
}

// ---------------------------------------------------------------------------
// Integration tests
// ---------------------------------------------------------------------------

func TestIntegration_SimpleRun(t *testing.T) {
	jsonLines := strings.Join([]string{
		`{"type":"init","session_id":"int-test-session","model":"premium"}`,
		`{"type":"message","role":"user","content":"hello"}`,
		`{"type":"message","role":"assistant","content":"Hello! ","delta":true}`,
		`{"type":"message","role":"assistant","content":"How can I help?","delta":true}`,
		`{"type":"result","status":"success","stats":{}}`,
	}, "\n") + "\n"

	script := writeMockBob(t, jsonLines)
	events, err := runBob(t, script, runner.RunOptions{
		Config:  config.Config{},
		Prompts: runner.Prompts{User: "hello"},
	})

	require.NoError(t, err)
	// Expected: system event, assistant text event, result event
	require.Len(t, events, 3)
	require.Equal(t, "system", events[0].Type)
	require.Equal(t, "int-test-session", events[0].SessionID())
	require.Equal(t, "assistant", events[1].Type)
	require.Equal(t, "Hello! How can I help?", events[1].TextContent())
	require.Equal(t, "result", events[2].Type)
	require.False(t, events[2].IsError())
}

func TestIntegration_ToolUseFlow(t *testing.T) {
	jsonLines := strings.Join([]string{
		`{"type":"init","session_id":"tool-test-session","model":"premium"}`,
		`{"type":"message","role":"assistant","content":"I will read the file.","delta":true}`,
		`{"type":"tool_use","tool_name":"read_file","tool_id":"tool-1","parameters":{"file_path":"/README.md"}}`,
		`{"type":"tool_result","tool_id":"tool-1","status":"success","output":""}`,
		`{"type":"message","role":"assistant","content":"Done.","delta":true}`,
		`{"type":"result","status":"success","stats":{}}`,
	}, "\n") + "\n"

	script := writeMockBob(t, jsonLines)
	events, err := runBob(t, script, runner.RunOptions{
		Config:  config.Config{},
		Prompts: runner.Prompts{User: "read README"},
	})

	require.NoError(t, err)
	// Expected: system, text-before-tool, tool_use, text-after-tool, result
	require.Len(t, events, 5)
	require.Equal(t, "system", events[0].Type)

	// Text before tool use
	require.Equal(t, "assistant", events[1].Type)
	require.Equal(t, "I will read the file.", events[1].TextContent())

	// Tool use event
	require.Equal(t, "assistant", events[2].Type)
	tools := events[2].ToolUses()
	require.Len(t, tools, 1)
	require.Equal(t, "read_file", tools[0]["name"])

	// Text after tool use
	require.Equal(t, "assistant", events[3].Type)
	require.Equal(t, "Done.", events[3].TextContent())

	// Result
	require.Equal(t, "result", events[4].Type)
	require.False(t, events[4].IsError())
}

func TestIntegration_QuestionFlow(t *testing.T) {
	question := `<!--QUESTION:{"questions":[{"question":"Which approach?","header":"Approach","type":"choice","options":[{"label":"A","description":"Option A"}]}]}-->`
	jsonLines := strings.Join([]string{
		`{"type":"init","session_id":"question-session","model":"premium"}`,
		fmt.Sprintf(`{"type":"message","role":"assistant","content":"%s","delta":true}`, strings.ReplaceAll(question, `"`, `\"`)),
		`{"type":"result","status":"success","stats":{}}`,
	}, "\n") + "\n"

	script := writeMockBob(t, jsonLines)
	events, err := runBob(t, script, runner.RunOptions{
		Config:  config.Config{},
		Prompts: runner.Prompts{User: "plan this"},
	})

	require.NoError(t, err)
	// system, assistant (with question marker), result
	require.Len(t, events, 3)

	// The question should be detectable in the assistant text
	text := events[1].TextContent()
	questions := runner.DetectQuestions(text)
	require.Len(t, questions, 1)
	require.Equal(t, "Which approach?", questions[0].Question)
}

func TestIntegration_FinishedMarker(t *testing.T) {
	jsonLines := strings.Join([]string{
		`{"type":"init","session_id":"finished-session","model":"premium"}`,
		`{"type":"message","role":"assistant","content":"Work complete. <!-- FINISHED -->","delta":true}`,
		`{"type":"result","status":"success","stats":{}}`,
	}, "\n") + "\n"

	script := writeMockBob(t, jsonLines)
	events, err := runBob(t, script, runner.RunOptions{
		Config:  config.Config{},
		Prompts: runner.Prompts{User: "do work"},
	})

	require.NoError(t, err)
	require.Len(t, events, 3)

	text := events[1].TextContent()
	require.True(t, runner.DetectFinished(text))
}

func TestIntegration_ErrorResult(t *testing.T) {
	jsonLines := strings.Join([]string{
		`{"type":"init","session_id":"error-session","model":"premium"}`,
		`{"type":"result","status":"error","stats":{}}`,
	}, "\n") + "\n"

	script := writeMockBob(t, jsonLines)
	events, err := runBob(t, script, runner.RunOptions{
		Config:  config.Config{},
		Prompts: runner.Prompts{User: "do something"},
	})

	require.NoError(t, err)
	require.Len(t, events, 2)
	require.Equal(t, "result", events[1].Type)
	require.True(t, events[1].IsError())
}

func TestIntegration_SessionIDCaptured(t *testing.T) {
	jsonLines := strings.Join([]string{
		`{"type":"init","session_id":"captured-session-id","model":"premium"}`,
		`{"type":"result","status":"success","stats":{}}`,
	}, "\n") + "\n"

	script := writeMockBob(t, jsonLines)
	events, err := runBob(t, script, runner.RunOptions{
		Config:  config.Config{},
		Prompts: runner.Prompts{User: "hello"},
	})

	require.NoError(t, err)
	require.GreaterOrEqual(t, len(events), 1)
	require.Equal(t, "system", events[0].Type)
	require.Equal(t, "captured-session-id", events[0].SessionID())
}

func TestIntegration_ThinkingStrippedEndToEnd(t *testing.T) {
	jsonLines := strings.Join([]string{
		`{"type":"init","session_id":"thinking-session","model":"premium"}`,
		`{"type":"message","role":"assistant","content":"<thinking>\nLet me consider...\n</thinking>\nThe answer is 42.","delta":true}`,
		`{"type":"result","status":"success","stats":{}}`,
	}, "\n") + "\n"

	script := writeMockBob(t, jsonLines)
	events, err := runBob(t, script, runner.RunOptions{
		Config:  config.Config{},
		Prompts: runner.Prompts{User: "what is the answer?"},
	})

	require.NoError(t, err)
	require.Len(t, events, 3)

	text := events[1].TextContent()
	require.NotContains(t, text, "<thinking>")
	require.NotContains(t, text, "Let me consider...")
	require.Equal(t, "The answer is 42.", text)
}
