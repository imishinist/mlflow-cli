package main

import (
	"context"
	"encoding/json"
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

	// Prepare multiple artifact sources
	var sources []uploader.ArtifactSource

	// Add file sources
	fileSource1, err := uploader.NewFileSource("/path/to/model.pkl", "model/model.pkl")
	if err != nil {
		log.Fatal("Failed to create file source 1:", err)
	}
	sources = append(sources, fileSource1)

	fileSource2, err := uploader.NewFileSource("/path/to/config.yaml", "config/config.yaml")
	if err != nil {
		log.Fatal("Failed to create file source 2:", err)
	}
	sources = append(sources, fileSource2)

	// Add memory-based sources
	metricsData := map[string]float64{
		"accuracy":  0.95,
		"precision": 0.92,
		"recall":    0.89,
		"f1_score":  0.905,
	}

	metricsJSON, err := json.Marshal(metricsData)
	if err != nil {
		log.Fatal("Failed to marshal metrics:", err)
	}

	metricsSource := uploader.NewBytesSource(metricsJSON, "metrics/final_metrics.json")
	sources = append(sources, metricsSource)

	// Add configuration data
	configData := map[string]interface{}{
		"learning_rate": 0.01,
		"batch_size":    32,
		"epochs":        10,
		"optimizer":     "adam",
	}

	configJSON, err := json.Marshal(configData)
	if err != nil {
		log.Fatal("Failed to marshal config:", err)
	}

	configSource := uploader.NewBytesSource(configJSON, "config/hyperparameters.json")
	sources = append(sources, configSource)

	// Upload all artifacts in batch
	err = client.UploadArtifacts(ctx, runID, sources)
	if err != nil {
		log.Fatal("Failed to upload artifacts:", err)
	}

	// Clean up resources
	for _, source := range sources {
		if err := source.Close(); err != nil {
			log.Printf("Warning: Failed to close source %s: %v", source.Name(), err)
		}
	}

	log.Printf("Successfully uploaded %d artifacts", len(sources))
}
