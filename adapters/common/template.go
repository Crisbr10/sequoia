package common

import (
	"bytes"
	"embed"
	"fmt"
	"sync"
	"text/template"
)

// templateCache stores parsed templates keyed by (FS pointer, name).
// text/template.Template is safe for concurrent use. embed.FS is
// immutable (embedded in the binary), making this cache safe.
var templateCache sync.Map

// RenderTemplate reads the named file from fs, parses it as a text/template,
// and executes it with data. Parsed templates are cached in a sync.Map so
// each (fs, name) pair is only parsed once. The data parameter is passed
// directly to template.Execute and can be any type that the template references.
func RenderTemplate(fs embed.FS, name string, data interface{}) (string, error) {
	key := fmt.Sprintf("%p:%s", fs, name)
	if cached, ok := templateCache.Load(key); ok {
		tmpl := cached.(*template.Template)
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("execute template %q: %w", name, err)
		}
		return buf.String(), nil
	}

	raw, err := fs.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("read template %q: %w", name, err)
	}
	tmpl, err := template.New(name).Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parse template %q: %w", name, err)
	}
	templateCache.Store(key, tmpl)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %q: %w", name, err)
	}
	return buf.String(), nil
}
