package mlflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/databricks/databricks-sdk-go/service/ml"
)

// UploadArtifact uploads a file as an artifact to the specified run
func (c *Client) UploadArtifact(ctx context.Context, runID, filePath, artifactPath string) error {
	// Get the artifact URI from the run info
	artifactURI, err := c.getArtifactURI(ctx, runID)
	if err != nil {
		return fmt.Errorf("failed to get artifact URI: %w", err)
	}

	// Use filename if artifact path is not specified
	if artifactPath == "" {
		artifactPath = filepath.Base(filePath)
	}

	// Upload to the appropriate storage based on artifact URI
	return c.uploadToStorage(ctx, artifactURI, filePath, artifactPath)
}

// UploadArtifacts uploads multiple files as artifacts to the specified run
func (c *Client) UploadArtifacts(ctx context.Context, runID string, files map[string]string) error {
	for filePath, artifactPath := range files {
		if err := c.UploadArtifact(ctx, runID, filePath, artifactPath); err != nil {
			return fmt.Errorf("failed to upload %s: %w", filePath, err)
		}
	}
	return nil
}

// openFileWithInfo opens a file and returns the file handle and file info
func (c *Client) openFileWithInfo(filePath string) (*os.File, os.FileInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return file, fileInfo, nil
}

// createPutRequest creates a PUT HTTP request with common headers
func (c *Client) createPutRequest(ctx context.Context, url string, body io.Reader, contentLength int64) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "PUT", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", contentLength))
	c.addAuthHeaders(req)

	return req, nil
}

// getArtifactURI retrieves the artifact URI for a given run
func (c *Client) getArtifactURI(ctx context.Context, runID string) (string, error) {
	// Use Databricks SDK if available (works for both Databricks and regular MLflow)
	if c.client != nil {
		resp, err := c.client.Experiments.GetRun(ctx, ml.GetRunRequest{
			RunId: runID,
		})
		if err != nil {
			return "", fmt.Errorf("failed to get run: %w", err)
		}

		if resp.Run.Info.ArtifactUri == "" {
			return "", fmt.Errorf("artifact URI not found for run %s", runID)
		}

		return resp.Run.Info.ArtifactUri, nil
	}

	// Fallback to HTTP API if SDK client is not available
	return c.getArtifactURIFromHTTP(ctx, runID)
}

// getArtifactURIFromHTTP retrieves artifact URI using HTTP API for regular MLflow server
func (c *Client) getArtifactURIFromHTTP(ctx context.Context, runID string) (string, error) {
	url := fmt.Sprintf("%s/api/2.0/mlflow/runs/get?run_id=%s", strings.TrimSuffix(c.config.TrackingURI, "/"), runID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get run request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var runResponse struct {
		Run struct {
			Info struct {
				ArtifactURI string `json:"artifact_uri"`
			} `json:"info"`
		} `json:"run"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&runResponse); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if runResponse.Run.Info.ArtifactURI == "" {
		return "", fmt.Errorf("artifact URI not found for run %s", runID)
	}

	return runResponse.Run.Info.ArtifactURI, nil
}

// uploadToStorage uploads file to the appropriate storage based on URI scheme
func (c *Client) uploadToStorage(ctx context.Context, artifactURI, filePath, artifactPath string) error {
	if strings.HasPrefix(artifactURI, "mlflow-artifacts:/") {
		return c.uploadToMLflowArtifacts(ctx, artifactURI, filePath, artifactPath)
	} else if strings.HasPrefix(artifactURI, "file://") || strings.HasPrefix(artifactURI, "/") {
		return c.uploadToLocalFS(ctx, artifactURI, filePath, artifactPath)
	} else {
		return fmt.Errorf("unsupported artifact URI scheme: %s", artifactURI)
	}
}

// uploadToMLflowArtifacts uploads using MLflow Artifacts Service
func (c *Client) uploadToMLflowArtifacts(ctx context.Context, artifactURI, filePath, artifactPath string) error {
	// Extract experiment_id and run_id from artifact URI
	experimentID, runID, err := c.extractIDsFromArtifactURI(artifactURI)
	if err != nil {
		return fmt.Errorf("failed to extract IDs from artifact URI: %w", err)
	}

	// Open file and get info
	file, fileInfo, err := c.openFileWithInfo(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Build URL: /api/2.0/mlflow-artifacts/artifacts/{experiment_id}/{run_id}/artifacts/{artifact_path}
	baseURL := strings.TrimSuffix(c.config.TrackingURI, "/")
	url := fmt.Sprintf("%s/api/2.0/mlflow-artifacts/artifacts/%s/%s/artifacts/%s", baseURL, experimentID, runID, artifactPath)

	// Create HTTP request
	req, err := c.createPutRequest(ctx, url, file, fileInfo.Size())
	if err != nil {
		return err
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload to MLflow Artifacts Service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MLflow Artifacts Service upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// uploadToLocalFS uploads file to local filesystem
func (c *Client) uploadToLocalFS(ctx context.Context, artifactURI, filePath, artifactPath string) error {
	localPath := strings.TrimPrefix(artifactURI, "file://")
	if !strings.HasSuffix(localPath, "/") {
		localPath += "/"
	}
	localPath += artifactPath

	// Create directory if it doesn't exist
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Copy file
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy content
	_, err = destFile.ReadFrom(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// extractIDsFromArtifactURI extracts experiment ID and run ID from mlflow-artifacts URI
func (c *Client) extractIDsFromArtifactURI(artifactURI string) (string, string, error) {
	// mlflow-artifacts:/0/47485d6a0b734e37aaddc60be04b7371/artifacts
	// Extract the experiment ID (first path component) and run ID (second path component)
	parts := strings.Split(strings.TrimPrefix(artifactURI, "mlflow-artifacts:"), "/")
	if len(parts) < 3 {
		return "", "", fmt.Errorf("invalid mlflow-artifacts URI format: %s", artifactURI)
	}

	// Remove empty first element if URI starts with /
	if parts[0] == "" && len(parts) > 3 {
		parts = parts[1:]
	}

	if len(parts) < 3 {
		return "", "", fmt.Errorf("invalid mlflow-artifacts URI format: %s", artifactURI)
	}

	experimentID := parts[0]
	runID := parts[1]

	return experimentID, runID, nil
}

// addAuthHeaders adds appropriate authentication headers to the request
func (c *Client) addAuthHeaders(req *http.Request) {
	// Handle Databricks authentication
	if c.config.IsDatabricks() && c.config.DatabricksToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.DatabricksToken)
	}
}
