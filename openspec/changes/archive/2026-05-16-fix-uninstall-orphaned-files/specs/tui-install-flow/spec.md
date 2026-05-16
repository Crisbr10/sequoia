# Delta for tui-install-flow

## MODIFIED Requirements

### Requirement: Complete Screen

The Complete screen MUST list succeeded tools. `q` exits the TUI. The heading SHALL display "Installation Complete!" when `OperationMode` is `"install"` and "Uninstallation Complete!" when `OperationMode` is `"uninstall"`. The summary line SHALL use `complete.installed_items` for install, `complete.uninstalled_items` for clean uninstall (zero warnings), and `complete.warnings_note` when uninstall completed with warnings.

(Previously: summary line used install-variant message for ALL cases — clean uninstalls displayed "Installed: Skills, Commands, System Prompt" under an "Uninstallation Complete" heading.)

| # | Scenario | GIVEN | WHEN | THEN |
|---|----------|-------|------|------|
| 14 | Install complete — summary | mode=install, all succeeded | Install finishes | Heading "Installation Complete!"; summary "Installed: Skills, Commands, System Prompt" |
| 15 | Partial success | 1 succeeded, 1 failed, mode=install | Install finishes | Succeeded listed with correct summary; failed deferred to Error screen |
| 16 | Clean uninstall — summary | mode=uninstall, all succeeded, warnedCount=0 | Uninstall finishes | Heading "Uninstallation Complete!"; summary "Uninstalled: Skills, Commands, System Prompt" |
| 17 | Uninstall with warnings | mode=uninstall, warnedCount>0 | Uninstall finishes | Heading "Uninstallation Complete! (N with warnings)"; warnings summary displayed — NOT item list |
| 18 | i18n key `complete.uninstalled_items` exists | i18n catalog loaded, lang="en" or "es" | T("complete.uninstalled_items", lang) | Returns localized "Uninstalled: Skills, Commands, System Prompt" (en) / "Desinstalado: Skills, Commands, System Prompt" (es) |
