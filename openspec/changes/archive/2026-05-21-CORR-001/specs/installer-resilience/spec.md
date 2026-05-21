# Installer Resilience Specification

## Purpose

Windows installer tolerates transient network failures with retry and backoff.

## Requirements

### Requirement: Download Retry

`Invoke-WebRequest` in `install.ps1` MUST retry up to 3 times on transient failures (timeout, 5xx). The installer SHALL exit non-zero after all retries exhausted.

#### Scenario: Retry recovers from transient failure

- GIVEN the installer downloads a binary
- WHEN the first request fails with timeout
- THEN the installer SHALL retry after a delay
- AND a subsequent attempt SHALL succeed

#### Scenario: All retries exhausted

- GIVEN the network is unreachable
- WHEN all 3 attempts fail
- THEN the installer SHALL exit with non-zero code
- AND SHALL output a clear error message

### Requirement: Exponential Backoff

Retry delays SHOULD follow exponential backoff (e.g., 2s, 4s, 8s).

#### Scenario: Delays increase between attempts

- GIVEN a download is retried
- WHEN the retry loop executes
- THEN delay before attempt 2 SHOULD be ~2s, before attempt 3 ~4s
