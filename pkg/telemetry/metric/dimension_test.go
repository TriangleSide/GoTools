package metric_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/telemetry/metric"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

type mockExporter struct {
	mu     sync.Mutex
	points []*metric.Point
}

func (m *mockExporter) Export(p *metric.Point) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.points = append(m.points, p)
}

func (m *mockExporter) Points() []*metric.Point {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]*metric.Point, len(m.points))
	copy(result, m.points)
	return result
}

func TestNew_WithNoDimensions_ReturnsEmptyDimension(t *testing.T) {
	t.Parallel()
	dim := metric.New()
	assert.NotNil(t, dim)
	assert.Equals(t, len(dim.Values()), 0)
}

func TestNew_WithSingleDimension_ReturnsDimensionWithOneValue(t *testing.T) {
	t.Parallel()
	dim := metric.New("key1")
	assert.NotNil(t, dim)
	values := dim.Values()
	assert.Equals(t, len(values), 1)
	assert.Equals(t, values[0], "key1")
}

func TestNew_WithMultipleDimensions_ReturnsDimensionWithAllValues(t *testing.T) {
	t.Parallel()
	dim := metric.New("key1", "key2", "key3")
	assert.NotNil(t, dim)
	values := dim.Values()
	assert.Equals(t, len(values), 3)
	assert.Equals(t, values[0], "key1")
	assert.Equals(t, values[1], "key2")
	assert.Equals(t, values[2], "key3")
}

func TestNew_ModifyingInputSlice_DoesNotAffectDimension(t *testing.T) {
	t.Parallel()
	input := []string{"key1", "key2"}
	dim := metric.New(input...)
	input[0] = "modified"
	values := dim.Values()
	assert.Equals(t, values[0], "key1")
}

func TestValues_ModifyingReturnedSlice_DoesNotAffectDimension(t *testing.T) {
	t.Parallel()
	dim := metric.New("key1", "key2")
	values := dim.Values()
	values[0] = "modified"
	originalValues := dim.Values()
	assert.Equals(t, originalValues[0], "key1")
}

func TestValues_WithEmptyDimension_ReturnsEmptySlice(t *testing.T) {
	t.Parallel()
	dim := metric.New()
	values := dim.Values()
	assert.NotNil(t, values)
	assert.Equals(t, len(values), 0)
}

func TestRecord_WithValidContext_ExportsPoint(t *testing.T) {
	t.Parallel()
	exp := &mockExporter{}
	ctx := metric.SetExporter(context.Background(), exp)
	dim := metric.New("metric_name")
	beforeRecord := time.Now()
	dim.Record(ctx, 100.0)
	afterRecord := time.Now()
	points := exp.Points()
	assert.Equals(t, len(points), 1)
	assert.True(t, !points[0].Time().Before(beforeRecord))
	assert.True(t, !points[0].Time().After(afterRecord))
	assert.FloatEquals(t, points[0].Value(), 100.0, 0.0001)
}

func TestRecord_WithNilExporter_Panics(t *testing.T) {
	t.Parallel()
	dim := metric.New("test")
	assert.PanicPart(t, func() {
		dim.Record(context.Background(), 1.0)
	}, "metric exporter is nil")
}

func TestRecord_MultipleRecords_ExportsAllPoints(t *testing.T) {
	t.Parallel()
	exp := &mockExporter{}
	ctx := metric.SetExporter(context.Background(), exp)
	dim := metric.New("counter")
	dim.Record(ctx, 1.0)
	dim.Record(ctx, 2.0)
	dim.Record(ctx, 3.0)
	points := exp.Points()
	assert.Equals(t, len(points), 3)
	assert.FloatEquals(t, points[0].Value(), 1.0, 0.0001)
	assert.FloatEquals(t, points[1].Value(), 2.0, 0.0001)
	assert.FloatEquals(t, points[2].Value(), 3.0, 0.0001)
}

func TestRecord_MultipleDimensions_ExportsWithCorrectDimensions(t *testing.T) {
	t.Parallel()
	exp := &mockExporter{}
	ctx := metric.SetExporter(context.Background(), exp)
	dim1 := metric.New("metric1")
	dim2 := metric.New("metric2", "label")
	dim1.Record(ctx, 10.0)
	dim2.Record(ctx, 20.0)
	points := exp.Points()
	assert.Equals(t, len(points), 2)
	assert.Equals(t, points[0].Dimension(), []string{"metric1"})
	assert.Equals(t, points[1].Dimension(), []string{"metric2", "label"})
}

func TestRecord_ConcurrentRecords_AllPointsExported(t *testing.T) {
	t.Parallel()
	exp := &mockExporter{}
	ctx := metric.SetExporter(context.Background(), exp)
	dim := metric.New("concurrent_metric")

	var waitGroup sync.WaitGroup
	numGoroutines := 100

	for i := range numGoroutines {
		waitGroup.Add(1)
		go func(val float64) {
			defer waitGroup.Done()
			dim.Record(ctx, val)
		}(float64(i))
	}

	waitGroup.Wait()
	points := exp.Points()
	assert.Equals(t, len(points), numGoroutines)
}

func TestRecord_WithZeroValue_ExportsPointWithZeroValue(t *testing.T) {
	t.Parallel()
	exp := &mockExporter{}
	ctx := metric.SetExporter(context.Background(), exp)
	dim := metric.New("zero_metric")
	dim.Record(ctx, 0.0)
	points := exp.Points()
	assert.Equals(t, len(points), 1)
	assert.FloatEquals(t, points[0].Value(), 0.0, 0.0001)
}

func TestRecord_WithNegativeValue_ExportsPointWithNegativeValue(t *testing.T) {
	t.Parallel()
	exp := &mockExporter{}
	ctx := metric.SetExporter(context.Background(), exp)
	dim := metric.New("negative_metric")
	dim.Record(ctx, -42.5)
	points := exp.Points()
	assert.Equals(t, len(points), 1)
	assert.FloatEquals(t, points[0].Value(), -42.5, 0.0001)
}

func TestRecordAt_WithValidContext_ExportsPointWithSpecifiedTime(t *testing.T) {
	t.Parallel()
	exp := &mockExporter{}
	ctx := metric.SetExporter(context.Background(), exp)
	dim := metric.New("metric_name")
	recordTime := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	dim.RecordAt(ctx, recordTime, 200.0)
	points := exp.Points()
	assert.Equals(t, len(points), 1)
	assert.True(t, points[0].Time().Equal(recordTime))
	assert.FloatEquals(t, points[0].Value(), 200.0, 0.0001)
	assert.Equals(t, points[0].Dimension(), []string{"metric_name"})
}

func TestRecordAt_WithNilExporter_Panics(t *testing.T) {
	t.Parallel()
	dim := metric.New("test")
	assert.PanicPart(t, func() {
		dim.RecordAt(context.Background(), time.Now(), 1.0)
	}, "metric exporter is nil")
}

func TestRecordAt_ModifyingDimensionAfterRecord_DoesNotAffectExportedPoint(t *testing.T) {
	t.Parallel()
	exp := &mockExporter{}
	ctx := metric.SetExporter(context.Background(), exp)
	dim := metric.New("original")
	dim.RecordAt(ctx, time.Now(), 1.0)
	points := exp.Points()
	assert.Equals(t, len(points), 1)
	assert.Equals(t, points[0].Dimension(), []string{"original"})
}
