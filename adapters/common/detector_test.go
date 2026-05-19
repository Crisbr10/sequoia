package common_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Crisbr10/sequoia/adapters/common"
)

// =========================================================================
// TestDetector_Detect_DelegatesToFunc
// =========================================================================

// TestDetector_Detect_DelegatesToFunc verifies that Detect() delegates to the
// detectFn provided at construction and returns its boolean result.
func TestDetector_Detect_DelegatesToFunc(t *testing.T) {
	t.Parallel()

	det := common.NewDetector(
		func() (string, error) { return "/tmp/test", nil },
		func(base string) bool { return true },
		func() bool { return true },
	)

	assert.True(t, det.Detect(), "Detect() should return true when detectFn returns true")
}

// =========================================================================
// TestDetector_Detect_ReturnsFalse
// =========================================================================

// TestDetector_Detect_ReturnsFalse triangulates: Detect() returns false when
// the injected detectFn returns false, proving delegation works in both directions.
func TestDetector_Detect_ReturnsFalse(t *testing.T) {
	t.Parallel()

	det := common.NewDetector(
		func() (string, error) { return "/tmp/test", nil },
		func(base string) bool { return true },
		func() bool { return false },
	)

	assert.False(t, det.Detect(), "Detect() should return false when detectFn returns false")
}

// =========================================================================
// TestDetector_Detect_NilFunc_ReturnsFalse
// =========================================================================

// TestDetector_Detect_NilFunc_ReturnsFalse verifies that a nil detectFn does
// not panic and returns false as a safe default.
func TestDetector_Detect_NilFunc_ReturnsFalse(t *testing.T) {
	t.Parallel()

	det := common.NewDetector(
		func() (string, error) { return "/tmp/test", nil },
		func(base string) bool { return true },
		nil, // nil detectFn — safe default
	)

	// Must not panic.
	assert.False(t, det.Detect(), "Detect() should return false when detectFn is nil")
}

// =========================================================================
// TestDetector_IsInstalled_DelegatesSuccess
// =========================================================================

// TestDetector_IsInstalled_DelegatesSuccess verifies that IsInstalled() calls
// baseFn to get the base directory, then passes it to isInstalledFn, and
// returns the result. The test also verifies the base directory is correctly
// threaded between the two functions.
func TestDetector_IsInstalled_DelegatesSuccess(t *testing.T) {
	t.Parallel()

	baseDir := "/tmp/test-base"

	det := common.NewDetector(
		func() (string, error) { return baseDir, nil },
		func(base string) bool {
			assert.Equal(t, baseDir, base, "isInstalledFn should receive the base directory from baseFn")
			return true
		},
		func() bool { return false },
	)

	assert.True(t, det.IsInstalled(), "IsInstalled() should return true when isInstalledFn returns true")
}

// =========================================================================
// TestDetector_IsInstalled_ReturnsFalse
// =========================================================================

// TestDetector_IsInstalled_ReturnsFalse triangulates: IsInstalled() returns
// false when isInstalledFn returns false, proving delegation works in both
// directions for the installed check.
func TestDetector_IsInstalled_ReturnsFalse(t *testing.T) {
	t.Parallel()

	det := common.NewDetector(
		func() (string, error) { return "/tmp/test", nil },
		func(base string) bool { return false },
		func() bool { return false },
	)

	assert.False(t, det.IsInstalled(), "IsInstalled() should return false when isInstalledFn returns false")
}

// =========================================================================
// TestDetector_IsInstalled_BaseError_ReturnsFalse
// =========================================================================

// TestDetector_IsInstalled_BaseError_ReturnsFalse verifies that when baseFn
// returns an error, IsInstalled() returns false without calling isInstalledFn.
// This prevents false positives when the base directory cannot be resolved.
func TestDetector_IsInstalled_BaseError_ReturnsFalse(t *testing.T) {
	t.Parallel()

	baseErr := errors.New("cannot resolve base")

	det := common.NewDetector(
		func() (string, error) { return "", baseErr },
		func(base string) bool {
			t.Error("isInstalledFn should NOT be called when baseFn errors")
			return true
		},
		func() bool { return false },
	)

	assert.False(t, det.IsInstalled(), "IsInstalled() should return false when baseFn errors")
}

// =========================================================================
// TestDetector_IsInstalled_NilIsInstalledFn_ReturnsFalse
// =========================================================================

// TestDetector_IsInstalled_NilIsInstalledFn_ReturnsFalse verifies that a nil
// isInstalledFn does not panic and returns false as a safe default.
func TestDetector_IsInstalled_NilIsInstalledFn_ReturnsFalse(t *testing.T) {
	t.Parallel()

	det := common.NewDetector(
		func() (string, error) { return "/tmp/test", nil },
		nil, // nil isInstalledFn — safe default
		func() bool { return false },
	)

	// Must not panic.
	assert.False(t, det.IsInstalled(), "IsInstalled() should return false when isInstalledFn is nil")
}
