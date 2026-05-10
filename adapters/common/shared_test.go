package common_test

import (
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sequoia-ai/adapters/common"
)

//go:embed testdata/test.tmpl
var testFS embed.FS

// TestVersion_IsCorrect verifies the shared Version constant.
func TestVersion_IsCorrect(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "0.1.0", common.Version)
}

// TestCommandFiles_HasExpectedEntries verifies the shared command file list.
func TestCommandFiles_HasExpectedEntries(t *testing.T) {
	t.Parallel()
	expected := []string{
		"sequoia-init.md",
		"sequoia-audit.md",
		"sequoia-review.md",
		"sequoia-fix.md",
		"sequoia-diff.md",
	}
	assert.Equal(t, expected, common.CommandFiles)
}

// TestStageFile_WritesContent creates a temp dir and stages a file.
func TestStageFile_WritesContent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	err := common.StageFile(dir, "hello.txt", []byte("world"))
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	require.NoError(t, err)
	assert.Equal(t, "world", string(data))
}

// TestStageFile_CreatesParentDir verifies StageFile creates missing parent dirs.
func TestStageFile_CreatesParentDir(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "nested", "deep")
	err := common.StageFile(dir, "hello.txt", []byte("ok"))
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(dir, "hello.txt"))
}

// TestRenderTemplate_RendersWithData tests the shared template renderer.
func TestRenderTemplate_RendersWithData(t *testing.T) {
	t.Parallel()

	type data struct {
		Name    string
		Version string
	}
	d := data{Name: "World", Version: "0.1.0"}

	result, err := common.RenderTemplate(testFS, "testdata/test.tmpl", d)
	require.NoError(t, err)
	assert.Equal(t, "Hello World! Version: 0.1.0\n", result)
}

// TestRenderTemplate_FileNotFound returns error for missing template.
func TestRenderTemplate_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := common.RenderTemplate(testFS, "nonexistent.tmpl", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read template")
}

// TestInstaller_Run_Success tests the full Run() lifecycle on a clean install.
func TestInstaller_Run_Success(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "backup")

	writeFile(t, srcDir, "alpha.txt", "content-alpha")
	writeFile(t, srcDir, "beta.txt", "content-beta")

	cfg := common.InstallerConfig{
		SourceDir: srcDir,
		TargetDir: dstDir,
		BackupDir: backupDir,
		Files:     []string{"alpha.txt", "beta.txt"},
	}
	inst := common.NewInstaller(cfg)

	require.NoError(t, inst.Run())

	assert.Equal(t, "content-alpha", readFile(t, dstDir, "alpha.txt"))
	assert.Equal(t, "content-beta", readFile(t, dstDir, "beta.txt"))
}

// TestInstaller_Run_ApplyFailureRollsBack tests Run() when Apply fails.
func TestInstaller_Run_ApplyFailureRollsBack(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "backup")

	// Only one of two files in source — Apply will fail on the second.
	writeFile(t, srcDir, "alpha.txt", "src-alpha")
	writeFile(t, dstDir, "alpha.txt", "orig-alpha")
	writeFile(t, dstDir, "beta.txt", "orig-beta")

	cfg := common.InstallerConfig{
		SourceDir: srcDir,
		TargetDir: dstDir,
		BackupDir: backupDir,
		Files:     []string{"alpha.txt", "beta.txt"},
	}
	inst := common.NewInstaller(cfg)

	err := inst.Run()
	require.Error(t, err, "Run should fail when source file is missing")

	// TargetDir restored to original state.
	assert.Equal(t, "orig-alpha", readFile(t, dstDir, "alpha.txt"))
	assert.Equal(t, "orig-beta", readFile(t, dstDir, "beta.txt"))
}
