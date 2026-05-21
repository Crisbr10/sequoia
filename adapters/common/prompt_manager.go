package common

import "github.com/Crisbr10/sequoia/adapters"

// PromptManager handles the system prompt write/remove lifecycle for a
// tool adapter. It encapsulates the prompt strategy, the write and remove
// functions, and a rollback-on-error flag.
//
// PromptManager is constructed once per adapter and injected into BaseAdapter
// via SetPromptManager(). All methods are safe to call with nil function
// fields (they return nil, matching Codex's no-op system prompt behavior).
type PromptManager struct {
	promptStrategy         adapters.PromptStrategy
	writeSystemPrompt      func(base, content string) error
	removeSystemPrompt     func(base string) error
	rollbackOnSysPromptErr bool
}

// NewPromptManager creates a PromptManager with the given strategy and
// write/remove functions. Either function may be nil — Write() and Remove()
// will return nil gracefully in that case (used by Codex which has no
// system prompt step).
func NewPromptManager(
	strategy adapters.PromptStrategy,
	writeFn func(base, content string) error,
	removeFn func(base string) error,
) *PromptManager {
	return &PromptManager{
		promptStrategy:     strategy,
		writeSystemPrompt:  writeFn,
		removeSystemPrompt: removeFn,
	}
}

// PromptStrategy returns the injection strategy used by this adapter.
func (pm *PromptManager) PromptStrategy() adapters.PromptStrategy {
	return pm.promptStrategy
}

// Write writes the system prompt content to the base directory using the
// configured write function. If writeFn is nil (Codex case), Write()
// returns nil — a graceful no-op.
func (pm *PromptManager) Write(base, content string) error {
	if pm.writeSystemPrompt == nil {
		return nil
	}
	return pm.writeSystemPrompt(base, content)
}

// Remove removes or restores the system prompt in the base directory using
// the configured remove function. If removeFn is nil (Codex case), Remove()
// returns nil — a graceful no-op.
func (pm *PromptManager) Remove(base string) error {
	if pm.removeSystemPrompt == nil {
		return nil
	}
	return pm.removeSystemPrompt(base)
}

// RollbackOnError reports whether the shared Install() should roll back
// skill and command installers when the system prompt step fails.
func (pm *PromptManager) RollbackOnError() bool {
	return pm.rollbackOnSysPromptErr
}

// SetRollbackOnError enables or disables rollback of skill and command
// installers when the system prompt step fails during Install().
func (pm *PromptManager) SetRollbackOnError(v bool) {
	pm.rollbackOnSysPromptErr = v
}
