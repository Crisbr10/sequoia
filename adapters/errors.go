package adapters

import "errors"

// ErrUnknownAdapter is returned when an adapter ID is not registered
// in the Registry.
var ErrUnknownAdapter = errors.New("unknown adapter")

// ErrInstallFailed is returned when an adapter's Install() operation fails.
// The wrapped error provides the specific cause.
var ErrInstallFailed = errors.New("install failed")

// ErrUninstallFailed is returned when an adapter's Uninstall() operation fails.
// The wrapped error may contain multiple partial errors collected via errors.Join.
var ErrUninstallFailed = errors.New("uninstall failed")

// ErrNotDetected is returned when the target AI tool is not detected
// on this system (e.g. binary not found, config dir missing).
var ErrNotDetected = errors.New("adapter not detected")
