package codex

// templateData holds variables available to text/template rendering
// for OpenAI Codex templates.
type templateData struct {
	Version      string
	SkillsPath   string
	CommandsPath string
}
