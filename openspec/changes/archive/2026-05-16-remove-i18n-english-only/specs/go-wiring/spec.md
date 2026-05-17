# Delta Spec: go-wiring — MODIFIED

## Action: MODIFIED

Previously: `InstallOpts` had `Language` field; adapters accepted `Install(opts InstallOpts)`; `go.mod` declared `go-i18n/v2` and `x/text`.

| # | Requirement | Scenario |
|---|-------------|----------|
| M11 | `InstallOpts` SHALL NOT have `Language` field | GIVEN `adapters/interface.go` → WHEN compiled → THEN struct lacks any language field |
| M12 | `Install()` / `Uninstall()` SHALL revert to zero-parameter signatures | GIVEN 5 adapters + `_template` → WHEN `Install()` called → THEN no `InstallOpts` argument required |
| M13 | `go.mod` SHALL NOT depend on `go-i18n/v2` or `x/text`; `BurntSushi/toml` SHALL remain | GIVEN `go.mod` after `go mod tidy` → WHEN inspected → THEN `go-i18n/v2` and `x/text` absent; `BurntSushi/toml` present |
| M14 | Pipeline runner SHALL NOT construct language options | GIVEN runner goroutine → WHEN `adapter.Install()` called → THEN no `InstallOpts` argument passed |

## Preserved Requirements (unchanged by this change)

The backup file/directory isolation requirements (0o700/0o600 permissions) remain in the main spec unchanged.
