package common

import (
	"path/filepath"
)

// ResolveHome resolves base to its canonical path by evaluating all symlinks.
// On Windows, calling filepath.EvalSymlinks on a non-existent path returns
// an error. On Unix, it returns the path as-is.
//
// Returns an error if the path does not exist or cannot be resolved.
func ResolveHome(base string) (string, error) {
	return filepath.EvalSymlinks(base)
}

// IsSymlink reports whether resolving home produces a different path,
// which implies the presence of a symlink somewhere in the path.
func IsSymlink(home string) bool {
	resolved, err := ResolveHome(home)
	if err != nil {
		return false
	}
	return home != resolved
}
