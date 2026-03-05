package workflow

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var testSteps = []StepConfig{
	{Name: "one", Src: []string{"new"}, Dst: "one"},
	{Name: "two", Src: []string{"one"}, Dst: "two"},
	{Name: "three", Src: []string{"two"}, Dst: "three"},
}

func TestNew(t *testing.T) {
	sp := filepath.Join(t.TempDir(), "state.json")
	wf := New(testSteps, sp, Config{}, nil)

	require.Equal(t, "new", wf.Current())
	require.False(t, wf.IsComplete())
}

func TestNextAdvancesThroughAllSteps(t *testing.T) {
	sp := filepath.Join(t.TempDir(), "state.json")
	wf := New(testSteps, sp, Config{}, nil)

	err := wf.Next() // new → one
	require.NoError(t, err)
	require.Equal(t, "one", wf.Current())

	err = wf.Next() // one → two
	require.NoError(t, err)
	require.Equal(t, "two", wf.Current())

	err = wf.Next() // two → three
	require.NoError(t, err)
	require.Equal(t, "three", wf.Current())

	err = wf.Next() // three → done
	require.NoError(t, err)
	require.True(t, wf.IsComplete())
}

func TestNextOnCompleteErrors(t *testing.T) {
	sp := filepath.Join(t.TempDir(), "state.json")
	wf := New(testSteps, sp, Config{}, nil)

	for i := 0; i <= len(testSteps); i++ {
		err := wf.Next()
		require.NoError(t, err)
	}

	err := wf.Next()
	require.Error(t, err)
}

func TestGotoForward(t *testing.T) {
	sp := filepath.Join(t.TempDir(), "state.json")
	wf := New(testSteps, sp, Config{}, nil)

	wf.Next() // → one

	err := wf.Goto("two")
	require.NoError(t, err)
	require.Equal(t, "two", wf.Current())
}

func TestGotoSameStepIsNoop(t *testing.T) {
	sp := filepath.Join(t.TempDir(), "state.json")
	wf := New(testSteps, sp, Config{}, nil)

	wf.Next() // → one

	err := wf.Goto("one")
	require.NoError(t, err)
	require.Equal(t, "one", wf.Current())
}

func TestGotoInvalidStepFails(t *testing.T) {
	sp := filepath.Join(t.TempDir(), "state.json")
	wf := New(testSteps, sp, Config{}, nil)

	err := wf.Goto("nonexistent")
	require.Error(t, err)
}

func TestAutoSaveOnTransition(t *testing.T) {
	sp := filepath.Join(t.TempDir(), "state.json")
	wf := New(testSteps, sp, Config{}, nil)

	wf.Next() // → one
	wf.Next() // → two

	// Rebuild from persisted state (auto-saved by enter_state).
	loaded := New(testSteps, sp, Config{}, nil)
	require.Equal(t, "two", loaded.Current())
	require.Equal(t, []string{"one"}, loaded.State().CompletedSteps)
}

func TestStepStatus(t *testing.T) {
	sp := filepath.Join(t.TempDir(), "state.json")
	wf := New(testSteps, sp, Config{}, nil)

	wf.Next() // → one
	wf.Next() // → two

	infos := wf.StepStatus()
	require.Len(t, infos, 3)
	require.Equal(t, "completed", infos[0].Status)
	require.Equal(t, "current", infos[1].Status)
	require.Equal(t, "pending", infos[2].Status)
}

func TestGotoBackwardFails(t *testing.T) {
	sp := filepath.Join(t.TempDir(), "state.json")
	wf := New(testSteps, sp, Config{}, nil)

	wf.Next() // → one
	wf.Next() // → two
	wf.Next() // → three

	err := wf.Goto("one")
	require.Error(t, err)
}

func TestNextStepName(t *testing.T) {
	sp := filepath.Join(t.TempDir(), "state.json")
	wf := New(testSteps, sp, Config{}, nil)

	wf.Next() // → one
	require.Equal(t, "two", wf.NextStepName())

	wf.Next() // → two
	require.Equal(t, "three", wf.NextStepName())

	wf.Next() // → three
	require.Equal(t, "", wf.NextStepName())
}

func TestCompletedStepsTracked(t *testing.T) {
	sp := filepath.Join(t.TempDir(), "state.json")
	wf := New(testSteps, sp, Config{}, nil)

	wf.Next() // → one
	wf.Next() // → two
	require.Equal(t, []string{"one"}, wf.State().CompletedSteps)

	wf.Next() // → three
	require.Equal(t, []string{"one", "two"}, wf.State().CompletedSteps)
}
