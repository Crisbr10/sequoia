# agent-p4-quality Specification

## Purpose

Define the P4 Quality agent's deep dependency scanning sub-domain: CVE scanning with severity triage, license compliance across transitive dependency trees, and SBOM generation methodology.

## Requirements

### Requirement: Deep CVE Scanning with Severity Triage
P4 MUST expand its CVE scanning methodology to include severity triage per dependency usage scope. The agent SHALL cross-reference lock files against at least NVD, GitHub Advisory, and OSV databases. CVEs SHALL be prioritized by: severity × usage_scope × exploitability.

#### Scenario: Detects critical CVE in direct production dependency
- GIVEN `go.sum` lists `golang.org/x/net v0.33.0` with CVE-2025-XXXX (critical)
- AND the project imports `golang.org/x/net/http2` in server startup path
- WHEN P4 performs deep CVE scan
- THEN a critical-severity finding SHALL be produced
- AND the finding SHALL cite the CVE ID, patched version, and usage scope (direct→server)

#### Scenario: Downgrades severity for transitive dev dependency CVE
- GIVEN a transitive test-only dependency has a medium CVE
- WHEN P4 performs deep CVE scan
- THEN the finding severity SHALL be downgraded to low
- AND the recommendation SHALL note the dependency is test-only with limited exploit surface

### Requirement: License Compliance with Transitive Tree
P4 MUST extend license verification to cover the full transitive dependency tree, not only direct dependencies. The agent SHALL detect license conflicts where a copyleft transitive dependency affects the project's license obligations.

#### Scenario: Detects GPL-3.0 transitive dependency
- GIVEN a project with MIT license that transitively depends on a GPL-3.0 library
- WHEN P4 scans the full dependency tree
- THEN a high-severity finding SHALL flag the GPL-3.0 dependency
- AND the finding SHALL explain copyleft contagion risk

### Requirement: SBOM Generation Methodology
P4 SHALL include an SBOM generation workflow section in its agent document. The methodology MUST describe extracting dependency metadata from lock files into CycloneDX or SPDX format. The agent document SHALL reference industry-standard SBOM tools appropriate for each language ecosystem but SHALL NOT implement SBOM generation logic.

#### Scenario: SBOM workflow documented for Go projects
- GIVEN the expanded `docs/agents/sequoia-quality.md`
- WHEN the SBOM methodology section is read
- THEN it SHALL describe using `govulncheck` or `syft` for Go SBOM generation
- AND it SHALL specify CycloneDX JSON as the recommended output format
