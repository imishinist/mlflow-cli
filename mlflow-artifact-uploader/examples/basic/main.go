package main

import (
	"context"
	"log"

	"github.com/imishinist/mlflow-artifact-uploader"
)

func main() {
	// Create client configuration
	config := &uploader.Config{
		TrackingURI: "http://localhost:5000",
	}

	// Create client
	client, err := uploader.NewClient(config)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()
	runID := "your-run-id"

	// Example 1: Upload a file
	err = client.UploadFile(ctx, runID, "/path/to/model.pkl", "model/model.pkl")
	if err != nil {
		log.Fatal("Failed to upload file:", err)
	}
	log.Println("File uploaded successfully")

	// Example 2: Upload using ArtifactSource
	source, err := uploader.NewFileSource("/path/to/config.yaml", "config/config.yaml")
	if err != nil {
		log.Fatal("Failed to create file source:", err)
	}
	defer source.Close()

	err = client.UploadArtifact(ctx, runID, source)
	if err != nil {
		log.Fatal("Failed to upload artifact:", err)
	}
	log.Println("Artifact uploaded successfully")

	// Example 3: Upload multiple files
	files := map[string]string{
		"/path/to/model.pkl":   "model/model.pkl",
		"/path/to/config.yaml": "config/config.yaml",
		"/path/to/metrics.json": "metrics/metrics.json",
	}

	err = client.UploadFiles(ctx, runID, files)
	if err != nil {
		log.Fatal("Failed to upload files:", err)
	}
	log.Println("All files uploaded successfully")
}
