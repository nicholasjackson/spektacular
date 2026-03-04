package workflow

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/looplab/fsm"
)

// StepConfig defines a single step in a workflow.
// Name is the event name (and step identifier).
// Src lists valid source states. Dst is the destination state.
type StepConfig struct {
	Name     string
	Src      []string
	Dst      string
	Callback fsm.Callback
}

// Workflow is a linear state machine with persistence.
type Workflow struct {
	FSM       *fsm.FSM
	steps     []StepConfig
	state     *State
	statePath string
}

// New builds a workflow FSM from step configs.
// If a state file exists at statePath, its CurrentStep is used as the initial
// FSM state. Otherwise the initial state defaults to steps[0].Src[0].
// An implicit "done" transition is appended after the last step.
// State is automatically persisted on every transition.
func New(steps []StepConfig, statePath string) *Workflow {
	events := make([]fsm.EventDesc, 0, len(steps)+1)
	callbacks := make(fsm.Callbacks)

	for _, s := range steps {
		events = append(events, fsm.EventDesc{
			Name: s.Name,
			Src:  s.Src,
			Dst:  s.Dst,
		})
		if s.Callback != nil {
			callbacks["after_"+s.Name] = s.Callback
		}
	}

	// Implicit final transition to "done".
	if len(steps) > 0 {
		last := steps[len(steps)-1]
		events = append(events, fsm.EventDesc{
			Name: "done",
			Src:  []string{last.Dst},
			Dst:  "done",
		})
	}

	// Load existing state or create new.
	var state *State
	var initialState string
	if s, err := loadState(statePath); err == nil {
		state = s
		initialState = s.CurrentStep
	} else {
		initialState = steps[0].Src[0]
		now := time.Now().UTC()
		state = &State{
			CurrentStep:    initialState,
			CompletedSteps: []string{},
			CreatedAt:      now,
			UpdatedAt:      now,
		}
	}

	w := &Workflow{
		steps:     steps,
		state:     state,
		statePath: statePath,
	}

	// Auto-update and persist state on every transition.
	callbacks["enter_state"] = func(_ context.Context, e *fsm.Event) {
		if w.validStep(e.Src) {
			w.state.markCompleted(e.Src)
		}
		w.state.CurrentStep = e.Dst
		w.state.UpdatedAt = time.Now().UTC()
		_ = saveState(w.statePath, w.state)
	}

	w.FSM = fsm.NewFSM(initialState, events, callbacks)
	return w
}

// Next fires the first available transition.
func (w *Workflow) Next() error {
	transitions := w.FSM.AvailableTransitions()
	if len(transitions) == 0 {
		return fmt.Errorf("workflow is already complete")
	}
	return w.FSM.Event(context.Background(), transitions[0])
}

// Goto jumps to a named step by firing the corresponding FSM event.
// The step's Src list must include the current state.
func (w *Workflow) Goto(name string) error {
	if w.Current() == name {
		return nil
	}

	if !w.validStep(name) {
		return fmt.Errorf("unknown step %q", name)
	}

	return w.FSM.Event(context.Background(), name)
}

// Current returns the current state name.
func (w *Workflow) Current() string {
	return w.FSM.Current()
}

// State returns the persisted state.
func (w *Workflow) State() *State {
	return w.state
}

// IsComplete returns true if the workflow has reached "done".
func (w *Workflow) IsComplete() bool {
	return w.Current() == "done"
}

// StepNames returns the ordered step names from the config.
func (w *Workflow) StepNames() []string {
	names := make([]string, len(w.steps))
	for i, s := range w.steps {
		names[i] = s.Name
	}
	return names
}

// NextStepName returns the name of the next step, or "" if at the end.
func (w *Workflow) NextStepName() string {
	cur := w.Current()
	for i, s := range w.steps {
		if s.Dst == cur && i+1 < len(w.steps) {
			return w.steps[i+1].Name
		}
	}
	return ""
}

// StepStatus returns per-step status info.
func (w *Workflow) StepStatus() []StepInfo {
	infos := make([]StepInfo, len(w.steps))
	for i, s := range w.steps {
		status := "pending"
		if slices.Contains(w.state.CompletedSteps, s.Name) {
			status = "completed"
		}
		if s.Dst == w.Current() {
			status = "current"
		}
		infos[i] = StepInfo{Name: s.Name, Status: status}
	}
	return infos
}

// StepInfo is a step with its current status.
type StepInfo struct {
	Name   string
	Status string // "pending", "current", "completed"
}

func (w *Workflow) validStep(name string) bool {
	for _, s := range w.steps {
		if s.Name == name {
			return true
		}
	}
	return false
}
