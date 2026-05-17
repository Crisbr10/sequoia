# Specs: TUI Core

## Domain: tui-core

### Requirement: Operation Mode Tracking

The Model SHALL track the current operation mode via an `OperationMode` field with values `"install"` or `"uninstall"`. The mode MUST be set when entering the InstallProgress screen from any entry point. Views SHALL read `OperationMode` to render correct operation-specific labels. An empty or unknown mode SHALL be treated as `"install"` for safe fallback.

**Scenario: Mode set on install entry**
- GIVEN the user confirms selections on the Configuration screen
- WHEN the model starts the install pipeline
- THEN `Model.OperationMode` SHALL be `"install"`

**Scenario: Mode set on uninstall entry**
- GIVEN the user confirms uninstallation on the Uninstall screen
- WHEN the model starts the uninstall pipeline
- THEN `Model.OperationMode` SHALL be `"uninstall"`

**Scenario: Safe default for empty mode**
- GIVEN `OperationMode` is its zero value (empty string)
- WHEN a view renders operation-specific labels
- THEN the view SHALL display install-variant labels

### Requirement: Model Without Language State

The Model SHALL NOT carry a `Language` field. All language-related types, constants, and fields SHALL be removed from `model/types.go` and `app/model.go`.

**Scenario: No Language field in structs**
- GIVEN the model types and app model
- WHEN compiled
- THEN no `Language string` field exists in any struct

### Requirement: English-Only Screen Views

All screen view functions SHALL render hardcoded English strings. No view function SHALL accept a `lang string` parameter. No view function SHALL call `i18n.T()`.

**Scenario: Views use English labels**
- GIVEN any screen view function
- WHEN called
- THEN all visible labels are hardcoded English text
- AND the function signature lacks a `lang` parameter

### Requirement: Configuration Screen Single Field

The Configuration screen SHALL have only the persistence backend selector (1 field). The language selector dropdown, `languageOptions`, `languageIndex`, and language field SHALL be removed. The `toggleField` function SHALL be simplified for single-field toggling. The `i18n.Initialized()` guard SHALL be removed.

**Scenario: Configuration shows only backend selector**
- GIVEN the Configuration screen
- WHEN rendered
- THEN no language dropdown exists; only persistence backend selector shown
