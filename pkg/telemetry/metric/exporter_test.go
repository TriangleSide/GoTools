package metric_test

import (
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/telemetry/metric"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

type mockExporterForExporterTest struct {
	mu     sync.Mutex
	points []*metric.Point
}

func (m *mockExporterForExporterTest) Export(p *metric.Point) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.points = append(m.points, p)
}

func (m *mockExporterForExporterTest) Points() []*metric.Point {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]*metric.Point, len(m.points))
	copy(result, m.points)
	return result
}

func TestSetExporter_WithValidExporter_ExporterIsStoredInContext(t *testing.T) {
	t.Parallel()
	exp := &mockExporterForExporterTest{}
	ctx := metric.SetExporter(t.Context(), exp)
	assert.NotNil(t, ctx)
	dim := metric.New("test")
	dim.Record(ctx, 1.0)
	points := exp.Points()
	assert.Equals(t, len(points), 1)
}

func TestSetExporter_OverwriteExporter_UsesNewExporter(t *testing.T) {
	t.Parallel()
	exp1 := &mockExporterForExporterTest{}
	exp2 := &mockExporterForExporterTest{}

	ctx := metric.SetExporter(t.Context(), exp1)
	dim := metric.New("test")
	dim.Record(ctx, 1.0)

	ctx = metric.SetExporter(ctx, exp2)
	dim.Record(ctx, 2.0)

	points1 := exp1.Points()
	points2 := exp2.Points()

	assert.Equals(t, len(points1), 1)
	assert.FloatEquals(t, points1[0].Value(), 1.0, 0.0001)
	assert.Equals(t, len(points2), 1)
	assert.FloatEquals(t, points2[0].Value(), 2.0, 0.0001)
}
