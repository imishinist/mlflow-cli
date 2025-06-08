package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Databricks domain suffixes for URL detection
var databricksDomains = []string{
	".cloud.databricks.com",
	".azuredatabricks.net",
	".gcp.databricks.com",
}

// Valid configuration values
var (
	validTimeResolutions = map[string]bool{
		"1m": true, "5m": true, "1h": true,
	}
	validTimeAlignments = map[string]bool{
		"floor": true, "ceil": true, "round": true,
	}
	validStepModes = map[string]bool{
		"auto": true, "timestamp": true, "sequence": true,
	}
)

type Config struct {
	TrackingURI     string
	ExperimentID    string
	TimeResolution  string
	TimeAlignment   string
	StepMode        string
	DatabricksHost  string
	DatabricksToken string
}

func New() *Config {
	return &Config{
		TrackingURI:     viper.GetString("tracking_uri"),
		ExperimentID:    viper.GetString("experiment_id"),
		TimeResolution:  viper.GetString("time_resolution"),
		TimeAlignment:   viper.GetString("time_alignment"),
		StepMode:        viper.GetString("step_mode"),
		DatabricksHost:  viper.GetString("databricks_host"),
		DatabricksToken: viper.GetString("databricks_token"),
	}
}

func (c *Config) Validate() error {
	if c.TrackingURI == "" {
		return fmt.Errorf("tracking URI is required")
	}

	// Validate time resolution
	if !validTimeResolutions[c.TimeResolution] {
		return fmt.Errorf("invalid time resolution: %s (valid: 1m, 5m, 1h)", c.TimeResolution)
	}

	// Validate time alignment
	if !validTimeAlignments[c.TimeAlignment] {
		return fmt.Errorf("invalid time alignment: %s (valid: floor, ceil, round)", c.TimeAlignment)
	}

	// Validate step mode
	if !validStepModes[c.StepMode] {
		return fmt.Errorf("invalid step mode: %s (valid: auto, timestamp, sequence)", c.StepMode)
	}

	return nil
}

// IsDatabricks checks if the tracking URI points to Databricks
func (c *Config) IsDatabricks() bool {
	if c.TrackingURI == "databricks" {
		return true
	}

	// Check for databricks:// protocol
	if strings.HasPrefix(c.TrackingURI, "databricks://") {
		return true
	}

	// Check for Databricks URLs
	if strings.HasPrefix(c.TrackingURI, "https://") {
		host := c.extractHostFromURL(c.TrackingURI)
		return c.isDatabricksHost(host)
	}

	return false
}

// extractHostFromURL extracts the hostname from a URL
func (c *Config) extractHostFromURL(url string) string {
	host := strings.TrimPrefix(url, "https://")
	// Remove any path components
	if idx := strings.Index(host, "/"); idx != -1 {
		host = host[:idx]
	}
	return host
}

// isDatabricksHost checks if a hostname belongs to Databricks
func (c *Config) isDatabricksHost(host string) bool {
	for _, domain := range databricksDomains {
		if strings.HasSuffix(host, domain) {
			return true
		}
	}
	return false
}

// GetDatabricksProfile extracts the profile name from databricks://{profile} URI
func (c *Config) GetDatabricksProfile() string {
	if !strings.HasPrefix(c.TrackingURI, "databricks://") {
		return ""
	}

	profile := strings.TrimPrefix(c.TrackingURI, "databricks://")
	// Remove any trailing slashes or paths
	if idx := strings.Index(profile, "/"); idx != -1 {
		profile = profile[:idx]
	}
	return profile
}
