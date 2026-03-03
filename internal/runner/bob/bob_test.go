package bob

import (
	"strings"
	"testing"

	"github.com/jumppad-labs/spektacular/internal/config"
	"github.com/jumppad-labs/spektacular/internal/runner"
	"github.com/stretchr/testify/require"
)

// Compile-time check that Bob implements runner.Runner.
var _ runner.Runner = (*Bob)(nil)

func TestNew_ReturnsNonNil(t *testing.T) {
	b := New()
	require.NotNil(t, b)
}

// ---------------------------------------------------------------------------
// Model and constants
// ---------------------------------------------------------------------------

func TestDefaultModel(t *testing.T) {
	require.Equal(t, "premium", DefaultModel)
}

func TestModels_ContainsExpected(t *testing.T) {
	require.Contains(t, Models, "premium")
	require.Contains(t, Models, "standard")
}

// ---------------------------------------------------------------------------
// Translation helpers
// ---------------------------------------------------------------------------

// collectEvents runs translate over a JSONL string and returns all emitted events.
func collectEvents(jsonLines string) []runner.Event {
	events := make(chan runner.Event, 100)
	translate(strings.NewReader(jsonLines), events)
	close(events)
	var got []runner.Event
	for e := range events {
		got = append(got, e)
	}
	return got
}

// testConfig returns a minimal Config with the given agent command.
func testConfig(cmd string) config.Config {
	cfg := config.Config{}
	cfg.Agent.Command = cmd
	return cfg
}

// indexOf returns the index of s in slice, or -1.
func indexOf(slice []string, s string) int {
	for i, v := range slice {
		if v == s {
			return i
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// Translation tests
// ---------------------------------------------------------------------------

func TestTranslateInitEvent(t *testing.T) {
	input := `{"type":"init","session_id":"test-session-123","model":"premium"}` + "\n"
	events := collectEvents(input)
	require.Len(t, events, 1)
	require.Equal(t, "system", events[0].Type)
	require.Equal(t, "test-session-123", events[0].Data["session_id"])
}

func TestSessionID_ExtractedFromInit(t *testing.T) {
	input := `{"type":"init","session_id":"abc-def-ghi","model":"premium"}` + "\n"
	events := collectEvents(input)
	require.Len(t, events, 1)
	require.Equal(t, "abc-def-ghi", events[0].SessionID())
}

func TestTranslateAssistantDeltas(t *testing.T) {
	input := strings.Join([]string{
		`{"type":"message","role":"assistant","content":"Hello ","delta":true}`,
		`{"type":"message","role":"assistant","content":"world","delta":true}`,
		`{"type":"result","status":"success"}`,
	}, "\n") + "\n"
	events := collectEvents(input)
	// assistant text is flushed on result → two events total
	require.Len(t, events, 2)
	require.Equal(t, "assistant", events[0].Type)
	require.Equal(t, "Hello world", events[0].TextContent())
	require.Equal(t, "result", events[1].Type)
}

func TestUserMessage_Skipped(t *testing.T) {
	input := `{"type":"message","role":"user","content":"Hello Bob"}` + "\n"
	events := collectEvents(input)
	require.Empty(t, events)
}

func TestTranslateToolUse(t *testing.T) {
	input := `{"type":"tool_use","tool_name":"read_file","tool_id":"tool-1","parameters":{"file_path":"/README.md"}}` + "\n"
	events := collectEvents(input)
	require.Len(t, events, 1)
	require.Equal(t, "assistant", events[0].Type)
	tools := events[0].ToolUses()
	require.Len(t, tools, 1)
	require.Equal(t, "tool_use", tools[0]["type"])
	require.Equal(t, "read_file", tools[0]["name"])
	input_, _ := tools[0]["input"].(map[string]any)
	require.Equal(t, "/README.md", input_["file_path"])
}

func TestTranslateToolResult_Skipped(t *testing.T) {
	input := `{"type":"tool_result","tool_id":"tool-1","status":"success","output":""}` + "\n"
	events := collectEvents(input)
	require.Empty(t, events)
}

func TestTranslateResultSuccess(t *testing.T) {
	input := `{"type":"result","status":"success","stats":{}}` + "\n"
	events := collectEvents(input)
	require.Len(t, events, 1)
	require.Equal(t, "result", events[0].Type)
	require.False(t, events[0].IsError())
	require.Equal(t, "completed", events[0].ResultText())
}

func TestTranslateResultError(t *testing.T) {
	input := `{"type":"result","status":"error","stats":{}}` + "\n"
	events := collectEvents(input)
	require.Len(t, events, 1)
	require.Equal(t, "result", events[0].Type)
	require.True(t, events[0].IsError())
	require.Equal(t, "error", events[0].ResultText())
}

func TestThinkingBlocksStripped(t *testing.T) {
	input := strings.Join([]string{
		`{"type":"message","role":"assistant","content":"<thinking>\n","delta":true}`,
		`{"type":"message","role":"assistant","content":"Let me analyze this carefully.","delta":true}`,
		`{"type":"message","role":"assistant","content":"\n</thinking>\n","delta":true}`,
		`{"type":"message","role":"assistant","content":"Hello!","delta":true}`,
		`{"type":"result","status":"success"}`,
	}, "\n") + "\n"
	events := collectEvents(input)

	var assistantEvents []runner.Event
	for _, e := range events {
		if e.Type == "assistant" {
			assistantEvents = append(assistantEvents, e)
		}
	}
	require.Len(t, assistantEvents, 1)
	text := assistantEvents[0].TextContent()
	require.NotContains(t, text, "<thinking>")
	require.NotContains(t, text, "Let me analyze this carefully.")
	require.Contains(t, text, "Hello!")
}

func TestToolAnnouncementStripped(t *testing.T) {
	input := strings.Join([]string{
		`{"type":"message","role":"assistant","content":"[using tool read_file: README.md]\n","delta":true}`,
		`{"type":"message","role":"assistant","content":"Here is the content.","delta":true}`,
		`{"type":"result","status":"success"}`,
	}, "\n") + "\n"
	events := collectEvents(input)

	var assistantEvents []runner.Event
	for _, e := range events {
		if e.Type == "assistant" {
			assistantEvents = append(assistantEvents, e)
		}
	}
	require.Len(t, assistantEvents, 1)
	text := assistantEvents[0].TextContent()
	require.NotContains(t, text, "[using tool")
	require.Contains(t, text, "Here is the content.")
}

func TestFlushOnToolUse(t *testing.T) {
	input := strings.Join([]string{
		`{"type":"message","role":"assistant","content":"Before tool.","delta":true}`,
		`{"type":"tool_use","tool_name":"read_file","tool_id":"tool-1","parameters":{}}`,
	}, "\n") + "\n"
	events := collectEvents(input)
	// accumulated text is flushed before tool_use → two events
	require.Len(t, events, 2)
	require.Equal(t, "assistant", events[0].Type)
	require.Equal(t, "Before tool.", events[0].TextContent())
	require.Equal(t, "assistant", events[1].Type)
	tools := events[1].ToolUses()
	require.Len(t, tools, 1)
	require.Equal(t, "read_file", tools[0]["name"])
}

func TestFlushOnResult(t *testing.T) {
	input := strings.Join([]string{
		`{"type":"message","role":"assistant","content":"Before result.","delta":true}`,
		`{"type":"result","status":"success"}`,
	}, "\n") + "\n"
	events := collectEvents(input)
	// accumulated text is flushed before result → two events
	require.Len(t, events, 2)
	require.Equal(t, "assistant", events[0].Type)
	require.Equal(t, "Before result.", events[0].TextContent())
	require.Equal(t, "result", events[1].Type)
}

// ---------------------------------------------------------------------------
// Command building tests
// ---------------------------------------------------------------------------

func TestDefaultModel_UsedWhenEmpty(t *testing.T) {
	opts := runner.RunOptions{
		Config:  testConfig("bob"),
		Prompts: runner.Prompts{User: "hello"},
	}
	cmd := buildCmd(opts)
	idx := indexOf(cmd, "-m")
	require.Greater(t, idx, -1, "expected -m flag")
	require.Equal(t, "premium", cmd[idx+1])
}

func TestModelOverride_UsedWhenSet(t *testing.T) {
	opts := runner.RunOptions{
		Config:  testConfig("bob"),
		Prompts: runner.Prompts{User: "hello"},
		Model:   "standard",
	}
	cmd := buildCmd(opts)
	idx := indexOf(cmd, "-m")
	require.Greater(t, idx, -1, "expected -m flag")
	require.Equal(t, "standard", cmd[idx+1])
}

func TestResume_IncludedInCommand(t *testing.T) {
	opts := runner.RunOptions{
		Config:    testConfig("bob"),
		Prompts:   runner.Prompts{User: "answer"},
		SessionID: "session-abc-123",
	}
	cmd := buildCmd(opts)
	idx := indexOf(cmd, "--resume")
	require.Greater(t, idx, -1, "expected --resume flag")
	require.Equal(t, "session-abc-123", cmd[idx+1])
}

func TestResume_OmittedWhenEmpty(t *testing.T) {
	opts := runner.RunOptions{
		Config:  testConfig("bob"),
		Prompts: runner.Prompts{User: "hello"},
	}
	cmd := buildCmd(opts)
	require.NotContains(t, cmd, "--resume")
}

func TestOutputFormat_FromConfigArgs(t *testing.T) {
	cfg := testConfig("bob")
	cfg.Agent.Args = []string{"--output-format", "stream-json"}
	opts := runner.RunOptions{
		Config:  cfg,
		Prompts: runner.Prompts{User: "hello"},
	}
	cmd := buildCmd(opts)
	require.Contains(t, cmd, "--output-format")
	idx := indexOf(cmd, "--output-format")
	require.Equal(t, "stream-json", cmd[idx+1])
}

func TestOutputFormat_NotDuplicated(t *testing.T) {
	cfg := testConfig("bob")
	cfg.Agent.Args = []string{"--output-format", "stream-json"}
	opts := runner.RunOptions{
		Config:  cfg,
		Prompts: runner.Prompts{User: "hello"},
	}
	cmd := buildCmd(opts)
	count := 0
	for _, arg := range cmd {
		if arg == "--output-format" {
			count++
		}
	}
	require.Equal(t, 1, count, "expected exactly one --output-format flag")
}

func TestDisallowedTools_PrefixedToPrompt(t *testing.T) {
	cfg := testConfig("bob")
	cfg.Agent.DisallowedTools = []string{"ask_followup_question"}
	opts := runner.RunOptions{
		Config:  cfg,
		Prompts: runner.Prompts{User: "hello"},
	}
	cmd := buildCmd(opts)
	idx := indexOf(cmd, "-p")
	require.Greater(t, idx, -1, "expected -p flag")
	prompt := cmd[idx+1]
	require.Contains(t, prompt, "Do NOT use the following tools: ask_followup_question")
	require.True(t, strings.HasSuffix(prompt, "hello"))
}

func TestDisallowedTools_EmptyNoPrefix(t *testing.T) {
	opts := runner.RunOptions{
		Config:  testConfig("bob"),
		Prompts: runner.Prompts{User: "hello"},
	}
	cmd := buildCmd(opts)
	idx := indexOf(cmd, "-p")
	require.Greater(t, idx, -1, "expected -p flag")
	require.Equal(t, "hello", cmd[idx+1])
}

// ---------------------------------------------------------------------------
// Strip content unit tests
// ---------------------------------------------------------------------------

func TestStripContent_ThinkingBlock(t *testing.T) {
	input := "<thinking>\nLet me think carefully.\n</thinking>\nHello!"
	result := stripContent(input)
	require.Equal(t, "Hello!", result)
}

func TestStripContent_ToolAnnouncement(t *testing.T) {
	input := "[using tool read_file: README.md]\nHere is the content."
	result := stripContent(input)
	require.Equal(t, "Here is the content.", result)
}

func TestStripContent_BothThinkingAndAnnouncement(t *testing.T) {
	input := "<thinking>\nSome thoughts.\n</thinking>\n[using tool write_file: out.txt]\nDone."
	result := stripContent(input)
	require.Equal(t, "Done.", result)
}

func TestStripContent_OnlyThinkingBlock_EmptyResult(t *testing.T) {
	input := "<thinking>\nOnly thoughts.\n</thinking>"
	result := stripContent(input)
	require.Equal(t, "", result)
}
