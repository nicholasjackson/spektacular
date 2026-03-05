package spec

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cbroglie/mustache"
	"github.com/jumppad-labs/spektacular/internal/store"
	"github.com/jumppad-labs/spektacular/internal/workflow"
	"github.com/jumppad-labs/spektacular/templates"
)

// SpecFilePath returns the store-relative path for a spec file.
func SpecFilePath(name string) string {
	return "specs/" + name + ".md"
}

// Steps returns the ordered step configs for a spec workflow.
// Each step has an explicit named callback — no string-based dispatch.
// The first step "new" is internal: it creates the spec file and produces no
// output, allowing the caller to automatically advance to "overview".
func Steps() []workflow.StepConfig {
	return []workflow.StepConfig{
		{Name: "new", Src: []string{"start"}, Dst: "new", Callback: new()},
		{Name: "overview", Src: []string{"new"}, Dst: "overview", Callback: overview()},
		{Name: "requirements", Src: []string{"overview"}, Dst: "requirements", Callback: requirements()},
		{Name: "acceptance_criteria", Src: []string{"requirements"}, Dst: "acceptance_criteria", Callback: acceptanceCriteria()},
		{Name: "constraints", Src: []string{"acceptance_criteria"}, Dst: "constraints", Callback: constraints()},
		{Name: "technical_approach", Src: []string{"constraints"}, Dst: "technical_approach", Callback: technicalApproach()},
		{Name: "success_metrics", Src: []string{"technical_approach"}, Dst: "success_metrics", Callback: successMetrics()},
		{Name: "non_goals", Src: []string{"success_metrics"}, Dst: "non_goals", Callback: nonGoals()},
		{Name: "verification", Src: []string{"non_goals"}, Dst: "verification", Callback: verification()},
		{Name: "finished", Src: []string{"verification"}, Dst: "finished", Callback: finished()},
	}
}

// new creates the spec file and produces no output.
// The caller is expected to immediately advance to "overview".
func new() workflow.StepCallback {
	return func(data workflow.Data, out workflow.ResultWriter, st store.Store, cfg workflow.Config) error {
		if cfg.DryRun {
			return nil
		}
		if st == nil {
			return fmt.Errorf("store required for new step")
		}
		name := getString(data, "name")
		rendered, err := renderTemplate("spec-scaffold.md", map[string]any{"name": name})
		if err != nil {
			return err
		}
		return st.Write(SpecFilePath(name), []byte(rendered))
	}
}

func overview() workflow.StepCallback {
	return func(data workflow.Data, out workflow.ResultWriter, st store.Store, cfg workflow.Config) error {
		return writeStepResult("overview", "requirements", "spec-steps/overview.md", data, out, st, cfg)
	}
}

func requirements() workflow.StepCallback {
	return func(data workflow.Data, out workflow.ResultWriter, st store.Store, cfg workflow.Config) error {
		return writeStepResult("requirements", "acceptance_criteria", "spec-steps/requirements.md", data, out, st, cfg)
	}
}

func acceptanceCriteria() workflow.StepCallback {
	return func(data workflow.Data, out workflow.ResultWriter, st store.Store, cfg workflow.Config) error {
		return writeStepResult("acceptance_criteria", "constraints", "spec-steps/acceptance_criteria.md", data, out, st, cfg)
	}
}

func constraints() workflow.StepCallback {
	return func(data workflow.Data, out workflow.ResultWriter, st store.Store, cfg workflow.Config) error {
		return writeStepResult("constraints", "technical_approach", "spec-steps/constraints.md", data, out, st, cfg)
	}
}

func technicalApproach() workflow.StepCallback {
	return func(data workflow.Data, out workflow.ResultWriter, st store.Store, cfg workflow.Config) error {
		return writeStepResult("technical_approach", "success_metrics", "spec-steps/technical_approach.md", data, out, st, cfg)
	}
}

func successMetrics() workflow.StepCallback {
	return func(data workflow.Data, out workflow.ResultWriter, st store.Store, cfg workflow.Config) error {
		return writeStepResult("success_metrics", "non_goals", "spec-steps/success_metrics.md", data, out, st, cfg)
	}
}

func nonGoals() workflow.StepCallback {
	return func(data workflow.Data, out workflow.ResultWriter, st store.Store, cfg workflow.Config) error {
		return writeStepResult("non_goals", "verification", "spec-steps/non_goals.md", data, out, st, cfg)
	}
}

func verification() workflow.StepCallback {
	return func(data workflow.Data, out workflow.ResultWriter, st store.Store, cfg workflow.Config) error {
		specName := getString(data, "name")
		scaffold, err := renderTemplate("spec-scaffold.md", map[string]any{"name": specName})
		if err != nil {
			return err
		}
		return writeStepResult("verification", "finished", "spec-steps/verification.md", data, out, st, cfg,
			map[string]any{"spec_template": scaffold})
	}
}

func finished() workflow.StepCallback {
	return func(data workflow.Data, out workflow.ResultWriter, st store.Store, cfg workflow.Config) error {
		specName := getString(data, "name")
		specPath := SpecFilePath(specName)
		if content := getString(data, "spec_template"); content != "" {
			if err := st.Write(specPath, []byte(content)); err != nil {
				return err
			}
		}
		return writeStepResult("finished", "", "spec-steps/finished.md", data, out, st, cfg)
	}
}

// writeStepResult renders the step template and writes the result to out.
// The spec path is derived from name and the store root rather than read from data.
func writeStepResult(name, nextStep, templatePath string, data workflow.Data, out workflow.ResultWriter, st store.Store, cfg workflow.Config, extra ...map[string]any) error {
	specName := getString(data, "name")
	specPath := filepath.Join(st.Root(), SpecFilePath(specName))

	vars := map[string]any{
		"step":      name,
		"title":     stepTitle(name),
		"spec_path": specPath,
		"next_step": nextStep,
		"config":    map[string]any{"command": cfg.Command},
	}
	for _, m := range extra {
		for k, v := range m {
			vars[k] = v
		}
	}

	instruction, err := renderTemplate(templatePath, vars)
	if err != nil {
		return err
	}
	return out.WriteResult(Result{
		Step:        name,
		SpecPath:    specPath,
		SpecName:    specName,
		Instruction: instruction,
	})
}

// getString is a convenience helper for reading string values from Data.
func getString(data workflow.Data, key string) string {
	v, ok := data.Get(key)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// renderTemplate reads a mustache template from the embedded FS and renders it.
func renderTemplate(templatePath string, data map[string]any) (string, error) {
	tmplBytes, err := templates.FS.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("loading template %s: %w", templatePath, err)
	}
	return mustache.Render(string(tmplBytes), data)
}

// stepTitle converts a step name like "acceptance_criteria" to "Acceptance Criteria".
func stepTitle(name string) string {
	words := strings.Split(name, "_")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
