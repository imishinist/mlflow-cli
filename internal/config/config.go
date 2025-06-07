package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
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
	validResolutions := map[string]bool{
		"1m": true, "5m": true, "1h": true,
	}
	if !validResolutions[c.TimeResolution] {
		return fmt.Errorf("invalid time resolution: %s (valid: 1m, 5m, 1h)", c.TimeResolution)
	}

	// Validate time alignment
	validAlignments := map[string]bool{
		"floor": true, "ceil": true, "round": true,
	}
	if !validAlignments[c.TimeAlignment] {
		return fmt.Errorf("invalid time alignment: %s (valid: floor, ceil, round)", c.TimeAlignment)
	}

	// Validate step mode
	validStepModes := map[string]bool{
		"auto": true, "timestamp": true, "sequence": true,
	}
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
		host := strings.TrimPrefix(c.TrackingURI, "https://")
		// Remove any path components
		if idx := strings.Index(host, "/"); idx != -1 {
			host = host[:idx]
		}

		// Check for common Databricks domain patterns
		return strings.HasSuffix(host, ".cloud.databricks.com") ||
			strings.HasSuffix(host, ".azuredatabricks.net") ||
			strings.HasSuffix(host, ".gcp.databricks.com")
	}

	return false
}

// GetDatabricksProfile extracts the profile name from databricks://{profile} URI
func (c *Config) GetDatabricksProfile() string {
	if strings.HasPrefix(c.TrackingURI, "databricks://") {
		profile := strings.TrimPrefix(c.TrackingURI, "databricks://")
		// Remove any trailing slashes or paths
		if idx := strings.Index(profile, "/"); idx != -1 {
			profile = profile[:idx]
		}
		return profile
	}
	return ""
}
