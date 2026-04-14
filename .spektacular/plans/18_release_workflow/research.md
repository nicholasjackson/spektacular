# Research: 18_release_workflow

## Alternatives considered and rejected

### Option 1: Module Import/Reuse

**Description**: Import jumppad's Dagger module as a dependency and create a thin wrapper that translates spektacular-specific parameters into jumppad's expected format.

**Rejected**: While Dagger supports module imports, this approach introduces a runtime dependency on jumppad's module stability and interface compatibility. Any breaking changes in jumppad's module would require immediate updates to spektacular. Additionally, the user explicitly requested copying the implementation and reusing only common modules, not importing jumppad's entire module.

**Evidence**: User feedback: "nah, we should just copy the implementation from jumppad and re-use common modules that exist"

### Option 2: Minimal Dagger Module with Deferred Packaging

**Description**: Create a simplified Dagger module that handles cross-platform builds, macOS signing/notarization, and GitHub releases, but defers Homebrew and GemFury publishing to a future iteration.

**Rejected**: Doesn't meet spec requirement "Release function publishes spektacular.rb formula to jumppad-labs/homebrew-repo" and "pushes packages to GemFury". Partial implementation creates confusion about distribution channels and forces users to wait for Homebrew support. Since jumppad's implementation already covers all channels, porting it delivers complete functionality immediately.

**Evidence**: Spec requirements at `.spektacular/specs/18_release_workflow.md:15-16` and `.spektacular/specs/18_release_workflow.md:18-19`.

### Option 3: Hybrid Approach - Dagger for Builds, GitHub Actions for Distribution

**Description**: Use Dagger only for cross-platform builds and macOS signing/notarization, but handle GitHub releases, Homebrew, and GemFury publishing directly in GitHub Actions using existing tools (gh CLI, curl).

**Rejected**: Spec explicitly requires "Dagger module exposes a Release function" with side-effects (`.spektacular/specs/18_release_workflow.md:14-15`). Split architecture diverges from jumppad's proven pattern and makes debugging harder (distribution logic scattered between Dagger and YAML). Porting jumppad's unified approach keeps all logic in one place and maintains consistency.

**Evidence**: Spec requirement at `.spektacular/specs/18_release_workflow.md:14-15`, jumppad's unified approach in `github.com/jumppad-labs/jumppad/dagger/main.go`.

## Chosen approach — evidence

**Port and Adapt Jumppad's Implementation**: Copy jumppad's Dagger module (~800 lines) and adapt it for spektacular-specific configuration (binary name, repo, formula name). Reuse common Dagger modules (GitHub module for PR labels, Quill for signing, cli.Deb() for package building) as dependencies while owning the core orchestration logic.

**Evidence**:
- Jumppad's module structure is well-documented and proven: `github.com/jumppad-labs/jumppad/dagger/main.go`
- User explicitly requested this approach: "we should just copy the implementation from jumppad and re-use common modules that exist"
- Provides full control over build pipeline without dependency on jumppad's module stability
- Allows spektacular-specific optimizations while maintaining proven patterns

**Proven Architecture**: Jumppad's Dagger module has been running in production, handling releases for jumppad itself. The architecture is battle-tested for cross-platform builds, macOS signing, and multi-channel publishing.

**Evidence**:
- Jumppad releases on GitHub: https://github.com/jumppad-labs/jumppad/releases
- Jumppad Homebrew formula: https://github.com/jumppad-labs/homebrew-repo/blob/main/Formula/jumppad.rb
- Jumppad's workflow: `github.com/jumppad-labs/jumppad/.github/workflows/build_and_deploy.yaml`

**Shared Infrastructure**: The `jumppad-labs/homebrew-repo` tap already exists and hosts jumppad's formula. Adding spektacular's formula to the same tap reuses infrastructure and provides a consistent user experience (single tap for both tools).

**Evidence**:
- Existing tap: https://github.com/jumppad-labs/homebrew-repo
- Tap structure supports multiple formulas: `Formula/jumppad.rb` can coexist with `Formula/spektacular.rb`

**Common Module Reuse**: While the core orchestration is ported, common Dagger modules are consumed as standard dependencies:
- Dagger GitHub module for version resolution (reading PR labels)
- Quill for macOS signing and notarization
- Dagger `cli.Deb()` module for creating Debian packages
- Standard Dagger SDK for container orchestration

**Evidence**:
- Dagger module ecosystem: https://docs.dagger.io/manuals/developer/modules/
- Quill integration pattern in jumppad: `github.com/jumppad-labs/jumppad/dagger/main.go:300-400`
- Debian package creation in jumppad: `github.com/jumppad-labs/jumppad/dagger/main.go:177-199` using `cli.Deb().Build()`

**Package Building Approach**: Jumppad uses the Dagger `cli.Deb()` module to create Debian packages, not fpm/nfpm directly. Only .deb packages are created (no RPM). This simplifies the implementation and reduces external dependencies.

**Evidence**:
- Jumppad's Package function: `github.com/jumppad-labs/jumppad/dagger/main.go:177-199`
- Uses `cli.Deb().Build(pkg, arch, "jumppad", version, "Nic Jackson", "Jumppad application")`
- Only creates .deb packages for linux/amd64 and linux/arm64

## Files examined

- `.github/workflows/build.yml:1-38` — Current workflow runs lint, test, and cross-compilation. Will be replaced.
- `Makefile:19-23` — Cross-compilation targets match spec requirements (darwin/linux on amd64/arm64).
- `go.mod:3` — Go version 1.25.0, compatible with Dagger SDK.
- `.spektacular/plans/6_convert_to_go/research.md:1-200` — Prior research on Go conversion, no release workflow details.
- `main.go:1-5` — Simple entry point, no build/release logic.
- `cmd/root.go` — CLI structure, no release-related commands.
- `github.com/jumppad-labs/jumppad/dagger/main.go` (external) — Reference implementation for porting (~800 lines)
- `github.com/jumppad-labs/jumppad/dagger/main.go:177-199` (external) — Package function using Dagger cli.Deb() module

## External references

- **Dagger Go SDK**: https://docs.dagger.io/api/reference/go — Container orchestration and module system
- **Dagger Module System**: https://docs.dagger.io/manuals/developer/modules/ — How to create and use modules
- **Dagger cli.Deb() Module**: Used by jumppad for creating Debian packages without fpm/nfpm
- **Quill**: https://github.com/anchore/quill — Apple code signing and notarization tool used by jumppad
- **Jumppad's Dagger Module**: https://github.com/jumppad-labs/jumppad/blob/main/dagger/main.go — Reference implementation (~800 lines) to port
- **Jumppad's Workflow**: https://github.com/jumppad-labs/jumppad/blob/main/.github/workflows/build_and_deploy.yaml — GitHub Actions integration pattern
- **Homebrew Formula Format**: https://docs.brew.sh/Formula-Cookbook — Structure of `.rb` formula files

## Prior plans / specs consulted

- `.spektacular/specs/18_release_workflow.md` — Source spec defining requirements, constraints, and acceptance criteria
- `.spektacular/plans/6_convert_to_go/research.md` — Prior research on Go conversion, confirmed no existing release infrastructure

## Open assumptions

**Quill credentials are available**: Assumes the seven required GitHub secrets (`GH_TOKEN`, `FURY_TOKEN`, `QUILL_SIGN_P12`, `QUILL_SIGN_PASSWORD`, `QUILL_NOTORY_KEY`, `QUILL_NOTARY_KEY_ID`, `QUILL_NOTARY_ISSUER`) are already configured at the org level and accessible to spektacular's workflow. If not, the implementer must STOP and ask the user to configure them.

**Homebrew tap write permissions**: Assumes `GH_TOKEN` has write access to `jumppad-labs/homebrew-repo`. If permissions are insufficient, the implementer must STOP and ask the user to verify token permissions.

**Dagger cli.Deb() module availability**: Assumes the Dagger `cli.Deb()` module used by jumppad is available and stable. This module creates Debian packages without requiring fpm/nfpm installation. If the module is unavailable or incompatible, the implementer must STOP and ask the user whether to use an alternative approach (fpm/nfpm directly) or adjust scope.

**Debian-only packaging**: Jumppad only creates .deb packages (no RPM). Assumes this is sufficient for spektacular's distribution needs. GemFury accepts .deb packages for APT repository hosting.

**Jumppad's implementation is portable**: Assumes jumppad's Dagger module code can be copied and adapted without significant refactoring. The error chaining pattern, build matrix, and publishing functions should port cleanly with only naming changes (binary, repo, formula).

## Rehydration cues

**To rebuild context from cold**:

1. **Read the spec**: `.spektacular/specs/18_release_workflow.md` — Requirements, constraints, acceptance criteria
2. **Examine jumppad's module**: `github.com/jumppad-labs/jumppad/dagger/main.go` — Reference implementation to port
3. **Check current state**: `.github/workflows/build.yml` — Existing workflow to be replaced
4. **Review Dagger docs**: https://docs.dagger.io/manuals/developer/modules/ — Module creation pattern
5. **Verify secrets**: GitHub org settings → Secrets → Check for `GH_TOKEN`, `FURY_TOKEN`, `QUILL_*` secrets
6. **Understand porting approach**: Copy jumppad's implementation, adapt naming, reuse common modules (GitHub, Quill, cli.Deb())

**Key decision points**:
- Phase 1.1: Port jumppad's module structure and error chaining pattern
- Phase 1.2-1.3: Port build and archive functions with spektacular-specific naming
- Phase 1.4: Port Package function using Dagger cli.Deb() module (not fpm/nfpm)
- Phase 2.1: Port Quill signing integration (most complex phase)
- Phase 3.1-3.5: Port version resolution and multi-channel publishing
- Phase 4.1-4.3: Create GitHub Actions workflow and cleanup old workflow

**Porting checklist**:
- Replace "jumppad" with "spektacular" in binary names, paths, and formulas
- Replace "jumppad-labs/jumppad" with "jumppad-labs/spektacular" in repo references
- Remove Windows-specific logic (no Windows builds)
- Remove website update logic (no spektacular website)
- Keep error chaining pattern, build matrix, and Quill integration unchanged
- Use Dagger cli.Deb() module for package creation (same as jumppad)
- Only create .deb packages (no RPM, matching jumppad's approach)
