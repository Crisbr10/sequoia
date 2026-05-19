package adapters_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/testutil"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		adapter adapters.ToolAdapter
		getID   string
		wantErr bool
	}{
		{
			name:    "registered adapter is retrievable by ID",
			adapter: &testutil.MockAdapter{IDVal: "claude-code", NameVal: "Claude Code"},
			getID:   "claude-code",
			wantErr: false,
		},
		{
			name:    "get unknown ID returns ErrUnknownAdapter",
			adapter: &testutil.MockAdapter{IDVal: "known-tool", NameVal: "Known Tool"},
			getID:   "does-not-exist",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := &adapters.Registry{}
			r.Register(tc.adapter)

			got, err := r.Get(tc.getID)
			if tc.wantErr {
				assert.ErrorIs(t, err, adapters.ErrUnknownAdapter)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.adapter.ID(), got.ID())
			}
		})
	}
}

func TestRegistry_All_ReturnsAllInOrder(t *testing.T) {
	t.Parallel()

	r := &adapters.Registry{}
	a1 := &testutil.MockAdapter{IDVal: "alpha", NameVal: "Alpha"}
	a2 := &testutil.MockAdapter{IDVal: "beta", NameVal: "Beta"}
	a3 := &testutil.MockAdapter{IDVal: "gamma", NameVal: "Gamma"}

	r.Register(a1)
	r.Register(a2)
	r.Register(a3)

	all := r.All()
	require.Len(t, all, 3)
	// All() returns adapters in registration order.
	assert.Equal(t, "alpha", all[0].ID())
	assert.Equal(t, "beta", all[1].ID())
	assert.Equal(t, "gamma", all[2].ID())
}

func TestRegistry_RegisterDuplicate_ReplacesExisting(t *testing.T) {
	t.Parallel()

	r := &adapters.Registry{}
	original := &testutil.MockAdapter{IDVal: "tool-x", NameVal: "Original Name"}
	replacement := &testutil.MockAdapter{IDVal: "tool-x", NameVal: "Replacement Name"}

	r.Register(original)
	r.Register(replacement)

	got, err := r.Get("tool-x")
	require.NoError(t, err)
	// Second registration replaces the first.
	assert.Equal(t, "Replacement Name", got.Name())

	// All() should still contain only one entry for this ID.
	all := r.All()
	count := 0
	for _, a := range all {
		if a.ID() == "tool-x" {
			count++
		}
	}
	assert.Equal(t, 1, count, "duplicate ID should appear exactly once in All()")
}

func TestFactory_NewAdapter_KnownID(t *testing.T) {
	t.Parallel()

	// Register into DefaultRegistry directly for the factory test.
	// Use a unique ID to avoid collisions with parallel tests.
	a := &testutil.MockAdapter{IDVal: "factory-test-known", NameVal: "Factory Known"}
	adapters.DefaultRegistry.Register(a)

	got, err := adapters.DefaultRegistry.Get("factory-test-known")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "factory-test-known", got.ID())
}

func TestFactory_NewAdapter_UnknownID(t *testing.T) {
	t.Parallel()

	_, err := adapters.DefaultRegistry.Get("this-id-was-never-registered-xyz123")
	assert.ErrorIs(t, err, adapters.ErrUnknownAdapter)
}

func TestRegistry_ConcurrentAccess_NoRace(t *testing.T) {
	t.Parallel()

	r := &adapters.Registry{}
	const goroutines = 20

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Concurrent Register calls.
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			id := "concurrent-adapter"
			r.Register(&testutil.MockAdapter{IDVal: id, NameVal: "Concurrent"})
			_ = i
		}()
	}

	// Concurrent Get calls (interleaved with Register).
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, _ = r.Get("concurrent-adapter")
		}()
	}

	wg.Wait()
}

// =========================================================================
// RegisterFactory — lazy construction
// =========================================================================

// TestRegistry_RegisterFactory_StoresFactoryNoConstruction verifies that
// RegisterFactory stores the factory function without invoking it, and that
// the factory is only called on first Get().
func TestRegistry_RegisterFactory_StoresFactoryNoConstruction(t *testing.T) {
	t.Parallel()

	r := adapters.NewRegistry()
	called := false
	r.RegisterFactory("lazy-adapter", func() adapters.ToolAdapter {
		called = true
		return &testutil.MockAdapter{IDVal: "lazy-adapter", NameVal: "Lazy"}
	})

	// Factory must NOT have been invoked during registration.
	assert.False(t, called, "factory should not be invoked during RegisterFactory")

	// Get() must trigger construction.
	got, err := r.Get("lazy-adapter")
	require.NoError(t, err)
	assert.True(t, called, "factory should be invoked on first Get()")
	assert.Equal(t, "lazy-adapter", got.ID())

	// Second Get() must return the SAME instance without re-calling factory.
	calledAgain := false
	originalCalled := called
	r.RegisterFactory("lazy-adapter", func() adapters.ToolAdapter {
		calledAgain = true
		return &testutil.MockAdapter{IDVal: "lazy-adapter", NameVal: "Different"}
	})
	got2, err := r.Get("lazy-adapter")
	require.NoError(t, err)
	assert.False(t, calledAgain, "factory must not be re-invoked after construction")
	assert.True(t, originalCalled, "original factory must have been called")
	assert.Same(t, got, got2, "second Get() must return the same cached instance")
}

// TestRegistry_RegisterFactory_RegisterStillEager verifies that Register()
// remains eager (backward-compatible) and coexists with RegisterFactory.
func TestRegistry_RegisterFactory_RegisterStillEager(t *testing.T) {
	t.Parallel()

	r := adapters.NewRegistry()
	eager := &testutil.MockAdapter{IDVal: "eager-adapter", NameVal: "Eager"}
	r.Register(eager)

	// Eager-registered adapter must be available immediately — no factory needed.
	got, err := r.Get("eager-adapter")
	require.NoError(t, err)
	assert.Same(t, eager, got)

	// All() includes it.
	all := r.All()
	assert.Len(t, all, 1)
	assert.Same(t, eager, all[0])
}

// =========================================================================
// All() triggers pending factories
// =========================================================================

// TestRegistry_RegisterFactory_AllTriggersConstruction verifies that All()
// invokes all pending factories before returning the snapshot.
func TestRegistry_RegisterFactory_AllTriggersConstruction(t *testing.T) {
	t.Parallel()

	r := adapters.NewRegistry()
	callCount := make(map[string]int)
	registerLazy := func(id, name string) {
		r.RegisterFactory(id, func() adapters.ToolAdapter {
			callCount[id]++
			return &testutil.MockAdapter{IDVal: id, NameVal: name}
		})
	}
	registerLazy("alpha", "Alpha")
	registerLazy("beta", "Beta")
	registerLazy("gamma", "Gamma")

	// Before All(), no factory should have been called.
	require.Zero(t, callCount["alpha"])
	require.Zero(t, callCount["beta"])
	require.Zero(t, callCount["gamma"])

	all := r.All()
	require.Len(t, all, 3)

	// Each factory must have been called exactly once.
	assert.Equal(t, 1, callCount["alpha"])
	assert.Equal(t, 1, callCount["beta"])
	assert.Equal(t, 1, callCount["gamma"])

	// Second All() must NOT re-trigger factories.
	all2 := r.All()
	assert.Len(t, all2, 3)
	assert.Equal(t, 1, callCount["alpha"])
	assert.Equal(t, 1, callCount["beta"])
	assert.Equal(t, 1, callCount["gamma"])
}

// TestRegistry_RegisterFactory_AllSkipsAlreadyConstructed verifies that
// when some factories have already been triggered via Get(), All() only
// invokes the remaining.
func TestRegistry_RegisterFactory_AllSkipsAlreadyConstructed(t *testing.T) {
	t.Parallel()

	r := adapters.NewRegistry()
	callCount := make(map[string]int)
	registerLazy := func(id, name string) {
		r.RegisterFactory(id, func() adapters.ToolAdapter {
			callCount[id]++
			return &testutil.MockAdapter{IDVal: id, NameVal: name}
		})
	}
	registerLazy("a", "A")
	registerLazy("b", "B")

	// Trigger "a" via Get().
	_, err := r.Get("a")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount["a"])
	assert.Equal(t, 0, callCount["b"])

	// All() must only trigger "b", not "a" again.
	all := r.All()
	assert.Len(t, all, 2)
	assert.Equal(t, 1, callCount["a"], "already-constructed adapter must not be re-triggered")
	assert.Equal(t, 1, callCount["b"], "pending adapter must be triggered")
}

// TestRegistry_RegisterFactory_AllOrderPreserved verifies that All() returns
// adapters in registration order regardless of construction timing.
func TestRegistry_RegisterFactory_AllOrderPreserved(t *testing.T) {
	t.Parallel()

	r := adapters.NewRegistry()
	r.RegisterFactory("second", func() adapters.ToolAdapter {
		return &testutil.MockAdapter{IDVal: "second", NameVal: "Second"}
	})
	r.RegisterFactory("first", func() adapters.ToolAdapter {
		return &testutil.MockAdapter{IDVal: "first", NameVal: "First"}
	})

	// Triggers "second" first via Get, then All() triggers "first" but
	// order must remain registration order.
	_, _ = r.Get("second")

	all := r.All()
	require.Len(t, all, 2)
	assert.Equal(t, "second", all[0].ID(), "order must follow registration order")
	assert.Equal(t, "first", all[1].ID())
}

// =========================================================================
// Concurrent Get safety
// =========================================================================

// TestRegistry_RegisterFactory_ConcurrentGet verifies that concurrent Get()
// calls on the same factory invoke it exactly once (sync.Once semantics).
func TestRegistry_RegisterFactory_ConcurrentGet(t *testing.T) {
	t.Parallel()

	r := adapters.NewRegistry()
	var callCount int32
	r.RegisterFactory("shared-lazy", func() adapters.ToolAdapter {
		callCount++
		return &testutil.MockAdapter{IDVal: "shared-lazy", NameVal: "Shared"}
	})

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	results := make([]adapters.ToolAdapter, goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			got, err := r.Get("shared-lazy")
			if err == nil {
				results[i] = got
			}
		}()
	}
	wg.Wait()

	// Factory must have been called exactly once despite 100 concurrent Gets.
	assert.Equal(t, int32(1), callCount, "factory should be invoked exactly once under concurrency")

	// All retrieved instances must be the same object.
	first := results[0]
	for i := 1; i < goroutines; i++ {
		assert.Same(t, first, results[i], "all goroutines must receive the same cached instance")
	}
}

// TestRegistry_RegisterFactory_ConcurrentRegistrationAndGet verifies that
// concurrent RegisterFactory + Get does not race.
func TestRegistry_RegisterFactory_ConcurrentRegistrationAndGet(t *testing.T) {
	t.Parallel()

	r := adapters.NewRegistry()
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: repeatedly register new factories.
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			id := string(rune('A' + (i % 26)))
			r.RegisterFactory(id, func() adapters.ToolAdapter {
				return &testutil.MockAdapter{IDVal: id, NameVal: "Test"}
			})
		}
	}()

	// Goroutine 2: repeatedly call Get on a random ID.
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_, _ = r.Get("Z")
		}
	}()

	wg.Wait()
	// No race = success.
}
