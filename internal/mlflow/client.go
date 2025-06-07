package mlflow

import (
	"fmt"

	"github.com/databricks/databricks-sdk-go"
	"github.com/imishinist/mlflow-cli/internal/config"
)

type Client struct {
	client *databricks.WorkspaceClient
	config *config.Config
}

func NewClient(cfg *config.Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var databricksConfig *databricks.Config

	if cfg.IsDatabricks() {
		// Databricks MLflow configuration
		databricksConfig = &databricks.Config{}

		// Handle different Databricks URI formats
		if cfg.TrackingURI == "databricks" {
			// Use DATABRICKS_HOST if available
			if cfg.DatabricksHost != "" {
				databricksConfig.Host = cfg.DatabricksHost
			}
			// Use default profile if no explicit profile specified
		} else if profile := cfg.GetDatabricksProfile(); profile != "" {
			// Handle databricks://{profile} format
			databricksConfig.Profile = profile
		} else {
			// Use the tracking URI as Databricks host (direct URL)
			databricksConfig.Host = cfg.TrackingURI
		}

		// Set authentication token if available (overrides profile)
		if cfg.DatabricksToken != "" {
			databricksConfig.Token = cfg.DatabricksToken
		}

		// Validate Databricks configuration
		if databricksConfig.Host == "" && databricksConfig.Profile == "" {
			return nil, fmt.Errorf("Databricks host or profile is required when using Databricks MLflow. Set DATABRICKS_HOST environment variable, use a full Databricks URL as tracking URI, or specify a profile with databricks://{profile}")
		}
	} else {
		// Regular MLflow server configuration
		databricksConfig = &databricks.Config{
			Host: cfg.TrackingURI,
			// For regular MLflow server, use a dummy token to bypass authentication
			Token: "dummy-token-for-regular-mlflow",
		}
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
