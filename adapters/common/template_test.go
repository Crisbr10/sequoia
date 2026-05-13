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

	r1, err := common.RenderTemplate(templateTestFS, "testdata/test.tmpl", d)
	require.NoError(t, err)
	r1 = strings.ReplaceAll(r1, "\r\n", "\n")

	r2, err := common.RenderTemplate(templateTestFS, "testdata/test.tmpl", d)
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

	r1, err := common.RenderTemplate(templateTestFS, "testdata/test.tmpl", d1)
	require.NoError(t, err)
	r1 = strings.ReplaceAll(r1, "\r\n", "\n")

	r2, err := common.RenderTemplate(templateTestFS, "testdata/test2.tmpl", d2)
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
// not stale data from a previous call. This ensures the cache does not
// store rendered output — only the parsed template is cached.
func TestRenderTemplate_CacheIntegrity(t *testing.T) {
	t.Parallel()

	type data struct {
		Name  string
		Count int
	}

	dA := data{Name: "First", Count: 1}
	dB := data{Name: "Second", Count: 99}

	rA, err := common.RenderTemplate(templateTestFS, "testdata/test2.tmpl", dA)
	require.NoError(t, err)
	rA = strings.ReplaceAll(rA, "\r\n", "\n")

	rB, err := common.RenderTemplate(templateTestFS, "testdata/test2.tmpl", dB)
	require.NoError(t, err)
	rB = strings.ReplaceAll(rB, "\r\n", "\n")

	assert.Equal(t, "Goodbye First! Count: 1\n", rA, "first call with data A")
	assert.Equal(t, "Goodbye Second! Count: 99\n", rB, "second call with data B")
	assert.NotEqual(t, rA, rB, "different data must produce different output")
}

// =========================================================================
// RenderTemplateLang tests
// =========================================================================

// TestRenderTemplateLang_EnglishFindsFile verifies that when the language is
// "en" and a test.en.tmpl file exists, that language-specific template is
// used instead of the base template.
func TestRenderTemplateLang_EnglishFindsFile(t *testing.T) {
	t.Parallel()

	type data struct {
		Name    string
		Version string
	}
	d := data{Name: "Alice", Version: "1.0"}

	result, err := common.RenderTemplateLang(templateTestFS, "testdata/test", "en", d)
	require.NoError(t, err)
	result = strings.ReplaceAll(result, "\r\n", "\n")

	// The test.en.tmpl appends "(English)" to distinguish it from test.tmpl.
	assert.Equal(t, "Hello Alice! Version: 1.0 (English)\n", result,
		"English template should be used when test.en.tmpl exists")
}

// TestRenderTemplateLang_SpanishFallsBack verifies that when the language is
// "es" but no test.es.tmpl exists, RenderTemplateLang falls back to the base
// test.tmpl (backward compatibility).
func TestRenderTemplateLang_SpanishFallsBack(t *testing.T) {
	t.Parallel()

	type data struct {
		Name    string
		Version string
	}
	d := data{Name: "Bob", Version: "2.0"}

	result, err := common.RenderTemplateLang(templateTestFS, "testdata/test", "es", d)
	require.NoError(t, err)
	result = strings.ReplaceAll(result, "\r\n", "\n")

	// Falls back to test.tmpl (no "(English)" suffix).
	assert.Equal(t, "Hello Bob! Version: 2.0\n", result,
		"base template should be used when test.es.tmpl does not exist")
}

// TestRenderTemplateLang_UnknownLangFallsBack verifies that an unrecognized
// language code also triggers the fallback to the base template.
func TestRenderTemplateLang_UnknownLangFallsBack(t *testing.T) {
	t.Parallel()

	type data struct {
		Name    string
		Version string
	}
	d := data{Name: "Charlie", Version: "3.0"}

	result, err := common.RenderTemplateLang(templateTestFS, "testdata/test", "zh", d)
	require.NoError(t, err)
	result = strings.ReplaceAll(result, "\r\n", "\n")

	// Falls back to test.tmpl (no "(English)" suffix).
	assert.Equal(t, "Hello Charlie! Version: 3.0\n", result,
		"base template should be used when test.zh.tmpl does not exist")
}

// =========================================================================
// BenchmarkRenderTemplate — measures performance of RenderTemplate
// including the first (cold) and subsequent (warm cache) calls.
// =========================================================================

func BenchmarkRenderTemplate(b *testing.B) {
	type data struct {
		Name    string
		Version string
	}
	d := data{Name: "Bench", Version: "1.0"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := common.RenderTemplate(templateTestFS, "testdata/test.tmpl", d)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// =========================================================================
// BenchmarkRenderTemplate_Cached — measures performance after the cache
// is warm. We call RenderTemplate once outside the timing loop to prime
// the cache, then measure only the cached path.
// =========================================================================

func BenchmarkRenderTemplate_Cached(b *testing.B) {
	type data struct {
		Name    string
		Version string
	}
	d := data{Name: "Bench", Version: "1.0"}

	// Prime the cache — first call parses and caches the template.
	_, err := common.RenderTemplate(templateTestFS, "testdata/test.tmpl", d)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := common.RenderTemplate(templateTestFS, "testdata/test.tmpl", d)
		if err != nil {
			b.Fatal(err)
		}
	}
}
