package metric_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/telemetry/metric"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

type mockExporterForPointTest struct {
	mu     sync.Mutex
	points []*metric.Point
}

func (m *mockExporterForPointTest) Export(p *metric.Point) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.points = append(m.points, p)
}

func (m *mockExporterForPointTest) Points() []*metric.Point {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]*metric.Point, len(m.points))
	copy(result, m.points)
	return result
}

func TestPointDimension_ReturnsCorrectDimension(t *testing.T) {
	t.Parallel()
	exp := &mockExporterForPointTest{}
	ctx := metric.SetExporter(context.Background(), exp)
	dim := metric.New("dim1", "dim2")
	dim.Record(ctx, 1.0)
	points := exp.Points()
	assert.Equals(t, len(points), 1)
	assert.Equals(t, points[0].Dimension(), []string{"dim1", "dim2"})
}

func TestPointTime_ReturnsCorrectTime(t *testing.T) {
	t.Parallel()
	exp := &mockExporterForPointTest{}
	ctx := metric.SetExporter(context.Background(), exp)
	dim := metric.New("dim1")
	expectedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	dim.RecordAt(ctx, expectedTime, 1.0)
	points := exp.Points()
	assert.Equals(t, len(points), 1)
	assert.True(t, points[0].Time().Equal(expectedTime))
}

func TestPointValue_ReturnsCorrectValue(t *testing.T) {
	t.Parallel()
	exp := &mockExporterForPointTest{}
	ctx := metric.SetExporter(context.Background(), exp)
	dim := metric.New("dim1")
	dim.Record(ctx, 42.5)
	points := exp.Points()
	assert.Equals(t, len(points), 1)
	assert.FloatEquals(t, points[0].Value(), 42.5, 0.0001)
}
