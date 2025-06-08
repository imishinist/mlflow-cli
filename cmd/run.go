package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/imishinist/mlflow-cli/internal/config"
	"github.com/imishinist/mlflow-cli/internal/mlflow"
	"github.com/imishinist/mlflow-cli/internal/models"
	"github.com/spf13/cobra"
)

// Valid run statuses
var validRunStatuses = map[string]models.RunStatus{
	"FINISHED": models.RunStatusFinished,
	"FAILED":   models.RunStatusFailed,
	"KILLED":   models.RunStatusKilled,
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Manage MLflow runs",
	Long:  "Create, update, and manage MLflow runs",
}

var runStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new MLflow run",
	Long:  "Create and start a new MLflow run",
	RunE:  runStart,
}

var runEndCmd = &cobra.Command{
	Use:   "end",
	Short: "End an MLflow run",
	Long:  "End an existing MLflow run",
	RunE:  runEnd,
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.AddCommand(runStartCmd)
	runCmd.AddCommand(runEndCmd)

	// Start command flags
	runStartCmd.Flags().String("experiment-id", "", "Experiment ID (overrides MLFLOW_EXPERIMENT_ID)")
	runStartCmd.Flags().String("run-name", "", "Run name (default: timestamp-based)")
	runStartCmd.Flags().StringArray("tag", []string{}, "Tags in key=value format")
	runStartCmd.Flags().String("description", "", "Run description")

	// End command flags
	runEndCmd.Flags().String("run-id", "", "Run ID to end (required)")
	runEndCmd.Flags().String("status", "FINISHED", "End status (FINISHED/FAILED/KILLED)")
	runEndCmd.MarkFlagRequired("run-id")
}

func runStart(cmd *cobra.Command, args []string) error {
	cfg := config.New()
	client, err := mlflow.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create MLflow client: %w", err)
	}

	runConfig, err := buildRunConfig(cmd, cfg)
	if err != nil {
		return err
	}

	// Create run
	ctx := context.Background()
	runInfo, err := client.CreateRun(ctx, runConfig)
	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}

	// Output only run ID for shell scripting
	fmt.Printf("%s\n", runInfo.RunID)

	return nil
}

// buildRunConfig constructs RunConfig from command flags and configuration
func buildRunConfig(cmd *cobra.Command, cfg *config.Config) (*models.RunConfig, error) {
	// Parse flags
	experimentID, _ := cmd.Flags().GetString("experiment-id")
	runName, _ := cmd.Flags().GetString("run-name")
	tags, _ := cmd.Flags().GetStringArray("tag")
	description, _ := cmd.Flags().GetString("description")

	// Use experiment ID from flag, environment variable, or config
	if experimentID == "" {
		experimentID = cfg.ExperimentID
	}

	// Validate experiment ID
	if experimentID == "" {
		return nil, fmt.Errorf("experiment ID must be specified via --experiment-id flag or MLFLOW_EXPERIMENT_ID environment variable")
	}

	// Parse tags
	tagMap, err := parseTags(tags)
	if err != nil {
		return nil, err
	}

	// Build run config
	runConfig := &models.RunConfig{
		ExperimentID: &experimentID,
		Tags:         tagMap,
	}

	if runName != "" {
		runConfig.RunName = &runName
	}

	if description != "" {
		// Process escape sequences in description
		processedDescription := processEscapeSequences(description)
		runConfig.Description = &processedDescription
	}

	return runConfig, nil
}

// parseTags parses tag strings in key=value format
func parseTags(tags []string) (map[string]string, error) {
	tagMap := make(map[string]string)
	for _, tag := range tags {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid tag format: %s (expected key=value)", tag)
		}
		tagMap[parts[0]] = parts[1]
	}
	return tagMap, nil
}

func runEnd(cmd *cobra.Command, args []string) error {
	cfg := config.New()
	client, err := mlflow.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create MLflow client: %w", err)
	}

	// Parse flags
	runID, _ := cmd.Flags().GetString("run-id")
	status, _ := cmd.Flags().GetString("status")

	// Validate status
	runStatus, valid := validRunStatuses[status]
	if !valid {
		return fmt.Errorf("invalid status: %s (valid: FINISHED, FAILED, KILLED)", status)
	}

	// Update run
	ctx := context.Background()
	err = client.UpdateRun(ctx, runID, runStatus)
	if err != nil {
		return fmt.Errorf("failed to end run: %w", err)
	}

	fmt.Printf("Run ended successfully\n")
	fmt.Printf("Run ID: %s\n", runID)
	fmt.Printf("Status: %s\n", status)

	return nil
}

// processEscapeSequences processes common escape sequences in strings
func processEscapeSequences(s string) string {
	// Replace common escape sequences
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\r", "\r")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}
