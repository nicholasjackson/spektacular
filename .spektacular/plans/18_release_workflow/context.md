# Context: 18_release_workflow

## Current State Analysis

**Existing Build Infrastructure**:
- `Makefile:19-23` — Cross-compilation targets for darwin/linux on amd64/arm64 using `CGO_ENABLED=0`
- `.github/workflows/build.yml:1-38` — Current workflow runs `make lint`, `make test`, and `make cross` on every push
- `go.mod:3` — Go version 1.25.0
- No Dagger module exists in the repository
- No release automation exists — users must clone and build manually

**Jumppad's Dagger Module** (reference implementation):
- `jumppad-labs/jumppad/dagger/main.go` — ~800 lines implementing `JumppadCI` struct with `All()` and `Release()` functions
- Handles cross-platform builds, Quill signing, archive creation, GitHub releases, Homebrew updates, and GemFury publishing
- Uses error chaining pattern with `lastError` field for fail-fast behavior
- Supports module reuse via Dagger's import system

**Secret Requirements**:
- Seven GitHub org-level secrets must exist: `GH_TOKEN`, `FURY_TOKEN`, `QUILL_SIGN_P12`, `QUILL_SIGN_PASSWORD`, `QUILL_NOTORY_KEY`, `QUILL_NOTARY_KEY_ID`, `QUILL_NOTARY_ISSUER`
- These are already configured for jumppad and can be reused for spektacular

## Per-Phase Technical Notes

### Phase 1.1: Dagger Module Scaffolding

Create `dagger/dagger.json` declaring the module and its dependency on jumppad's module:
```json
{
  "name": "spektacular",
  "sdk": "go",
  "dependencies": [
    "github.com/jumppad-labs/jumppad/dagger"
  ]
}
```

Create `dagger/main.go` with the wrapper struct:
```go
package main

import (
    "context"
    "dagger/jumppad/dagger"
)

type SpektacularCI struct {
    jumppadCI *dagger.JumppadCI
}

func New() *SpektacularCI {
    return &SpektacularCI{
        jumppadCI: dagger.New(),
    }
}
```

Create `dagger/go.mod` and `dagger/go.sum` via `dagger develop`.

**Complexity**: Low
**Token estimate**: ~5k
**Agent strategy**: Single agent, sequential execution

**Fallback if module import fails**: If jumppad's module cannot be imported directly (e.g., not published, incompatible interface), fall back to creating a standalone module that ports the necessary functions. This decision point happens during this phase.

### Phase 1.2: Cross-Platform Build Function

Implement the `All` wrapper function in `dagger/main.go`:
```go
func (s *SpektacularCI) All(
    ctx context.Context,
    src *dagger.Directory,
    version string,
) (*dagger.Directory, error) {
    return s.jumppadCI.All(ctx, src, version, dagger.AllOpts{
        BinaryName: "spektacular",
        Repository: "jumppad-labs/spektacular",
        Platforms: []string{
            "darwin/amd64",
            "darwin/arm64",
            "linux/amd64",
            "linux/arm64",
        },
    })
}
```

The wrapper translates spektacular's domain into jumppad's expected parameters. If jumppad's `All` function doesn't support these options, adapt the wrapper to call lower-level functions (e.g., `Build`, `Archive`) separately.

**Complexity**: Low
**Token estimate**: ~8k
**Agent strategy**: Single agent, sequential execution

**Testing approach**: Run `dagger call all --src=. --version=0.0.1` locally and verify:
- Four archives are created in the output directory
- Each archive extracts to a binary named `spektacular`
- Binaries report version `0.0.1` when run with `--version`

### Phase 2.1: macOS Signing and Notarization

Configure the wrapper to pass Quill credentials to jumppad's signing function. Modify the `All` function to accept signing credentials:
```go
func (s *SpektacularCI) All(
    ctx context.Context,
    src *dagger.Directory,
    version string,
    quillP12 *dagger.Secret,
    quillPassword *dagger.Secret,
    quillNotaryKey *dagger.Secret,
    quillNotaryKeyID string,
    quillNotaryIssuer string,
) (*dagger.Directory, error) {
    return s.jumppadCI.All(ctx, src, version, dagger.AllOpts{
        BinaryName: "spektacular",
        Repository: "jumppad-labs/spektacular",
        Platforms: []string{"darwin/amd64", "darwin/arm64", "linux/amd64", "linux/arm64"},
        QuillP12: quillP12,
        QuillPassword: quillPassword,
        QuillNotaryKey: quillNotaryKey,
        QuillNotaryKeyID: quillNotaryKeyID,
        QuillNotaryIssuer: quillNotaryIssuer,
    })
}
```

**Complexity**: Low
**Token estimate**: ~6k
**Agent strategy**: Single agent, sequential execution

**Testing approach**: Run `dagger call all` with valid Quill credentials from environment variables and verify macOS binaries are signed using `codesign -dv` and `spctl -a -vv`.

### Phase 3.1: Release Orchestration

Implement the `Release` wrapper function in `dagger/main.go`:
```go
func (s *SpektacularCI) Release(
    ctx context.Context,
    archives *dagger.Directory,
    githubToken *dagger.Secret,
    gemfuryToken *dagger.Secret,
) (string, error) {
    return s.jumppadCI.Release(ctx, archives, dagger.ReleaseOpts{
        Repository: "jumppad-labs/spektacular",
        FormulaName: "spektacular.rb",
        BrewRepo: "jumppad-labs/homebrew-repo",
        GithubToken: githubToken,
        GemfuryToken: gemfuryToken,
    })
}
```

The wrapper translates spektacular's release configuration into jumppad's expected parameters. If jumppad's `Release` function doesn't support these options, adapt the wrapper to call individual publishing functions (e.g., `GithubRelease`, `UpdateBrew`, `UpdateGemFury`) separately.

**Complexity**: Medium
**Token estimate**: ~12k
**Agent strategy**: Single agent, sequential execution

**Testing approach**: Run `dagger call release` against a test branch and verify:
- GitHub Release is created with all archives
- Homebrew formula is updated in `jumppad-labs/homebrew-repo`
- GemFury packages are uploaded
- Function returns the version string

### Phase 4.1: Build Workflow

Create `.github/workflows/build_and_deploy.yaml`:
```yaml
name: Build and Deploy

on:
  push:
    branches: ['**']

permissions:
  contents: write
  packages: write

jobs:
  dagger_build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Dagger
        uses: dagger/dagger-for-github@v8.2.0
        with:
          version: "0.19.8"
      
      - name: Build all platforms
        run: |
          dagger call all \
            --src=. \
            --version=$(git describe --tags --always) \
            --quill-p12=env:QUILL_SIGN_P12 \
            --quill-password=env:QUILL_SIGN_PASSWORD \
            --quill-notary-key=env:QUILL_NOTORY_KEY \
            --quill-notary-key-id=${{ secrets.QUILL_NOTARY_KEY_ID }} \
            --quill-notary-issuer=${{ secrets.QUILL_NOTARY_ISSUER }} \
            export --path=./archives
        env:
          QUILL_SIGN_P12: ${{ secrets.QUILL_SIGN_P12 }}
          QUILL_SIGN_PASSWORD: ${{ secrets.QUILL_SIGN_PASSWORD }}
          QUILL_NOTORY_KEY: ${{ secrets.QUILL_NOTORY_KEY }}
      
      - name: Upload archives
        uses: actions/upload-artifact@v4
        with:
          name: archives
          path: archives/
```

**Complexity**: Low
**Token estimate**: ~5k
**Agent strategy**: Single agent, sequential execution

### Phase 4.2: Release Workflow

Add the release job to `.github/workflows/build_and_deploy.yaml`:
```yaml
  release:
    needs: dagger_build
    if: ${{ github.ref == 'refs/heads/main' }}
    runs-on: ubuntu-latest
    outputs:
      version: ${{ steps.release.outputs.version }}
    steps:
      - uses: actions/checkout@v4
      
      - name: Download archives
        uses: actions/download-artifact@v4
        with:
          name: archives
          path: archives/
      
      - name: Setup Dagger
        uses: dagger/dagger-for-github@v8.2.0
        with:
          version: "0.19.8"
      
      - name: Release
        id: release
        run: |
          VERSION=$(dagger call release \
            --archives=./archives \
            --github-token=env:GH_TOKEN \
            --gemfury-token=env:FURY_TOKEN)
          echo "version=$VERSION" >> $GITHUB_OUTPUT
        env:
          GH_TOKEN: ${{ secrets.GH_TOKEN }}
          FURY_TOKEN: ${{ secrets.FURY_TOKEN }}
```

**Complexity**: Low
**Token estimate**: ~5k
**Agent strategy**: Single agent, sequential execution

### Phase 4.3: Cleanup Old Workflow

Delete `.github/workflows/build.yml`.

**Complexity**: Low
**Token estimate**: ~1k
**Agent strategy**: Single agent, sequential execution

## Testing Strategy

All testing is manual verification:
1. **Local Dagger testing**: Run `dagger call all` and `dagger call release` locally with test inputs
2. **Archive verification**: Extract archives and verify binary naming and structure
3. **Signing verification**: Run `codesign` and `spctl` on macOS binaries
4. **Release verification**: After first release to `main`, verify GitHub Release, Homebrew formula, and GemFury packages

## Project References

- Jumppad Dagger module: `github.com/jumppad-labs/jumppad/dagger/main.go`
- Dagger documentation: https://docs.dagger.io/
- Quill documentation: https://github.com/anchore/quill
- GitHub Actions Dagger integration: https://github.com/dagger/dagger-for-github

## Token Management Strategy

| Tier | Token Budget | Agent Strategy |
|------|-------------|----------------|
| Low | ~10k | Single agent, sequential |
| Medium | ~25k | 2-3 parallel agents |
| High | ~50k+ | Parallel analysis, sequential integration |

All phases in this plan are Low complexity, suitable for single-agent sequential execution.

## Migration Notes

No migration needed — this is a new feature. The existing `.github/workflows/build.yml` is replaced outright.

## Performance Considerations

- Dagger caching: The Go build cache is persisted across runs via Dagger's cache volumes
- Parallel builds: Cross-platform builds run in parallel within the Dagger container
- Artifact size: Four archives totaling ~50-100MB depending on binary size
- Workflow runtime: Expected ~5-10 minutes for build job, ~2-3 minutes for release job
