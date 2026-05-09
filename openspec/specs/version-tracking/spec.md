# Specs: Version Tracking

## Domain: version-tracking

### Requirement: Version Marker Written During Install

Each adapter's `Install()` method MUST write a `.sequoia-version` file into the adapter's skills directory containing the adapter's `Version` constant. The file SHALL be removed during `Uninstall()` as part of cleanup.

**Scenario: Version file written on fresh install**
- GIVEN a clean adapter with no prior Sequoia installation
- WHEN `Install()` completes successfully
- THEN `filepath.Join(SkillsPath(), ".sequoia-version")` MUST exist
- AND its first line MUST equal the adapter's `Version` constant

**Scenario: Version file overwritten on reinstall**
- GIVEN Sequoia is already installed with version `0.0.9`
- WHEN `Install()` is called again with version `0.1.0`
- THEN the `.sequoia-version` file MUST contain `0.1.0`

**Scenario: Version file removed on uninstall**
- GIVEN a `.sequoia-version` file exists
- WHEN `Uninstall()` completes successfully
- THEN the file MUST NOT exist

### Requirement: Version Read During Status Check

Each adapter's `Status()` method MUST read the `.sequoia-version` file and populate `AdapterStatus.Version` with its trimmed content. If the file is missing (legacy install), `Version` SHALL be `""`.

**Scenario: Version present**
- GIVEN `.sequoia-version` contains `0.1.0\n`
- WHEN `Status()` is called
- THEN `AdapterStatus.Version` MUST equal `"0.1.0"`

**Scenario: Legacy install with no version file**
- GIVEN Sequoia is installed but `.sequoia-version` is missing
- WHEN `Status()` is called
- THEN `AdapterStatus.Version` MUST be `""`
- AND `AdapterStatus.Installed` MUST be `true`
