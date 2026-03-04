package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"
)

// State is the persisted progress of a workflow.
type State struct {
	Name           string    `json:"name"`
	ArtifactPath   string    `json:"artifact_path"`
	CurrentStep    string    `json:"current_step"`
	CompletedSteps []string  `json:"completed_steps"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (s *State) markCompleted(step string) {
	if !slices.Contains(s.CompletedSteps, step) {
		s.CompletedSteps = append(s.CompletedSteps, step)
	}
}

func loadState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}
	return &s, nil
}

func saveState(path string, s *State) error {
	s.UpdatedAt = time.Now().UTC()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}
