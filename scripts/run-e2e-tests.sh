#!/bin/bash
# Comprehensive E2E Test Runner
# This script runs all e2e tests and collects results

set -e

echo "🚀 Running Comprehensive vendetta E2E Tests"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0
ISSUES_FOUND=()

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

record_issue() {
    local issue="$1"
    local severity="${2:-medium}"
    ISSUES_FOUND+=("$severity: $issue")
    log_error "Issue found: $issue"
}

# Prerequisites check
check_prerequisites() {
    log_info "Checking prerequisites..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker is required but not installed"
        exit 1
    fi

    if ! command -v git &> /dev/null; then
        log_error "Git is required but not installed"
        exit 1
    fi

    if ! command -v go &> /dev/null; then
        log_error "Go is required but not installed"
        exit 1
    fi

    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi

    log_success "Prerequisites check passed"
}

# Build vendetta binary
build_vendetta() {
    log_info "Building vendetta binary..."
    if ! go build -o vendetta ./cmd/vendetta; then
        log_error "Failed to build vendetta binary"
        exit 1
    fi
    log_success "vendetta binary built successfully"
}

# Run a single test
run_test() {
    local test_name="$1"
    local test_timeout="${2:-10m}"

    log_info "Running test: $test_name"
    TESTS_RUN=$((TESTS_RUN + 1))

    if go test -v "./e2e/" -run "$test_name" -timeout "$test_timeout" 2>&1; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        log_success "Test $test_name passed"
        return 0
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        log_error "Test $test_name failed"
        record_issue "Test $test_name failed" "high"
        return 1
    fi
}

# Run all e2e tests
run_e2e_tests() {
    log_info "Starting E2E test suite..."

    # Basic functionality tests
    run_test "TestWorkspaceLifecycle" "15m" || true
    run_test "TestWorkspaceList" "10m" || true

    # Plugin and provider tests
    run_test "TestPluginSystem" "10m" || true
    run_test "TestDockerProvider" "10m" || true

    # LXC provider tests (only if LXC_TEST is set)
    if [ "$LXC_TEST" = "1" ]; then
        run_test "TestLXCProvider" "15m" || true
    else
        log_info "Skipping LXC tests - set LXC_TEST=1 to run"
    fi

    # Error handling and performance tests
    run_test "TestErrorHandling" "10m" || true
    run_test "TestPerformanceBenchmarks" "5m" || true

    log_info "E2E test suite completed"
}

# Analyze test results
analyze_results() {
    log_info "Analyzing test results..."

    echo
    echo "=== Test Results Summary ==="
    echo "Tests Run: $TESTS_RUN"
    echo "Tests Passed: $TESTS_PASSED"
    echo "Tests Failed: $TESTS_FAILED"
    echo

    if [ ${#ISSUES_FOUND[@]} -gt 0 ]; then
        echo "=== Issues Found ==="
        for issue in "${ISSUES_FOUND[@]}"; do
            echo "• $issue"
        done
        echo

        # Categorize issues
        local critical_issues=()
        local high_issues=()
        local medium_issues=()
        local low_issues=()

        for issue in "${ISSUES_FOUND[@]}"; do
            if [[ $issue == critical:* ]]; then
                critical_issues+=("${issue#critical: }")
            elif [[ $issue == high:* ]]; then
                high_issues+=("${issue#high: }")
            elif [[ $issue == medium:* ]]; then
                medium_issues+=("${issue#medium: }")
            elif [[ $issue == low:* ]]; then
                low_issues+=("${issue#low: }")
            fi
        done

        echo "=== Issues by Severity ==="
        if [ ${#critical_issues[@]} -gt 0 ]; then
            echo "Critical Issues (${#critical_issues[@]}):"
            for issue in "${critical_issues[@]}"; do
                echo "  • $issue"
            done
        fi

        if [ ${#high_issues[@]} -gt 0 ]; then
            echo "High Priority Issues (${#high_issues[@]}):"
            for issue in "${high_issues[@]}"; do
                echo "  • $issue"
            done
        fi

        if [ ${#medium_issues[@]} -gt 0 ]; then
            echo "Medium Priority Issues (${#medium_issues[@]}):"
            for issue in "${medium_issues[@]}"; do
                echo "  • $issue"
            done
        fi

        if [ ${#low_issues[@]} -gt 0 ]; then
            echo "Low Priority Issues (${#low_issues[@]}):"
            for issue in "${low_issues[@]}"; do
                echo "  • $issue"
            done
        fi
    else
        log_success "No issues found!"
    fi
}

# Main execution
main() {
    echo "========================================"
    echo "🧪 vendetta E2E Test Suite"
    echo "========================================"

    check_prerequisites
    build_vendetta
    run_e2e_tests
    analyze_results

    echo
    echo "========================================"

    if [ $TESTS_FAILED -eq 0 ]; then
        log_success "All tests passed! ✅"
        exit 0
    else
        log_error "Some tests failed. See issues above for details. ❌"
        exit 1
    fi
}

# Run main function
main "$@"
