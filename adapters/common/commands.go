package common

// CommandFiles is the ordered list of command template filenames shared by all
// adapters. Each adapter's embed.FS must contain these files under a
// "templates/commands/" directory.
var CommandFiles = []string{
	"sequoia-init.md",
	"sequoia-audit.md",
	"sequoia-review.md",
	"sequoia-fix.md",
	"sequoia-diff.md",
}
