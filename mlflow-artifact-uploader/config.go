package uploader

import (
	"fmt"
	"strings"
	"time"
)

// Config holds configuration for MLflow artifact uploader
type Config struct {
	TrackingURI     string        // MLflow tracking URI
	DatabricksHost  string        // Databricks host (optional)
	DatabricksToken string        // Databricks token (optional)
	Profile         string        // Databricks profile (optional)
	Timeout         time.Duration // Default timeout for operations
}

// Databricks domain suffixes for URL detection
var databricksDomains = []string{
	".cloud.databricks.com",
	".azuredatabricks.net",
	".gcp.databricks.com",
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.TrackingURI == "" {
		return fmt.Errorf("tracking URI is required")
	}

	if c.IsDatabricks() {
		if c.TrackingURI == "databricks" && c.DatabricksHost == "" {
			return fmt.Errorf("Databricks host is required when using 'databricks' tracking URI")
		}
	}

	return nil
}

// IsDatabricks returns true if this is a Databricks MLflow configuration
func (c *Config) IsDatabricks() bool {
	if c.TrackingURI == "databricks" {
		return true
	}

	if strings.HasPrefix(c.TrackingURI, "databricks://") {
		return true
	}

	// Check if tracking URI contains Databricks domain
	for _, domain := range databricksDomains {
		if strings.Contains(c.TrackingURI, domain) {
			return true
		}
	}

	return false
}

// GetDatabricksProfile extracts profile name from databricks:// URI
func (c *Config) GetDatabricksProfile() string {
	if strings.HasPrefix(c.TrackingURI, "databricks://") {
		return strings.TrimPrefix(c.TrackingURI, "databricks://")
	}
	return c.Profile
}

// SetDefaults sets default values for configuration
func (c *Config) SetDefaults() {
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
}
