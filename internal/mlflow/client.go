package mlflow

import (
	"fmt"

	"github.com/databricks/databricks-sdk-go"

	"github.com/imishinist/mlflow-cli/internal/config"
)

// Client wraps the Databricks SDK client for MLflow operations
type Client struct {
	client *databricks.WorkspaceClient
	config *config.Config
}

// NewClient creates a new MLflow client with appropriate configuration
func NewClient(cfg *config.Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	databricksConfig, err := buildDatabricksConfig(cfg)
	if err != nil {
		return nil, err
	}

	client, err := databricks.NewWorkspaceClient(databricksConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create MLflow client: %w", err)
	}

	return &Client{
		client: client,
		config: cfg,
	}, nil
}

// buildDatabricksConfig creates appropriate Databricks configuration based on tracking URI
func buildDatabricksConfig(cfg *config.Config) (*databricks.Config, error) {
	if cfg.IsDatabricks() {
		return buildDatabricksMLflowConfig(cfg)
	}
	return buildRegularMLflowConfig(cfg), nil
}

// buildDatabricksMLflowConfig creates configuration for Databricks MLflow
func buildDatabricksMLflowConfig(cfg *config.Config) (*databricks.Config, error) {
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
		return nil, fmt.Errorf("Databricks configuration required: set DATABRICKS_HOST environment variable, use full Databricks URL, or specify profile with databricks://{profile}")
	}

	return databricksConfig, nil
}

// buildRegularMLflowConfig creates configuration for regular MLflow server
func buildRegularMLflowConfig(cfg *config.Config) *databricks.Config {
	return &databricks.Config{
		Host: cfg.TrackingURI,
		// For regular MLflow server, use a dummy token to bypass authentication
		Token: "dummy-token-for-regular-mlflow",
	}
}
