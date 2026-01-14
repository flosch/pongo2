#!/bin/bash
#
# Run all fuzz tests in parallel with reduced workers.

set -euo pipefail

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS] [duration] [parallel] [workers]

Run all fuzz tests in parallel with configurable parameters.

Arguments:
  duration    How long to run each fuzz test (default: 10s)
  parallel    Number of fuzz tests to run in parallel (default: 4)
  workers     Number of fuzz workers per test (default: 2)

Options:
  -h, --help      Show this help message and exit
  -u, --unlimited Run continuously, restarting after all tests complete

Examples:
  $(basename "$0")              # Run with defaults (10s, 4 parallel, 2 workers)
  $(basename "$0") 30s 2        # Run for 30s each, 2 tests at a time
  $(basename "$0") 1m 4 1       # Run for 1 minute, 4 parallel, 1 worker each
  $(basename "$0") -u 10s 2 1   # Run continuously with 10s per test
EOF
    exit 0
}

# Parse options
UNLIMITED=0
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            usage
            ;;
        -u|--unlimited)
            UNLIMITED=1
            shift
            ;;
        -*)
            echo "Error: Unknown option $1" >&2
            exit 1
            ;;
        *)
            break
            ;;
    esac
done

DURATION="${1:-10s}"
PARALLEL="${2:-4}"
FUZZ_WORKERS="${3:-2}"

# Check for required commands
if ! command -v go &>/dev/null; then
    echo "Error: 'go' command not found. Please install Go." >&2
    exit 1
fi

# Validate duration format (Go duration: e.g., 10s, 1m, 2h, 500ms)
if ! [[ "$DURATION" =~ ^[0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h)$ ]]; then
    echo "Error: Invalid duration format '$DURATION'" >&2
    echo "Expected Go duration format, e.g., 10s, 1m, 2h, 500ms" >&2
    exit 1
fi

# Validate parallel is a positive integer
if ! [[ "$PARALLEL" =~ ^[1-9][0-9]*$ ]]; then
    echo "Error: parallel must be a positive integer, got '$PARALLEL'" >&2
    exit 1
fi

# Validate workers is a positive integer
if ! [[ "$FUZZ_WORKERS" =~ ^[1-9][0-9]*$ ]]; then
    echo "Error: workers must be a positive integer, got '$FUZZ_WORKERS'" >&2
    exit 1
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo -e "${BLUE}=== Fuzz Test Runner ===${NC}"
echo -e "Duration per test: ${YELLOW}$DURATION${NC}"
echo -e "Parallel tests:    ${YELLOW}$PARALLEL${NC}"
echo -e "Fuzz workers:      ${YELLOW}$FUZZ_WORKERS${NC}"
if [[ $UNLIMITED -eq 1 ]]; then
    echo -e "Mode:              ${YELLOW}unlimited (Ctrl+C to stop)${NC}"
fi
echo ""

# Find all fuzz test functions
FUZZ_TESTS=$(grep -rh "^func Fuzz" --include="*_test.go" . | sed 's/func \(Fuzz[a-zA-Z0-9_]*\).*/\1/' | sort -u)
TOTAL=$(echo "$FUZZ_TESTS" | wc -l | tr -d ' ')

echo -e "Found ${GREEN}$TOTAL${NC} fuzz tests"
echo ""

# Create temporary directory for results and PID tracking
RESULTS_DIR=$(mktemp -d)
PIDS_FILE="$RESULTS_DIR/pids"
touch "$PIDS_FILE"

# Interrupt flag
INTERRUPTED=0

cleanup() {
    INTERRUPTED=1
    echo ""
    echo -e "${YELLOW}Interrupted, cleaning up...${NC}"
    trap - INT TERM EXIT  # Prevent recursive signals
    # Kill tracked PIDs and their children
    if [[ -f "$PIDS_FILE" ]]; then
        while read -r pid; do
            # Kill children first (go test spawns pongo2.test)
            pkill -P "$pid" 2>/dev/null || true
            # Then kill the subshell
            kill "$pid" 2>/dev/null || true
        done < "$PIDS_FILE"
    fi
    # Also kill any pongo2.test processes directly
    pkill -f "pongo2.test.*-test.fuzz" 2>/dev/null || true
    # Wait briefly
    sleep 0.3
    # Force kill any remaining
    if [[ -f "$PIDS_FILE" ]]; then
        while read -r pid; do
            pkill -9 -P "$pid" 2>/dev/null || true
            kill -9 "$pid" 2>/dev/null || true
        done < "$PIDS_FILE"
    fi
    pkill -9 -f "pongo2.test.*-test.fuzz" 2>/dev/null || true
    rm -rf "$RESULTS_DIR"
    exit 130
}

trap cleanup INT TERM
trap 'rm -rf "$RESULTS_DIR"' EXIT

# Function to run a single fuzz test (runs in subshell when backgrounded)
run_fuzz() {
    local test_name="$1"
    local result_file="$RESULTS_DIR/$test_name.result"
    local log_file="$RESULTS_DIR/$test_name.log"

    echo -e "${BLUE}Starting${NC} $test_name"

    if go test -fuzz="^${test_name}\$" -fuzztime="$DURATION" -parallel="$FUZZ_WORKERS" ./... > "$log_file" 2>&1; then
        echo "PASS" > "$result_file"
        echo -e "${GREEN}PASS${NC} $test_name"
    else
        echo "FAIL" > "$result_file"
        echo -e "${RED}FAIL${NC} $test_name"
        echo -e "${YELLOW}Log:${NC} $log_file"
    fi
}

# Count running jobs from tracked PIDs
count_running() {
    local count=0
    if [[ ${#ACTIVE_PIDS[@]} -gt 0 ]]; then
        for pid in "${ACTIVE_PIDS[@]}"; do
            if kill -0 "$pid" 2>/dev/null; then
                ((count++)) || true
            fi
        done
    fi
    echo "$count"
}

# Main test execution loop
ROUND=0
while true; do
    ((ROUND++)) || true  # Prevent exit when ROUND was 0

    # Run fuzz tests in parallel using background jobs
    if [[ $UNLIMITED -eq 1 ]]; then
        echo -e "${BLUE}=== Round $ROUND ===${NC}"
    fi
    echo -e "${BLUE}Running fuzz tests...${NC}"
    echo ""

    # Clear previous results
    find "$RESULTS_DIR" -name "*.result" -delete 2>/dev/null || true
    find "$RESULTS_DIR" -name "*.log" -delete 2>/dev/null || true
    : > "$PIDS_FILE"

    ACTIVE_PIDS=()
    for test_name in $FUZZ_TESTS; do
        # Check if interrupted
        [[ $INTERRUPTED -eq 1 ]] && break

        # Wait if we've reached max parallel jobs
        while [[ $(count_running) -ge $PARALLEL ]] && [[ $INTERRUPTED -eq 0 ]]; do
            sleep 0.2
        done

        # Check again after waiting
        [[ $INTERRUPTED -eq 1 ]] && break

        # Start fuzz test in background and track PID
        run_fuzz "$test_name" &
        pid=$!
        ACTIVE_PIDS+=("$pid")
        echo "$pid" >> "$PIDS_FILE"
    done

    # Wait for all remaining jobs (unless interrupted)
    [[ $INTERRUPTED -eq 0 ]] && wait

    # Check if interrupted during execution
    [[ $INTERRUPTED -eq 1 ]] && break

    echo ""
    echo -e "${BLUE}=== Results ===${NC}"

    # Count results
    PASSED=0
    FAILED=0
    FAILED_TESTS=""

    for test_name in $FUZZ_TESTS; do
        result_file="$RESULTS_DIR/$test_name.result"
        if [[ -f "$result_file" ]] && [[ "$(cat "$result_file")" == "PASS" ]]; then
            ((PASSED++)) || true
        else
            ((FAILED++)) || true
            FAILED_TESTS="$FAILED_TESTS $test_name"
        fi
    done

    echo -e "${GREEN}Passed:${NC} $PASSED"
    echo -e "${RED}Failed:${NC} $FAILED"

    if [[ $FAILED -gt 0 ]]; then
        echo ""
        echo -e "${RED}Failed tests:${NC}"
        for test in $FAILED_TESTS; do
            echo "  - $test"
            log_file="$RESULTS_DIR/$test.log"
            if [[ -f "$log_file" ]]; then
                echo -e "${YELLOW}    Last 10 lines of log:${NC}"
                tail -10 "$log_file" | sed 's/^/    /'
            fi
        done
        # In unlimited mode, continue despite failures
        if [[ $UNLIMITED -eq 0 ]]; then
            exit 1
        fi
    else
        echo ""
        echo -e "${GREEN}All fuzz tests passed!${NC}"
    fi

    # Exit loop if not in unlimited mode
    if [[ $UNLIMITED -eq 0 ]]; then
        break
    fi

    echo ""
    echo -e "${BLUE}Starting next round...${NC}"
    echo ""
done
