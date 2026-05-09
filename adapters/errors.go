package adapters

import "errors"

// ErrUnknownAdapter is returned when an adapter ID is not registered
// in the Registry.
var ErrUnknownAdapter = errors.New("unknown adapter")
