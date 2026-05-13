// Package translations embeds the TOML message catalogs for English and Spanish.
// The //go:embed directive bundles *.toml files into the binary at compile time
// so the i18n engine can load them without external dependencies.
package translations

import "embed"

//go:embed *.toml
var FS embed.FS
