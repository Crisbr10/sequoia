#!/usr/bin/env bash
# =============================================================================
# Tests: CodeGraph auto-install block in install.sh
#
# TDD: RED phase — install_codegraph() not yet in install.sh
# GREEN phase — after adding the function, these tests exercise it
#
# Usage: bash scripts/install_codegraph_test.sh
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TESTS_RUN=0
TESTS_FAILED=0
OUTPUT_DIR=""

# ---- Helpers: logging (match install.sh patterns, without color) ------------
log_info() { printf "[INFO]  %s\n" "$*"; }
log_warn() { printf "[WARN]  %s\n" "$*" >&2; }

# ---- Helpers: assertions ----------------------------------------------------
assert_output_contains() {
    local haystack="$1"
    local needle="$2"
    local msg="$3"
    if [[ "$haystack" != *"$needle"* ]]; then
        printf "  FAIL: %s\n" "$msg"
        printf "    expected: %s\n" "$needle"
        printf "    got:      %s\n" "$haystack"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    printf "  PASS: %s\n" "$msg"
    return 0
}

assert_output_not_contains() {
    local haystack="$1"
    local needle="$2"
    local msg="$3"
    if [[ "$haystack" == *"$needle"* ]]; then
        printf "  FAIL: %s\n" "$msg"
        printf "    unexpected match: %s\n" "$needle"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    printf "  PASS: %s\n" "$msg"
    return 0
}

assert_exit_code() {
    local expected="$1"
    local actual="$2"
    local msg="$3"
    if [ "$actual" -ne "$expected" ]; then
        printf "  FAIL: %s\n" "$msg"
        printf "    expected exit: %d, got: %d\n" "$expected" "$actual"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    printf "  PASS: %s\n" "$msg"
    return 0
}

# ---- Load install_codegraph() from install.sh --------------------------------
# Use sed to extract the function definition without executing install.sh.
# Returns non-zero if the function is not found (RED phase).
load_codegraph_function() {
    local func_body
    func_body="$(sed -n '/^install_codegraph()/,/^}/p' "${SCRIPT_DIR}/install.sh" 2>/dev/null || true)"
    if [ -z "$func_body" ]; then
        return 1
    fi
    eval "$func_body"
    return 0
}

# ---- Mocking infrastructure --------------------------------------------------
# Each test defines its own mock overrides. The mock state is isolated via
# subshell execution. Mock functions shadow real commands:
#   command()  — shadows the 'command' builtin
#   curl()     — shadows the real curl binary
#   codegraph() — shadows the real codegraph binary

# ---- Test: Already installed (T1.2) -----------------------------------------
test_already_installed() {
    TESTS_RUN=$((TESTS_RUN + 1))
    echo ""
    echo "--- Test: CodeGraph already installed → skip ---"

    # Capture stderr separately for warn assertions
    local stderr_file
    stderr_file="$(mktemp)"

    # Mock: command -v codegraph → success (returns 0)
    command() {
        if [ "$1" = "-v" ] && [ "$2" = "codegraph" ]; then
            return 0
        fi
        builtin command "$@"
    }
    export -f command

    # Mock: curl should NOT be called — flag if it is
    curl() { echo "UNEXPECTED CURL CALL" >&2; return 1; }
    export -f curl

    # Mock: codegraph --version for the log message
    codegraph() {
        if [ "$1" = "--version" ]; then
            echo "v1.0.0-test"
            return 0
        fi
        return 1
    }
    export -f codegraph

    local output
    output="$(install_codegraph 2>"$stderr_file")" || true
    local exit_code=$?
    local stderr_output
    stderr_output="$(cat "$stderr_file")"
    rm -f "$stderr_file"

    assert_exit_code 0 "$exit_code" "exit code is 0 (non-blocking)"
    assert_output_contains "$output" "already installed" "logs 'already installed'"
    assert_output_not_contains "$output" "Installing CodeGraph" "does NOT attempt download"
    assert_output_not_contains "$stderr_output" "skipped" "no warn output on stderr"
}

# ---- Test: Fresh install success (T1.2 — download + config both OK) --------
test_fresh_install_success() {
    TESTS_RUN=$((TESTS_RUN + 1))
    echo ""
    echo "--- Test: Fresh install — download + config succeed ---"

    # Mock: command -v codegraph → not found (returns 1)
    command() {
        if [ "$1" = "-v" ] && [ "$2" = "codegraph" ]; then
            return 1
        fi
        builtin command "$@"
    }
    export -f command

    # Mock: curl → success (returns 0)
    curl() {
        if [[ "$*" == *"raw.githubusercontent.com"* ]]; then
            return 0
        fi
        builtin curl "$@"
    }
    export -f curl

    # Mock: codegraph install → success (returns 0)
    codegraph() {
        if [ "$1" = "install" ]; then
            return 0
        fi
        return 1
    }
    export -f codegraph

    local output
    output="$(install_codegraph 2>/dev/null)" || true
    local exit_code=$?

    assert_exit_code 0 "$exit_code" "exit code is 0 (non-blocking)"
    assert_output_contains "$output" "Installing CodeGraph" "logs download attempt"
    assert_output_contains "$output" "Auto-configuring agents" "logs config step"
    assert_output_contains "$output" "CodeGraph integration ready" "logs success"
}

# ---- Test: Download failure (T1.3) ------------------------------------------
test_download_failure() {
    TESTS_RUN=$((TESTS_RUN + 1))
    echo ""
    echo "--- Test: Download failure → warning + continue ---"

    local stderr_file
    stderr_file="$(mktemp)"

    # Mock: command -v codegraph → not found
    command() {
        if [ "$1" = "-v" ] && [ "$2" = "codegraph" ]; then
            return 1
        fi
        builtin command "$@"
    }
    export -f command

    # Mock: curl → failure (returns 1)
    curl() {
        if [[ "$*" == *"raw.githubusercontent.com"* ]]; then
            return 1
        fi
        builtin curl "$@"
    }
    export -f curl

    local output
    output="$(install_codegraph 2>"$stderr_file")" || true
    local exit_code=$?
    local stderr_output
    stderr_output="$(cat "$stderr_file")"
    rm -f "$stderr_file"

    assert_exit_code 0 "$exit_code" "exit code is 0 (non-blocking on failure)"
    assert_output_contains "$stderr_output" "installation skipped" "warns about skipped installation"
    assert_output_contains "$stderr_output" "Sequoia works fine without it" "warns Sequoia still works"
    assert_output_contains "$stderr_output" "curl -fsSL" "includes manual install command"
    assert_output_not_contains "$output" "CodeGraph installed" "does NOT claim install succeeded"
}

# ---- Test: Config failure (T1.4) — codegraph install fails, flow continues --
test_config_failure() {
    TESTS_RUN=$((TESTS_RUN + 1))
    echo ""
    echo "--- Test: Download OK but codegraph install fails → warns and continues ---"

    local stderr_file
    stderr_file="$(mktemp)"

    # Mock: command -v codegraph → not found
    command() {
        if [ "$1" = "-v" ] && [ "$2" = "codegraph" ]; then
            return 1
        fi
        builtin command "$@"
    }
    export -f command

    # Mock: curl → success
    curl() {
        if [[ "$*" == *"raw.githubusercontent.com"* ]]; then
            return 0
        fi
        builtin curl "$@"
    }
    export -f curl

    # Mock: codegraph install → failure (returns 1)
    codegraph() {
        if [ "$1" = "install" ]; then
            echo "no agents detected" >&2
            return 1
        fi
        return 1
    }
    export -f codegraph

    local output
    output="$(install_codegraph 2>"$stderr_file")" || true
    local exit_code=$?
    local stderr_output
    stderr_output="$(cat "$stderr_file")"
    rm -f "$stderr_file"

    assert_exit_code 0 "$exit_code" "exit code is 0 (non-blocking on config failure)"
    assert_output_contains "$stderr_output" "configuration failed" "warns about configuration failure"
    assert_output_contains "$stderr_output" "Run manually" "includes manual fix command"
}

# ---- Test runner ------------------------------------------------------------
run_tests() {
    echo "======================================"
    echo " CodeGraph Auto-Install Test Suite"
    echo "======================================"

    if ! load_codegraph_function; then
        echo ""
        echo "=== RED: install_codegraph() not found in install.sh ==="
        echo "This is expected in the RED phase."
        echo "After adding the function to install.sh, re-run these tests."
        echo ""
        TESTS_FAILED=$((TESTS_FAILED + 1))
    else
        echo "install_codegraph() loaded from install.sh"
        echo ""

        test_already_installed
        test_fresh_install_success
        test_download_failure
        test_config_failure
    fi

    echo ""
    echo "======================================"
    echo " Results: ${TESTS_RUN} run, ${TESTS_FAILED} failed"
    echo "======================================"

    return "$TESTS_FAILED"
}

run_tests
