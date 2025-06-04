package config

import (
	"fmt"
	"os"

	"github.com/TriangleSide/GoTools/pkg/stringcase"
	"github.com/TriangleSide/GoTools/pkg/structs"
)

// envSource fetches configuration values from environment variables. The variable name is derived from the
// struct field name converted to SNAKE_CASE. If a prefix is provided, it is prepended followed by an underscore.
func envSource(fieldName string, _ *structs.FieldMetadata, cfg Options) (string, bool, error) {
	formattedEnvName := stringcase.CamelToSnake(fieldName)
	if cfg.Prefix != "" {
		formattedEnvName = fmt.Sprintf("%s_%s", cfg.Prefix, formattedEnvName)
	}
	envValue, hasEnvValue := os.LookupEnv(formattedEnvName)
	return envValue, hasEnvValue, nil
}

// init registers the environment processor.
func init() {
	MustRegisterProcessor(ProcessorTypeEnv, envSource)
}
