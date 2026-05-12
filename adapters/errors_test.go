package adapters

import (
	"errors"
	"testing"
)

// TestSentinelErrors_Defined verifies all sentinel errors exist and are
// distinct, non-nil error values.
func TestSentinelErrors_Defined(t *testing.T) {
	t.Parallel()

	t.Run("ErrUnknownAdapter exists and is non-nil", func(t *testing.T) {
		if ErrUnknownAdapter == nil {
			t.Fatal("ErrUnknownAdapter is nil")
		}
		if ErrUnknownAdapter.Error() == "" {
			t.Fatal("ErrUnknownAdapter has empty message")
		}
	})

	t.Run("ErrInstallFailed exists and is non-nil", func(t *testing.T) {
		if ErrInstallFailed == nil {
			t.Fatal("ErrInstallFailed is nil")
		}
		if ErrInstallFailed.Error() == "" {
			t.Fatal("ErrInstallFailed has empty message")
		}
	})

	t.Run("ErrUninstallFailed exists and is non-nil", func(t *testing.T) {
		if ErrUninstallFailed == nil {
			t.Fatal("ErrUninstallFailed is nil")
		}
		if ErrUninstallFailed.Error() == "" {
			t.Fatal("ErrUninstallFailed has empty message")
		}
	})

	t.Run("ErrNotDetected exists and is non-nil", func(t *testing.T) {
		if ErrNotDetected == nil {
			t.Fatal("ErrNotDetected is nil")
		}
		if ErrNotDetected.Error() == "" {
			t.Fatal("ErrNotDetected has empty message")
		}
	})

	t.Run("all sentinel errors are distinct", func(t *testing.T) {
		sentinels := []error{ErrUnknownAdapter, ErrInstallFailed, ErrUninstallFailed, ErrNotDetected}
		for i := 0; i < len(sentinels); i++ {
			for j := i + 1; j < len(sentinels); j++ {
				if sentinels[i] == sentinels[j] {
					t.Errorf("sentinel errors at index %d and %d are the same pointer", i, j)
				}
				if errors.Is(sentinels[i], sentinels[j]) {
					t.Errorf("sentinel errors at index %d and %d are considered equal via errors.Is", i, j)
				}
			}
		}
	})
}
