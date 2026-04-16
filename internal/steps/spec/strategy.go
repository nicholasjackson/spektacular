package spec

import (
	"path/filepath"

	"github.com/jumppad-labs/spektacular/internal/stepkit"
)

// strategy implements stepkit.PathStrategy for the spec workflow.
type strategy struct{}

func (strategy) PrimaryPathField() string { return "spec_path" }

func (strategy) PathVars(instanceName, storeRoot string) map[string]any {
	return map[string]any{
		"spec_path": filepath.Join(storeRoot, SpecFilePath(instanceName)),
		"spec_name": instanceName,
	}
}

// Compile-time interface check.
var _ stepkit.PathStrategy = strategy{}
