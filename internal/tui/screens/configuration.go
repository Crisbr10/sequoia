package screens

import (
	"fmt"
	"strings"

	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
)

// languageOptions lists the selectable UI languages in display order.
var languageOptions = []struct {
	value model.Language
	label string
}{
	{model.LangEN, "English"},
	{model.LangES, "Español"},
}

// persistenceOptions lists the selectable persistence backends in display order.
var persistenceOptions = []struct {
	value model.PersistenceBackend
	label string
}{
	{model.PersistenceEngram, "Engram"},
	{model.PersistenceFiles, "Files"},
	{model.PersistenceBoth, "Both"},
}

// ConfigurationView renders the Configuration screen showing language and
// persistence selectors. activeField indicates which field is currently
// focused (0 = language, 1 = persistence). engramAvailable controls whether
// the Engram persistence option is selectable or greyed out.
func ConfigurationView(config model.TUIConfig, activeField int, engramAvailable bool) string {
	var b strings.Builder

	// Title.
	b.WriteString(styles.Title().Render("Configuration"))
	b.WriteString("\n\n")

	// === Language field ===
	b.WriteString(renderFieldLabel("Language", activeField == 0))
	b.WriteString("\n")

	// Render language options.
	langIdx := languageIndex(config.Language)
	for i, opt := range languageOptions {
		cursorMark := "  "
		if activeField == 0 && i == langIdx {
			cursorMark = styles.Accent().Render("► ")
		}
		highlighted := i == langIdx
		label := opt.label
		if highlighted && activeField == 0 {
			label = styles.Accent().Render(label)
		} else if highlighted {
			label = styles.Success().Render(label)
		}
		fmt.Fprintf(&b, "  %s%s\n", cursorMark, label)
	}
	b.WriteString("\n")

	// === Persistence field ===
	b.WriteString(renderFieldLabel("Persistence", activeField == 1))
	b.WriteString("\n")

	// Render persistence options.
	persIdx := persistenceIndex(config.Persistence)
	for i, opt := range persistenceOptions {
		cursorMark := "  "
		if activeField == 1 && i == persIdx {
			cursorMark = styles.Accent().Render("► ")
		}
		highlighted := i == persIdx
		label := opt.label
		if highlighted && activeField == 1 {
			label = styles.Accent().Render(label)
		} else if highlighted {
			label = styles.Success().Render(label)
		}

		// Engram unavailable note.
		extra := ""
		if opt.value == model.PersistenceEngram && !engramAvailable {
			extra = styles.Muted().Render(" (not detected)")
			if highlighted {
				label = styles.Muted().Render(label)
			} else {
				label = styles.Muted().Render(label)
			}
		}

		fmt.Fprintf(&b, "  %s%s%s\n", cursorMark, label, extra)
	}
	b.WriteString("\n")

	// Footer hints.
	b.WriteString(styles.Muted().Render("  "))
	b.WriteString(styles.Accent().Render("Tab"))
	b.WriteString(styles.Muted().Render(" switch field  "))
	b.WriteString(styles.Accent().Render("↑/↓/←/→"))
	b.WriteString(styles.Muted().Render(" change option  "))
	b.WriteString(styles.Accent().Render("Enter"))
	b.WriteString(styles.Muted().Render(" confirm  "))
	b.WriteString(styles.Accent().Render("Esc"))
	b.WriteString(styles.Muted().Render(" back"))

	return b.String()
}

// renderFieldLabel renders a field label, optionally highlighted when active.
func renderFieldLabel(name string, active bool) string {
	label := fmt.Sprintf("  %s:", name)
	if active {
		return styles.Accent().Render(label)
	}
	return styles.Body().Render(label)
}

// ConfigurationUpdate processes key events for the Configuration screen.
// activeField is 0 (language) or 1 (persistence). config holds current selections.
// Returns new active field, updated config, and action ("confirm", "back", "quit", or "").
func ConfigurationUpdate(msg tea.KeyMsg, activeField int, config model.TUIConfig, engramAvailable bool) (newActiveField int, newConfig model.TUIConfig, action string) {
	switch msg.Type {
	case tea.KeyTab:
		return toggleField(activeField), config, ""

	case tea.KeyUp:
		return cycleOption(activeField, config, engramAvailable, -1)

	case tea.KeyDown:
		return cycleOption(activeField, config, engramAvailable, 1)

	case tea.KeyLeft:
		return cycleOption(activeField, config, engramAvailable, -1)

	case tea.KeyRight:
		return cycleOption(activeField, config, engramAvailable, 1)

	case tea.KeyEnter:
		return activeField, config, "confirm"

	case tea.KeyEsc:
		return activeField, config, "back"
	}

	return activeField, config, ""
}

// toggleField switches activeField between 0 and 1 (language ↔ persistence).
func toggleField(field int) int {
	if field == 0 {
		return 1
	}
	return 0
}

// cycleOption advances or retreats the option within the active field.
// direction: +1 for right, -1 for left.
func cycleOption(activeField int, config model.TUIConfig, engramAvailable bool, direction int) (int, model.TUIConfig, string) {
	switch activeField {
	case 0:
		// Language field: cycle between EN and ES.
		idx := languageIndex(config.Language)
		idx = (idx + direction + len(languageOptions)) % len(languageOptions)
		config.Language = string(languageOptions[idx].value)
	case 1:
		// Persistence field: cycle through Engram/Files/Both.
		idx := persistenceIndex(config.Persistence)
		idx = (idx + direction + len(persistenceOptions)) % len(persistenceOptions)
		// Skip Engram if unavailable.
		if !engramAvailable && persistenceOptions[idx].value == model.PersistenceEngram {
			idx = (idx + direction + len(persistenceOptions)) % len(persistenceOptions)
		}
		config.Persistence = string(persistenceOptions[idx].value)
	}
	return activeField, config, ""
}

// languageIndex returns the index of the given language value in languageOptions.
func languageIndex(lang string) int {
	for i, opt := range languageOptions {
		if string(opt.value) == lang {
			return i
		}
	}
	return 0
}

// persistenceIndex returns the index of the given persistence value in persistenceOptions.
func persistenceIndex(pers string) int {
	for i, opt := range persistenceOptions {
		if string(opt.value) == pers {
			return i
		}
	}
	return 0
}
