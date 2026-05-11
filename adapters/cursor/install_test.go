package cursor_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
	"github.com/Crisbr10/sequoia/adapters/cursor"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstall_CreatesAllFiles(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := cursor.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))

	assert.FileExists(t, filepath.Join(a.SkillsPath(), "SKILL.md"))

	for _, cmd := range []string{
		"sequoia-init.md",
		"sequoia-audit.md",
		"sequoia-review.md",
		"sequoia-fix.md",
		"sequoia-diff.md",
	} {
		assert.FileExists(t, filepath.Join(a.CommandsPath(), cmd))
	}

	assert.FileExists(t, a.SystemPromptPath())
}

func TestInstall_IsInstalledAfterInstall(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := cursor.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))
	assert.True(t, a.IsInstalled())
}

func TestInstall_RulesMDHasVersion(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := cursor.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))

	raw, err := os.ReadFile(a.SystemPromptPath())
	require.NoError(t, err)
	assert.Contains(t, string(raw), common.Version)
}

func TestStatus_AfterInstall(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := cursor.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))

	status := a.Status()
	assert.True(t, status.Installed)
	assert.NotEmpty(t, status.Path)
	assert.Equal(t, common.Version, status.Version)
}

func TestInstall_WritesVersionFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := cursor.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))

	versionFile := filepath.Join(a.SkillsPath(), ".sequoia-version")
	data, err := os.ReadFile(versionFile)
	require.NoError(t, err)
	assert.Equal(t, common.Version, string(data))
}

func TestUninstall_RemovesVersionFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := cursor.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))

	versionFile := filepath.Join(a.SkillsPath(), ".sequoia-version")
	_, err := os.Stat(versionFile)
	require.NoError(t, err, "version file must exist before uninstall")

	require.NoError(t, a.Uninstall(adapters.InstallOpts{}))

	_, err = os.Stat(versionFile)
	assert.True(t, os.IsNotExist(err), "version file should be removed by Uninstall")
}

func TestVerify_AllFilesReadable(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := cursor.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))

	skillPath := filepath.Join(a.SkillsPath(), "SKILL.md")
	assert.FileExists(t, skillPath)
	raw, err := os.ReadFile(skillPath)
	require.NoError(t, err)
	assert.NotEmpty(t, raw)

	for _, cmd := range []string{
		"sequoia-init.md",
		"sequoia-audit.md",
		"sequoia-review.md",
		"sequoia-fix.md",
		"sequoia-diff.md",
	} {
		cmdPath := filepath.Join(a.CommandsPath(), cmd)
		assert.FileExists(t, cmdPath)
		raw, err := os.ReadFile(cmdPath)
		require.NoError(t, err)
		assert.NotEmpty(t, raw, "command file %s should not be empty", cmd)
	}

	raw, err = os.ReadFile(a.SystemPromptPath())
	require.NoError(t, err)
	assert.NotEmpty(t, raw)
}
