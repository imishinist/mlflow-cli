package parser

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/imishinist/mlflow-cli/internal/models"
)

func ParseJSONParams(reader io.Reader) (map[string]string, error) {
	var data models.ParametersFile
	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON parameters: %w", err)
	}

	return data.Parameters, nil
}

func ParseJSONMetrics(reader io.Reader) (*models.MetricsFile, error) {
	var data models.MetricsFile
	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON metrics: %w", err)
	}

	return &data, nil
}
