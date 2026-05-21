package common

import "embed"

//go:embed templates/commands
var CommandFS embed.FS

//go:embed templates/.sequoia-dev.yaml.default
var ConfigDefaultFS embed.FS
