package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"

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

	// Example 1: Upload JSON data from memory
	data := map[string]interface{}{
		"accuracy":   0.95,
		"precision":  0.92,
		"recall":     0.89,
		"model_type": "random_forest",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal("Failed to marshal JSON:", err)
	}

	source := uploader.NewBytesSource(jsonData, "metrics/results.json")
	err = client.UploadArtifact(ctx, runID, source)
	if err != nil {
		log.Fatal("Failed to upload JSON data:", err)
	}
	log.Println("JSON data uploaded successfully")

	// Example 2: Upload text content from string
	logContent := `Training started at 2024-01-01 10:00:00
Epoch 1/10: loss=0.5, accuracy=0.8
Epoch 2/10: loss=0.3, accuracy=0.85
...
Training completed successfully`

	reader := strings.NewReader(logContent)
	source2 := uploader.NewReaderSource(reader, "logs/training.log", int64(len(logContent)))
	defer source2.Close()

	err = client.UploadArtifact(ctx, runID, source2)
	if err != nil {
		log.Fatal("Failed to upload log content:", err)
	}
	log.Println("Log content uploaded successfully")

	// Example 3: Upload using convenience method
	configContent := `{
  "learning_rate": 0.01,
  "batch_size": 32,
  "epochs": 10
}`

	err = client.UploadReader(ctx, runID, strings.NewReader(configContent), "config/hyperparams.json", int64(len(configContent)))
	if err != nil {
		log.Fatal("Failed to upload config:", err)
	}
	log.Println("Config uploaded successfully")
}
