#!/usr/bin/env bash
# =============================================================================
# Sequoia One-Line Installer (Unix: macOS & Linux)
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/Crisbr10/sequoia/main/scripts/install.sh | bash
#
# Or with custom options:
#   curl -sSL ... | REPO=myfork/sequoia VERSION=v0.2.0 bash
#
# Environment variables:
#   REPO         GitHub org/repo (default: Crisbr10/sequoia)
#   VERSION      Release version tag (default: latest, resolved via GitHub API)
#   INSTALL_DIR  Target directory for the binary (default: /usr/local/bin)
# =============================================================================

set -euo pipefail

# -- Configuration ------------------------------------------------------------
BINARY="sequoia"
REPO="${REPO:-Crisbr10/sequoia}"
VERSION_INPUT="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

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

CHECKSUMS_DOWNLOADED=""
if [ "$DOWNLOADER" = "curl" ]; then
    curl -fsSL --retry 3 --retry-delay 2 -o "${TMPDIR}/checksums.txt" "$CHECKSUM_URL" 2>/dev/null && CHECKSUMS_DOWNLOADED="1" || true
else
    wget -q --retry-connrefused --tries=3 -O "${TMPDIR}/checksums.txt" "$CHECKSUM_URL" 2>/dev/null && CHECKSUMS_DOWNLOADED="1" || true
fi

if [ -n "$CHECKSUMS_DOWNLOADED" ] && [ -f "${TMPDIR}/checksums.txt" ]; then
    COMPUTED_HASH=""
    if [ "$HASH_TOOL" = "sha256sum" ]; then
        COMPUTED_HASH="$(sha256sum "${TMPDIR}/${TARBALL}" | awk '{print $1}')"
    else
        COMPUTED_HASH="$(shasum -a 256 "${TMPDIR}/${TARBALL}" | awk '{print $1}')"
    fi

    EXPECTED_HASH="$(grep "${TARBALL}" "${TMPDIR}/checksums.txt" | awk '{print $1}' | head -1)" || EXPECTED_HASH=""

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
    log_warn "Could not download checksums.txt. Skipping checksum verification."
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
