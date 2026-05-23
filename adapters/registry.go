package adapters

import (
	"fmt"
	"sync"
)

// Registry holds all registered ToolAdapters indexed by their ID.
// Use Register to add adapters via dependency injection — call adapter.RegisterIn(reg)
// to populate a registry.
//
// Use RegisterFactory to defer construction until first Get() or All().
// Factories are guarded by per-key sync.Once for thread-safe lazy construction.
//
// Registry is safe for concurrent use.
type Registry struct {
	mu    sync.RWMutex
	items map[string]ToolAdapter
	order []string // preserves registration order for All()

	// factories stores lazy constructors registered via RegisterFactory.
	// Each factory is paired with a sync.Once to guarantee single construction.
	factories map[string]func() ToolAdapter
	onces     map[string]*sync.Once
}

// NewRegistry creates an initialized Registry ready for adapter registration.
// Use this constructor instead of &Registry{} to ensure internal maps and
// slices are allocated correctly (Register() does lazy-init, but NewRegistry
// is explicit and test-friendly).
func NewRegistry() *Registry {
	return &Registry{
		items:     make(map[string]ToolAdapter),
		order:     make([]string, 0),
		factories: make(map[string]func() ToolAdapter),
		onces:     make(map[string]*sync.Once),
	}
}

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

// RegisterFactory stores a lazy constructor for an adapter.
// The factory is NOT invoked until the first Get(id) or All() call.
// If a factory was already registered for id, it is replaced.
// Replacement does NOT invalidate a previously constructed adapter.
func (r *Registry) RegisterFactory(id string, factory func() ToolAdapter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.factories == nil {
		r.factories = make(map[string]func() ToolAdapter)
	}
	if r.onces == nil {
		r.onces = make(map[string]*sync.Once)
	}

	if _, exists := r.factories[id]; !exists {
		// Only append to order if no factory OR item has this ID.
		// Check items map too — Register may have already added this ID.
		if _, inItems := r.items[id]; !inItems {
			r.order = append(r.order, id)
		}
	}
	r.factories[id] = factory
	r.onces[id] = &sync.Once{}
}

// Get returns the adapter registered under id.
// If the adapter was registered via RegisterFactory, this triggers
// lazy construction (at most once, guarded by sync.Once).
// If no adapter or factory with that ID exists, it returns (nil, ErrUnknownAdapter).
func (r *Registry) Get(id string) (ToolAdapter, error) {
	// Fast path: check if already materialized.
	r.mu.RLock()
	a, ok := r.items[id]
	r.mu.RUnlock()
	if ok {
		return a, nil
	}

	// Slow path: check if a factory exists for lazy construction.
	r.mu.RLock()
	once, hasFactory := r.onces[id]
	r.mu.RUnlock()

	if !hasFactory {
		return nil, fmt.Errorf("%w: %s", ErrUnknownAdapter, id)
	}

	// Trigger lazy construction. once.Do guarantees thread safety:
	// only one goroutine executes the factory; others wait.
	once.Do(func() {
		r.mu.RLock()
		fn := r.factories[id]
		r.mu.RUnlock()
		adapter := fn()
		// Store the materialized adapter.
		r.mu.Lock()
		r.items[id] = adapter
		r.mu.Unlock()
	})

	// Return the now-materialized adapter.
	r.mu.RLock()
	a = r.items[id]
	r.mu.RUnlock()
	return a, nil
}

// All returns all registered adapters in registration order.
// Pending factories are triggered via Get(id) before the snapshot is taken.
// The returned slice is a snapshot — subsequent Register/RegisterFactory calls
// do not affect it.
func (r *Registry) All() []ToolAdapter {
	// Trigger construction for all pending factories.
	// Copy the order slice under read lock, then call Get outside the lock
	// to avoid deadlock (Get acquires the write lock internally).
	r.mu.RLock()
	ids := make([]string, len(r.order))
	copy(ids, r.order)
	r.mu.RUnlock()

	for _, id := range ids {
		// Get will trigger factory construction if pending, or return
		// the already-materialized adapter. Ignore errors — factories may
		// not exist for every ID in order (e.g., Register added directly).
		_, _ = r.Get(id)
	}

	// Now snapshot the materialized adapters in registration order.
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
