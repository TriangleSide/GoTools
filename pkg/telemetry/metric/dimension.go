package metric

import (
	"context"
	"errors"
	"time"
)

// Dimension represents a set of string labels for categorizing metrics.
type Dimension struct {
	values []string
}

// New creates a new Dimension with the given string values.
func New(dimension ...string) *Dimension {
	v := make([]string, len(dimension))
	copy(v, dimension)
	return &Dimension{
		values: v,
	}
}

// Values returns a copy of the dimension values.
func (d *Dimension) Values() []string {
	result := make([]string, len(d.values))
	copy(result, d.values)
	return result
}

// Record creates a metric point with this dimension and the given value at the current time,
// and exports it using the exporter stored in the context.
func (d *Dimension) Record(ctx context.Context, value float64) {
	d.RecordAt(ctx, time.Now(), value)
}

// RecordAt creates a metric point with this dimension, the given time, and value,
// and exports it using the exporter stored in the context.
func (d *Dimension) RecordAt(ctx context.Context, givenTime time.Time, value float64) {
	exp := exporterFromContext(ctx)
	if exp == nil {
		panic(errors.New("metric exporter is nil"))
	}
	exp.Export(&Point{
		dimension: d.Values(),
		time:      givenTime,
		value:     value,
	})
}
