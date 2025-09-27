package main

import (
	"context"
	"log"
	"os"

	"github.com/imishinist/mlflow-artifact-uploader"
)

func main() {
	// Example 1: Using Databricks with explicit host and token
	config := &uploader.Config{
		TrackingURI:     "databricks",
		DatabricksHost:  "https://your-workspace.cloud.databricks.com",
		DatabricksToken: os.Getenv("DATABRICKS_TOKEN"),
	}

	client, err := uploader.NewClient(config)
	if err != nil {
		log.Fatal("Failed to create Databricks client:", err)
	}

	ctx := context.Background()
	runID := "your-run-id"

	// Upload artifact to Databricks MLflow
	err = client.UploadFile(ctx, runID, "/path/to/model.pkl", "model/model.pkl")
	if err != nil {
		log.Fatal("Failed to upload to Databricks:", err)
	}
	log.Println("Artifact uploaded to Databricks successfully")

	// Example 2: Using Databricks profile
	config2 := &uploader.Config{
		TrackingURI: "databricks://your-profile",
	}

	client2, err := uploader.NewClient(config2)
	if err != nil {
		log.Fatal("Failed to create client with profile:", err)
	}

	// Upload using profile-based authentication
	err = client2.UploadFile(ctx, runID, "/path/to/results.json", "results/results.json")
	if err != nil {
		log.Fatal("Failed to upload with profile:", err)
	}
	log.Println("Artifact uploaded using profile successfully")

	// Example 3: Direct Databricks URL
	config3 := &uploader.Config{
		TrackingURI:     "https://your-workspace.cloud.databricks.com",
		DatabricksToken: os.Getenv("DATABRICKS_TOKEN"),
	}

	client3, err := uploader.NewClient(config3)
	if err != nil {
		log.Fatal("Failed to create client with direct URL:", err)
	}

	// Upload to direct Databricks URL
	err = client3.UploadFile(ctx, runID, "/path/to/metrics.csv", "data/metrics.csv")
	if err != nil {
		log.Fatal("Failed to upload to direct URL:", err)
	}
	log.Println("Artifact uploaded to direct URL successfully")
}
