## v0.1.2

Patch release — fixes version detection and installer reliability.

### Fixes

- **Version detection**: `sequoia version` now correctly reports the release version via `debug.ReadBuildInfo` fallback
- **Installer arch detection**: replaced fragile OSArchitecture API with simple, reliable `PROCESSOR_ARCHITECTURE` check (same as gentle-ai)
- **Installer asset URLs**: fixed `v` prefix mismatch between tags and asset filenames
- **PowerShell installer**: terminal no longer closes immediately after install

### Install

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/Crisbr10/sequoia/main/scripts/install.sh | bash

# Windows
irm https://raw.githubusercontent.com/Crisbr10/sequoia/main/scripts/install.ps1 | iex
```

### Upgrade

```bash
sequoia status                # check current version
# re-run the install script — it detects and upgrades automatically
```
