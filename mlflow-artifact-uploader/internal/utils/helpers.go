package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateArtifactPath validates artifact path format
func ValidateArtifactPath(path string) error {
	if path == "" {
		return fmt.Errorf("artifact path cannot be empty")
	}

	// Remove leading slash for consistency
	path = strings.TrimPrefix(path, "/")

	// Check for invalid characters (basic validation)
	if strings.Contains(path, "..") {
		return fmt.Errorf("artifact path cannot contain '..'")
	}

	return nil
}

// NormalizeArtifactPath normalizes artifact path
func NormalizeArtifactPath(path string) string {
	// Remove leading slash and clean path
	path = strings.TrimPrefix(path, "/")
	return filepath.Clean(path)
}

// SanitizeFileName sanitizes filename for cross-platform compatibility
func SanitizeFileName(filename string) string {
	// Replace problematic characters
	replacements := map[string]string{
		"<":  "_",
		">":  "_",
		":":  "_",
		"\"": "_",
		"|":  "_",
		"?":  "_",
		"*":  "_",
	}

	for old, new := range replacements {
		filename = strings.ReplaceAll(filename, old, new)
	}

	return filename
}

// IsValidRunID checks if run ID format is valid
func IsValidRunID(runID string) bool {
	if runID == "" {
		return false
	}

	// Basic validation - MLflow run IDs are typically 32-character hex strings
	if len(runID) != 32 {
		return false
	}

	// Check if all characters are valid hex
	for _, char := range runID {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}

	return true
}
