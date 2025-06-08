package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/imishinist/mlflow-cli/internal/config"
	"github.com/imishinist/mlflow-cli/internal/mlflow"
	"github.com/imishinist/mlflow-cli/internal/parser"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Log parameters, metrics, and artifacts",
	Long:  "Log parameters, metrics, and artifacts to MLflow runs",
}

var logParamsCmd = &cobra.Command{
	Use:   "params",
	Short: "Log parameters to MLflow run",
	Long:  "Log parameters to an existing MLflow run",
	RunE:  logParams,
}

func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.AddCommand(logParamsCmd)

	// Params command flags
	logParamsCmd.Flags().String("run-id", "", "Run ID to log parameters to (required)")
	logParamsCmd.Flags().StringArray("param", []string{}, "Parameters in key=value format")
	logParamsCmd.Flags().String("from-file", "", "Load parameters from file (JSON/YAML)")
	logParamsCmd.MarkFlagRequired("run-id")
}

func logParams(cmd *cobra.Command, args []string) error {
	cfg := config.New()
	client, err := mlflow.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create MLflow client: %w", err)
	}

	// Parse flags
	runID, _ := cmd.Flags().GetString("run-id")
	params, _ := cmd.Flags().GetStringArray("param")
	fromFile, _ := cmd.Flags().GetString("from-file")

	ctx := context.Background()

	// Log parameters from command line
	if len(params) > 0 {
		paramMap := make(map[string]string)
		for _, param := range params {
			parts := strings.SplitN(param, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid parameter format: %s (expected key=value)", param)
			}
			paramMap[parts[0]] = parts[1]
		}

		if err := client.LogParamsFromMap(ctx, runID, paramMap); err != nil {
			return fmt.Errorf("failed to log parameters: %w", err)
		}

		fmt.Printf("Successfully logged %d parameters\n", len(paramMap))
		for key, value := range paramMap {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	// Log parameters from file
	if fromFile != "" {
		file, err := os.Open(fromFile)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", fromFile, err)
		}
		defer file.Close()

		var paramMap map[string]string
		ext := strings.ToLower(filepath.Ext(fromFile))

		switch ext {
		case ".json":
			paramMap, err = parser.ParseJSONParams(file)
		case ".yaml", ".yml":
			paramMap, err = parser.ParseYAMLParams(file)
		default:
			return fmt.Errorf("unsupported file format: %s (supported: .json, .yaml, .yml)", ext)
		}

		if err != nil {
			return fmt.Errorf("failed to parse parameters file: %w", err)
		}

		if err := client.LogParamsFromMap(ctx, runID, paramMap); err != nil {
			return fmt.Errorf("failed to log parameters from file: %w", err)
		}

		fmt.Printf("Successfully logged %d parameters from %s\n", len(paramMap), fromFile)
		for key, value := range paramMap {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	if len(params) == 0 && fromFile == "" {
		return fmt.Errorf("either --param or --from-file must be specified")
	}

	return nil
}
