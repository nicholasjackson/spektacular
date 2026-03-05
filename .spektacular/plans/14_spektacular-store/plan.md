---
date: 2026-03-05T10:51:23Z
git_commit: a005c89b1dfbc0a69f279889aa194657f47714bc
branch: skills
repository: jumppad-labs/spektacular
feature: 14_spektacular-store
---

# Plan: Spektacular Store Interface

## Overview

Create a `Store` interface abstraction over the `.spektacular` project directory. The interface provides backend-agnostic file operations (read, write, delete, list, exists) so that steps interact with project storage through a contract rather than raw `os` calls. The first implementation is a `FileStore` rooted at `.spektacular/`. Future backends (HTTP API, database) can swap in without changing step logic.

## Current State Analysis

- `internal/spec/steps.go:41-48` calls `os.MkdirAll` and `os.WriteFile` directly to create spec files.
- `internal/project/init.go:33-52` also uses raw `os` calls for directory/file creation.
- `workflow.Config` (`internal/workflow/workflow.go:13-16`) holds `Command string` and `DryRun bool` — no storage abstraction.
- No package exists for project storage.

## Desired End State

- `internal/store/store.go` defines a `Store` interface and `FileStore` struct.
- `StepCallback` signature updated: `func(data Data, out ResultWriter, store Store, cfg Config) error` — store is a first-class call-time dependency like `out`.
- `Workflow` struct holds a `store Store` field, set at construction; `workflow.Config` is **not** changed.
- `spec/steps.go` `new()` step uses the `store` parameter instead of direct `os` calls.
- `cmd/spec.go` constructs a `FileStore` and passes it to `workflow.New`.
- All existing tests pass; new unit tests cover `FileStore`.

## What We're NOT Doing

- **Not refactoring `project/init.go`** — init runs before a store is needed and creates the root directory structure; it stays as-is.
- **Not adding `Update` method** — will be added later when HTTP/DB backends are introduced.
- **Not adding search** — out of scope for this plan; noted as future work.
- **Not putting Store in `workflow.Config`** — store is a service like `ResultWriter`, not a config value.

## Implementation Approach

1. Define a minimal `Store` interface in a new `internal/store` package.
2. Implement `FileStore` with path validation (no escaping the root).
3. Update `StepCallback` to `func(data Data, out ResultWriter, store Store, cfg Config) error`.
4. Add `store Store` field to `Workflow`; pass it in the FSM callback when invoking steps.
5. Update `workflow.New` to accept a `store Store` parameter.
6. Update `spec/steps.go` all steps to accept the new signature; `new()` uses the store.
7. Update `cmd/spec.go` to construct a `FileStore` and pass it to `workflow.New`.
8. Write unit tests for `FileStore`.

## Project References

- Build & test: `make build`, `make test`, `go test ./...`
- Relevant files:
  - [internal/store/store.go](../../../internal/store/store.go) ← to create
  - [internal/workflow/workflow.go](../../../internal/workflow/workflow.go) — `Config` struct
  - [internal/spec/steps.go](../../../internal/spec/steps.go) — `new()` step
  - [cmd/spec.go](../../../cmd/spec.go) — `workflow.Config` construction

---

## Milestone 1: Store Interface & Implementation

**Goal**: A working `Store` interface and `FileStore` that can be used by any code in the project.
**Testable**: Unit tests for `FileStore` pass; `Read`/`Write`/`Delete`/`List`/`Exists` work correctly on the filesystem.

### Phase 1.1: Create `internal/store` package

**Overview**

Define the `Store` interface and implement `FileStore`.

**Changes Required**

Create `internal/store/store.go`:

```go
package store

import (
    "errors"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
)

// ErrNotFound is returned by Read when the file does not exist.
var ErrNotFound = errors.New("not found")

// Store provides read/write access to a project's data directory.
// All paths are relative to the store root.
type Store interface {
    // Read returns the contents of the file at path.
    Read(path string) ([]byte, error)
    // Write creates or overwrites the file at path.
    // Parent directories are created automatically.
    Write(path string, content []byte) error
    // Delete removes the file at path.
    Delete(path string) error
    // List returns the names of direct children within the directory at path.
    List(path string) ([]string, error)
    // Exists reports whether a file or directory exists at path.
    Exists(path string) bool
}

// FileStore implements Store over the local filesystem.
// All paths are resolved relative to root and must not escape it.
type FileStore struct {
    root string
}

// NewFileStore creates a FileStore rooted at root.
func NewFileStore(root string) *FileStore {
    return &FileStore{root: filepath.Clean(root)}
}

// abs resolves path relative to root and guards against path traversal.
func (f *FileStore) abs(path string) (string, error) {
    cleaned := filepath.Join(f.root, filepath.Clean("/"+path))
    if !strings.HasPrefix(cleaned, f.root) {
        return "", errors.New("path escapes store root")
    }
    return cleaned, nil
}

func (f *FileStore) Read(path string) ([]byte, error) {
    abs, err := f.abs(path)
    if err != nil {
        return nil, err
    }
    data, err := os.ReadFile(abs)
    if errors.Is(err, fs.ErrNotExist) {
        return nil, ErrNotFound
    }
    return data, err
}

func (f *FileStore) Write(path string, content []byte) error {
    abs, err := f.abs(path)
    if err != nil {
        return err
    }
    if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
        return err
    }
    return os.WriteFile(abs, content, 0644)
}

func (f *FileStore) Delete(path string) error {
    abs, err := f.abs(path)
    if err != nil {
        return err
    }
    if err := os.Remove(abs); errors.Is(err, fs.ErrNotExist) {
        return nil // idempotent
    } else {
        return err
    }
}

func (f *FileStore) List(path string) ([]string, error) {
    abs, err := f.abs(path)
    if err != nil {
        return nil, err
    }
    entries, err := os.ReadDir(abs)
    if errors.Is(err, fs.ErrNotExist) {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    names := make([]string, len(entries))
    for i, e := range entries {
        names[i] = e.Name()
    }
    return names, nil
}

func (f *FileStore) Exists(path string) bool {
    abs, err := f.abs(path)
    if err != nil {
        return false
    }
    _, err = os.Stat(abs)
    return err == nil
}
```

**Success Criteria**

Automated:
- [ ] `go build ./internal/store/...` passes
- [ ] `go test ./internal/store/...` passes

---

### Phase 1.2: Unit tests for `FileStore`

**Overview**

Test each method of `FileStore` using a temp directory.

**Changes Required**

Create `internal/store/store_test.go` covering:
- `Write` creates file and parent dirs
- `Read` returns content; returns `ErrNotFound` for missing
- `Delete` removes file; is idempotent on missing
- `List` returns entry names; returns `ErrNotFound` for missing dir
- `Exists` returns true/false correctly
- Path traversal (`../escape`) is rejected

**Success Criteria**

Automated:
- [ ] `go test ./internal/store/...` passes with all subtests green

---

## Milestone 2: Wire Store into Workflow & Steps

**Goal**: Steps receive a `Store` as a call-time parameter (like `out`) and the `new()` step uses it to create spec files.
**Testable**: `spektacular spec new --data '{"name":"test-store"}' --dry-run` succeeds; `spektacular spec new --data '{"name":"test-store"}'` creates the spec file via the store.

### Phase 2.1: Update `StepCallback` signature and `Workflow`

**Overview**

Add `store Store` as a parameter to `StepCallback` — analogous to `out ResultWriter`. Add `store Store` field to the `Workflow` struct and wire it through `New` and the FSM callback.

**Changes Required**

In `internal/workflow/workflow.go`:

```go
import "github.com/jumppad-labs/spektacular/internal/store"

// StepCallback receives the data store, output writer, project store, and config.
type StepCallback func(data Data, out ResultWriter, store store.Store, cfg Config) error

// Workflow struct gains a store field:
type Workflow struct {
    // ...existing fields...
    store store.Store
}

// New gains a store parameter:
func New(steps []StepConfig, statePath string, cfg Config, st store.Store) *Workflow {
    // ...
    w := &Workflow{
        // ...
        store: st,
    }
    // FSM callback passes w.store:
    callbacks["after_"+s.Name] = func(_ context.Context, e *fsm.Event) {
        // ...
        if err := step.Callback(w.data, out, w.store, w.cfg); err != nil {
            e.Cancel(err)
        }
    }
}
```

**Note on nil store**: When `store` is nil (e.g. schema-only queries like `spec steps`), steps that call the store will panic. Steps that don't need the store simply ignore the parameter. The `new()` step guards with `if store == nil { return fmt.Errorf("store required") }`.

**Success Criteria**

Automated:
- [ ] `go build ./...` passes

---

### Phase 2.2: Update all step callbacks in `spec/steps.go`

**Overview**

Update every step function to the new signature. Only `new()` actually uses the store; others pass it through unused.

**Changes Required**

In `internal/spec/steps.go`, all callbacks change from:
```go
func(data workflow.Data, out workflow.ResultWriter, cfg workflow.Config) error
```
to:
```go
func(data workflow.Data, out workflow.ResultWriter, st store.Store, cfg workflow.Config) error
```

The `new()` step uses `st.Write("specs/"+name+".md", []byte(rendered))` instead of the current `os` calls. Store `spec_path` as a store-relative path (`specs/foo.md`) in data; the full path reported in `Result.SpecPath` is reconstructed as `filepath.Join(dataDir, specRelPath)` in `cmd/spec.go`.

Remove the `os`, `path/filepath` imports from `steps.go` once the direct calls are gone.

**Success Criteria**

Automated:
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes

---

### Phase 2.3: Inject `FileStore` in `cmd/spec.go`

**Overview**

Construct a `FileStore` rooted at `dataDir` (the `.spektacular` directory) and pass it to `workflow.New`.

**Changes Required**

In `cmd/spec.go`, all `workflow.New(...)` calls gain the store argument:

```go
st := store.NewFileStore(dataDir)
wf := workflow.New(steps, statePath, wfCfg, st)
```

For schema-only paths (e.g. `runSpecSteps`, `runSpecStatus`), pass `nil` or a no-op store — steps in those paths don't invoke the store.

Update `specPath` in `runSpecNew` to be store-relative (`"specs/" + input.Name + ".md"`) and reconstruct the absolute path for `wf.SetData("spec_path", ...)` output by joining `dataDir`.

**Success Criteria**

Automated:
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes

Manual:
- [ ] `spektacular spec new --data '{"name":"test-store-plan"}' --dry-run` prints JSON instruction
- [ ] `spektacular spec new --data '{"name":"test-store-plan"}'` creates `.spektacular/specs/test-store-plan.md`

---

## Testing Strategy

**Unit**: `internal/store/store_test.go` — all `FileStore` methods with temp dirs; path traversal rejection.

**Integration**: Existing `cmd` and `workflow` tests continue to pass after wiring.

**Manual**: Run `spec new` end-to-end and verify the spec file is created.

## References

- Current spec steps: [internal/spec/steps.go](../../../internal/spec/steps.go)
- Workflow config: [internal/workflow/workflow.go:13](../../../internal/workflow/workflow.go#L13)
- Spec command: [cmd/spec.go:142](../../../cmd/spec.go#L142)
