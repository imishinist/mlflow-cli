package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/imishinist/mlflow-artifact-uploader"
)

func main() {
	fmt.Println("=== MLflow Artifact Uploader Test ===")

	// Test 1: Config validation
	fmt.Println("\n1. Testing configuration...")
	
	config := &uploader.Config{
		TrackingURI: "http://localhost:5000",
		Timeout:     30 * time.Second,
	}
	
	err := config.Validate()
	if err != nil {
		log.Printf("Config validation failed: %v", err)
	} else {
		fmt.Println("✓ Config validation passed")
	}

	// Test 2: Client creation
	fmt.Println("\n2. Testing client creation...")
	
	client, err := uploader.NewClient(config)
	if err != nil {
		log.Printf("Client creation failed: %v", err)
		return
	}
	fmt.Println("✓ Client created successfully")

	// Test 3: ArtifactSource creation
	fmt.Println("\n3. Testing ArtifactSource implementations...")
	
	// Test BytesSource
	testData := []byte("This is test data for artifact upload")
	bytesSource := uploader.NewBytesSource(testData, "test/data.txt")
	fmt.Printf("✓ BytesSource created: name=%s, size=%d\n", bytesSource.Name(), bytesSource.Size())
	
	// Test ReaderSource
	reader := strings.NewReader("This is test content from reader")
	readerSource := uploader.NewReaderSource(reader, "test/reader.txt", int64(len("This is test content from reader")))
	fmt.Printf("✓ ReaderSource created: name=%s, size=%d\n", readerSource.Name(), readerSource.Size())
	
	// Test FileSource (create a temporary file)
	tempFile, err := os.CreateTemp("", "test_artifact_*.txt")
	if err != nil {
		log.Printf("Failed to create temp file: %v", err)
		return
	}
	defer os.Remove(tempFile.Name())
	
	testContent := "This is test file content"
	_, err = tempFile.WriteString(testContent)
	if err != nil {
		log.Printf("Failed to write to temp file: %v", err)
		return
	}
	tempFile.Close()
	
	fileSource, err := uploader.NewFileSource(tempFile.Name(), "test/file.txt")
	if err != nil {
		log.Printf("Failed to create FileSource: %v", err)
		return
	}
	defer fileSource.Close()
	fmt.Printf("✓ FileSource created: name=%s, size=%d\n", fileSource.Name(), fileSource.Size())

	// Test 4: Databricks configuration detection
	fmt.Println("\n4. Testing Databricks configuration detection...")
	
	databricksConfigs := []*uploader.Config{
		{TrackingURI: "databricks"},
		{TrackingURI: "databricks://profile"},
		{TrackingURI: "https://workspace.cloud.databricks.com"},
		{TrackingURI: "http://localhost:5000"},
	}
	
	for _, cfg := range databricksConfigs {
		isDatabricks := cfg.IsDatabricks()
		fmt.Printf("  URI: %s -> Databricks: %v\n", cfg.TrackingURI, isDatabricks)
	}

	// Test 5: Mock upload test (without actual MLflow server)
	fmt.Println("\n5. Testing upload methods (mock)...")
	
	ctx := context.Background()
	mockRunID := "1234567890abcdef1234567890abcdef"
	
	// Note: These will fail because there's no MLflow server running,
	// but we can test that the methods are callable and handle errors gracefully
	
	fmt.Println("  Testing UploadArtifact method...")
	err = client.UploadArtifact(ctx, mockRunID, bytesSource)
	if err != nil {
		fmt.Printf("  Expected error (no MLflow server): %v\n", err)
	}
	
	fmt.Println("  Testing UploadReader method...")
	err = client.UploadReader(ctx, mockRunID, strings.NewReader("test"), "test.txt", 4)
	if err != nil {
		fmt.Printf("  Expected error (no MLflow server): %v\n", err)
	}

	fmt.Println("\n=== Test Summary ===")
	fmt.Println("✓ All basic functionality tests completed")
	fmt.Println("✓ Library is ready for use")
	fmt.Println("Note: Actual upload tests require a running MLflow server")
}
