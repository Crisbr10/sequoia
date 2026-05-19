package adapters

import (
	"fmt"
	"sync"
)

// Registry holds all registered ToolAdapters indexed by their ID.
// Use Register to add adapters — typically called from adapter init() functions,
// following the database/sql self-registration pattern.
//
// Registry is safe for concurrent use.
type Registry struct {
	mu    sync.RWMutex
	items map[string]ToolAdapter
	order []string // preserves registration order for All()
}

// NewRegistry creates an initialized Registry ready for adapter registration.
// Use this constructor instead of &Registry{} to ensure internal maps and
// slices are allocated correctly (Register() does lazy-init, but NewRegistry
// is explicit and test-friendly).
func NewRegistry() *Registry {
	return &Registry{
		items: make(map[string]ToolAdapter),
		order: make([]string, 0),
	}
}

// DefaultRegistry is the global adapter registry used by NewAdapter.
// Deprecated: prefer NewRegistry() constructor and dependency injection.
// init() functions in adapter sub-packages still register here for backward
// compatibility. New code should use constructor DI via NewRegistry().
var DefaultRegistry = NewRegistry()

// Register adds a to the registry.
// If an adapter with the same ID already exists, it is replaced.
// Replacement preserves the original registration order position.
func (r *Registry) Register(a ToolAdapter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.items == nil {
		r.items = make(map[string]ToolAdapter)
	}

	id := a.ID()
	if _, exists := r.items[id]; !exists {
		// New ID — append to order slice.
		r.order = append(r.order, id)
	}
	// Overwrite (or insert) the adapter.
	r.items[id] = a
}

// Get returns the adapter registered under id.
// If no adapter with that ID exists, it returns (nil, ErrUnknownAdapter).
func (r *Registry) Get(id string) (ToolAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, ok := r.items[id]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownAdapter, id)
	}
	return a, nil
}

// All returns all registered adapters in registration order.
// The returned slice is a snapshot — subsequent Register calls do not affect it.
func (r *Registry) All() []ToolAdapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]ToolAdapter, 0, len(r.order))
	for _, id := range r.order {
		if a, ok := r.items[id]; ok {
			result = append(result, a)
		}
	}
	return result
}
