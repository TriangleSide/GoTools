package metric

import (
	"time"
)

// Point represents a metric data point with a dimension, time, and value.
type Point struct {
	dimension []string
	time      time.Time
	value     float64
}

// Dimension returns the dimension associated with this point.
func (p *Point) Dimension() []string {
	result := make([]string, len(p.dimension))
	copy(result, p.dimension)
	return result
}

// Time returns the timestamp of this point.
func (p *Point) Time() time.Time {
	return p.time
}

// Value returns the metric value of this point.
func (p *Point) Value() float64 {
	return p.value
}
