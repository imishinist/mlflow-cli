package models

type Parameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ParametersFile struct {
	Parameters map[string]string `json:"parameters"`
}
