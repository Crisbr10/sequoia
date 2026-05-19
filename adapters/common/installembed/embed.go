// Package installembed provides test-only embed.FS instances with template
// files at paths matching what BaseAdapter.Install() expects. This enables
// testing the Install pipeline (including context cancellation checkpoints)
// without modifying production code.
//
// This package MUST NOT be imported by production code.
package installembed

import "embed"

//go:embed templates
var FS embed.FS
