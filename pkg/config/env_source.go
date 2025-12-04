package config

import (
	"os"

	"github.com/TriangleSide/GoTools/pkg/structs"
)

const (
	// ProcessorTypeEnv identifies the environment variable processor.
	ProcessorTypeEnv = "ENV"
)

// envSource fetches configuration values from environment variables. The variable name is derived from the
// struct field name converted to SNAKE_CASE.
func envSource(fieldName string, _ *structs.FieldMetadata) (string, bool, error) {
	formattedEnvName := camelToSnake(fieldName)
	envValue, hasEnvValue := os.LookupEnv(formattedEnvName)
	return envValue, hasEnvValue, nil
}

// init registers the environment processor.
func init() {
	MustRegisterProcessor(ProcessorTypeEnv, envSource)
}
