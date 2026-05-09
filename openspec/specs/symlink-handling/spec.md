# Specs: Symlink Handling

## Domain: symlink-handling

### Requirement: Symlink Resolution in Base Paths

Each adapter's `base()` function MUST apply `filepath.EvalSymlinks()` to the resolved home directory before constructing the tool config path. The status output SHALL display the resolved (real) path, not the symlink path.

**Scenario: macOS home directory is a symlink**
- GIVEN `/Users/jane` is a symlink to `/Volumes/External/Users/jane`
- WHEN `claudeBase(homeDir)` resolves the home directory
- THEN the returned path MUST be `/Volumes/External/Users/jane/.claude`
- AND `sequoia status` MUST display the real path

**Scenario: No symlinks — normal path unchanged**
- GIVEN the home directory is not a symlink
- WHEN `base()` resolves the path
- THEN the returned path MUST be identical to `filepath.Join(homeDir, ".claude")` (or `.config/opencode`)

### Requirement: Graceful Fallback on Resolution Failure

If `filepath.EvalSymlinks()` returns an error, the adapter MUST fall back to the unresolved path and SHALL NOT propagate the error.

**Scenario: EvalSymlinks fails on Windows**
- GIVEN `filepath.EvalSymlinks()` returns an error
- WHEN `base()` is called
- THEN the unresolved path MUST be returned without error
- AND `sequoia status` MUST display the unresolved path
