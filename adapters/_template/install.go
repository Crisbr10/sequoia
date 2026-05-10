package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"sequoia-ai/adapters/common"
)

// Version is the Sequoia version embedded in installed skill and rules files.
//
// TODO: Set to the current Sequoia version (match other adapters).
const Version = "0.1.0"

// commandFiles is the ordered list of command template filenames.
//
// TODO: Add or remove commands based on what your tool supports.
// Most adapters use these five commands:
//
//	"sequoia-init.md"
//	"sequoia-audit.md"
//	"sequoia-review.md"
//	"sequoia-fix.md"
//	"sequoia-diff.md"
var commandFiles = []string{
	"sequoia-init.md",
	"sequoia-audit.md",
	"sequoia-review.md",
	"sequoia-fix.md",
	"sequoia-diff.md",
}

// templateData holds variables available to text/template rendering.
//
// TODO: Add tool-specific template variables if needed
// (e.g. tool version, config paths, feature flags).
type templateData struct {
	Version string
}

// renderTemplate renders the named template from templateFS with data.
func renderTemplate(name string, data templateData) (string, error) {
	raw, err := templateFS.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("read template %q: %w", name, err)
	}
	tmpl, err := template.New(name).Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parse template %q: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %q: %w", name, err)
	}
	return buf.String(), nil
}

// runInstaller runs the Prepare → Apply → Verify cycle.
// On Apply or Verify failure it calls Rollback (best-effort) and returns the original error.
func runInstaller(inst *common.Installer) error {
	if err := inst.Prepare(); err != nil {
		return err
	}
	if err := inst.Apply(); err != nil {
		_ = inst.Rollback()
		return err
	}
	if err := inst.Verify(); err != nil {
		_ = inst.Rollback()
		return err
	}
	return nil
}

// stageFile writes content to filepath.Join(dir, name), creating dir if needed.
func stageFile(dir, name string, content []byte) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), content, 0o644)
}
