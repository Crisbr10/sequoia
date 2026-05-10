package template

// templateData holds variables available to text/template rendering
// for this tool's templates.
//
// TODO: Add tool-specific template variables if needed
// (e.g. tool version, config paths, feature flags).
type templateData struct {
	Version string
}
