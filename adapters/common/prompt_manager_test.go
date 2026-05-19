package common_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
)

// =========================================================================
// TestNewPromptManager_ConstructsWithFields
// =========================================================================

// TestNewPromptManager_ConstructsWithFields verifies that NewPromptManager
// stores the strategy, writeFn, and removeFn provided at construction, and
// that PromptStrategy() returns the stored strategy.
func TestNewPromptManager_ConstructsWithFields(t *testing.T) {
	t.Parallel()

	pm := common.NewPromptManager(
		adapters.StrategyMarkdownSections,
		func(base, content string) error { return nil },
		func(base string) error { return nil },
	)

	assert.Equal(t, adapters.StrategyMarkdownSections, pm.PromptStrategy(),
		"PromptStrategy() should return the strategy set at construction")
}

// =========================================================================
// TestPromptManager_PromptStrategy_AllValues
// =========================================================================

// TestPromptManager_PromptStrategy_AllValues triangulates: verifies that
// each PromptStrategy constant is stored and retrieved correctly.
func TestPromptManager_PromptStrategy_AllValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		strategy adapters.PromptStrategy
	}{
		{"MarkdownSections", adapters.StrategyMarkdownSections},
		{"FileReplace", adapters.StrategyFileReplace},
		{"ConfigMerge", adapters.StrategyConfigMerge},
		{"TOMLMerge", adapters.StrategyTOMLMerge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := common.NewPromptManager(tt.strategy, nil, nil)
			assert.Equal(t, tt.strategy, pm.PromptStrategy(),
				"PromptStrategy() should return %v", tt.strategy)
		})
	}
}

// =========================================================================
// TestPromptManager_Write_DelegatesSuccess
// =========================================================================

// TestPromptManager_Write_DelegatesSuccess verifies that Write() delegates
// to the writeFn with the provided base and content, and returns nil on success.
func TestPromptManager_Write_DelegatesSuccess(t *testing.T) {
	t.Parallel()

	var receivedBase, receivedContent string
	pm := common.NewPromptManager(
		adapters.StrategyMarkdownSections,
		func(base, content string) error {
			receivedBase = base
			receivedContent = content
			return nil
		},
		nil,
	)

	err := pm.Write("/tmp/test-base", "test content")
	assert.NoError(t, err, "Write() should return nil on success")
	assert.Equal(t, "/tmp/test-base", receivedBase, "writeFn should receive the base directory")
	assert.Equal(t, "test content", receivedContent, "writeFn should receive the content")
}

// =========================================================================
// TestPromptManager_Write_PropagatesError
// =========================================================================

// TestPromptManager_Write_PropagatesError triangulates: when writeFn returns
// an error, Write() propagates that error to the caller.
func TestPromptManager_Write_PropagatesError(t *testing.T) {
	t.Parallel()

	writeErr := errors.New("disk full")
	pm := common.NewPromptManager(
		adapters.StrategyFileReplace,
		func(base, content string) error { return writeErr },
		nil,
	)

	err := pm.Write("/tmp/base", "content")
	assert.Error(t, err, "Write() should propagate the error from writeFn")
	assert.True(t, errors.Is(err, writeErr), "Write() should propagate the exact error")
}

// =========================================================================
// TestPromptManager_Remove_DelegatesSuccess
// =========================================================================

// TestPromptManager_Remove_DelegatesSuccess verifies that Remove() delegates
// to the removeFn with the provided base, and returns nil on success.
func TestPromptManager_Remove_DelegatesSuccess(t *testing.T) {
	t.Parallel()

	var receivedBase string
	pm := common.NewPromptManager(
		adapters.StrategyMarkdownSections,
		nil,
		func(base string) error {
			receivedBase = base
			return nil
		},
	)

	err := pm.Remove("/tmp/test-base")
	assert.NoError(t, err, "Remove() should return nil on success")
	assert.Equal(t, "/tmp/test-base", receivedBase, "removeFn should receive the base directory")
}

// =========================================================================
// TestPromptManager_Remove_PropagatesError
// =========================================================================

// TestPromptManager_Remove_PropagatesError triangulates: when removeFn
// returns an error, Remove() propagates that error to the caller.
func TestPromptManager_Remove_PropagatesError(t *testing.T) {
	t.Parallel()

	removeErr := errors.New("permission denied")
	pm := common.NewPromptManager(
		adapters.StrategyConfigMerge,
		nil,
		func(base string) error { return removeErr },
	)

	err := pm.Remove("/tmp/base")
	assert.Error(t, err, "Remove() should propagate the error from removeFn")
	assert.True(t, errors.Is(err, removeErr), "Remove() should propagate the exact error")
}

// =========================================================================
// TestPromptManager_Write_NilFunc_ReturnsNil
// =========================================================================

// TestPromptManager_Write_NilFunc_ReturnsNil verifies the Codex edge case:
// when writeFn is nil (Codex uses StrategyTOMLMerge with nil write/remove
// functions), Write() must NOT panic and must return nil.
func TestPromptManager_Write_NilFunc_ReturnsNil(t *testing.T) {
	t.Parallel()

	pm := common.NewPromptManager(
		adapters.StrategyTOMLMerge,
		nil, // nil writeFn — Codex case
		nil, // nil removeFn
	)

	// Must not panic.
	err := pm.Write("/tmp/base", "content")
	assert.NoError(t, err, "Write() with nil writeFn should return nil, not panic")
}

// =========================================================================
// TestPromptManager_Remove_NilFunc_ReturnsNil
// =========================================================================

// TestPromptManager_Remove_NilFunc_ReturnsNil verifies the Codex edge case:
// when removeFn is nil, Remove() must NOT panic and must return nil.
func TestPromptManager_Remove_NilFunc_ReturnsNil(t *testing.T) {
	t.Parallel()

	pm := common.NewPromptManager(
		adapters.StrategyTOMLMerge,
		nil,
		nil, // nil removeFn — Codex case
	)

	// Must not panic.
	err := pm.Remove("/tmp/base")
	assert.NoError(t, err, "Remove() with nil removeFn should return nil, not panic")
}

// =========================================================================
// TestPromptManager_RollbackOnError_DefaultFalse
// =========================================================================

// TestPromptManager_RollbackOnError_DefaultFalse verifies that the rollback
// flag defaults to false (the zero value) when not explicitly set. Most
// adapters (Claude, Gemini, Codex) should not rollback on system prompt error.
func TestPromptManager_RollbackOnError_DefaultFalse(t *testing.T) {
	t.Parallel()

	pm := common.NewPromptManager(
		adapters.StrategyMarkdownSections,
		func(base, content string) error { return nil },
		func(base string) error { return nil },
	)

	assert.False(t, pm.RollbackOnError(), "RollbackOnError() should default to false")
}

// =========================================================================
// TestPromptManager_SetRollbackOnError
// =========================================================================

// TestPromptManager_SetRollbackOnError verifies that SetRollbackOnError(true)
// sets the flag so RollbackOnError() returns true. This is used by Cursor and
// OpenCode adapters which need rollback on system prompt failure.
func TestPromptManager_SetRollbackOnError(t *testing.T) {
	t.Parallel()

	pm := common.NewPromptManager(
		adapters.StrategyFileReplace,
		func(base, content string) error { return nil },
		func(base string) error { return nil },
	)

	pm.SetRollbackOnError(true)
	assert.True(t, pm.RollbackOnError(), "RollbackOnError() should return true after SetRollbackOnError(true)")

	// Triangulate: set back to false.
	pm.SetRollbackOnError(false)
	assert.False(t, pm.RollbackOnError(), "RollbackOnError() should return false after SetRollbackOnError(false)")
}
