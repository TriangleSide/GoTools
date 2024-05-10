package metrics

import (
	"time"
)

// A Metric is a standard of measurement used to evaluate the performance and efficiency of a system.
type Metric struct {
	Namespace   string            `json:"namespace" validate:"required"`
	Scopes      map[string]string `json:"scopes"    validate:"required"`
	Timestamp   time.Time         `json:"timestamp" validate:"required"`
	Measurement *float32          `json:"measurement,omitempty"`
}
