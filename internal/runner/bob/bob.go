// Package bob implements the Runner interface for the Bob CLI agent.
package bob

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jumppad-labs/spektacular/internal/runner"
)

func init() {
	runner.Register("bob", func() runner.Runner { return New() })
}

// DefaultModel is the default Bob model tier used when none is specified.
const DefaultModel = "premium"

// Models is the list of available Bob model tiers.
var Models = []string{"premium", "standard"}

var (
	thinkingPattern         = regexp.MustCompile(`(?s)<thinking>.*?</thinking>`)
	toolAnnouncementPattern = regexp.MustCompile(`\[using tool [^\]]+\]\n?`)
)

// Bob implements runner.Runner by spawning the Bob CLI subprocess.
type Bob struct{}

// New returns a new Bob runner.
func New() *Bob { return &Bob{} }

// Run spawns the bob subprocess and returns a channel of events and an error channel.
func (b *Bob) Run(opts runner.RunOptions) (<-chan runner.Event, <-chan error) {
	events := make(chan runner.Event, 64)
	errc := make(chan error, 1)

	go func() {
		defer close(events)
		if err := run(opts, events); err != nil {
			errc <- err
		}
		close(errc)
	}()

	return events, errc
}

// buildCmd constructs the Bob CLI command slice from RunOptions.
func buildCmd(opts runner.RunOptions) []string {
	cfg := opts.Config
	model := opts.Model
	if model == "" {
		model = DefaultModel
	}

	prompt := opts.Prompts.User
	if len(cfg.Agent.DisallowedTools) > 0 {
		prompt = disallowedToolsPrefix(cfg.Agent.DisallowedTools) + prompt
	}

	cmd := []string{cfg.Agent.Command, "-p", prompt, "-m", model}

	if cfg.Agent.DangerouslySkipPermissions {
		cmd = append(cmd, "-y")
	}

	cmd = append(cmd, cfg.Agent.Args...)

	if opts.SessionID != "" {
		cmd = append(cmd, "--resume", opts.SessionID)
	}

	return cmd
}

func run(opts runner.RunOptions, events chan<- runner.Event) error {
	cmd := buildCmd(opts)

	cwd := opts.CWD
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}
	}

	proc := exec.Command(cmd[0], cmd[1:]...) //nolint:gosec
	proc.Dir = cwd
	proc.Stderr = io.Discard

	stdout, err := proc.StdoutPipe()
	if err != nil {
		return fmt.Errorf("creating stdout pipe: %w", err)
	}
	if err := proc.Start(); err != nil {
		return fmt.Errorf("starting bob process: %w", err)
	}

	var reader io.Reader = stdout
	if opts.LogFile != "" {
		if mkdirErr := os.MkdirAll(filepath.Dir(opts.LogFile), 0755); mkdirErr == nil {
			if f, openErr := os.OpenFile(opts.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); openErr == nil {
				defer f.Close()
				if opts.SessionID == "" {
					fmt.Fprintf(f, "\n\n========== NEW SESSION: %s ==========\n", time.Now().Format("15:04:05"))
				}
				reader = io.TeeReader(stdout, f)
			}
		}
	}

	translate(reader, events)

	if err := proc.Wait(); err != nil {
		return fmt.Errorf("bob process exited with error: %w", err)
	}
	return nil
}

// translate reads Bob JSONL from r, translates each event into Spektacular's
// Event format, and sends translated events to the events channel.
// It blocks until r is exhausted.
func translate(r io.Reader, events chan<- runner.Event) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1 MiB lines

	var textBuf strings.Builder

	flushText := func() {
		if textBuf.Len() == 0 {
			return
		}
		text := stripContent(textBuf.String())
		textBuf.Reset()
		if text == "" {
			return
		}
		events <- runner.Event{
			Type: "assistant",
			Data: map[string]any{
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "text", "text": text},
					},
				},
			},
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var data map[string]any
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}
		eventType, _ := data["type"].(string)

		switch eventType {
		case "init":
			// Bob init → Spektacular system event with session_id.
			sessionID, _ := data["session_id"].(string)
			events <- runner.Event{
				Type: "system",
				Data: map[string]any{"session_id": sessionID},
			}

		case "message":
			// User messages are echoed input — skip them.
			// Assistant messages are streaming deltas — accumulate.
			role, _ := data["role"].(string)
			if role != "assistant" {
				continue
			}
			content, _ := data["content"].(string)
			textBuf.WriteString(content)

		case "tool_use":
			// Flush any accumulated assistant text before the tool event.
			flushText()
			toolName, _ := data["tool_name"].(string)
			params, _ := data["parameters"].(map[string]any)
			events <- runner.Event{
				Type: "assistant",
				Data: map[string]any{
					"message": map[string]any{
						"content": []any{
							map[string]any{
								"type":  "tool_use",
								"name":  toolName,
								"input": params,
							},
						},
					},
				},
			}

		case "tool_result":
			// Tool results are internal to Bob; not forwarded to the TUI.

		case "result":
			// Flush any remaining accumulated text before the terminal event.
			flushText()
			status, _ := data["status"].(string)
			isError := status == "error"
			resultText := "completed"
			if isError {
				resultText = "error"
			}
			events <- runner.Event{
				Type: "result",
				Data: map[string]any{
					"result":   resultText,
					"is_error": isError,
				},
			}
		}
	}

	// Flush any remaining accumulated text after stream ends.
	flushText()
}

// disallowedToolsPrefix returns a prompt prefix instructing the agent to avoid
// specific tools. Bob CLI has no --disallowedTools flag, so we inject this as
// a prompt-level instruction instead.
func disallowedToolsPrefix(tools []string) string {
	return fmt.Sprintf("IMPORTANT: Do NOT use the following tools: %s. "+
		"Output text directly in your response instead.\n\n", strings.Join(tools, ", "))
}

// stripContent removes <thinking>...</thinking> blocks and [using tool ...] tool
// announcement lines from assistant text before forwarding to the TUI.
func stripContent(text string) string {
	text = thinkingPattern.ReplaceAllString(text, "")
	text = toolAnnouncementPattern.ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}
