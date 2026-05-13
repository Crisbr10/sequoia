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

	got, err := adapters.NewAdapter("factory-test-known")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "factory-test-known", got.ID())
}

func TestFactory_NewAdapter_UnknownID(t *testing.T) {
	t.Parallel()

	_, err := adapters.NewAdapter("this-id-was-never-registered-xyz123")
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
