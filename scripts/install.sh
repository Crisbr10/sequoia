#!/usr/bin/env bash
# =============================================================================
# Sequoia One-Line Installer (Unix: macOS & Linux)
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/Crisbr10/sequoia/main/scripts/install.sh | bash
#
# Or with custom options:
#   curl -sSL ... | REPO=myfork/sequoia VERSION=v0.2.0 bash
#   curl -sSL ... | SKIP_CHECKSUMS=true bash   (air-gapped / no checksums.txt)
#
# Environment variables:
#   REPO           GitHub org/repo (default: Crisbr10/sequoia)
#   VERSION        Release version tag (default: latest, resolved via GitHub API)
#   INSTALL_DIR    Target directory for the binary (default: /usr/local/bin)
#   SKIP_CHECKSUMS If set to "true", bypass SHA-256 verification (opt-in, for
#                  air-gapped environments where checksums.txt is unreachable)
# =============================================================================

set -euo pipefail

# -- Configuration ------------------------------------------------------------
BINARY="sequoia"
REPO="${REPO:-Crisbr10/sequoia}"
VERSION_INPUT="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
SKIP_CHECKSUMS="${SKIP_CHECKSUMS:-false}"

# Support --skip-checksums flag when running script directly (not piped)
for arg in "$@"; do
    case "$arg" in
        --skip-checksums) SKIP_CHECKSUMS="true" ;;
        --help|-h)
            echo "Sequoia One-Line Installer"
            echo ""
            echo "Environment variables:"
            echo "  REPO           GitHub org/repo (default: Crisbr10/sequoia)"
            echo "  VERSION        Release version tag (default: latest)"
            echo "  INSTALL_DIR    Target directory (default: /usr/local/bin)"
            echo "  SKIP_CHECKSUMS Set to 'true' to bypass SHA-256 verification"
            echo "                 (opt-in for air-gapped environments)"
            echo ""
            echo "Flags (when running script directly):"
            echo "  --skip-checksums  Bypass SHA-256 verification"
            echo "  --help, -h        Show this help message"
            exit $EXIT_OK
            ;;
    esac
done

# Exit codes (matched to design contract)
EXIT_OK=0
EXIT_GENERAL=1
EXIT_CHECKSUM=2
EXIT_NETWORK=3

# -- Color helpers ------------------------------------------------------------
if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BOLD='\033[1m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    BOLD=''
    NC=''
fi

log_info()  { printf "${GREEN}[INFO]${NC}  %s\n" "$*"; }
log_warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$*" >&2; }
log_error() { printf "${RED}[ERROR]${NC} %s\n" "$*" >&2; }

# -- Path sanitization ---------------------------------------------------------
validate_path() {
    path="$1"
    context="${2:-install directory}"

    if [ -z "$path" ]; then
        log_error "$context must not be empty"
        exit $EXIT_GENERAL
    fi

    case "$path" in
        /*) ;;
        *)  log_error "$context must be absolute (start with /)"
            exit $EXIT_GENERAL ;;
    esac

    case "$path" in *\.\.*)
        log_error "$context contains '..' traversal"
        exit $EXIT_GENERAL ;;
    esac

    case "$path" in *\;*)
        log_error "$context contains forbidden: ';'"
        exit $EXIT_GENERAL ;;
    esac

    case "$path" in *\|*)
        log_error "$context contains forbidden: '|'"
        exit $EXIT_GENERAL ;;
    esac

    case "$path" in *\&*)
        log_error "$context contains forbidden: '&'"
        exit $EXIT_GENERAL ;;
    esac

    case "$path" in *\`*)
        log_error "$context contains forbidden: backtick"
        exit $EXIT_GENERAL ;;
    esac

    case "$path" in *\$\(*)
        log_error "$context contains \$() substitution"
        exit $EXIT_GENERAL ;;
    esac
}

# -- Temporary directory ------------------------------------------------------
TMPDIR=""
cleanup() {
    if [ -n "${TMPDIR:-}" ] && [ -d "$TMPDIR" ]; then
        rm -rf "$TMPDIR"
    fi
}
trap cleanup EXIT

TMPDIR="$(mktemp -d 2>/dev/null || mktemp -d -t sequoia-install)"

# -- OS / Arch detection ------------------------------------------------------
detect_os() {
    local os
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$os" in
        darwin)  echo "darwin" ;;
        linux)   echo "linux"  ;;
        *)
            log_error "Unsupported OS: $os"
            log_error "Supported platforms: Darwin (macOS), Linux"
            exit $EXIT_GENERAL
            ;;
    esac
}

detect_arch() {
    local arch
    arch="$(uname -m | tr '[:upper:]' '[:lower:]')"
    case "$arch" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *)
            log_error "Unsupported architecture: $arch"
            log_error "Supported architectures: x86_64/amd64, arm64/aarch64"
            exit $EXIT_GENERAL
            ;;
    esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"

# -- Tool detection -----------------------------------------------------------
find_downloader() {
    if command -v curl >/dev/null 2>&1; then
        echo "curl"
    elif command -v wget >/dev/null 2>&1; then
        echo "wget"
    else
        log_error "Neither curl nor wget is available. Please install one to continue."
        exit $EXIT_GENERAL
    fi
}

find_hash_tool() {
    if command -v sha256sum >/dev/null 2>&1; then
        echo "sha256sum"
    elif command -v shasum >/dev/null 2>&1; then
        echo "shasum"
    else
        log_error "Neither sha256sum nor shasum found. Please install a SHA-256 utility."
        exit $EXIT_GENERAL
    fi
}

DOWNLOADER="$(find_downloader)"
HASH_TOOL="$(find_hash_tool)"

# -- Version resolution (GitHub API for "latest") -----------------------------
resolve_version() {
    local version="$1"

    if [ "$version" != "latest" ]; then
        echo "$version"
        return 0
    fi

    log_info "Resolving latest version for ${REPO}..."
    local api_url="https://api.github.com/repos/${REPO}/releases/latest"
    local response

    if [ "$DOWNLOADER" = "curl" ]; then
        response="$(curl -fsSL "$api_url" 2>/dev/null)" || {
            log_error "Failed to fetch latest release info from GitHub."
            log_error "Check your internet connection or set VERSION explicitly (e.g. VERSION=v0.1.0)."
            exit $EXIT_NETWORK
        }
    else
        response="$(wget -qO- "$api_url" 2>/dev/null)" || {
            log_error "Failed to fetch latest release info from GitHub."
            log_error "Check your internet connection or set VERSION explicitly (e.g. VERSION=v0.1.0)."
            exit $EXIT_NETWORK
        }
    fi

    local tag
    tag="$(echo "$response" | tr ',' '\n' | grep '"tag_name"' | head -1 | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')"
    if [ -z "$tag" ]; then
        log_error "Could not parse version tag from GitHub API response."
        log_error "Set VERSION explicitly (e.g. VERSION=v0.1.0)."
        exit $EXIT_GENERAL
    fi

    echo "$tag"
}

VERSION="$(resolve_version "$VERSION_INPUT")"

# Strip "v" prefix for asset filenames (tags are v0.1.1, assets use 0.1.1)
VERSION_NUMBER="${VERSION#v}"

# Validate install path before any I/O
validate_path "$INSTALL_DIR" "INSTALL_DIR"

# -- Construct download URLs --------------------------------------------------
TARBALL="sequoia_${VERSION_NUMBER}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}"
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/sequoia_${VERSION_NUMBER}_checksums.txt"

# -- Idempotency check --------------------------------------------------------
check_existing() {
    local target="${INSTALL_DIR}/${BINARY}"

    if [ ! -x "$target" ]; then
        return 1
    fi

    local installed_version
    installed_version="$("$target" version 2>/dev/null)" || {
        log_warn "Existing binary at ${target} but 'version' command failed. Reinstalling..."
        return 1
    }

    if [ "$installed_version" = "$VERSION" ]; then
        printf "${BOLD}Sequoia %s is already installed at %s${NC}\n" "$VERSION" "$target"
        return 0
    fi

    log_info "Sequoia ${installed_version} found at ${target}, upgrading to ${VERSION}..."
    return 1
}

if check_existing; then
    exit $EXIT_OK
fi

# -- Download -----------------------------------------------------------------
log_info "Downloading ${BOLD}Sequoia ${VERSION}${NC} for ${OS}/${ARCH}..."
log_info "  URL: ${DOWNLOAD_URL}"

if [ "$DOWNLOADER" = "curl" ]; then
    if ! curl -fsSL --retry 3 --retry-delay 2 -o "${TMPDIR}/${TARBALL}" "$DOWNLOAD_URL"; then
        log_error "Download failed. Please check:"
        log_error "  - Internet connectivity"
        log_error "  - REPO=${REPO} (correct GitHub org/repo?)"
        log_error "  - VERSION=${VERSION} (tag exists?)"
        exit $EXIT_NETWORK
    fi
else
    if ! wget -q --retry-connrefused --tries=3 -O "${TMPDIR}/${TARBALL}" "$DOWNLOAD_URL"; then
        log_error "Download failed. Please check:"
        log_error "  - Internet connectivity"
        log_error "  - REPO=${REPO} (correct GitHub org/repo?)"
        log_error "  - VERSION=${VERSION} (tag exists?)"
        exit $EXIT_NETWORK
    fi
fi

# -- SHA-256 checksum verification --------------------------------------------
log_info "Verifying SHA-256 checksum..."

CHECKSUMS_FILE="${TMPDIR}/checksums.txt"

# Download checksums with retry. If download fails, ABORT unless user opted in
# to skip verification (--skip-checksums / SKIP_CHECKSUMS=true).
if [ "$DOWNLOADER" = "curl" ]; then
    curl -fsSL --retry 3 --retry-delay 2 -o "$CHECKSUMS_FILE" "$CHECKSUM_URL" 2>/dev/null
    CHECKSUMS_DOWNLOAD_EXIT=$?
else
    wget -q --retry-connrefused --tries=3 -O "$CHECKSUMS_FILE" "$CHECKSUM_URL" 2>/dev/null
    CHECKSUMS_DOWNLOAD_EXIT=$?
fi

if [ "$CHECKSUMS_DOWNLOAD_EXIT" -ne 0 ]; then
    if [ "$SKIP_CHECKSUMS" = "true" ]; then
        log_warn "Could not download checksums.txt. Skipping verification (--skip-checksums)."
    else
        log_error "Could not download checksums.txt from:"
        log_error "  ${CHECKSUM_URL}"
        log_error ""
        log_error "Checksum verification is mandatory. The binary cannot be verified."
        log_error "To bypass this check (air-gapped environments), set SKIP_CHECKSUMS=true:"
        log_error ""
        log_error "  curl -sSL ... | SKIP_CHECKSUMS=true bash"
        log_error ""
        log_error "  Or run: ./install.sh --skip-checksums"
        exit $EXIT_CHECKSUM
    fi
elif [ -f "$CHECKSUMS_FILE" ]; then
    COMPUTED_HASH=""
    if [ "$HASH_TOOL" = "sha256sum" ]; then
        COMPUTED_HASH="$(sha256sum "${TMPDIR}/${TARBALL}" | awk '{print $1}')"
    else
        COMPUTED_HASH="$(shasum -a 256 "${TMPDIR}/${TARBALL}" | awk '{print $1}')"
    fi

    EXPECTED_HASH="$(grep "${TARBALL}" "$CHECKSUMS_FILE" | awk '{print $1}' | head -1)" || EXPECTED_HASH=""

    if [ -z "$EXPECTED_HASH" ]; then
        log_warn "No checksum entry found for ${TARBALL} in checksums.txt. Skipping verification."
    elif [ "$COMPUTED_HASH" != "$EXPECTED_HASH" ]; then
        log_error "SHA-256 checksum mismatch!"
        log_error "  Expected: ${EXPECTED_HASH}"
        log_error "  Got:      ${COMPUTED_HASH}"
        log_error "The downloaded file may be corrupt or tampered with. Aborting."
        exit $EXIT_CHECKSUM
    else
        log_info "Checksum verified: ${COMPUTED_HASH}"
    fi
else
    # Download reported success but file is absent — edge case, warn and continue
    log_warn "Checksums file missing after download. Skipping verification."
fi

# -- Cosign signature verification --------------------------------------------
# Attempt cryptographic signature verification (REQ-IV-001).
# cosign absence → warn + offer install (SHA-256 remains the baseline)
# .sigstore.json bundle download failure → fall back to .sig/.cert (backward compat)
# Verification failure → error exit (signature does not match → tampered binary)
# Cosign is additive — SKIP_CHECKSUMS does NOT skip this step.
BUNDLE_URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}.sigstore.json"
COSIGN_SIG_URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}.sig"
CERT_URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}.cert"
BUNDLE_FILE="${TMPDIR}/${TARBALL}.sigstore.json"
COSIGN_SIG_FILE="${TMPDIR}/${TARBALL}.sig"
CERT_FILE="${TMPDIR}/${TARBALL}.cert"

# Detect cosign under any known binary name
COSIGN_CMD=""
for name in cosign cosign-windows-amd64 cosign-linux-amd64 cosign-darwin-amd64 cosign-linux-arm64 cosign-darwin-arm64; do
    if command -v "$name" >/dev/null 2>&1; then
        COSIGN_CMD="$name"
        break
    fi
done

if [ -n "$COSIGN_CMD" ]; then
    log_info "Cosign detected ($COSIGN_CMD) — verifying cryptographic signature..."

    # Try .sigstore.json bundle first (cosign v3 format)
    BUNDLE_DOWNLOAD_OK=true
    if [ "$DOWNLOADER" = "curl" ]; then
        curl -fsSL --retry 2 --retry-delay 2 -o "$BUNDLE_FILE" "$BUNDLE_URL" 2>/dev/null || BUNDLE_DOWNLOAD_OK=false
    else
        wget -q --retry-connrefused --tries=2 -O "$BUNDLE_FILE" "$BUNDLE_URL" 2>/dev/null || BUNDLE_DOWNLOAD_OK=false
    fi

    if [ "$BUNDLE_DOWNLOAD_OK" = "true" ] && [ -f "$BUNDLE_FILE" ]; then
        # New format: verify with --bundle
        if "$COSIGN_CMD" verify-blob \
            --bundle "$BUNDLE_FILE" \
            --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
            --certificate-identity "https://github.com/${REPO}/.github/workflows/release.yml@refs/tags/${VERSION}" \
            "${TMPDIR}/${TARBALL}" 2>/dev/null; then
            log_info "Cosign signature verified (sigstore bundle)"
        else
            log_error "Cosign signature verification FAILED."
            log_error "The downloaded binary may have been tampered with. Aborting."
            exit $EXIT_CHECKSUM
        fi
    else
        # Fallback: try legacy .sig + .cert (pre-migration releases)
        SIG_DOWNLOAD_OK=true
        if [ "$DOWNLOADER" = "curl" ]; then
            curl -fsSL --retry 2 --retry-delay 2 -o "$COSIGN_SIG_FILE" "$COSIGN_SIG_URL" 2>/dev/null || SIG_DOWNLOAD_OK=false
            curl -fsSL --retry 2 --retry-delay 2 -o "$CERT_FILE" "$CERT_URL" 2>/dev/null || SIG_DOWNLOAD_OK=false
        else
            wget -q --retry-connrefused --tries=2 -O "$COSIGN_SIG_FILE" "$COSIGN_SIG_URL" 2>/dev/null || SIG_DOWNLOAD_OK=false
            wget -q --retry-connrefused --tries=2 -O "$CERT_FILE" "$CERT_URL" 2>/dev/null || SIG_DOWNLOAD_OK=false
        fi

        if [ "$SIG_DOWNLOAD_OK" != "true" ]; then
            log_warn "Failed to download signature files. Skipping cosign verification."
            log_warn "SHA-256 checksum remains the integrity baseline."
        else
            if "$COSIGN_CMD" verify-blob \
                --signature "$COSIGN_SIG_FILE" \
                --certificate "$CERT_FILE" \
                --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
                --certificate-identity "https://github.com/${REPO}/.github/workflows/release.yml@refs/tags/${VERSION}" \
                "${TMPDIR}/${TARBALL}" 2>/dev/null; then
                log_info "Cosign signature verified (legacy format)"
            else
                log_error "Cosign signature verification FAILED."
                log_error "The downloaded binary may have been tampered with. Aborting."
                exit $EXIT_CHECKSUM
            fi
        fi
    fi
else
    log_warn "Cosign is not installed. Skipping cryptographic signature verification."

    # Offer automatic installation when running interactively
    if [ -t 0 ]; then
        printf "${YELLOW}Install cosign now for stronger integrity guarantees?${NC} [y/N] " >&2
        read -r INSTALL_CHOICE
        if [ "$INSTALL_CHOICE" = "y" ] || [ "$INSTALL_CHOICE" = "Y" ]; then
            # Detect package manager
            COSIGN_INSTALLED=false
            if command -v brew >/dev/null 2>&1; then
                log_info "Installing cosign via Homebrew..."
                brew install cosign >/dev/null 2>&1 && COSIGN_INSTALLED=true
            elif command -v apt-get >/dev/null 2>&1; then
                log_info "Installing cosign via apt..."
                sudo apt-get update -qq && sudo apt-get install -y cosign >/dev/null 2>&1 && COSIGN_INSTALLED=true
            elif command -v dnf >/dev/null 2>&1; then
                log_info "Installing cosign via dnf..."
                sudo dnf install -y cosign >/dev/null 2>&1 && COSIGN_INSTALLED=true
            elif command -v pacman >/dev/null 2>&1; then
                log_info "Installing cosign via pacman..."
                sudo pacman -S --noconfirm cosign >/dev/null 2>&1 && COSIGN_INSTALLED=true
            elif command -v apk >/dev/null 2>&1; then
                log_info "Installing cosign via apk..."
                sudo apk add cosign >/dev/null 2>&1 && COSIGN_INSTALLED=true
            fi

            if [ "$COSIGN_INSTALLED" = "true" ]; then
                # Re-detect
                COSIGN_CMD=""
                for name in cosign cosign-windows-amd64 cosign-linux-amd64 cosign-darwin-amd64 cosign-linux-arm64 cosign-darwin-arm64; do
                    if command -v "$name" >/dev/null 2>&1; then
                        COSIGN_CMD="$name"
                        break
                    fi
                done
                if [ -n "$COSIGN_CMD" ]; then
                    log_info "Cosign installed successfully. Verifying signature..."
                    # Try .sigstore.json bundle first (cosign v3 format)
                    BUNDLE_DOWNLOAD_OK=true
                    if [ "$DOWNLOADER" = "curl" ]; then
                        curl -fsSL --retry 2 --retry-delay 2 -o "$BUNDLE_FILE" "$BUNDLE_URL" 2>/dev/null || BUNDLE_DOWNLOAD_OK=false
                    else
                        wget -q --retry-connrefused --tries=2 -O "$BUNDLE_FILE" "$BUNDLE_URL" 2>/dev/null || BUNDLE_DOWNLOAD_OK=false
                    fi

                    if [ "$BUNDLE_DOWNLOAD_OK" = "true" ] && [ -f "$BUNDLE_FILE" ]; then
                        if "$COSIGN_CMD" verify-blob \
                            --bundle "$BUNDLE_FILE" \
                            --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
                            --certificate-identity "https://github.com/${REPO}/.github/workflows/release.yml@refs/tags/${VERSION}" \
                            "${TMPDIR}/${TARBALL}" 2>/dev/null; then
                            log_info "Cosign signature verified (sigstore bundle)"
                        else
                            log_error "Cosign signature verification FAILED."
                            log_error "The downloaded binary may have been tampered with. Aborting."
                            exit $EXIT_CHECKSUM
                        fi
                    else
                        # Fallback: legacy .sig + .cert
                        SIG_DOWNLOAD_OK=true
                        if [ "$DOWNLOADER" = "curl" ]; then
                            curl -fsSL --retry 2 --retry-delay 2 -o "$COSIGN_SIG_FILE" "$COSIGN_SIG_URL" 2>/dev/null || SIG_DOWNLOAD_OK=false
                            curl -fsSL --retry 2 --retry-delay 2 -o "$CERT_FILE" "$CERT_URL" 2>/dev/null || SIG_DOWNLOAD_OK=false
                        else
                            wget -q --retry-connrefused --tries=2 -O "$COSIGN_SIG_FILE" "$COSIGN_SIG_URL" 2>/dev/null || SIG_DOWNLOAD_OK=false
                            wget -q --retry-connrefused --tries=2 -O "$CERT_FILE" "$CERT_URL" 2>/dev/null || SIG_DOWNLOAD_OK=false
                        fi
                        if [ "$SIG_DOWNLOAD_OK" != "true" ]; then
                            log_warn "Failed to download signature files. Skipping cosign verification."
                            log_warn "SHA-256 checksum remains the integrity baseline."
                        else
                            if "$COSIGN_CMD" verify-blob \
                                --signature "$COSIGN_SIG_FILE" \
                                --certificate "$CERT_FILE" \
                                --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
                                --certificate-identity "https://github.com/${REPO}/.github/workflows/release.yml@refs/tags/${VERSION}" \
                                "${TMPDIR}/${TARBALL}" 2>/dev/null; then
                                log_info "Cosign signature verified (legacy format)"
                            else
                                log_error "Cosign signature verification FAILED."
                                log_error "The downloaded binary may have been tampered with. Aborting."
                                exit $EXIT_CHECKSUM
                            fi
                        fi
                    fi
                fi
            else
                log_warn "Automatic cosign installation failed."
            fi
        fi
    fi

    if [ -z "$COSIGN_CMD" ]; then
        log_warn "Install cosign for stronger integrity guarantees:"
        if [ "$OS" = "darwin" ]; then
            log_warn "  brew install cosign"
        else
            log_warn "  See https://docs.sigstore.dev/cosign/installation/"
        fi
        log_warn "SHA-256 checksum remains the integrity baseline."
    fi
fi

# -- Extract ------------------------------------------------------------------
log_info "Extracting ${TARBALL}..."
tar -xzf "${TMPDIR}/${TARBALL}" -C "$TMPDIR"

EXTRACTED_BINARY="$(find "$TMPDIR" -type f -name "$BINARY" -not -path "${TMPDIR}/${TARBALL}" | head -1)"
if [ -z "$EXTRACTED_BINARY" ]; then
    log_error "Could not find '${BINARY}' binary in the downloaded archive."
    log_error "Archive contents:"
    find "$TMPDIR" -not -name "${TARBALL}" -not -name "checksums.txt" -ls >&2 2>/dev/null || true
    exit $EXIT_GENERAL
fi

# -- Install ------------------------------------------------------------------
if [ ! -d "$INSTALL_DIR" ]; then
    log_info "Creating install directory: ${INSTALL_DIR}"
    mkdir -p "$INSTALL_DIR" || {
        log_error "Cannot create ${INSTALL_DIR}. Try running with sudo or set INSTALL_DIR."
        exit $EXIT_GENERAL
    }
fi

cp "$EXTRACTED_BINARY" "${INSTALL_DIR}/${BINARY}" || {
    log_error "Failed to copy binary to ${INSTALL_DIR}. Permission denied? Try sudo."
    exit $EXIT_GENERAL
}
chmod +x "${INSTALL_DIR}/${BINARY}"

log_info "Installed ${BOLD}${BINARY}${NC} -> ${INSTALL_DIR}/${BINARY}"

# -- Run sequoia install ------------------------------------------------------
log_info "Running '${BINARY} install --no-tui'..."
if ! "${INSTALL_DIR}/${BINARY}" install --no-tui; then
    log_warn "'sequoia install' completed with warnings. Check output above."
fi

# -- Done ---------------------------------------------------------------------
echo ""
printf "${GREEN}%s${NC}\n" "=============================================="
printf "${GREEN}%s${NC}\n" "  Sequoia ${VERSION} installed successfully!"
printf "${GREEN}%s${NC}\n" "=============================================="
echo ""

if ! command -v "$BINARY" >/dev/null 2>&1; then
    log_warn "${INSTALL_DIR} is not in your PATH."
    echo "  Add this to your shell profile to use 'sequoia' globally:"
    echo ""
    echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
    echo ""
else
    printf "Run ${BOLD}sequoia status${NC} to verify your installation.\n"
fi

exit $EXIT_OK
