# Delta for tui-core

## ADDED Requirements

### Requirement: Operation Mode Tracking

The Model SHALL track the current operation mode via an `OperationMode` field with values `"install"` or `"uninstall"`. The mode MUST be set when entering the InstallProgress screen from any entry point. Views SHALL read `OperationMode` to render correct operation-specific labels. An empty or unknown mode SHALL be treated as `"install"` for safe fallback.

#### Scenario: Mode set on install entry

- GIVEN the user confirms selections on the Configuration screen
- WHEN the model starts the install pipeline
- THEN `Model.OperationMode` SHALL be `"install"`

#### Scenario: Mode set on uninstall entry

- GIVEN the user confirms uninstallation on the Uninstall screen
- WHEN the model starts the uninstall pipeline
- THEN `Model.OperationMode` SHALL be `"uninstall"`

#### Scenario: Safe default for empty mode

- GIVEN `OperationMode` is its zero value (empty string)
- WHEN a view renders operation-specific labels
- THEN the view SHALL display install-variant labels
