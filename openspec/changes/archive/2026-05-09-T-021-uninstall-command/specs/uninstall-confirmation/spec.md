# Specs: Uninstall Confirmation

## Domain: uninstall-confirmation

### Requirement: Confirmation Gate Before Destructive Uninstall

The system MUST display an interactive confirmation prompt before removing Sequoia files when `--yes` is not set. The user SHALL confirm intent before any filesystem mutation occurs.

| Flag | Behavior |
|------|----------|
| `--yes` / `-y` present | Skip prompt; proceed directly |
| `--yes` absent, stdin is a TTY | Show interactive prompt |
| `--yes` absent, stdin is NOT a TTY | Return error directing user to `--yes` |

**Scenario: --yes skips confirmation**
- GIVEN `--yes` flag is set
- WHEN `sequoia uninstall` runs
- THEN the uninstall MUST proceed without displaying a prompt
- AND no interactive input MUST be read from stdin

**Scenario: Interactive prompt on terminal**
- GIVEN stdin is a terminal AND `--yes` is absent
- WHEN `sequoia uninstall` runs
- THEN the system MUST display a confirmation prompt listing affected tools
- AND MUST wait for user input before proceeding

**Scenario: Confirm with "y" or "Y"**
- GIVEN the confirmation prompt is displayed
- WHEN the user enters "y" or "Y"
- THEN the uninstall MUST proceed with file removal
- AND each adapter's `Uninstall()` MUST be called

**Scenario: Deny with "n" or any other input**
- GIVEN the confirmation prompt is displayed
- WHEN the user enters "n", "N", any other string, or EOF
- THEN the uninstall MUST abort immediately
- AND no files MUST be modified
- AND the user MUST see an "aborted" message

**Scenario: Piped/non-interactive stdin without --yes**
- GIVEN stdin is not a terminal (piped or redirected) AND `--yes` is absent
- WHEN `sequoia uninstall` runs
- THEN an error message MUST be returned
- AND the error MUST mention `--yes` as the required flag
- AND no files MUST be modified

**Scenario: --all with confirmation shows affected tools**
- GIVEN `--all` is set AND `--yes` is absent AND stdin is a terminal
- WHEN `sequoia uninstall` runs
- THEN the confirmation prompt MUST indicate the number of tools affected
- AND the prompt MUST list which tools will have Sequoia removed

### Requirement: Invalid Tool ID Rejection Prior to Confirmation

The system MUST validate `--tool` values before presenting any confirmation prompt. Invalid tool IDs SHALL be rejected regardless of `--yes` flag state.

**Scenario: Invalid --tool with --yes set**
- GIVEN `--tool=no-such-adapter` AND `--yes` is set
- WHEN `sequoia uninstall` runs
- THEN the system MUST return an error mentioning the invalid adapter
- AND no prompt MUST be shown
- AND no files MUST be modified

**Scenario: Invalid --tool without --yes**
- GIVEN `--tool=no-such-adapter` AND `--yes` is absent
- WHEN `sequoia uninstall` runs
- THEN the system MUST return an error before any prompt
- AND the error MUST mention "unknown adapter"
