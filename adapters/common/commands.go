package common

// commandFiles is the ordered list of command template filenames shared by all
// adapters. Each adapter's embed.FS must contain these files under a
// "templates/commands/" directory.
var commandFiles = []string{
	"sequoia-init.md",
	"sequoia-audit.md",
	"sequoia-review.md",
	"sequoia-diff.md",
	"sequoia-dev.md",
}

// CommandFiles returns a defensive copy of the ordered list of command template
// filenames. Callers receive an independent slice — mutating the returned slice
// does not affect the backing list.
func CommandFiles() []string {
	return append([]string{}, commandFiles...)
}

// Commands is an alias for CommandFiles() for backwards compatibility with
// callers that used bare CommandFiles as a value.
func Commands() []string {
	return CommandFiles()
}

// ConfigFiles lists static configuration files that are copied to the
// adapter's base directory (not the commands/ directory). Files are only
// copied if the target does not already exist, preserving user edits.
var ConfigFiles = []struct {
	Source string // filename relative to templates/ in the embed.FS
	Target string // filename relative to the adapter's base directory
}{
	{Source: ".sequoia-dev.yaml.default", Target: ".sequoia-dev.yaml"},
}
