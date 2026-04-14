# Research: 18_release_workflow

## Alternatives considered and rejected

### Option 1: Full Port of Jumppad's Dagger Module

**Description**: Create a complete `dagger/` module by porting `jumppad-labs/jumppad/dagger/main.go` to spektacular, preserving all patterns and structure (~800 lines).

**Rejected**: High initial effort (3-5 days) to port and maintain duplicate code when Dagger's module system allows direct reuse. Module reuse provides the same functionality with significantly less code and automatic benefit from jumppad's improvements. The only scenario where porting makes sense is if jumppad's module proves impossible to import or adapt, which should be discovered during Phase 1.1.

**Evidence**: Dagger documentation on module imports (https://docs.dagger.io/manuals/developer/modules/), jumppad's module structure at `github.com/jumppad-labs/jumppad/dagger/main.go`.

### Option 2: Minimal Dagger Module with Deferred Packaging

**Description**: Create a simplified Dagger module that handles cross-platform builds, macOS signing/notarization, and GitHub releases, but defers Homebrew and GemFury publishing to a future iteration.

**Rejected**: Doesn't meet spec requirement "Release function publishes spektacular.rb formula to jumppad-labs/homebrew-repo" and "pushes packages to GemFury". Partial implementation creates confusion about distribution channels and forces users to wait for Homebrew support. Since jumppad's module already implements all channels, reusing it delivers complete functionality immediately.

**Evidence**: Spec requirements at `.spektacular/specs/18_release_workflow.md:15-16` and `.spektacular/specs/18_release_workflow.md:18-19`.

### Option 3: Hybrid Approach - Dagger for Builds, GitHub Actions for Distribution

**Description**: Use Dagger only for cross-platform builds and macOS signing/notarization, but handle GitHub releases, Homebrew, and GemFury publishing directly in GitHub Actions using existing tools (gh CLI, curl).

**Rejected**: Spec explicitly requires "Dagger module exposes a Release function" with side-effects (`.spektacular/specs/18_release_workflow.md:14-15`). Split architecture diverges from jumppad's proven pattern and makes debugging harder (distribution logic scattered between Dagger and YAML). Module reuse keeps all logic in one place and maintains consistency with jumppad.

**Evidence**: Spec requirement at `.spektacular/specs/18_release_workflow.md:14-15`, jumppad's unified approach in `github.com/jumppad-labs/jumppad/dagger/main.go`.

## Chosen approach — evidence

**Module Reuse via Dagger Import System**: Dagger supports importing modules as dependencies, allowing spektacular to reuse jumppad's entire build pipeline by declaring a dependency in `dagger.json` and wrapping the imported functions with spektacular-specific configuration.

**Evidence**:
- Dagger module system documentation: https://docs.dagger.io/manuals/developer/modules/
- Jumppad's module structure is self-contained and parameterizable: `github.com/jumppad-labs/jumppad/dagger/main.go` exposes functions that accept configuration (binary name, repo, formula name)
- Existing pattern in Dagger ecosystem: modules commonly wrap other modules to provide domain-specific configuration

**Proven Architecture**: Jumppad's Dagger module has been running in production, handling releases for jumppad itself. The architecture is battle-tested for cross-platform builds, macOS signing, and multi-channel publishing.

**Evidence**:
- Jumppad releases on GitHub: https://github.com/jumppad-labs/jumppad/releases
- Jumppad Homebrew formula: https://github.com/jumppad-labs/homebrew-repo/blob/main/Formula/jumppad.rb
- Jumppad's workflow: `github.com/jumppad-labs/jumppad/.github/workflows/build_and_deploy.yaml`

**Shared Infrastructure**: The `jumppad-labs/homebrew-repo` tap already exists and hosts jumppad's formula. Adding spektacular's formula to the same tap reuses infrastructure and provides a consistent user experience (single tap for both tools).

**Evidence**:
- Existing tap: https://github.com/jumppad-labs/homebrew-repo
- Tap structure supports multiple formulas: `Formula/jumppad.rb` can coexist with `Formula/spektacular.rb`

## Files examined

- `.github/workflows/build.yml:1-38` — Current workflow runs lint, test, and cross-compilation. Will be replaced.
- `Makefile:19-23` — Cross-compilation targets match spec requirements (darwin/linux on amd64/arm64).
- `go.mod:3` — Go version 1.25.0, compatible with Dagger SDK.
- `.spektacular/plans/6_convert_to_go/research.md:1-200` — Prior research on Go conversion, no release workflow details.
- `main.go:1-5` — Simple entry point, no build/release logic.
- `cmd/root.go` — CLI structure, no release-related commands.

## External references

- **Dagger Go SDK**: https://docs.dagger.io/api/reference/go — Container orchestration and module system used by jumppad
- **Dagger Module System**: https://docs.dagger.io/manuals/developer/modules/ — How to import and reuse modules
- **Quill**: https://github.com/anchore/quill — Apple code signing and notarization tool used by jumppad
- **Jumppad's Dagger Module**: https://github.com/jumppad-labs/jumppad/blob/main/dagger/main.go — Reference implementation (~800 lines)
- **Jumppad's Workflow**: https://github.com/jumppad-labs/jumppad/blob/main/.github/workflows/build_and_deploy.yaml — GitHub Actions integration pattern
- **Homebrew Formula Format**: https://docs.brew.sh/Formula-Cookbook — Structure of `.rb` formula files

## Prior plans / specs consulted

- `.spektacular/specs/18_release_workflow.md` — Source spec defining requirements, constraints, and acceptance criteria
- `.spektacular/plans/6_convert_to_go/research.md` — Prior research on Go conversion, confirmed no existing release infrastructure

## Open assumptions

**Jumppad's module is importable**: Assumes `github.com/jumppad-labs/jumppad/dagger` can be imported as a Dagger module dependency. If jumppad's module is not published or has incompatible interfaces, the implementation will need to fall back to porting or wrapping. This will be verified during Phase 1.1.

**Quill credentials are available**: Assumes the seven required GitHub secrets (`GH_TOKEN`, `FURY_TOKEN`, `QUILL_SIGN_P12`, `QUILL_SIGN_PASSWORD`, `QUILL_NOTORY_KEY`, `QUILL_NOTARY_KEY_ID`, `QUILL_NOTARY_ISSUER`) are already configured at the org level and accessible to spektacular's workflow. If not, the implementer must STOP and ask the user to configure them.

**Homebrew tap write permissions**: Assumes `GH_TOKEN` has write access to `jumppad-labs/homebrew-repo`. If permissions are insufficient, the implementer must STOP and ask the user to verify token permissions.

**GemFury package format**: Assumes jumppad's `UpdateGemFury` function produces deb/rpm packages in a format GemFury accepts. If package format is incompatible, the implementer must STOP and ask the user for guidance on package metadata or scope adjustment.

## Rehydration cues

**To rebuild context from cold**:

1. **Read the spec**: `.spektacular/specs/18_release_workflow.md` — Requirements, constraints, acceptance criteria
2. **Examine jumppad's module**: `github.com/jumppad-labs/jumppad/dagger/main.go` — Reference implementation
3. **Check current state**: `.github/workflows/build.yml` — Existing workflow to be replaced
4. **Review Dagger docs**: https://docs.dagger.io/manuals/developer/modules/ — Module import pattern
5. **Verify secrets**: GitHub org settings → Secrets → Check for `GH_TOKEN`, `FURY_TOKEN`, `QUILL_*` secrets
6. **Test module import**: Run `dagger install github.com/jumppad-labs/jumppad/dagger` to verify importability

**Key decision points**:
- Phase 1.1: Can jumppad's module be imported directly? If no, fall back to porting.
- Phase 3.1: Does jumppad's `Release` function support spektacular-specific configuration? If no, adapt the wrapper.
- Phase 4.1: Are all seven secrets configured? If no, STOP and ask user.
