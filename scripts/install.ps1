#Requires -Version 5.1
<#
.SYNOPSIS
    Sequoia One-Line Installer (Windows PowerShell)
.DESCRIPTION
    Downloads, verifies, and installs the Sequoia CLI on Windows.
    Equivalent to the Unix install.sh, adapted for PowerShell 5.1+.
.PARAMETER Repo
    GitHub org/repo (default: Crisbr10/sequoia).
.PARAMETER Version
    Release version tag (default: latest, resolved via GitHub API).
.PARAMETER InstallDir
    Target directory for the binary (default: $env:LOCALAPPDATA\sequoia).
.PARAMETER SkipChecksum
    Skip SHA-256 verification of the downloaded archive.
    Opt-in flag for air-gapped environments where checksums.txt is unreachable.
    Without this flag, checksum verification is MANDATORY — the installer
    will abort if checksums.txt cannot be downloaded.
.PARAMETER NoPath
    Skip adding INSTALL_DIR to the user-level PATH environment variable.
    By default, the installer adds INSTALL_DIR to PATH so 'sequoia' is
    available globally from any terminal. Use -NoPath to opt out.
.EXAMPLE
    irm https://raw.githubusercontent.com/Crisbr10/sequoia/main/scripts/install.ps1 | iex
.EXAMPLE
    .\install.ps1 -Version v0.2.0 -InstallDir "C:\tools\sequoia"
.EXAMPLE
    .\install.ps1 -NoPath
#>

param(
    [string]$Repo = "Crisbr10/sequoia",
    [string]$Version = "latest",
    [string]$InstallDir = "$env:LOCALAPPDATA\sequoia",
    [switch]$SkipChecksum,
    [switch]$NoPath
)

# -- Configuration ------------------------------------------------------------
$Binary   = "sequoia.exe"
$ProgressPreference = "SilentlyContinue"  # Speed up Invoke-WebRequest

# Exit codes (matched to design contract)
Set-Variable -Name EXIT_OK      -Value 0  -Option ReadOnly
Set-Variable -Name EXIT_GENERAL -Value 1  -Option ReadOnly
Set-Variable -Name EXIT_CHECKSUM -Value 2  -Option ReadOnly
Set-Variable -Name EXIT_NETWORK -Value 3  -Option ReadOnly

# -- Color helpers ------------------------------------------------------------
function Write-Info  { Write-Host "[INFO]  $args" -ForegroundColor Green }
function Write-Warn  { Write-Host "[WARN]  $args" -ForegroundColor Yellow }
function Write-Err   { Write-Host "[ERROR] $args" -ForegroundColor Red }

# -- Retry wrapper (exponential backoff: 2s, 4s, 8s) -------------------------
function Invoke-WebRequestWithRetry {
    param([string]$Uri, [string]$OutFile)
    $delays = @(2, 4, 8)
    for ($i = 0; $i -lt 3; $i++) {
        try {
            Invoke-WebRequest -Uri $Uri -OutFile $OutFile -UseBasicParsing -ErrorAction Stop
            return
        } catch {
            if ($i -eq 2) { throw }
            Write-Warn "Attempt $($i+1) failed, retrying in $($delays[$i])s..."
            Start-Sleep -Seconds $delays[$i]
        }
    }
}

# -- Path sanitization --------------------------------------------------------
function Resolve-SafePath {
    param([string]$Path, [string]$Context)

    if ([string]::IsNullOrWhiteSpace($Path)) {
        Write-Err "$Context must not be empty"
        exit $EXIT_GENERAL
    }

    if (-not [System.IO.Path]::IsPathRooted($Path)) {
        Write-Err "$Context must be absolute (e.g., C:\Tools\sequoia)"
        exit $EXIT_GENERAL
    }

    $forbidden = @('..', ';', '|', '&', '`', '$(')
    foreach ($bad in $forbidden) {
        if ($Path.Contains($bad)) {
            Write-Err "$Context contains forbidden: '$bad'"
            exit $EXIT_GENERAL
        }
    }

    if ($Path -match '[<>\"*?\x00-\x1f]') {
        Write-Err "$Context has invalid Windows path chars"
        exit $EXIT_GENERAL
    }
}

# -- OS / Arch detection ------------------------------------------------------
function Get-NormalizedArch {
    # Simple, reliable detection (same approach as gentle-ai)
    if (-not [Environment]::Is64BitOperatingSystem) {
        Write-Err "32-bit Windows is not supported"
        exit $EXIT_GENERAL
    }
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
        return "arm64"
    }
    return "amd64"
}

$OS   = "windows"
$Arch = Get-NormalizedArch

# -- Version resolution -------------------------------------------------------
function Resolve-Version {
    param([string]$VersionInput)

    if ($VersionInput -ne "latest") {
        return $VersionInput
    }

    Write-Info "Resolving latest version for $Repo..."
    $apiUrl = "https://api.github.com/repos/$Repo/releases/latest"

    try {
        $response = Invoke-WebRequest -Uri $apiUrl -UseBasicParsing -ErrorAction Stop
        $json = $response.Content | ConvertFrom-Json
        $tag = $json.tag_name

        if (-not $tag) {
            throw "tag_name not found in API response"
        }

        return $tag
    } catch {
        Write-Err "Failed to fetch latest release info from GitHub."
        Write-Err "Check your internet connection or set -Version explicitly (e.g. -Version v0.1.0)."
        Write-Err "Error: $_"
        exit $EXIT_NETWORK
    }
}

$ResolvedVersion = Resolve-Version -VersionInput $Version

# Strip "v" prefix for asset filenames (tags are v0.1.1, assets use 0.1.1)
$VersionNumber = $ResolvedVersion.TrimStart("v")

# Validate install path before any I/O
Resolve-SafePath -Path $InstallDir -Context "InstallDir"

# -- Download URLs ------------------------------------------------------------
$Tarball     = "sequoia_${VersionNumber}_${OS}_${Arch}.zip"
$DownloadUrl = "https://github.com/$Repo/releases/download/$ResolvedVersion/$Tarball"
$ChecksumUrl = "https://github.com/$Repo/releases/download/$ResolvedVersion/sequoia_${VersionNumber}_checksums.txt"

# -- Idempotency check --------------------------------------------------------
function Test-SequoiaInstalled {
    $target = Join-Path -Path $InstallDir -ChildPath $Binary

    if (-not (Test-Path -Path $target)) {
        return $false
    }

    try {
        $installedVersion = & $target version 2>&1 | Out-String
        $installedVersion = $installedVersion.Trim()
    } catch {
        Write-Warn "Existing binary at $target but 'version' command failed. Reinstalling..."
        return $false
    }

    if ($installedVersion -eq $ResolvedVersion) {
        Write-Info "Sequoia $ResolvedVersion is already installed at $target"
        return $true
    }

    Write-Info "Sequoia $installedVersion found at $target, upgrading to $ResolvedVersion..."
    return $false
}

if (Test-SequoiaInstalled) {
    # Success — already up to date. Keep terminal open so user can see the message.
    Write-Host ""
    Read-Host "Press Enter to exit"
    return  # return, not exit — preserves caller's PowerShell session
}

# -- Temp directory -----------------------------------------------------------
$TempDir = Join-Path -Path $env:TEMP -ChildPath "sequoia-install-$(Get-Random)"
Resolve-SafePath -Path $TempDir -Context "TempDir"
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

try {
    # -- Download -------------------------------------------------------------
    Write-Info "Downloading Sequoia $ResolvedVersion for windows/$Arch..."
    Write-Info "  URL: $DownloadUrl"

    try {
        Invoke-WebRequestWithRetry -Uri $DownloadUrl -OutFile (Join-Path $TempDir $Tarball)
    } catch {
        # REQ-IER-01: Distinguish HTTP 404 (source-only release or missing tag)
        # from generic network errors. PS 5.1's $_.Exception.Response may be
        # null on some failures, so we also string-match as a fallback.
        $is404 = $false
        if ($_.Exception.Response -and $_.Exception.Response.StatusCode) {
            if ($_.Exception.Response.StatusCode.value__ -eq 404) {
                $is404 = $true
            }
        }
        if ($_ -match '404|Not Found') {
            $is404 = $true
        }

        if ($is404) {
            # Query the GitHub Releases API to see if the tag exists.
            # If it does, this is a source-only release (tag exists, no asset).
            try {
                $releasesApiUrl = "https://api.github.com/repos/$Repo/releases/tags/$ResolvedVersion"
                $apiResponse = Invoke-WebRequest -Uri $releasesApiUrl -UseBasicParsing -TimeoutSec 10 -ErrorAction Stop
                if ($apiResponse.StatusCode -eq 200) {
                    # Tag exists but the binary asset was not found — source-only release.
                    Write-Err "Release $ResolvedVersion exists on GitHub but no precompiled binary asset was found."
                    Write-Err "This is a source-only release (published without GoReleaser binaries)."
                    Write-Err "Try installing a previous version, e.g.:"
                    Write-Err "  irm https://raw.githubusercontent.com/$Repo/main/scripts/install.ps1 | iex -Version \$PrevVersion"
                    Write-Err "Or report this at: https://github.com/Crisbr10/sequoia/issues"
                    exit $EXIT_NETWORK
                }
            } catch {
                # API lookup failed (timeout, DNS, rate limit, etc.) —
                # fall through to the generic error message.
            }
        }

        Write-Err "Download failed. Please check:"
        Write-Err "  - Internet connectivity"
        Write-Err "  - Repo=$Repo (correct GitHub org/repo?)"
        Write-Err "  - Version=$ResolvedVersion (tag exists?)"
        Write-Err "  - Error: $_"
        exit $EXIT_NETWORK
    }

    # -- SHA-256 checksum verification ----------------------------------------
    if (-not $SkipChecksum) {
        Write-Info "Verifying SHA-256 checksum..."

        $checksumsPath = Join-Path $TempDir "checksums.txt"
        $checksumsDownloaded = $false

        try {
            Invoke-WebRequestWithRetry -Uri $ChecksumUrl -OutFile $checksumsPath
            $checksumsDownloaded = $true
        } catch {
            if ($SkipChecksum) {
                Write-Warn "Could not download checksums.txt. Skipping verification (-SkipChecksum)."
            } else {
                Write-Err "Could not download checksums.txt from:"
                Write-Err "  $ChecksumUrl"
                Write-Err ""
                Write-Err "Checksum verification is mandatory. The binary cannot be verified."
                Write-Err "To bypass this check (air-gapped environments), download the script"
                Write-Err "and run it with -SkipChecksum:"
                Write-Err ""
                Write-Err "  .\install.ps1 -SkipChecksum"
                exit $EXIT_CHECKSUM
            }
        }

        if ($checksumsDownloaded -and (Test-Path $checksumsPath)) {
            $archivePath = Join-Path $TempDir $Tarball
            $computedHash = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLower()

            # Search for the tarball name in checksums.txt
            $expectedLine = Get-Content $checksumsPath | Where-Object { $_ -match [regex]::Escape($Tarball) } | Select-Object -First 1
            if ($expectedLine) {
                $expectedHash = ($expectedLine -split '\s+')[0].ToLower()

                if ($computedHash -ne $expectedHash) {
                    Write-Err "SHA-256 checksum mismatch!"
                    Write-Err "  Expected: $expectedHash"
                    Write-Err "  Got:      $computedHash"
                    Write-Err "The downloaded file may be corrupt or tampered with. Aborting."
                    exit $EXIT_CHECKSUM
                }

                Write-Info "Checksum verified: $computedHash"
            } else {
                Write-Warn "No checksum entry found for $Tarball in checksums.txt. Skipping verification."
            }
        }
    } else {
        Write-Warn "Checksum verification skipped (--SkipChecksum)."
    }

    # -- Cosign signature verification ----------------------------------------
    # Attempt cryptographic signature verification (REQ-IV-002).
    # cosign absence → warn + offer install (SHA-256 remains the baseline)
    # .sigstore.json bundle download failure → fall back to .sig/.cert (backward compat)
    # Verification failure → error exit (signature does not match → tampered binary)
    # Cosign is additive — -SkipChecksum does NOT skip this step.
    $bundleUrl = "https://github.com/$Repo/releases/download/$ResolvedVersion/$Tarball.sigstore.json"
    $sigUrl  = "https://github.com/$Repo/releases/download/$ResolvedVersion/$Tarball.sig"
    $certUrl = "https://github.com/$Repo/releases/download/$ResolvedVersion/$Tarball.cert"
    $bundleFile = Join-Path $TempDir "$Tarball.sigstore.json"
    $sigFile = Join-Path $TempDir "$Tarball.sig"
    $certFile = Join-Path $TempDir "$Tarball.cert"

    # Detect cosign under any known binary name (winget installs as cosign-windows-amd64)
    $cosignCmd = $null
    foreach ($name in @('cosign', 'cosign-windows-amd64')) {
        $found = Get-Command $name -ErrorAction SilentlyContinue
        if ($found) { $cosignCmd = $found.Name; break }
    }

    if ($cosignCmd) {
        Write-Info "Cosign detected ($cosignCmd) — verifying cryptographic signature..."

        # Try .sigstore.json bundle first (cosign v3 format)
        $bundleDownloadOk = $true
        try {
            Invoke-WebRequest -Uri $bundleUrl -OutFile $bundleFile -UseBasicParsing -ErrorAction Stop
        } catch {
            $bundleDownloadOk = $false
        }

        if ($bundleDownloadOk -and (Test-Path $bundleFile)) {
            # New format: verify with --bundle
            $archivePath = Join-Path $TempDir $Tarball
            $cosignArgs = @(
                "verify-blob",
                "--bundle", $bundleFile,
                "--certificate-oidc-issuer", "https://token.actions.githubusercontent.com",
                "--certificate-identity", "https://github.com/$Repo/.github/workflows/release.yml@refs/tags/$ResolvedVersion",
                $archivePath
            )

            & $cosignCmd $cosignArgs *>$null
            if ($LASTEXITCODE -eq 0) {
                Write-Info "Cosign signature verified (sigstore bundle)"
            } else {
                Write-Err "Cosign signature verification FAILED."
                Write-Err "The downloaded binary may have been tampered with. Aborting."
                exit $EXIT_CHECKSUM
            }
        } else {
            # Fallback: try legacy .sig + .cert (pre-migration releases)
            $sigDownloadOk = $true
            try {
                Invoke-WebRequest -Uri $sigUrl -OutFile $sigFile -UseBasicParsing -ErrorAction Stop
            } catch {
                $sigDownloadOk = $false
            }
            try {
                Invoke-WebRequest -Uri $certUrl -OutFile $certFile -UseBasicParsing -ErrorAction Stop
            } catch {
                $sigDownloadOk = $false
            }

            if (-not $sigDownloadOk) {
                Write-Warn "Failed to download signature files. Skipping cosign verification."
                Write-Warn "SHA-256 checksum remains the integrity baseline."
            } else {
                $archivePath = Join-Path $TempDir $Tarball
                $cosignArgs = @(
                    "verify-blob",
                    "--signature", $sigFile,
                    "--certificate", $certFile,
                    "--certificate-oidc-issuer", "https://token.actions.githubusercontent.com",
                    "--certificate-identity", "https://github.com/$Repo/.github/workflows/release.yml@refs/tags/$ResolvedVersion",
                    $archivePath
                )

                & $cosignCmd $cosignArgs *>$null
                if ($LASTEXITCODE -eq 0) {
                    Write-Info "Cosign signature verified (legacy format)"
                } else {
                    Write-Err "Cosign signature verification FAILED."
                    Write-Err "The downloaded binary may have been tampered with. Aborting."
                    exit $EXIT_CHECKSUM
                }
            }
        }
    } else {
        Write-Warn "Cosign is not installed. Skipping cryptographic signature verification."

        # Offer automatic installation if winget is available
        $wingetFound = Get-Command winget -ErrorAction SilentlyContinue
        if ($wingetFound) {
            $choice = Read-Host "Would you like to install Cosign automatically via winget? (y/N)"
            if ($choice -match '^[yY]') {
                Write-Info "Installing Cosign via winget..."
                winget install sigstore.cosign --accept-source-agreements --accept-package-agreements 2>&1 | Out-Null
                if ($LASTEXITCODE -eq 0) {
                    Write-Info "Cosign installed successfully. Re-running signature verification..."
                    # Re-detect after install — winget may create cosign-windows-amd64
                    foreach ($name in @('cosign', 'cosign-windows-amd64')) {
                        $found = Get-Command $name -ErrorAction SilentlyContinue
                        if ($found) { $cosignCmd = $found.Name; break }
                    }
                    if ($cosignCmd) {
                        # Try .sigstore.json bundle first (cosign v3 format)
                        $bundleDownloadOk = $true
                        try { Invoke-WebRequest -Uri $bundleUrl -OutFile $bundleFile -UseBasicParsing -ErrorAction Stop } catch { $bundleDownloadOk = $false }

                        if ($bundleDownloadOk -and (Test-Path $bundleFile)) {
                            $archivePath = Join-Path $TempDir $Tarball
                            $cosignArgs = @(
                                "verify-blob",
                                "--bundle", $bundleFile,
                                "--certificate-oidc-issuer", "https://token.actions.githubusercontent.com",
                                "--certificate-identity", "https://github.com/$Repo/.github/workflows/release.yml@refs/tags/$ResolvedVersion",
                                $archivePath
                            )
                            & $cosignCmd $cosignArgs *>$null
                            if ($LASTEXITCODE -eq 0) {
                                Write-Info "Cosign signature verified (sigstore bundle)"
                            } else {
                                Write-Err "Cosign signature verification FAILED."
                                Write-Err "The downloaded binary may have been tampered with. Aborting."
                                exit $EXIT_CHECKSUM
                            }
                        } else {
                            # Fallback: legacy .sig + .cert
                            $sigDownloadOk = $true
                            try { Invoke-WebRequest -Uri $sigUrl -OutFile $sigFile -UseBasicParsing -ErrorAction Stop } catch { $sigDownloadOk = $false }
                            try { Invoke-WebRequest -Uri $certUrl -OutFile $certFile -UseBasicParsing -ErrorAction Stop } catch { $sigDownloadOk = $false }
                            if (-not $sigDownloadOk) {
                                Write-Warn "Failed to download signature files. Skipping cosign verification."
                                Write-Warn "SHA-256 checksum remains the integrity baseline."
                            } else {
                                $archivePath = Join-Path $TempDir $Tarball
                                $cosignArgs = @(
                                    "verify-blob",
                                    "--signature", $sigFile,
                                    "--certificate", $certFile,
                                    "--certificate-oidc-issuer", "https://token.actions.githubusercontent.com",
                                    "--certificate-identity", "https://github.com/$Repo/.github/workflows/release.yml@refs/tags/$ResolvedVersion",
                                    $archivePath
                                )
                                & $cosignCmd $cosignArgs *>$null
                                if ($LASTEXITCODE -eq 0) {
                                    Write-Info "Cosign signature verified (legacy format)"
                                } else {
                                    Write-Err "Cosign signature verification FAILED."
                                    Write-Err "The downloaded binary may have been tampered with. Aborting."
                                    exit $EXIT_CHECKSUM
                                }
                            }
                        }
                    }
                } else {
                    Write-Warn "Cosign installation via winget failed."
                    Write-Warn "Install cosign for stronger integrity guarantees:"
                    Write-Warn "  winget install sigstore.cosign"
                    Write-Warn "  Or see https://docs.sigstore.dev/cosign/installation/"
                    Write-Warn "SHA-256 checksum remains the integrity baseline."
                }
            } else {
                Write-Warn "Install cosign for stronger integrity guarantees:"
                Write-Warn "  winget install sigstore.cosign"
                Write-Warn "  Or see https://docs.sigstore.dev/cosign/installation/"
                Write-Warn "SHA-256 checksum remains the integrity baseline."
            }
        } else {
            Write-Warn "Install cosign for stronger integrity guarantees:"
            Write-Warn "  winget install sigstore.cosign"
            Write-Warn "  Or see https://docs.sigstore.dev/cosign/installation/"
            Write-Warn "SHA-256 checksum remains the integrity baseline."
        }
    }

    # -- Extract --------------------------------------------------------------
    Write-Info "Extracting $Tarball..."
    $extractDir = Join-Path $TempDir "extracted"
    Expand-Archive -Path (Join-Path $TempDir $Tarball) -DestinationPath $extractDir -Force

    # Find the binary (handles both flat and nested layouts)
    $extractedBinary = Get-ChildItem -Path $extractDir -Recurse -Filter $Binary | Select-Object -First 1
    if (-not $extractedBinary) {
        Write-Err "Could not find '$Binary' in the downloaded archive."
        Write-Err "Archive contents:"
        Get-ChildItem -Path $extractDir -Recurse | ForEach-Object { Write-Err "  $($_.FullName)" }
        exit $EXIT_GENERAL
    }

    # -- Install --------------------------------------------------------------
    if (-not (Test-Path -Path $InstallDir)) {
        Write-Info "Creating install directory: $InstallDir"
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    $targetPath = Join-Path $InstallDir $Binary
    Copy-Item -Path $extractedBinary.FullName -Destination $targetPath -Force

    Write-Info "Installed sequoia -> $targetPath"

    # -- Add to PATH (always, unless -NoPath is passed) -----------------------
    if (-not $NoPath) {
        Write-Info "Ensuring $InstallDir is in user PATH..."

        # Read the persistent user PATH from registry
        $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")

        # Split on ';' and filter empty entries to check for duplicates
        $entries = if ($currentPath) {
            $currentPath -split ';' | Where-Object { $_ }
        } else {
            @()
        }

        if ($InstallDir -notin $entries) {
            # Build the new PATH string
            if (-not $currentPath) {
                $newPath = $InstallDir
            } else {
                $newPath = "$currentPath;$InstallDir"
            }

            # Re-validate InstallDir before writing to PATH (IS-SEC-003)
            # Prevents semicolon-based PATH splitting injection.
            Resolve-SafePath -Path $InstallDir -Context "InstallDir for PATH"

            # Write to registry (persists across terminal sessions and reboots)
            [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
            Write-Info "Added to user PATH (persistent across sessions)."

            # Also update current session so 'sequoia' works immediately
            $sessionEntries = $env:Path -split ';' | Where-Object { $_ }
            if ($InstallDir -notin $sessionEntries) {
                $env:Path = "$env:Path;$InstallDir"
                Write-Info "Also available in current terminal session."
            }
        } else {
            Write-Info "$InstallDir is already in PATH."
        }
    } else {
        Write-Warn "PATH not modified (-NoPath was specified)."
        Write-Host "  Run 'sequoia' from: $InstallDir"
    }

    # -- Run sequoia install -------------------------------------------------
    Write-Info "Running 'sequoia install --no-tui'..."
    try {
        $installResult = & $targetPath install --no-tui 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Warn "'sequoia install' completed with warnings (exit code: $LASTEXITCODE)."
        }
    } catch {
        Write-Warn "'sequoia install' completed with warnings. Check output above."
    }

    # -- Install CodeGraph (non-blocking) ------------------------------------
    $codegraphCmd = Get-Command codegraph -ErrorAction SilentlyContinue
    if ($codegraphCmd) {
        Write-Info "CodeGraph is already installed ($($codegraphCmd.Source))."
    } else {
        Write-Info "Installing CodeGraph for enhanced code intelligence..."
        try {
            irm https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.ps1 | iex
            Write-Info "CodeGraph installed. Auto-configuring agents..."
            & codegraph install --target=auto --location=global --yes 2>&1 | Out-Null
            if ($global:LASTEXITCODE -ne 0) {
                Write-Warn "CodeGraph agent configuration failed (non-blocking)."
                Write-Warn "Run manually: codegraph install --target=auto --location=global --yes"
            } else {
                Write-Info "CodeGraph integration ready."
            }
        } catch {
            Write-Warn "CodeGraph installation skipped (network issue or unsupported platform)."
            Write-Warn "Sequoia works fine without it — /sequoia-dev will fall back to file-based exploration."
            Write-Warn "Install manually: irm https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.ps1 | iex"
        }
    }

    # -- Done -----------------------------------------------------------------
    Write-Host ""
    Write-Host "==============================================" -ForegroundColor Green
    Write-Host "  Sequoia $ResolvedVersion installed successfully!" -ForegroundColor Green
    Write-Host "==============================================" -ForegroundColor Green
    Write-Host ""

    if ($NoPath) {
        Write-Warn "$InstallDir is not in your PATH."
        Write-Host "  Run 'sequoia' directly from: $targetPath"
        Write-Host "  Or add it manually: `$env:Path += `";$InstallDir`""
    } else {
        Write-Host "Run 'sequoia status' to verify your installation."
    }

    # Success — let the script end naturally, don't close the terminal.
    # Explicit return keeps the caller's PowerShell session alive.
} finally {
    # -- Cleanup --------------------------------------------------------------
    if (Test-Path -Path $TempDir) {
        Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# -- Keep the terminal open so the user can read the output ---------------
Write-Host ""
Read-Host "Press Enter to exit"
