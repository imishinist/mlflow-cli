package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/imishinist/mlflow-cli/internal/config"
	"github.com/imishinist/mlflow-cli/internal/mlflow"
)

var logArtifactCmd = &cobra.Command{
	Use:   "artifact",
	Short: "Log artifact to MLflow run",
	Long: `Log a file as an artifact to an MLflow run.
The file will be uploaded with its original filename unless --artifact-path is specified.`,
	Example: `  # Upload a file with its original name
  mlflow-cli log artifact --run-id <run-id> --file model.pkl
  
  # Upload a file with a custom artifact path
  mlflow-cli log artifact --run-id <run-id> --file model.pkl --artifact-path models/final_model.pkl
  
  # Upload multiple files
  mlflow-cli log artifact --run-id <run-id> --file model.pkl --file config.yaml`,
	RunE: logArtifact,
}

func init() {
	logCmd.AddCommand(logArtifactCmd)

	// Artifact command flags
	logArtifactCmd.Flags().String("run-id", "", "Run ID to upload artifacts to (required)")
	logArtifactCmd.Flags().StringSlice("file", []string{}, "File path to upload (can be specified multiple times)")
	logArtifactCmd.Flags().String("artifact-path", "", "Custom artifact path (only valid when uploading a single file)")
	logArtifactCmd.MarkFlagRequired("run-id")
	logArtifactCmd.MarkFlagRequired("file")
}

func logArtifact(cmd *cobra.Command, args []string) error {
	cfg := config.New()
	client, err := mlflow.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create MLflow client: %w", err)
	}

	// Parse flags
	runID, _ := cmd.Flags().GetString("run-id")
	files, _ := cmd.Flags().GetStringSlice("file")
	artifactPath, _ := cmd.Flags().GetString("artifact-path")

	// Validation
	if len(files) == 0 {
		return fmt.Errorf("at least one file must be specified")
	}

	if len(files) > 1 && artifactPath != "" {
		return fmt.Errorf("--artifact-path can only be used when uploading a single file")
	}

	ctx := context.Background()
	successCount := 0

	for _, filePath := range files {
		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "File not found: %s\n", filePath)
			continue
		}

		// Determine artifact path
		var targetPath string
		if artifactPath != "" {
			targetPath = artifactPath
		} else {
			targetPath = filepath.Base(filePath)
		}

		err := client.UploadArtifact(ctx, runID, filePath, targetPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to upload %s: %v\n", filePath, err)
			continue
		}
		successCount++
	}

	if successCount == 0 {
		return fmt.Errorf("failed to upload any artifacts")
	}

	// Output success message
	if len(files) == 1 {
		fmt.Printf("Successfully uploaded artifact: %s\n", files[0])
		if artifactPath != "" {
			fmt.Printf("  Artifact path: %s\n", artifactPath)
		} else {
			fmt.Printf("  Artifact path: %s\n", filepath.Base(files[0]))
		}
	} else {
		fmt.Printf("Successfully uploaded %d/%d artifacts\n", successCount, len(files))
	}

	return nil
}
