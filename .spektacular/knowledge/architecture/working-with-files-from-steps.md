# Working with Files from Steps

Steps interact with the `.spektacular` directory through a `store.Store` interface — never via `os` directly.

## How the Store Reaches a Step

`store.Store` is the third parameter of every `StepCallback`:

```go
type StepCallback func(data Data, out ResultWriter, st store.Store, cfg Config) error
```

The store is set once at workflow construction and passed to every step automatically. Steps that don't need it simply ignore the parameter. Steps that require it should guard against `nil`:

```go
if st == nil {
    return fmt.Errorf("store required for this step")
}
```

## Interface

```go
type Store interface {
    Root()   string                       // absolute path to .spektacular/
    Read(path string) ([]byte, error)     // ErrNotFound if missing
    Write(path string, content []byte) error  // creates or overwrites; makes parent dirs
    Delete(path string) error             // idempotent on missing
    List(path string) ([]string, error)   // ErrNotFound if dir missing
    Exists(path string) bool
}
```

All paths are **store-relative** — relative to `.spektacular/`. The store rejects paths that escape the root (e.g. `../secret`).

## Path Conventions

File locations within `.spektacular/` are **constants, not data**. Do not store paths in `workflow.Data`. Instead, derive them from the spec name using a typed helper:

```go
// In internal/spec/steps.go
func SpecFilePath(name string) string {
    return "specs/" + name + ".md"
}
```

Use `st.Root()` only when you need the absolute path for output shown to agents:

```go
absPath := filepath.Join(st.Root(), SpecFilePath(name))
```

## Common Patterns

### Create a file

```go
return st.Write(SpecFilePath(name), []byte(content))
```

### Read a file

```go
content, err := st.Read(SpecFilePath(name))
if errors.Is(err, store.ErrNotFound) {
    // handle missing
}
```

### Check existence before acting

```go
if !st.Exists(SpecFilePath(name)) {
    return fmt.Errorf("spec %q not found", name)
}
```

### List files in a directory

```go
names, err := st.List("specs")
// names is []string of filenames, not full paths
```

## Injecting the Store

The store is constructed in `cmd/` and passed to `workflow.New`:

```go
wf := workflow.New(steps, statePath, wfCfg, store.NewFileStore(dataDir))
```

Pass `nil` for workflows that only query state and never touch files (e.g. `spec status`, `spec steps`).

## Future Backends

The `Store` interface is backend-agnostic. A future HTTP or database backend swaps in at the `workflow.New` call site — no step code changes.
