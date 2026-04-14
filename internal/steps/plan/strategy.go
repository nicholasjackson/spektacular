package plan

import (
	"path/filepath"

	"github.com/jumppad-labs/spektacular/internal/stepkit"
)

// strategy implements stepkit.PathStrategy for the plan workflow.
type strategy struct{}

func (strategy) PrimaryPathField() string { return "plan_path" }

func (strategy) PathVars(instanceName, storeRoot string) map[string]any {
	planPath := filepath.Join(storeRoot, PlanFilePath(instanceName))
	contextPath := filepath.Join(storeRoot, ContextFilePath(instanceName))
	researchPath := filepath.Join(storeRoot, ResearchFilePath(instanceName))
	specPath := filepath.Join(storeRoot, "specs", instanceName+".md")

	return map[string]any{
		"plan_path":     planPath,
		"context_path":  contextPath,
		"research_path": researchPath,
		"plan_dir":      filepath.Dir(planPath),
		"plan_name":     instanceName,
		"spec_path":     specPath,
	}
}

// Compile-time interface check.
var _ stepkit.PathStrategy = strategy{}
