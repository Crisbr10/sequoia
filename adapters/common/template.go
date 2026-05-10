package common

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

// RenderTemplate reads the named file from fs, parses it as a text/template,
// and executes it with data. The data parameter is passed directly to
// template.Execute and can be any type that the template references.
func RenderTemplate(fs embed.FS, name string, data interface{}) (string, error) {
	raw, err := fs.ReadFile(name)
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
