# MLflow CLI Tool

A command-line tool for MLflow tracking operations built with Go.

## Features

- ✅ Create and manage MLflow runs
- ✅ Log parameters (single or batch from file)
- ✅ Log metrics (single or batch from file)
- ✅ Time series processing with configurable resolution
- ✅ Log artifacts (local filesystem and DBFS)

## Installation

```bash
# Install from GitHub
go install github.com/imishinist/mlflow-cli@latest

# Or build from source
git clone https://github.com/imishinist/mlflow-cli.git
cd mlflow-cli
make build

# Install locally
make install

# Cross-compile for multiple platforms
make cross-compile
```

## Configuration

The tool uses environment variables for configuration:

```bash
export MLFLOW_TRACKING_URI=http://localhost:8885  # MLflow server URL
export MLFLOW_EXPERIMENT_ID=123456789             # Default experiment ID
export MLFLOW_TIME_RESOLUTION=1m                  # Time resolution (1m, 5m, 1h)
export MLFLOW_TIME_ALIGNMENT=floor                # Time alignment (floor, ceil, round)
export MLFLOW_STEP_MODE=auto                      # Step mode (auto, timestamp, sequence)
```

### Databricks MLflow

To use Databricks MLflow, you have several options:

#### Option 1: Use `databricks` with environment variables
```bash
export MLFLOW_TRACKING_URI=databricks
export DATABRICKS_HOST=https://your-workspace.cloud.databricks.com
export DATABRICKS_TOKEN=your-databricks-token
export MLFLOW_EXPERIMENT_ID=123456789
```

#### Option 2: Use `databricks://{profile}` with Databricks CLI profiles (Recommended)
```bash
export MLFLOW_TRACKING_URI=databricks://my-profile
export MLFLOW_EXPERIMENT_ID=123456789
```

#### Option 3: Use full Databricks URL
```bash
export MLFLOW_TRACKING_URI=https://your-workspace.cloud.databricks.com
export MLFLOW_EXPERIMENT_ID=123456789
```

The tool automatically detects Databricks URLs and uses appropriate authentication. When using profiles, make sure your Databricks CLI is configured with `databricks configure --token`.

**Note**: For DBFS artifact uploads, all authentication methods are supported. Profile-based authentication is recommended for ease of use.

## Usage

### 1. Start a new run

```bash
# Start run with environment variable experiment ID
mlflow-cli run start --run-name "test-run-1"

# Start run with explicit experiment ID
mlflow-cli run start --experiment-id "1" --run-name "test-run-1"

# With tags and description
mlflow-cli run start \
  --experiment-id "1" \
  --run-name "test-run-1" \
  --tag "version=1.0" \
  --tag "env=production" \
  --description "Performance test run"
```

### 2. Log parameters

```bash
# Log single parameters
mlflow-cli log params --run-id <run-id> --param batch_size=100 --param learning_rate=0.001

# Log parameters from file
mlflow-cli log params --run-id <run-id> --from-file test_params.json
```

### 3. Log metrics

```bash
# Log single metric
mlflow-cli log metric --run-id <run-id> --name "accuracy" --value 0.95 --step 1

# Log metrics from file
mlflow-cli log metrics --run-id <run-id> --from-file test_metrics.json

# With custom time processing
mlflow-cli log metrics \
  --run-id <run-id> \
  --from-file test_metrics.json \
  --time-resolution 5m \
  --time-alignment ceil \
  --step-mode timestamp
```

### 4. Log artifacts

```bash
# Log single artifact
mlflow-cli log artifact --run-id <run-id> --file model.pkl

# Log artifact with custom path
mlflow-cli log artifact --run-id <run-id> --file model.pkl --artifact-path models/final_model.pkl

# Log multiple artifacts
mlflow-cli log artifact --run-id <run-id> --file model.pkl --file config.yaml
```

#### DBFS Artifacts (Databricks)

DBFS artifact uploads support all Databricks authentication methods:

**Option 1: Profile-based authentication (Recommended)**
```bash
export MLFLOW_TRACKING_URI=databricks://my-profile
mlflow-cli log artifact --run-id <run-id> --file model.pkl
```

**Option 2: Direct Databricks URL**
```bash
export MLFLOW_TRACKING_URI=https://your-workspace.cloud.databricks.com
mlflow-cli log artifact --run-id <run-id> --file model.pkl
```

**Option 3: Environment variables**
```bash
export MLFLOW_TRACKING_URI=databricks
export DATABRICKS_HOST=https://your-workspace.cloud.databricks.com
mlflow-cli log artifact --run-id <run-id> --file model.pkl
```

**Supported credential types:**
- AWS S3 (AWS_PRESIGNED_URL)
- Azure Blob Storage (AZURE_SAS_URI)
- Google Cloud Storage (GCP_SIGNED_URL)
- Azure Data Lake Storage Gen2 (AZURE_ADLS_GEN2_SAS_URI)

**Note**: All authentication methods are fully supported for DBFS artifacts. Profile-based authentication is recommended for ease of use.

### 5. End a run

```bash
# End run successfully
mlflow-cli run end --run-id <run-id> --status FINISHED

# End run with failure
mlflow-cli run end --run-id <run-id> --status FAILED
```

## File Formats

### Parameters File (JSON)
```json
{
  "parameters": {
    "batch_size": "100",
    "learning_rate": "0.001",
    "epochs": "50"
  }
}
```

### Parameters File (YAML)
```yaml
parameters:
  batch_size: "100"
  learning_rate: "0.001"
  epochs: "50"
```

### Metrics File (JSON)
```json
{
  "metrics": [
    {
      "timestamp": "2025-06-07T14:01:00Z",
      "execution_time": 1.5,
      "success_rate": 0.95,
      "error_count": 2
    },
    {
      "timestamp": "2025-06-07T14:02:00Z",
      "execution_time": 1.3,
      "success_rate": 0.97,
      "error_count": 1
    }
  ]
}
```

### Metrics File (YAML)
```yaml
metrics:
  - timestamp: "2025-06-07T14:01:00Z"
    execution_time: 1.5
    success_rate: 0.95
    error_count: 2
  - timestamp: "2025-06-07T14:02:00Z"
    execution_time: 1.3
    success_rate: 0.97
    error_count: 1
```

## Testing

### Unit Tests
```bash
make test
```

### E2E Tests
The project includes comprehensive E2E tests that use Docker Compose to run a real MLflow server:

```bash
# Build MLflow Docker image (first time only)
make docker-build

# Recommended development workflow:
make docker-up       # Start MLflow server once
make e2e-test        # Run E2E tests (fast mode, default)
make e2e-test-debug  # Run with debug output (fast mode)
make docker-down     # Stop server when done

# Full E2E tests (with Docker setup/teardown)
make e2e-test-all       # Full mode with Docker management
make e2e-test-full      # Alias for e2e-test-all
make e2e-test-all-debug # Full mode with debug output

# Docker operations
make docker-up    # Start MLflow at http://localhost:5001
make docker-down  # Stop and clean up
make docker-logs  # View server logs

# Get help
make help         # Show all available targets
```

**Development workflow (recommended):**
1. Start MLflow server once: `make docker-up`
2. Run tests multiple times: `make e2e-test` (fast, ~2 seconds)
3. Stop server when done: `make docker-down`

**Test modes:**
- **Fast mode** (`make e2e-test`): Uses existing MLflow server, ~2 seconds
- **Full mode** (`make e2e-test-all`): Manages Docker lifecycle, ~30 seconds

The E2E tests use a custom Docker image with pre-installed MLflow for faster startup times.

The E2E tests cover:
- Run lifecycle (start/end)
- Parameter logging (single and batch from JSON/YAML files)
- Metric logging (single and batch from JSON/YAML files)
- Time series processing with different configurations
- **Data verification**: Confirms that logged data can be retrieved correctly
- **Parameter validation**: Verifies parameter values match expected inputs
- **Metric validation**: Confirms metric values and time series data points
- Error handling and validation
- API integration verification

## Time Series Processing

The tool automatically processes time series data to ensure consistency:

- **Time Resolution**: Aligns all timestamps to specified intervals (1m, 5m, 1h)
- **Time Alignment**: Controls how timestamps are rounded (floor, ceil, round)
- **Step Mode**: Determines how step numbers are generated
  - `auto`: Use timestamp-based steps if timestamps exist, otherwise sequence
  - `timestamp`: Convert timestamps to minutes from base time
  - `sequence`: Use sequential numbering (0, 1, 2, ...)

## Example Workflow

```bash
# Set environment (using profile-based authentication)
export MLFLOW_TRACKING_URI=databricks://my-profile
export MLFLOW_EXPERIMENT_ID=123456789

# Start run (outputs only run ID for shell scripting)
RUN_ID=$(mlflow-cli run start --run-name "batch-100-test")
echo "Started run: $RUN_ID"

# Log parameters
mlflow-cli log params --run-id $RUN_ID --param batch_size=100 --param timeout=300

# Run your experiment
./run_experiment.sh  # This should generate test_metrics.json

# Log metrics
mlflow-cli log metrics --run-id $RUN_ID --from-file test_metrics.json

# Log artifacts (including DBFS support)
mlflow-cli log artifact --run-id $RUN_ID --file model.pkl --file config.yaml

# End run
mlflow-cli run end --run-id $RUN_ID --status FINISHED
```

## Shell Integration

The `run start` command outputs only the Run ID to stdout, making it easy to capture in shell variables:

```bash
# Capture run ID
RUN_ID=$(mlflow-cli run start --run-name "my-run")

# Use in subsequent commands
mlflow-cli log params --run-id $RUN_ID --param key=value
mlflow-cli log metric --run-id $RUN_ID --name accuracy --value 0.95
mlflow-cli run end --run-id $RUN_ID --status FINISHED
```

## Development

```bash
# Install dependencies
make deps

# Build
make build

# Run tests
make test

# Development build with race detection
make dev
```

## Requirements

- Go 1.21+
- MLflow server running and accessible

## Notes

- Port 5000 is often used by Apple AirPlay on macOS. Use a different port (e.g., 5001) for MLflow server.
- The tool uses MLflow REST API directly for maximum compatibility.
- Time series processing ensures all metrics can be compared on the same timeline.
