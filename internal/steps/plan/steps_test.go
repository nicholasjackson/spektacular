package plan

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jumppad-labs/spektacular/internal/store"
	"github.com/jumppad-labs/spektacular/internal/workflow"
	"github.com/stretchr/testify/require"
)

type testData struct {
	values map[string]any
}

func (d *testData) Get(key string) (any, bool) {
	v, ok := d.values[key]
	return v, ok
}

func (d *testData) Set(key string, value any) {
	d.values[key] = value
}

type captureWriter struct {
	result Result
}

func (c *captureWriter) WriteResult(v any) error {
	c.result = v.(Result)
	return nil
}

func renderStep(t *testing.T, cb workflow.StepCallback) string {
	t.Helper()
	data := &testData{values: map[string]any{"name": "test"}}
	writer := &captureWriter{}
	st := store.NewFileStore(t.TempDir())
	_, err := cb(data, writer, st, workflow.Config{Command: "spektacular"})
	require.NoError(t, err)
	return writer.result.Instruction
}

func TestArchitectureStepContainsOptionsAndAgreementBeat(t *testing.T) {
	out := renderStep(t, architecture())
	require.Contains(t, strings.ToLower(out), "option", "architecture step must prompt the agent to present design options")
	require.Contains(t, strings.ToLower(out), "agreement", "architecture step must prompt the agent to get user agreement")
}

func TestImplementationDetailStepIsHighLevelOnly(t *testing.T) {
	out := renderStep(t, implementationDetail())
	require.Contains(t, strings.ToLower(out), "high-level", "implementation_detail step must enforce high-level only content")
	require.Contains(t, out, "context.md", "implementation_detail step must redirect per-phase detail to context.md")
}

func TestTestingApproachStepIsHighLevelOnly(t *testing.T) {
	out := renderStep(t, testingApproach())
	require.Contains(t, strings.ToLower(out), "high-level", "testing_approach step must enforce high-level only content")
	require.Contains(t, out, "context.md", "testing_approach step must redirect per-phase detail to context.md")
}

func TestOpenQuestionsStepRestrictsToImplTimeUncertainties(t *testing.T) {
	out := renderStep(t, openQuestions())
	require.Contains(t, strings.ToLower(out), "implementation", "open_questions step must restrict the section to impl-time uncertainties")
	require.Contains(t, strings.ToLower(out), "cannot be resolved", "open_questions step must state the cannot-resolve-now rule")
}

func TestOutOfScopeStepCoversExclusions(t *testing.T) {
	out := renderStep(t, outOfScope())
	require.Contains(t, out, "Out of Scope", "out_of_scope step must name the section it populates")
	require.Contains(t, strings.ToLower(out), "exclusion", "out_of_scope step must prompt for explicit exclusions")
}

func TestStepsOrderMatchesExpected(t *testing.T) {
	expected := []string{
		"new",
		"overview",
		"discovery",
		"architecture",
		"components",
		"data_structures",
		"implementation_detail",
		"dependencies",
		"testing_approach",
		"milestones",
		"phases",
		"open_questions",
		"out_of_scope",
		"verification",
		"write_plan",
		"write_context",
		"write_research",
		"finished",
	}
	got := Steps()
	require.Len(t, got, len(expected))
	for i, step := range got {
		require.Equal(t, expected[i], step.Name, "step %d name mismatch", i)
	}
}

func TestFSMWalkFromNewToFinished(t *testing.T) {
	tmp := t.TempDir()
	statePath := filepath.Join(tmp, "state.json")
	st := store.NewFileStore(tmp)
	writer := &captureWriter{}

	require.NoError(t, st.Write(PlanFilePath("test"), []byte("")))
	require.NoError(t, st.Write(ContextFilePath("test"), []byte("")))
	require.NoError(t, st.Write(ResearchFilePath("test"), []byte("")))

	wf := workflow.New(Steps(), statePath, workflow.Config{Command: "spektacular", DryRun: true}, st, writer)
	wf.SetData("name", "test")
	wf.SetData("plan_template", "plan content")
	wf.SetData("context_template", "context content")
	wf.SetData("research_template", "research content")

	require.Equal(t, "start", wf.Current())

	expectedStates := []string{
		"overview",
		"discovery",
		"architecture",
		"components",
		"data_structures",
		"implementation_detail",
		"dependencies",
		"testing_approach",
		"milestones",
		"phases",
		"open_questions",
		"out_of_scope",
		"verification",
		"write_plan",
		"write_context",
		"write_research",
		"finished",
	}

	for _, want := range expectedStates {
		require.NoError(t, wf.Next(), "transition to %s failed", want)
		require.Equal(t, want, wf.Current(), "expected state %s after transition", want)
	}
}
