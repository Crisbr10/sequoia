package gemini

// StrategyConfigMerge implements section injection for Gemini's GEMINI.md.
// It uses marker-based delimiters (<!-- sequoia:start -->) to inject and
// remove Sequoia content without modifying content outside the markers.
type StrategyConfigMerge struct {
	path string
}

// NewStrategy creates a StrategyConfigMerge targeting the given file path.
func NewStrategy(path string) *StrategyConfigMerge {
	return &StrategyConfigMerge{path: path}
}

// Inject writes the Sequoia content into the target file using marker injection.
func (s *StrategyConfigMerge) Inject(content string) error {
	return InjectSection(s.path, content)
}

// Remove deletes the Sequoia section from the target file.
func (s *StrategyConfigMerge) Remove() error {
	return RemoveSection(s.path)
}
