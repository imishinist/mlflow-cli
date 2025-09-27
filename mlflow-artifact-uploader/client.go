package uploader

import (
	"context"
	"fmt"
	"io"

	"github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/httpclient"
)

// Client wraps the Databricks SDK client for MLflow artifact operations
type Client struct {
	client    *databricks.WorkspaceClient
	config    *Config
	apiClient *httpclient.ApiClient
}

// NewClient creates a new MLflow artifact uploader client
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	// Set defaults and validate
	config.SetDefaults()
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	databricksConfig, err := buildDatabricksConfig(config)
	if err != nil {
		return nil, err
	}

	client, err := databricks.NewWorkspaceClient(databricksConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create MLflow client: %w", err)
	}

	// Create API client for DBFS artifacts if this is a Databricks client
	var apiClient *httpclient.ApiClient
	if config.IsDatabricks() && client != nil {
		apiClient, err = client.Config.NewApiClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create API client: %w", err)
		}
	}

	return &Client{
		client:    client,
		config:    config,
		apiClient: apiClient,
	}, nil
}

// UploadArtifact uploads a single artifact from an ArtifactSource
func (c *Client) UploadArtifact(ctx context.Context, runID string, source ArtifactSource, opts ...UploadOptions) error {
	if source == nil {
		return fmt.Errorf("source is required")
	}

	// Get the artifact URI from the run info
	artifactURI, err := c.getArtifactURI(ctx, runID)
	if err != nil {
		return fmt.Errorf("failed to get artifact URI: %w", err)
	}

	// Upload to the appropriate storage based on artifact URI
	return c.uploadToStorage(ctx, artifactURI, source, source.Name())
}

// UploadArtifacts uploads multiple artifacts from ArtifactSources
func (c *Client) UploadArtifacts(ctx context.Context, runID string, sources []ArtifactSource, opts ...UploadOptions) error {
	for _, source := range sources {
		if err := c.UploadArtifact(ctx, runID, source, opts...); err != nil {
			return fmt.Errorf("failed to upload %s: %w", source.Name(), err)
		}
	}
	return nil
}

// UploadFile uploads a file as an artifact (convenience method)
func (c *Client) UploadFile(ctx context.Context, runID, filePath, artifactPath string, opts ...UploadOptions) error {
	source, err := NewFileSource(filePath, artifactPath)
	if err != nil {
		return err
	}
	defer source.Close()

	return c.UploadArtifact(ctx, runID, source, opts...)
}

// UploadFiles uploads multiple files as artifacts (convenience method)
func (c *Client) UploadFiles(ctx context.Context, runID string, files map[string]string, opts ...UploadOptions) error {
	for filePath, artifactPath := range files {
		if err := c.UploadFile(ctx, runID, filePath, artifactPath, opts...); err != nil {
			return fmt.Errorf("failed to upload %s: %w", filePath, err)
		}
	}
	return nil
}

// UploadReader uploads from an io.Reader as an artifact (convenience method)
func (c *Client) UploadReader(ctx context.Context, runID string, reader io.Reader, artifactName string, size int64, opts ...UploadOptions) error {
	source := NewReaderSource(reader, artifactName, size)
	defer source.Close()

	return c.UploadArtifact(ctx, runID, source, opts...)
}

// buildDatabricksConfig creates appropriate Databricks configuration based on tracking URI
func buildDatabricksConfig(cfg *Config) (*databricks.Config, error) {
	if cfg.IsDatabricks() {
		return buildDatabricksMLflowConfig(cfg)
	}
	return buildRegularMLflowConfig(cfg), nil
}

// buildDatabricksMLflowConfig creates configuration for Databricks MLflow
func buildDatabricksMLflowConfig(cfg *Config) (*databricks.Config, error) {
	databricksConfig := &databricks.Config{}

	// Handle different Databricks URI formats
	switch {
	case cfg.TrackingURI == "databricks":
		// Use DATABRICKS_HOST if available
		if cfg.DatabricksHost != "" {
			databricksConfig.Host = cfg.DatabricksHost
		}
	case cfg.GetDatabricksProfile() != "":
		// Handle databricks://{profile} format
		databricksConfig.Profile = cfg.GetDatabricksProfile()
	default:
		// Use the tracking URI as Databricks host (direct URL)
		databricksConfig.Host = cfg.TrackingURI
	}

	// Set authentication token if available (overrides profile)
	if cfg.DatabricksToken != "" {
		databricksConfig.Token = cfg.DatabricksToken
	}

	// Validate Databricks configuration
	if databricksConfig.Host == "" && databricksConfig.Profile == "" {
		return nil, fmt.Errorf("Databricks configuration required: set DatabricksHost, use full Databricks URL, or specify profile with databricks://{profile}")
	}

	return databricksConfig, nil
}

// buildRegularMLflowConfig creates configuration for regular MLflow server
func buildRegularMLflowConfig(cfg *Config) *databricks.Config {
	return &databricks.Config{
		Host: cfg.TrackingURI,
		// For regular MLflow server, use a dummy token to bypass authentication
		Token: "dummy-token-for-regular-mlflow",
	}
}
