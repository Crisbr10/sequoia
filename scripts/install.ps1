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
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

try {
    # -- Download -------------------------------------------------------------
    Write-Info "Downloading Sequoia $ResolvedVersion for windows/$Arch..."
    Write-Info "  URL: $DownloadUrl"

    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile (Join-Path $TempDir $Tarball) -UseBasicParsing -ErrorAction Stop
    } catch {
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
            Invoke-WebRequest -Uri $ChecksumUrl -OutFile $checksumsPath -UseBasicParsing -ErrorAction Stop
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
