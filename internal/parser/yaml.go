package parser

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"

	"github.com/imishinist/mlflow-cli/internal/models"
)

func ParseYAMLParams(reader io.Reader) (map[string]string, error) {
	var data models.ParametersFile
	decoder := yaml.NewDecoder(reader)

	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse YAML parameters: %w", err)
	}

	return data.Parameters, nil
}

func ParseYAMLMetrics(reader io.Reader) (*models.MetricsFile, error) {
	var data models.MetricsFile
	decoder := yaml.NewDecoder(reader)

	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse YAML metrics: %w", err)
	}

	return &data, nil
}
