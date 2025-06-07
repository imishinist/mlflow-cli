package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/imishinist/mlflow-cli/internal/config"
	"github.com/imishinist/mlflow-cli/internal/mlflow"
	"github.com/imishinist/mlflow-cli/internal/models"
	"github.com/imishinist/mlflow-cli/internal/parser"
	timeutils "github.com/imishinist/mlflow-cli/internal/time"
	"github.com/spf13/cobra"
)

var logMetricCmd = &cobra.Command{
	Use:   "metric",
	Short: "Log a single metric to MLflow run",
	Long:  "Log a single metric to an existing MLflow run",
	RunE:  logMetric,
}

var logMetricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Log multiple metrics to MLflow run",
	Long:  "Log multiple metrics from file to an existing MLflow run",
	RunE:  logMetrics,
}

func init() {
	logCmd.AddCommand(logMetricCmd)
	logCmd.AddCommand(logMetricsCmd)

	// Single metric command flags
	logMetricCmd.Flags().String("run-id", "", "Run ID to log metric to (required)")
	logMetricCmd.Flags().String("name", "", "Metric name (required)")
	logMetricCmd.Flags().Float64("value", 0, "Metric value (required)")
	logMetricCmd.Flags().Int64("step", -1, "Step number (optional)")
	logMetricCmd.Flags().String("timestamp", "", "Timestamp in ISO8601 format (optional)")
	logMetricCmd.MarkFlagRequired("run-id")
	logMetricCmd.MarkFlagRequired("name")
	logMetricCmd.MarkFlagRequired("value")

	// Multiple metrics command flags
	logMetricsCmd.Flags().String("run-id", "", "Run ID to log metrics to (required)")
	logMetricsCmd.Flags().String("from-file", "", "Load metrics from file (JSON/YAML/CSV)")
	logMetricsCmd.Flags().String("time-resolution", "", "Time resolution (1m/5m/1h)")
	logMetricsCmd.Flags().String("time-alignment", "", "Time alignment (floor/ceil/round)")
	logMetricsCmd.Flags().String("step-mode", "", "Step mode (auto/timestamp/sequence)")
	logMetricsCmd.MarkFlagRequired("run-id")
	logMetricsCmd.MarkFlagRequired("from-file")
}

func logMetric(cmd *cobra.Command, args []string) error {
	cfg := config.New()
	client, err := mlflow.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create MLflow client: %w", err)
	}

	// Parse flags
	runID, _ := cmd.Flags().GetString("run-id")
	name, _ := cmd.Flags().GetString("name")
	value, _ := cmd.Flags().GetFloat64("value")
	step, _ := cmd.Flags().GetInt64("step")
	timestampStr, _ := cmd.Flags().GetString("timestamp")

	var timestamp *time.Time
	var stepPtr *int64

	// Parse timestamp if provided
	if timestampStr != "" {
		t, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			return fmt.Errorf("invalid timestamp format: %s (expected ISO8601)", timestampStr)
		}
		timestamp = &t
	}

	// Set step if provided
	if step >= 0 {
		stepPtr = &step
	}

	ctx := context.Background()
	if err := client.LogMetric(ctx, runID, name, value, timestamp, stepPtr); err != nil {
		return fmt.Errorf("failed to log metric: %w", err)
	}

	fmt.Printf("Successfully logged metric: %s = %f", name, value)
	if stepPtr != nil {
		fmt.Printf(" (step: %d)", *stepPtr)
	}
	if timestamp != nil {
		fmt.Printf(" (timestamp: %s)", timestamp.Format(time.RFC3339))
	}
	fmt.Println()

	return nil
}

func logMetrics(cmd *cobra.Command, args []string) error {
	cfg := config.New()
	client, err := mlflow.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create MLflow client: %w", err)
	}

	// Parse flags
	runID, _ := cmd.Flags().GetString("run-id")
	fromFile, _ := cmd.Flags().GetString("from-file")
	timeResolution, _ := cmd.Flags().GetString("time-resolution")
	timeAlignment, _ := cmd.Flags().GetString("time-alignment")
	stepMode, _ := cmd.Flags().GetString("step-mode")

	// Use config defaults if not specified
	if timeResolution == "" {
		timeResolution = cfg.TimeResolution
	}
	if timeAlignment == "" {
		timeAlignment = cfg.TimeAlignment
	}
	if stepMode == "" {
		stepMode = cfg.StepMode
	}

	// Open and parse file
	file, err := os.Open(fromFile)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", fromFile, err)
	}
	defer file.Close()

	var metricsFile *models.MetricsFile
	ext := strings.ToLower(filepath.Ext(fromFile))

	switch ext {
	case ".json":
		metricsFile, err = parser.ParseJSONMetrics(file)
	case ".yaml", ".yml":
		metricsFile, err = parser.ParseYAMLMetrics(file)
	default:
		return fmt.Errorf("unsupported file format: %s (supported: .json, .yaml, .yml)", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to parse metrics file: %w", err)
	}

	// Process metrics with time configuration
	timeConfig := models.TimeConfig{
		Resolution: timeResolution,
		Alignment:  timeAlignment,
		StepMode:   stepMode,
	}

	processedMetrics, err := timeutils.ProcessMetrics(metricsFile.Metrics, timeConfig, nil)
	if err != nil {
		return fmt.Errorf("failed to process metrics: %w", err)
	}

	// Log metrics using batch API for efficiency
	ctx := context.Background()
	if err := client.LogBatchMetrics(ctx, runID, processedMetrics); err != nil {
		return fmt.Errorf("failed to log metrics: %w", err)
	}

	fmt.Printf("Successfully logged %d metrics from %s\n", len(processedMetrics), fromFile)
	fmt.Printf("Time configuration: resolution=%s, alignment=%s, step_mode=%s\n",
		timeResolution, timeAlignment, stepMode)

	// Show summary of metrics
	metricCounts := make(map[string]int)
	for _, metric := range processedMetrics {
		metricCounts[metric.Key]++
	}

	fmt.Println("Metrics summary:")
	for key, count := range metricCounts {
		fmt.Printf("  %s: %d data points\n", key, count)
	}

	return nil
}
