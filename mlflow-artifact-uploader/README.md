# MLflow Artifact Uploader

A Go library for uploading artifacts to MLflow tracking servers, supporting multiple storage backends and flexible data sources.

## Features

- **Multiple Storage Backends**: MLflow Artifacts Service, DBFS, Local filesystem
- **Flexible Data Sources**: Upload from files, io.Reader, byte slices
- **Databricks Support**: Full support for Databricks MLflow with authentication
- **Batch Operations**: Upload multiple artifacts efficiently
- **Stream Support**: Upload data without saving to disk first

## Installation

```bash
go get github.com/imishinist/mlflow-artifact-uploader
```

## Quick Start

### Basic File Upload

```go
package main

import (
    "context"
    "log"
    
    "github.com/imishinist/mlflow-artifact-uploader"
)

func main() {
    config := &uploader.Config{
        TrackingURI: "http://localhost:5000",
    }
    
    client, err := uploader.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    err = client.UploadFile(ctx, "run-id", "/path/to/model.pkl", "model/model.pkl")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Upload from Memory

```go
import (
    "encoding/json"
    "strings"
)

// Upload JSON data
data := map[string]interface{}{
    "accuracy": 0.95,
    "model_type": "random_forest",
}

jsonData, _ := json.Marshal(data)
source := uploader.NewBytesSource(jsonData, "metrics/results.json")
err = client.UploadArtifact(ctx, "run-id", source)

// Upload text content
content := "Training log content..."
reader := strings.NewReader(content)
source = uploader.NewReaderSource(reader, "logs/training.log", int64(len(content)))
err = client.UploadArtifact(ctx, "run-id", source)
```

### Databricks Integration

```go
config := &uploader.Config{
    TrackingURI:     "databricks",
    DatabricksHost:  "https://your-workspace.cloud.databricks.com",
    DatabricksToken: "your-token",
}

client, err := uploader.NewClient(config)
// Upload works the same way
```

## API Reference

### Client Creation

```go
func NewClient(config *Config) (*Client, error)
```

### Configuration

```go
type Config struct {
    TrackingURI     string        // MLflow tracking URI
    DatabricksHost  string        // Databricks host (optional)
    DatabricksToken string        // Databricks token (optional)
    Profile         string        // Databricks profile (optional)
    Timeout         time.Duration // Default timeout
}
```

### Artifact Sources

```go
type ArtifactSource interface {
    Name() string      // Artifact name/path
    Reader() io.Reader // Data source
    Size() int64       // Size in bytes (-1 if unknown)
    Close() error      // Resource cleanup
}

// Create sources
func NewFileSource(filePath, artifactName string) (ArtifactSource, error)
func NewReaderSource(reader io.Reader, artifactName string, size int64) ArtifactSource
func NewBytesSource(data []byte, artifactName string) ArtifactSource
```

### Upload Methods

```go
// Primary methods
func (c *Client) UploadArtifact(ctx context.Context, runID string, source ArtifactSource, opts ...UploadOptions) error
func (c *Client) UploadArtifacts(ctx context.Context, runID string, sources []ArtifactSource, opts ...UploadOptions) error

// Convenience methods
func (c *Client) UploadFile(ctx context.Context, runID, filePath, artifactPath string, opts ...UploadOptions) error
func (c *Client) UploadFiles(ctx context.Context, runID string, files map[string]string, opts ...UploadOptions) error
func (c *Client) UploadReader(ctx context.Context, runID string, reader io.Reader, artifactName string, size int64, opts ...UploadOptions) error
```

## Supported Storage Backends

### MLflow Artifacts Service
- URI format: `mlflow-artifacts://{experiment_id}/{run_id}/artifacts`
- Direct upload to MLflow server

### DBFS (Databricks File System)
- URI format: `dbfs:/databricks/mlflow-tracking/{experiment_id}/{run_id}/artifacts`
- Uses Databricks credentials-for-write API
- Supports AWS S3, Azure Blob Storage, Google Cloud Storage

### Local File System
- URI format: `file:///path/to/artifacts` or `/path/to/artifacts`
- Direct file system operations

## Examples

See the [examples](./examples/) directory for complete working examples:

- [basic](./examples/basic/) - Basic file upload operations
- [stream](./examples/stream/) - Upload from memory and streams
- [databricks](./examples/databricks/) - Databricks-specific configurations
- [batch](./examples/batch/) - Batch upload operations

## Authentication

### Regular MLflow Server
No authentication required for most MLflow servers.

### Databricks MLflow

#### Option 1: Direct configuration
```go
config := &uploader.Config{
    TrackingURI:     "databricks",
    DatabricksHost:  "https://your-workspace.cloud.databricks.com",
    DatabricksToken: "your-token",
}
```

#### Option 2: Profile-based
```go
config := &uploader.Config{
    TrackingURI: "databricks://your-profile",
}
```

#### Option 3: Direct URL
```go
config := &uploader.Config{
    TrackingURI:     "https://your-workspace.cloud.databricks.com",
    DatabricksToken: "your-token",
}
```

## Error Handling

The library provides detailed error messages for common issues:

- Invalid configuration
- Network connectivity problems
- Authentication failures
- Storage-specific errors
- File system errors

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
