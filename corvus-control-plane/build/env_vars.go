package build

import (
	"encoding/json"
	"fmt"
)

// decodeEnvVarsToSlice converts the JSON-encoded environment variables string
// stored in the database into a []string of "KEY=VALUE" pairs that the Docker
// SDK expects for container.Config.Env.
//
// Returns nil (not an error) when the input pointer is nil or the JSON object
// is empty, meaning no environment variables were configured.
func decodeEnvVarsToSlice(encodedEnvVars *string) ([]string, error) {
	if encodedEnvVars == nil || *encodedEnvVars == "" {
		return nil, nil
	}

	var envVarsMap map[string]string
	unmarshalError := json.Unmarshal([]byte(*encodedEnvVars), &envVarsMap)
	if unmarshalError != nil {
		return nil, fmt.Errorf("failed to unmarshal environment variables JSON: %w", unmarshalError)
	}

	if len(envVarsMap) == 0 {
		return nil, nil
	}

	envVarsList := make([]string, 0, len(envVarsMap))
	for key, value := range envVarsMap {
		envVarsList = append(envVarsList, key+"="+value)
	}
	return envVarsList, nil
}
