package common

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveSymlink resolves path via filepath.EvalSymlinks.
// If resolution succeeds, the resolved path is returned with no warning.
// If resolution fails:
//   - If the path is a symlink (detected via os.Lstat), the original path is
//     returned along with a warning message containing the unresolved path.
//   - If the path is not a symlink (or Lstat itself fails), the original path
//     is returned without a warning.
func ResolveSymlink(path string) (resolved string, warning string) {
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved, ""
	}

	// EvalSymlinks failed. Check if the path is a symlink.
	info, lerr := os.Lstat(path)
	if lerr != nil {
		// Lstat failed (e.g., path doesn't exist) — treat as non-symlink.
		return path, ""
	}

	if info.Mode()&os.ModeSymlink != 0 {
		// The path IS a symlink but could not be resolved — emit a warning.
		return path, fmt.Sprintf("symlink could not be resolved: %s (using unresolved path)", path)
	}

	// Path exists but is not a symlink — no warning.
	return path, ""
}
