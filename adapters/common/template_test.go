package common_test

import (
	"embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters/common"
)

// templateTestFS embeds all test template files used by template tests.
//
//go:embed testdata/*.tmpl
var templateTestFS embed.FS

// cacheTestFSA and cacheTestFSB embed separate template directories with
// different content but the same template file name. Used by
// TestRenderTemplateCacheIsolation to verify that different *embed.FS
// instances produce correct, isolated cache entries — FS A never returns
// FS B's content and vice versa. The fix (pointer parameter) makes the
// cache key stable per source, preventing cross-adapter contamination.
//
//go:embed testdata/isol_a/templates
var cacheTestFSA embed.FS

//go:embed testdata/isol_b/templates
var cacheTestFSB embed.FS

// =========================================================================
// TestRenderTemplate_Caching
// =========================================================================

// TestRenderTemplate_Caching verifies that calling RenderTemplate twice with
// the same (fs, name, data) produces identical output. This is the core
// correctness invariant for the template cache – cached and uncached paths
// must produce the same result.
func TestRenderTemplate_Caching(t *testing.T) {
	t.Parallel()

	type data struct {
		Name    string
		Version string
	}
	d := data{Name: "World", Version: "0.1.0"}

	r1, err := common.RenderTemplate(&templateTestFS, "testdata/test.tmpl", d)
	require.NoError(t, err)
	r1 = strings.ReplaceAll(r1, "\r\n", "\n")

	r2, err := common.RenderTemplate(&templateTestFS, "testdata/test.tmpl", d)
	require.NoError(t, err)
	r2 = strings.ReplaceAll(r2, "\r\n", "\n")

	assert.Equal(t, r1, r2, "both calls should produce identical output")
	assert.Equal(t, "Hello World! Version: 0.1.0\n", r1)
}

// =========================================================================
// TestRenderTemplate_DifferentTemplates
// =========================================================================

// TestRenderTemplate_DifferentTemplates verifies that rendering two
// different template files from the same FS produces the correct output
// for each one, with no mixing or cross-contamination.
func TestRenderTemplate_DifferentTemplates(t *testing.T) {
	t.Parallel()

	type data1 struct {
		Name    string
		Version string
	}
	d1 := data1{Name: "Alpha", Version: "1.0"}

	type data2 struct {
		Name  string
		Count int
	}
	d2 := data2{Name: "Beta", Count: 42}

	r1, err := common.RenderTemplate(&templateTestFS, "testdata/test.tmpl", d1)
	require.NoError(t, err)
	r1 = strings.ReplaceAll(r1, "\r\n", "\n")

	r2, err := common.RenderTemplate(&templateTestFS, "testdata/test2.tmpl", d2)
	require.NoError(t, err)
	r2 = strings.ReplaceAll(r2, "\r\n", "\n")

	assert.Equal(t, "Hello Alpha! Version: 1.0\n", r1, "first template output")
	assert.Equal(t, "Goodbye Beta! Count: 42\n", r2, "second template output")
	assert.NotEqual(t, r1, r2, "different templates must produce different output")
}

// =========================================================================
// TestRenderTemplate_CacheIntegrity
// =========================================================================

// TestRenderTemplate_CacheIntegrity verifies that when the same template
// is rendered with different data, the output reflects the current data,
// not stale data from a previous call.
func TestRenderTemplate_CacheIntegrity(t *testing.T) {
	t.Parallel()

	type data struct {
		Name  string
		Count int
	}

	dA := data{Name: "First", Count: 1}
	dB := data{Name: "Second", Count: 99}

	rA, err := common.RenderTemplate(&templateTestFS, "testdata/test2.tmpl", dA)
	require.NoError(t, err)
	rA = strings.ReplaceAll(rA, "\r\n", "\n")

	rB, err := common.RenderTemplate(&templateTestFS, "testdata/test2.tmpl", dB)
	require.NoError(t, err)
	rB = strings.ReplaceAll(rB, "\r\n", "\n")

	assert.Equal(t, "Goodbye First! Count: 1\n", rA, "first call with data A")
	assert.Equal(t, "Goodbye Second! Count: 99\n", rB, "second call with data B")
	assert.NotEqual(t, rA, rB, "different data must produce different output")
}

// =========================================================================
// TestRenderTemplateCacheIsolation
// =========================================================================

// TestRenderTemplateCacheIsolation verifies that two different embed.FS
// instances with the same template name produce their own correct content.
// RenderTemplate reads and parses on every call, so different FS instances
// are always isolated regardless of pointer identity.
func TestRenderTemplateCacheIsolation(t *testing.T) {
	t.Parallel()

	aName := "testdata/isol_a/templates/skill.md.tmpl"
	bName := "testdata/isol_b/templates/skill.md.tmpl"

	// Call A first — caches template from cacheTestFSA.
	rA1, err := common.RenderTemplate(&cacheTestFSA, aName, nil)
	require.NoError(t, err)
	rA1 = strings.ReplaceAll(rA1, "\r\n", "\n")
	assert.Equal(t, "PARENT_A_SKILL\n", rA1, "first call with FS A")

	// Call B — caches template from cacheTestFSB.
	rB, err := common.RenderTemplate(&cacheTestFSB, bName, nil)
	require.NoError(t, err)
	rB = strings.ReplaceAll(rB, "\r\n", "\n")
	assert.Equal(t, "PARENT_B_SKILL\n", rB, "call with FS B — must NOT return FS A content")

	// Call A again — must still return A's content, proving isolation.
	rA2, err := common.RenderTemplate(&cacheTestFSA, aName, nil)
	require.NoError(t, err)
	rA2 = strings.ReplaceAll(rA2, "\r\n", "\n")
	assert.Equal(t, "PARENT_A_SKILL\n", rA2, "second call with FS A — must NOT be polluted by FS B")
}

// =========================================================================
// BenchmarkRenderTemplate — measures performance of RenderTemplate.
// =========================================================================

func BenchmarkRenderTemplate(b *testing.B) {
	type data struct {
		Name    string
		Version string
	}
	d := data{Name: "Bench", Version: "1.0"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := common.RenderTemplate(&templateTestFS, "testdata/test.tmpl", d)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// =========================================================================
// BenchmarkRenderTemplate_Repeated — measures repeated calls to ensure
// the no-cache path has stable performance across many iterations.
// =========================================================================

func BenchmarkRenderTemplate_Repeated(b *testing.B) {
	type data struct {
		Name    string
		Version string
	}
	d := data{Name: "Bench", Version: "1.0"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := common.RenderTemplate(&templateTestFS, "testdata/test.tmpl", d)
		if err != nil {
			b.Fatal(err)
		}
	}
}
