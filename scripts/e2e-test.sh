#!/bin/bash

set -e

# Colors using tput (safer for different terminals)
if command -v tput >/dev/null 2>&1 && tput setaf 1 >/dev/null 2>&1; then
    RED=$(tput setaf 1)
    GREEN=$(tput setaf 2)
    YELLOW=$(tput setaf 3)
    BLUE=$(tput setaf 4)
    BOLD=$(tput bold)
    NC=$(tput sgr0)  # No Color / Reset
else
    # Fallback to no colors if tput is not available or doesn't support colors
    RED=""
    GREEN=""
    YELLOW=""
    BLUE=""
    BOLD=""
    NC=""
fi

# Test configuration
MLFLOW_TRACKING_URI="http://localhost:5001"
TEST_EXPERIMENT_NAME="e2e-test-$(date +%s)"
BINARY_PATH="./mlflow-cli"
DEBUG_MODE=false
SKIP_DOCKER_SETUP=false

# Test counters
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --debug)
            DEBUG_MODE=true
            shift
            ;;
        --skip-docker)
            SKIP_DOCKER_SETUP=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--debug] [--skip-docker]"
            echo "  --debug       Enable debug output"
            echo "  --skip-docker Skip Docker Compose setup (use existing MLflow server)"
            exit 1
            ;;
    esac
done

# Logging functions
log_info() {
    echo "${BLUE}[INFO ]${NC} $1"
}

log_success() {
    echo "${GREEN}[ OK  ]${NC} $1"
}

log_error() {
    echo "${RED}[ERROR]${NC} $1"
}

log_debug() {
    if [ "$DEBUG_MODE" = true ]; then
        echo "${YELLOW}[DEBUG]${NC} $1"
    fi
}

# Test execution with simplified output
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_failure="${3:-false}"

    printf "%-50s " "$test_name"

    if [ "$expected_failure" = true ]; then
        # For error handling tests, we expect the command to fail
        if eval "$test_command" >/dev/null 2>&1; then
            echo "${RED}[FAIL]${NC}"
            echo "  Expected: Command should fail"
            echo "  Actual: Command succeeded"
            ((TESTS_FAILED++))
        else
            echo "${GREEN}[ OK ]${NC}"
            ((TESTS_PASSED++))
        fi
    else
        # For normal tests, we expect the command to succeed
        local output
        local exit_code
        output=$(eval "$test_command" 2>&1)
        exit_code=$?

        if [ $exit_code -eq 0 ]; then
            echo "${GREEN}[ OK ]${NC}"
            ((TESTS_PASSED++))
        else
            echo "${RED}[FAIL]${NC}"
            echo "  Command: $test_command"
            echo "  Exit code: $exit_code"
            echo "  Output: $output"
            ((TESTS_FAILED++))
        fi
    fi
    ((TESTS_TOTAL++))
}

# Verification functions
verify_parameter() {
    local run_details="$1"
    local param_key="$2"
    local expected_value="$3"

    if echo "$run_details" | grep -q "\"key\": \"$param_key\"" && echo "$run_details" | grep -q "\"value\": \"$expected_value\""; then
        return 0
    else
        return 1
    fi
}

verify_metric() {
    local run_id="$1"
    local metric_key="$2"
    local expected_value="$3"

    local metric_history
    metric_history=$(curl -s "$MLFLOW_TRACKING_URI/api/2.0/mlflow/metrics/get-history?run_id=$run_id&metric_key=$metric_key")

    if echo "$metric_history" | grep -q "\"value\": $expected_value"; then
        return 0
    else
        return 1
    fi
}

count_metric_points() {
    local run_id="$1"
    local metric_key="$2"

    local metric_history
    metric_history=$(curl -s "$MLFLOW_TRACKING_URI/api/2.0/mlflow/metrics/get-history?run_id=$run_id&metric_key=$metric_key")

    echo "$metric_history" | grep -o '"value":' | wc -l | tr -d ' '
}

wait_for_mlflow() {
    printf "%-50s " "Waiting for MLflow server"
    local max_attempts=15
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$MLFLOW_TRACKING_URI/api/2.0/mlflow/experiments/search" \
            -H "Content-Type: application/json" \
            -d '{"max_results": 1}' > /dev/null 2>&1; then
            echo "${GREEN}[ OK ]${NC}"
            return 0
        fi

        sleep 2
        ((attempt++))
    done

    echo "${RED}[FAIL]${NC}"
    echo "  MLflow server failed to start within timeout"
    return 1
}

cleanup() {
    if [ "$SKIP_DOCKER_SETUP" = false ]; then
        docker-compose down -v > /dev/null 2>&1 || true
        rm -rf mlflow-data > /dev/null 2>&1 || true
    fi
}

# Set up cleanup trap
trap cleanup EXIT

main() {
    echo "MLflow CLI E2E Tests"
    echo "===================="

    # Check if binary exists
    if [ ! -f "$BINARY_PATH" ]; then
        log_error "MLflow CLI binary not found at $BINARY_PATH"
        log_info "Please run 'make build' first"
        exit 1
    fi

    # Start MLflow server (if not skipping Docker setup)
    if [ "$SKIP_DOCKER_SETUP" = false ]; then
        printf "%-50s " "Starting MLflow server"
        if docker-compose up -d > /dev/null 2>&1; then
            echo "${GREEN}[ OK ]${NC}"
        else
            echo "${RED}[FAIL]${NC}"
            exit 1
        fi
    fi

    # Wait for MLflow server to be ready
    if ! wait_for_mlflow; then
        exit 1
    fi

    # Set environment variables
    export MLFLOW_TRACKING_URI="$MLFLOW_TRACKING_URI"

    # Create experiment
    printf "%-50s " "Creating test experiment"
    EXPERIMENT_RESPONSE=$(curl -s -X POST "$MLFLOW_TRACKING_URI/api/2.0/mlflow/experiments/create" \
        -H "Content-Type: application/json" \
        -d "{\"name\": \"$TEST_EXPERIMENT_NAME\"}")

    EXPERIMENT_ID=$(echo "$EXPERIMENT_RESPONSE" | grep -o '"experiment_id": "[^"]*"' | cut -d'"' -f4)

    if [ -n "$EXPERIMENT_ID" ]; then
        echo "${GREEN}[ OK ]${NC}"
        export MLFLOW_EXPERIMENT_ID="$EXPERIMENT_ID"
    else
        echo "${RED}[FAIL]${NC}"
        echo "  Response: $EXPERIMENT_RESPONSE"
        echo "  Parsed ID: '$EXPERIMENT_ID'"
        exit 1
    fi

    echo ""
    echo "Running Tests:"
    echo "--------------"

    # Test 1: Start run
    printf "%-50s " "Start run"
    RUN_ID=$($BINARY_PATH run start --run-name "e2e-test-run-$(date +%s)" 2>/dev/null)
    if [ $? -eq 0 ] && [ -n "$RUN_ID" ]; then
        echo "${GREEN}[ OK ]${NC}"
        ((TESTS_PASSED++))
    else
        echo "${RED}[FAIL]${NC}"
        echo "  Failed to start run"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    # Test 2-8: Basic operations
    run_test "Log single parameters" \
        "$BINARY_PATH log params --run-id $RUN_ID --param test_param=test_value --param batch_size=32"

    run_test "Log parameters from JSON file" \
        "$BINARY_PATH log params --run-id $RUN_ID --from-file test/fixtures/test_params.json"

    run_test "Log parameters from YAML file" \
        "$BINARY_PATH log params --run-id $RUN_ID --from-file test/fixtures/test_params.yaml"

    run_test "Log single metric" \
        "$BINARY_PATH log metric --run-id $RUN_ID --name accuracy --value 0.95 --step 1"

    run_test "Log metrics from JSON file" \
        "$BINARY_PATH log metrics --run-id $RUN_ID --from-file test/fixtures/test_metrics.json"

    run_test "Log metrics from YAML file" \
        "$BINARY_PATH log metrics --run-id $RUN_ID --from-file test/fixtures/test_metrics.yaml"

    run_test "Log metrics with time processing" \
        "$BINARY_PATH log metrics --run-id $RUN_ID --from-file test/fixtures/test_metrics.json --time-resolution 1m --time-alignment floor --step-mode timestamp"

    run_test "End run with FINISHED status" \
        "$BINARY_PATH run end --run-id $RUN_ID --status FINISHED"

    # Test 9: Start and end run with FAILED status
    printf "%-50s " "Start and end run with FAILED status"
    FAILED_RUN_ID=$($BINARY_PATH run start --run-name "e2e-test-failed-run-$(date +%s)" 2>/dev/null)
    if [ $? -eq 0 ] && [ -n "$FAILED_RUN_ID" ]; then
        if $BINARY_PATH run end --run-id $FAILED_RUN_ID --status FAILED >/dev/null 2>&1; then
            echo "${GREEN}[ OK ]${NC}"
            ((TESTS_PASSED++))
        else
            echo "${RED}[FAIL]${NC}"
            echo "  Failed to end run with FAILED status"
            ((TESTS_FAILED++))
        fi
    else
        echo "${RED}[FAIL]${NC}"
        echo "  Failed to start run for FAILED status test"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    # Test 10-11: Error handling
    run_test "Error handling - Invalid run ID" \
        "$BINARY_PATH log params --run-id invalid-run-id --param test=value" true

    run_test "Error handling - Missing experiment ID" \
        "unset MLFLOW_EXPERIMENT_ID && $BINARY_PATH run start --run-name test" true

    # Test 12-13: API verification
    printf "%-50s " "Verify experiment exists via API"
    EXPERIMENT_CHECK=$(curl -s "$MLFLOW_TRACKING_URI/api/2.0/mlflow/experiments/get?experiment_id=$EXPERIMENT_ID")
    if echo "$EXPERIMENT_CHECK" | grep -q "\"experiment_id\": \"$EXPERIMENT_ID\""; then
        echo "${GREEN}[ OK ]${NC}"
        ((TESTS_PASSED++))
    else
        echo "${RED}[FAIL]${NC}"
        echo "  Experiment not found via API"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    printf "%-50s " "Verify runs exist via API"
    RUNS_CHECK=$(curl -s "$MLFLOW_TRACKING_URI/api/2.0/mlflow/runs/search" \
        -H "Content-Type: application/json" \
        -d "{\"experiment_ids\": [\"$EXPERIMENT_ID\"], \"max_results\": 100}")
    if echo "$RUNS_CHECK" | grep -q "\"run_id\": \"$RUN_ID\""; then
        echo "${GREEN}[ OK ]${NC}"
        ((TESTS_PASSED++))
    else
        echo "${RED}[FAIL]${NC}"
        echo "  Run not found via API"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    # Test 14-16: Data verification
    RUN_DETAILS=$(curl -s "$MLFLOW_TRACKING_URI/api/2.0/mlflow/runs/get?run_id=$RUN_ID")

    printf "%-50s " "Verify single parameter"
    if verify_parameter "$RUN_DETAILS" "test_param" "test_value"; then
        echo "${GREEN}[ OK ]${NC}"
        ((TESTS_PASSED++))
    else
        echo "${RED}[FAIL]${NC}"
        echo "  Expected: test_param=test_value"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    printf "%-50s " "Verify JSON parameters"
    if verify_parameter "$RUN_DETAILS" "json_batch_size" "100"; then
        echo "${GREEN}[ OK ]${NC}"
        ((TESTS_PASSED++))
    else
        echo "${RED}[FAIL]${NC}"
        echo "  Expected: json_batch_size=100 from JSON file"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    printf "%-50s " "Verify YAML parameters"
    if verify_parameter "$RUN_DETAILS" "yaml_learning_rate" "0.01"; then
        echo "${GREEN}[ OK ]${NC}"
        ((TESTS_PASSED++))
    else
        echo "${RED}[FAIL]${NC}"
        echo "  Expected: yaml_learning_rate=0.01 from YAML file"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    printf "%-50s " "Verify single metric"
    if verify_metric "$RUN_ID" "accuracy" "0.95"; then
        echo "${GREEN}[ OK ]${NC}"
        ((TESTS_PASSED++))
    else
        echo "${RED}[FAIL]${NC}"
        echo "  Expected: accuracy=0.95"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    printf "%-50s " "Verify JSON metrics"
    ERROR_COUNT_HISTORY=$(curl -s "$MLFLOW_TRACKING_URI/api/2.0/mlflow/metrics/get-history?run_id=$RUN_ID&metric_key=error_count")
    if echo "$ERROR_COUNT_HISTORY" | grep -q "\"value\""; then
        echo "${GREEN}[ OK ]${NC}"
        ((TESTS_PASSED++))
    else
        echo "${RED}[FAIL]${NC}"
        echo "  Expected: error_count metric from JSON file"
        if [ "$DEBUG_MODE" = true ]; then
            echo "  Response: $ERROR_COUNT_HISTORY"
        fi
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    printf "%-50s " "Verify YAML metrics"
    ERROR_COUNT_YAML=$(curl -s "$MLFLOW_TRACKING_URI/api/2.0/mlflow/metrics/get-history?run_id=$RUN_ID&metric_key=error_count")
    if echo "$ERROR_COUNT_YAML" | grep -q "\"value\""; then
        echo "${GREEN}[ OK ]${NC}"
        ((TESTS_PASSED++))
    else
        echo "${RED}[FAIL]${NC}"
        echo "  Expected: error_count metric from YAML file"
        if [ "$DEBUG_MODE" = true ]; then
            echo "  Response: $ERROR_COUNT_YAML"
        fi
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    # Test 17-18: Time series and status verification
    printf "%-50s " "Verify time series processing (JSON)"
    JSON_METRIC_COUNT=$(count_metric_points "$RUN_ID" "error_count")
    if [ "$JSON_METRIC_COUNT" -ge 4 ]; then
        echo "${GREEN}[ OK ]${NC} ($JSON_METRIC_COUNT points)"
        ((TESTS_PASSED++))
    else
        echo "${RED}[FAIL]${NC}"
        echo "  Expected: ≥4 data points, got: $JSON_METRIC_COUNT"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    printf "%-50s " "Verify time series processing (YAML)"
    YAML_METRIC_COUNT=$(count_metric_points "$RUN_ID" "error_count")
    if [ "$YAML_METRIC_COUNT" -ge 3 ]; then
        echo "${GREEN}[ OK ]${NC} ($YAML_METRIC_COUNT points)"
        ((TESTS_PASSED++))
    else
        echo "${RED}[FAIL]${NC}"
        echo "  Expected: ≥3 data points, got: $YAML_METRIC_COUNT"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    printf "%-50s " "Verify run status"
    if echo "$RUN_DETAILS" | grep -q "\"status\": \"FINISHED\""; then
        echo "${GREEN}[ OK ]${NC}"
        ((TESTS_PASSED++))
    else
        echo "${RED}[FAIL]${NC}"
        echo "  Expected: status=FINISHED"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    printf "%-50s " "Verify parameter count"
    PARAM_COUNT=$(echo "$RUN_DETAILS" | grep -o "\"key\":" | wc -l | tr -d ' ')
    if [ "$PARAM_COUNT" -ge 12 ]; then
        echo "${GREEN}[ OK ]${NC} ($PARAM_COUNT parameters)"
        ((TESTS_PASSED++))
    else
        echo "${RED}[FAIL]${NC}"
        echo "  Expected: ≥12 parameters, got: $PARAM_COUNT"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    # Print test summary
    echo ""
    echo "Test Summary:"
    echo "============="
    echo "Total tests: $TESTS_TOTAL"
    echo "Passed: $TESTS_PASSED"
    echo "Failed: $TESTS_FAILED"

    if [ $TESTS_FAILED -eq 0 ]; then
        echo "${GREEN}All tests passed! 🎉${NC}"
        exit 0
    else
        echo "${RED}Some tests failed! ❌${NC}"
        exit 1
    fi
}

# Check dependencies
check_dependencies() {
    local missing_deps=()

    for dep in curl docker docker-compose; do
        if ! command -v "$dep" >/dev/null 2>&1; then
            missing_deps+=("$dep")
        fi
    done

    if [ ${#missing_deps[@]} -gt 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        log_info "Please install the missing dependencies and try again"
        exit 1
    fi
}

# Main execution
check_dependencies
main "$@"
