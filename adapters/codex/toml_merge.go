package codex

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// MergeSection parses existingTOML into a map, sets/replaces the [sequoia] table
// with section data, and marshals back to a TOML string. All pre-existing keys
// and sections are preserved. An empty existingTOML is treated as a valid empty document.
func MergeSection(existingTOML string, section map[string]interface{}) (string, error) {
	var data map[string]interface{}

	if existingTOML == "" {
		data = make(map[string]interface{})
	} else {
		if _, err := toml.Decode(existingTOML, &data); err != nil {
			return "", fmt.Errorf("toml_merge: parse existing: %w", err)
		}
	}

	// Set or replace the [sequoia] table.
	data["sequoia"] = section

	out, err := toml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("toml_merge: marshal: %w", err)
	}
	return string(out), nil
}

// RemoveSection parses existingTOML into a map, deletes the [sequoia] table
// if present, and marshals back to a TOML string. All other content is preserved.
// If the [sequoia] table is not present, the original content is returned unchanged.
func RemoveSection(existingTOML string) (string, error) {
	if existingTOML == "" {
		return "", nil
	}

	var data map[string]interface{}
	if _, err := toml.Decode(existingTOML, &data); err != nil {
		return "", fmt.Errorf("toml_merge: parse existing: %w", err)
	}

	if _, ok := data["sequoia"]; !ok {
		// No sequoia section — return original unchanged to preserve formatting.
		return existingTOML, nil
	}

	delete(data, "sequoia")

	out, err := toml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("toml_merge: marshal: %w", err)
	}
	return string(out), nil
}
